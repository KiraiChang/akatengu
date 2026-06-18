package projection_repo_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/shared/utils/test"
)

// insertActiveTxn 插入一筆 ACTIVE 交易及其分錄。
type testEntry struct {
	accountId string
	ledgerId  *int64
	debit     float64
	credit    float64
}

func insertActiveTxn(t *testing.T, db *sqlx.DB, txnID int64, entries []testEntry) {
	t.Helper()
	ctx := context.Background()
	var total float64
	for _, e := range entries {
		total += e.debit
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO transactions (txn_id, merchant_id, txn_date, description, total_amount, status, version)
		 VALUES (?, ?, '2025-01-15', 'test', ?, 'ACTIVE', 1)`,
		txnID, testMerchantID, total)
	if err != nil {
		t.Fatalf("insertActiveTxn txn_id=%d: %v", txnID, err)
	}
	for _, e := range entries {
		_, err = db.ExecContext(ctx,
			`INSERT INTO journal_entries (txn_id, merchant_id, ledger_id, account_id, debit, credit)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			txnID, testMerchantID, e.ledgerId, e.accountId, e.debit, e.credit)
		if err != nil {
			t.Fatalf("insertActiveTxn entry: %v", err)
		}
	}
}

type rbRow struct{ Debit, Credit float64 }

func queryRB(t *testing.T, db *sqlx.DB, accountId string) rbRow {
	t.Helper()
	var d, c float64
	db.QueryRowContext(context.Background(),
		`SELECT COALESCE(debit_total,0), COALESCE(credit_total,0)
		 FROM account_running_balances WHERE account_id=? AND merchant_id=?`,
		accountId, testMerchantID).Scan(&d, &c)
	return rbRow{d, c}
}

func queryLedgerRB(t *testing.T, db *sqlx.DB, ledgerId int64) rbRow {
	t.Helper()
	var d, c float64
	db.QueryRowContext(context.Background(),
		`SELECT COALESCE(debit_total,0), COALESCE(credit_total,0)
		 FROM ledger_running_balances WHERE ledger_id=? AND merchant_id=?`,
		ledgerId, testMerchantID).Scan(&d, &c)
	return rbRow{d, c}
}

func assertRB(t *testing.T, acct string, got rbRow, wantD, wantC float64) {
	t.Helper()
	if got.Debit != wantD {
		t.Errorf("%s debit_total: got %.2f, want %.2f", acct, got.Debit, wantD)
	}
	if got.Credit != wantC {
		t.Errorf("%s credit_total: got %.2f, want %.2f", acct, got.Credit, wantC)
	}
}

// applyBalance 套用一個分錄的 delta（sign=+1 為套用，sign=-1 為反轉）。
func applyBalance(
	t *testing.T, ctx context.Context,
	accRepo projection_repo.AccountRunningBalanceRepo,
	ledRepo projection_repo.LedgerRunningBalanceRepo,
	entries []testEntry,
	sign decimal.Decimal,
) {
	t.Helper()
	for _, e := range entries {
		debit := decimal.NewFromFloat(e.debit).Mul(sign)
		credit := decimal.NewFromFloat(e.credit).Mul(sign)

		if err := accRepo.Upsert(ctx, e.accountId, testMerchantID, debit, credit); err != nil {
			t.Fatalf("Upsert account %s: %v", e.accountId, err)
		}
		ancestors, err := accRepo.GetAncestorIds(ctx, e.accountId, testMerchantID)
		if err != nil {
			t.Fatalf("GetAncestorIds %s: %v", e.accountId, err)
		}
		for _, anc := range ancestors {
			if err := accRepo.Upsert(ctx, anc, testMerchantID, debit, credit); err != nil {
				t.Fatalf("Upsert ancestor %s: %v", anc, err)
			}
		}
		if e.ledgerId != nil {
			if err := ledRepo.Upsert(ctx, *e.ledgerId, testMerchantID, debit, credit); err != nil {
				t.Fatalf("Upsert ledger %d: %v", *e.ledgerId, err)
			}
		}
	}
}

var (
	pos = decimal.NewFromInt(1)
	neg = decimal.NewFromInt(-1)
)

// TestApplyAccountBalance_Create 一般交易建立後葉科目及祖先科目餘額正確。
func TestApplyAccountBalance_Create(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	accRepo := projection_repo.NewAccountRunningBalanceRepo(q)
	ledRepo := projection_repo.NewLedgerRunningBalanceRepo(q)
	ctx := context.Background()

	entries := []testEntry{
		{accountId: acctCash, debit: 100, credit: 0},    // DR 1101-01
		{accountId: acctExpense, debit: 0, credit: 100}, // CR 5101-01
	}
	applyBalance(t, ctx, accRepo, ledRepo, entries, pos)

	assertRB(t, acctCash, queryRB(t, db, acctCash), 100, 0)
	assertRB(t, acctCash1101Parent, queryRB(t, db, acctCash1101Parent), 100, 0)
	assertRB(t, acctCash110Parent, queryRB(t, db, acctCash110Parent), 100, 0)
	assertRB(t, acctExpense, queryRB(t, db, acctExpense), 0, 100)
	assertRB(t, acctExp5101Parent, queryRB(t, db, acctExp5101Parent), 0, 100)
}

