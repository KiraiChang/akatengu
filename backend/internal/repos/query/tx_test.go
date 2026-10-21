package query_test

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"akatengu/internal/repos/query"
	"akatengu/internal/shared/utils/test"
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// Fixture identifiers. 1101, 1203-02 and 5101-01 come from seeds/accounts.sql;
// 1101-92 is not in the seeds.
const (
	txTestAccountID      = "1101-92"
	txTestInvestmentUUID = "0192f3c0-0000-7000-8000-000000000001"
	txTestTemplateID     = int64(1)
	txTestOtherUserID    = int64(2) // a user that is not yet a member of the merchant
)

// txWriteTables are the tables the write methods of the four wrapped repos touch.
var txWriteTables = []string{
	"ledger_account_type_config",
	"asset_type_account_config",
	"merchants",
	"user_merchants",
	"users",
	"transaction_templates",
	"transaction_template_entries",
}

// ── helpers ──────────────────────────────────────────────────────────────────

// newTxTestFixture opens a transaction on the only connection and writes an
// account, an investment, users, a merchant and a template inside it. The
// transaction is rolled back at cleanup if the test has not ended it.
func newTxTestFixture(t *testing.T) (*sqlx.DB, *sqlx.Tx, int64) {
	t.Helper()
	sdb := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	sdb.SetMaxOpenConns(1)

	tx, err := sdb.Beginx()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	exec := func(query string, args ...any) sql.Result {
		t.Helper()
		res, err := tx.Exec(query, args...)
		if err != nil {
			t.Fatalf("exec %q: %v", query, err)
		}
		return res
	}
	exec(`INSERT INTO accounts(account_id, merchant_id, parent_id, name, type, normal_balance, is_summary, is_active, version)
		VALUES(?, ?, '1101', 'R0 交易內查詢科目', 'ASSET', 'DEBIT', 0, 1, 1)`, txTestAccountID, testMerchantID)
	res := exec(`INSERT INTO investments(merchant_id, account_id, asset_type, currency, symbol, name, cost_method, ifrs_category, is_active, version, uuid)
		VALUES(?, '1203-02', 'STOCK', 'TWD', 'R0-TX-TEST', 'R0 交易內查詢投資', 'FIFO', 'FVOCI', 1, 1, ?)`, testMerchantID, txTestInvestmentUUID)
	investmentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("investment id: %v", err)
	}

	exec(`INSERT INTO users(user_id, username, password, status) VALUES(1, 'r0-owner', 'unused', 'ACTIVE'), (?, 'r0-member', 'unused', 'ACTIVE')`,
		txTestOtherUserID)
	exec(`INSERT INTO merchants(merchant_id, name, display_name) VALUES(?, 'r0 test', 'r0 test')`, testMerchantID)
	exec(`INSERT INTO user_merchants(user_id, merchant_id, role) VALUES(1, ?, 'OWNER')`, testMerchantID)
	exec(`INSERT INTO transaction_templates(id, merchant_id, name) VALUES(?, ?, 'R0 測試範本')`, txTestTemplateID, testMerchantID)
	exec(`INSERT INTO transaction_template_entries(merchant_id, template_id, sort_order, account_id, debit, credit)
		VALUES(?, ?, 0, '5101-01', 100, 0)`, testMerchantID, txTestTemplateID)
	return sdb, tx, investmentID
}

// txTableRows returns every row of table as seen inside tx.
func txTableRows(t *testing.T, tx *sqlx.Tx, table string) []map[string]any {
	t.Helper()
	rows, err := tx.Queryx(`SELECT * FROM ` + table + ` ORDER BY rowid`)
	if err != nil {
		t.Fatalf("select %s: %v", table, err)
	}
	defer rows.Close()
	var result []map[string]any
	for rows.Next() {
		row := map[string]any{}
		if err := rows.MapScan(row); err != nil {
			t.Fatalf("scan %s: %v", table, err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows %s: %v", table, err)
	}
	return result
}

func txSnapshot(t *testing.T, tx *sqlx.Tx) map[string][]map[string]any {
	t.Helper()
	s := make(map[string][]map[string]any, len(txWriteTables))
	for _, table := range txWriteTables {
		s[table] = txTableRows(t, tx, table)
	}
	return s
}

// ── tests ────────────────────────────────────────────────────────────────────

// TestNewQueryRepositoryTx_Reads checks the sqlc path and both sqlx paths of a
// Repo bound to a transaction, why a normal Repo cannot be used while the
// transaction holds the only connection, and that the rows are gone after
// rollback.
func TestNewQueryRepositoryTx_Reads(t *testing.T) {
	sdb, tx, investmentID := newTxTestFixture(t)

	reads := []struct {
		name string
		find func(ctx context.Context, r *query.Repo) (bool, error)
	}{
		{
			name: "Account.GetAccount（sqlc）",
			find: func(ctx context.Context, r *query.Repo) (bool, error) {
				a, err := r.Account.GetAccount(ctx, txTestAccountID)
				if errors.Is(err, sql.ErrNoRows) {
					return false, nil
				}
				return a != nil, err
			},
		},
		{
			name: "Investment.GetAllSummaries（SelectContext）",
			find: func(ctx context.Context, r *query.Repo) (bool, error) {
				summaries, err := r.Investment.GetAllSummaries(ctx)
				for _, s := range summaries {
					if s.InvestmentId == investmentID {
						return true, err
					}
				}
				return false, err
			},
		},
		{
			name: "Investment.GetSummary（QueryRowxContext）",
			find: func(ctx context.Context, r *query.Repo) (bool, error) {
				s, err := r.Investment.GetSummary(ctx, investmentID)
				return s != nil, err
			},
		},
	}

	txRepo := query.NewQueryRepositoryTx(tx)
	for _, rd := range reads {
		t.Run("交易內 "+rd.name, func(t *testing.T) {
			found, err := rd.find(testCtx(), txRepo)
			if err != nil || !found {
				t.Fatalf("found = %v, err = %v, want the row written in the transaction", found, err)
			}
		})
	}

	t.Run("交易進行中 一般 Repo 等不到唯一的連線", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(testCtx(), 200*time.Millisecond)
		defer cancel()
		_, err := query.NewQueryRepository(sdb).Account.GetAccount(ctx, txTestAccountID)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want context.DeadlineExceeded", err)
		}
	})

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	dbRepo := query.NewQueryRepository(sdb)
	for _, rd := range reads {
		t.Run("回滾後 "+rd.name, func(t *testing.T) {
			found, err := rd.find(testCtx(), dbRepo)
			if err != nil || found {
				t.Fatalf("found = %v, err = %v, want no row after rollback", found, err)
			}
		})
	}
}