// TestApplyLedgerBalance_Create 帶 ledger_id 的交易建立後 ledger 餘額正確。
func TestApplyLedgerBalance_Create(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	accRepo := projection_repo.NewAccountRunningBalanceRepo(q)
	ledRepo := projection_repo.NewLedgerRunningBalanceRepo(q)
	ctx := context.Background()

	ledgerID := int64(ledgerCashId)
	entries := []testEntry{
		{accountId: acctCash, ledgerId: &ledgerID, debit: 200, credit: 0},
	}
	applyBalance(t, ctx, accRepo, ledRepo, entries, pos)

	assertRB(t, "ledger", queryLedgerRB(t, db, ledgerCashId), 200, 0)
}

// TestReverseAccountBalance_Voided 交易 voided 後餘額歸零。
func TestReverseAccountBalance_Voided(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	accRepo := projection_repo.NewAccountRunningBalanceRepo(q)
	ledRepo := projection_repo.NewLedgerRunningBalanceRepo(q)
	ctx := context.Background()

	entries := []testEntry{
		{accountId: acctCash, debit: 100, credit: 0},
		{accountId: acctExpense, debit: 0, credit: 100},
	}

	applyBalance(t, ctx, accRepo, ledRepo, entries, pos)
	assertRB(t, acctCash, queryRB(t, db, acctCash), 100, 0)

	applyBalance(t, ctx, accRepo, ledRepo, entries, neg)

	assertRB(t, acctCash, queryRB(t, db, acctCash), 0, 0)
	assertRB(t, acctCash1101Parent, queryRB(t, db, acctCash1101Parent), 0, 0)
	assertRB(t, acctCash110Parent, queryRB(t, db, acctCash110Parent), 0, 0)
	assertRB(t, acctExpense, queryRB(t, db, acctExpense), 0, 0)
}

// TestReverseAccountBalance_Corrected 更正後餘額僅保留更正交易金額。
func TestReverseAccountBalance_Corrected(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	accRepo := projection_repo.NewAccountRunningBalanceRepo(q)
	ledRepo := projection_repo.NewLedgerRunningBalanceRepo(q)
	ctx := context.Background()

	orig := []testEntry{{accountId: acctCash, debit: 100, credit: 0}}
	correction := []testEntry{{accountId: acctCash, debit: 90, credit: 0}}

	applyBalance(t, ctx, accRepo, ledRepo, orig, pos)
	applyBalance(t, ctx, accRepo, ledRepo, correction, pos)
	assertRB(t, acctCash, queryRB(t, db, acctCash), 190, 0)

	// 更正事件發生：反轉原始交易
	applyBalance(t, ctx, accRepo, ledRepo, orig, neg)
	assertRB(t, acctCash, queryRB(t, db, acctCash), 90, 0)
	assertRB(t, acctCash1101Parent, queryRB(t, db, acctCash1101Parent), 90, 0)
}

// TestApplyMultipleEntries_AncestorAccumulates 同層兩個葉科目的祖先正確累加。
func TestApplyMultipleEntries_AncestorAccumulates(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	accRepo := projection_repo.NewAccountRunningBalanceRepo(q)
	ledRepo := projection_repo.NewLedgerRunningBalanceRepo(q)
	ctx := context.Background()

	entries := []testEntry{
		{accountId: acctCash, debit: 100, credit: 0}, // 1101-01
		{accountId: acctCash2, debit: 50, credit: 0}, // 1101-02（同一 1101 下）
	}
	applyBalance(t, ctx, accRepo, ledRepo, entries, pos)

	assertRB(t, acctCash, queryRB(t, db, acctCash), 100, 0)
	assertRB(t, acctCash2, queryRB(t, db, acctCash2), 50, 0)
	assertRB(t, acctCash1101Parent, queryRB(t, db, acctCash1101Parent), 150, 0)
	assertRB(t, acctCash110Parent, queryRB(t, db, acctCash110Parent), 150, 0)
}

// TestReverseIdempotent Apply 一次再 Reverse 一次，餘額應為 0 不為負。
func TestReverseIdempotent(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	accRepo := projection_repo.NewAccountRunningBalanceRepo(q)
	ledRepo := projection_repo.NewLedgerRunningBalanceRepo(q)
	ctx := context.Background()

	entries := []testEntry{{accountId: acctCash, debit: 100, credit: 0}}
	applyBalance(t, ctx, accRepo, ledRepo, entries, pos)
	applyBalance(t, ctx, accRepo, ledRepo, entries, neg)

	got := queryRB(t, db, acctCash)
	if got.Debit != 0 {
		t.Errorf("debit should be 0 after apply+reverse, got %.2f", got.Debit)
	}
}

// TestGetEntriesByTxnId_ReturnsCorrectEntries 確認 GetEntriesByTxnId 正確回傳分錄。
func TestGetEntriesByTxnId_ReturnsCorrectEntries(t *testing.T) {
	db := test.NewTestDB(t)
	q := sqlcdb.New(db)
	txnRepo := projection_repo.NewTransactionRepo(q)
	ctx := context.Background()

	insertActiveTxn(t, db, 1, []testEntry{
		{accountId: acctCash, debit: 100, credit: 0},
		{accountId: acctExpense, debit: 0, credit: 100},
	})

	entries, err := txnRepo.GetEntriesByTxnId(ctx, 1, testMerchantID)
	if err != nil {
		t.Fatalf("GetEntriesByTxnId: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}