// TestNewQueryRepositoryTx_WritesRejected calls every write method listed in
// tx.go on a Repo bound to a transaction; each must return ErrTxBoundWrite and
// leave the tables unchanged inside the transaction.
func TestNewQueryRepositoryTx_WritesRejected(t *testing.T) {
	_, tx, _ := newTxTestFixture(t)
	r := query.NewQueryRepositoryTx(tx)
	entries := []query.TemplateEntryInput{{AccountID: "5101-01", Debit: decimal.NewFromInt(200)}}

	cases := []struct {
		name  string
		write func(ctx context.Context) error
	}{
		{
			name: "Config.UpsertLedgerAccountTypeConfig",
			write: func(ctx context.Context) error {
				return r.Config.UpsertLedgerAccountTypeConfig(ctx, enums.LedgerAccountTypeBankAccount.Enum(), "1101-02", nil)
			},
		},
		{
			name: "Config.UpsertAssetTypeAccountConfig",
			write: func(ctx context.Context) error {
				return r.Config.UpsertAssetTypeAccountConfig(ctx, sqlcdb.UpsertAssetTypeAccountConfigParams{
					MerchantID:              testMerchantID,
					AssetType:               enums.AssetTypeStock.Enum(),
					RealizedGainAccountID:   "4203-01",
					RealizedLossAccountID:   "5402-11",
					UnrealizedGainAccountID: "4204-01",
					UnrealizedLossAccountID: "5402-21",
					FeeAccountID:            "5402-01",
					TaxAccountID:            "5402-05",
				})
			},
		},
		{
			name: "Merchant.Create",
			write: func(ctx context.Context) error {
				_, err := r.Merchant.Create(ctx, "r0 new", "r0 new", "TWD")
				return err
			},
		},
		{
			name: "Merchant.AddUser",
			write: func(ctx context.Context) error {
				return r.Merchant.AddUser(ctx, txTestOtherUserID, testMerchantID, enums.MerchantRoleTypeMember.Enum())
			},
		},
		{
			name: "Merchant.Update",
			write: func(ctx context.Context) error {
				return r.Merchant.Update(ctx, testMerchantID, "r0 renamed", "r0 renamed", "TWD")
			},
		},
		{
			name: "Merchant.Deactivate",
			write: func(ctx context.Context) error {
				return r.Merchant.Deactivate(ctx, testMerchantID)
			},
		},
		{
			name: "User.Create",
			write: func(ctx context.Context) error {
				_, err := r.User.Create(ctx, &db.User{Username: "r0-new", Password: "unused", Status: enums.UserStatusTypeActive.Enum()})
				return err
			},
		},
		{
			name: "Template.CreateTemplate",
			write: func(ctx context.Context) error {
				_, err := r.Template.CreateTemplate(ctx, query.CreateTemplateCmd{Name: "R0 新範本", Entries: entries})
				return err
			},
		},
		{
			name: "Template.UpdateTemplate",
			write: func(ctx context.Context) error {
				return r.Template.UpdateTemplate(ctx, txTestTemplateID, query.UpdateTemplateCmd{Name: "R0 改名範本", Entries: entries})
			},
		},
		{
			name: "Template.DeleteTemplate",
			write: func(ctx context.Context) error {
				return r.Template.DeleteTemplate(ctx, txTestTemplateID)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := txSnapshot(t, tx)
			if err := tc.write(testCtx()); !errors.Is(err, query.ErrTxBoundWrite) {
				t.Fatalf("err = %v, want query.ErrTxBoundWrite", err)
			}
			if after := txSnapshot(t, tx); !reflect.DeepEqual(after, before) {
				t.Fatalf("tables changed:\nbefore %v\nafter  %v", before, after)
			}
		})
	}
}

// TestNewQueryRepository_WriteUnaffected checks that the wrappers only apply to
// a bound Repo: a normal Repo still writes.
func TestNewQueryRepository_WriteUnaffected(t *testing.T) {
	sdb := test.NewTestDB(t)
	sdb.SetMaxOpenConns(1)

	err := query.NewQueryRepository(sdb).Config.UpsertLedgerAccountTypeConfig(testCtx(),
		enums.LedgerAccountTypeBankAccount.Enum(), "1101-02", nil)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var accountID string
	err = sdb.Get(&accountID, `SELECT account_id FROM ledger_account_type_config WHERE merchant_id = ? AND type = 'BANK_ACCOUNT'`,
		testMerchantID)
	if err != nil || accountID != "1101-02" {
		t.Fatalf("account_id = %q, err = %v, want 1101-02", accountID, err)
	}
}
