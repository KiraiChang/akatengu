package services_test

import (
	"akatengu/internal/enums"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"akatengu/internal/shared/utils/test"
	"context"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

const (
	approveTestImportUUID  = "0192f3b0-0000-7000-8000-000000000010"
	approveTestBankTxnUUID = "0192f3b0-0000-7000-8000-000000000001"
)

// ─────────────────────────────────────────
// Stubs
// ─────────────────────────────────────────

// raceEventRepo runs onFirstLookup and then reports no events on the first
// aggregate lookup, simulating a concurrent approval that completes between this
// request's lookup and its append.
type raceEventRepo struct {
	query.EventRepo
	onFirstLookup func()
	lookups       int
}

func (r *raceEventRepo) GetByAggregate(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) ([]dbmodel.EventStore, error) {
	r.lookups++
	if r.lookups == 1 {
		r.onFirstLookup()
		return nil, nil
	}
	return r.EventRepo.GetByAggregate(ctx, aggregateType, aggregateID)
}

var _ query.EventRepo = (*raceEventRepo)(nil)

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

// newApproveTestDB prepares an open period, the bank ledger and one withdrawal
// bank txn waiting for approval, and returns the bank txn ID.
func newApproveTestDB(t *testing.T) (*sqlx.DB, int64) {
	t.Helper()
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)

	insertPeriodOpen(t, db, 1, "2026-10-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	res, err := db.Exec(`INSERT INTO bank_statement_imports(import_uuid, merchant_id, ledger_id, statement_date)
		VALUES(?, ?, 1, '2026-10-31')`, approveTestImportUUID, testMID)
	if err != nil {
		t.Fatalf("insert bank statement import: %v", err)
	}
	importID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("bank statement import id: %v", err)
	}

	res, err = db.Exec(`INSERT INTO bank_statement_txns(bank_txn_uuid, import_id, merchant_id, txn_date, description, debit, credit, ledger_id, ledger_uuid)
		VALUES(?, ?, ?, '2026-10-05', 'rent', 1000, 0, 1, ?)`, approveTestBankTxnUUID, importID, testMID, testLedgerUUID(1))
	if err != nil {
		t.Fatalf("insert bank statement txn: %v", err)
	}
	bankTxnID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("bank statement txn id: %v", err)
	}

	return db, bankTxnID
}

func newApproveService(db *sqlx.DB, q *query.Repo) services.BankStatementReviewService {
	return services.NewBankStatementReviewService(q, event_store.NewUnitOfWork(db), newSvc(db))
}

func approveRequest(bankTxnID int64) services.ApproveRequest {
	return services.ApproveRequest{
		BankTxnID:        bankTxnID,
		LedgerID:         1,
		AccountID:        "1101-02",
		CounterAccountID: "5101-01",
		TxnDate:          "2026-10-05",
		Description:      "rent",
	}
}

func approveTxnIDs(t *testing.T, db *sqlx.DB) []int64 {
	t.Helper()
	var ids []int64
	if err := db.Select(&ids, `SELECT txn_id FROM transactions WHERE merchant_id = ?`, testMID); err != nil {
		t.Fatalf("select transactions: %v", err)
	}
	return ids
}

type approveBankTxnRow struct {
	MatchStatus  string `db:"match_status"`
	CreatedTxnID *int64 `db:"created_txn_id"`
}

func approveBankTxn(t *testing.T, db *sqlx.DB, bankTxnID int64) approveBankTxnRow {
	t.Helper()
	var row approveBankTxnRow
	if err := db.Get(&row, `SELECT match_status, created_txn_id FROM bank_statement_txns WHERE bank_txn_id = ?`, bankTxnID); err != nil {
		t.Fatalf("get bank statement txn: %v", err)
	}
	return row
}

// expectApprovedOnce checks that exactly one transaction exists and the bank txn
// is APPROVED and linked to it; it returns the transaction ID.
func expectApprovedOnce(t *testing.T, db *sqlx.DB, bankTxnID int64) int64 {
	t.Helper()
	ids := approveTxnIDs(t, db)
	if len(ids) != 1 {
		t.Fatalf("transactions: got %v, want exactly 1", ids)
	}

	row := approveBankTxn(t, db, bankTxnID)
	if row.MatchStatus != "APPROVED" {
		t.Errorf("match_status: got %s, want APPROVED", row.MatchStatus)
	}
	if row.CreatedTxnID == nil {
		t.Errorf("created_txn_id: got NULL, want %d", ids[0])
	} else if *row.CreatedTxnID != ids[0] {
		t.Errorf("created_txn_id: got %d, want %d", *row.CreatedTxnID, ids[0])
	}
	return ids[0]
}

func mustApprove(t *testing.T, svc services.BankStatementReviewService, bankTxnID int64) {
	t.Helper()
	if err := svc.ApproveTxn(testCtx(), approveRequest(bankTxnID)); err != nil {
		t.Fatalf("ApproveTxn: %v", err)
	}
}

// ─────────────────────────────────────────
// ApproveTxn
// ─────────────────────────────────────────

func TestApproveTxn(t *testing.T) {
	cases := []struct {
		name string
		// prepare sets up the state before the approval under test.
		prepare func(t *testing.T, db *sqlx.DB, svc services.BankStatementReviewService, bankTxnID int64)
		// wantErr is a substring of the expected error; empty means success.
		wantErr string
		check   func(t *testing.T, db *sqlx.DB, bankTxnID int64)
	}{
		{
			name: "尚未核准的明細：建立 1 筆交易（2 筆分錄），明細為 APPROVED 並關聯到該交易",
			check: func(t *testing.T, db *sqlx.DB, bankTxnID int64) {
				txnID := expectApprovedOnce(t, db, bankTxnID)
				var entries int
				if err := db.Get(&entries, `SELECT COUNT(*) FROM journal_entries WHERE txn_id = ?`, txnID); err != nil {
					t.Fatalf("count journal entries: %v", err)
				}
				if entries != 2 {
					t.Errorf("journal entries: got %d, want 2", entries)
				}
			},
		},
		{
			name: "已核准的明細再次核准：回傳 already approved 錯誤，交易仍只有 1 筆",
			prepare: func(t *testing.T, db *sqlx.DB, svc services.BankStatementReviewService, bankTxnID int64) {
				mustApprove(t, svc, bankTxnID)
			},
			wantErr: "already approved",
			check: func(t *testing.T, db *sqlx.DB, bankTxnID int64) {
				expectApprovedOnce(t, db, bankTxnID)
			},
		},
		{
			name: "交易已建立但關聯回寫失敗後重試：不新增交易，關聯回原本的交易",
			prepare: func(t *testing.T, db *sqlx.DB, svc services.BankStatementReviewService, bankTxnID int64) {
				mustApprove(t, svc, bankTxnID)
				expectApprovedOnce(t, db, bankTxnID)
				if _, err := db.Exec(`UPDATE bank_statement_txns SET match_status = 'UNMATCHED', created_txn_id = NULL WHERE bank_txn_id = ?`, bankTxnID); err != nil {
					t.Fatalf("reset bank statement txn link: %v", err)
				}
			},
			check: func(t *testing.T, db *sqlx.DB, bankTxnID int64) {
				// Exactly one transaction still exists, so the link points to the original one.
				expectApprovedOnce(t, db, bankTxnID)
			},
		},
		{
			name: "固定 aggregate 已有版本紀錄但沒有交易事件：Append 失敗且重查不到事件，回傳錯誤、不建立交易、明細維持 UNMATCHED",
			prepare: func(t *testing.T, db *sqlx.DB, svc services.BankStatementReviewService, bankTxnID int64) {
				// Same derivation as ApproveTxn's fixed aggregate ID.
				aggID := uuidx.NewFromEvent(approveTestBankTxnUUID, "bank_approve")
				if _, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
					VALUES('TRANSACTION', ?, ?, 1)`, aggID, testMID); err != nil {
					t.Fatalf("insert aggregate version: %v", err)
				}
			},
			wantErr: "create transaction",
			check: func(t *testing.T, db *sqlx.DB, bankTxnID int64) {
				if ids := approveTxnIDs(t, db); len(ids) != 0 {
					t.Errorf("transactions: got %v, want none", ids)
				}
				row := approveBankTxn(t, db, bankTxnID)
				if row.MatchStatus != "UNMATCHED" {
					t.Errorf("match_status: got %s, want UNMATCHED", row.MatchStatus)
				}
				if row.CreatedTxnID != nil {
					t.Errorf("created_txn_id: got %d, want NULL", *row.CreatedTxnID)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, bankTxnID := newApproveTestDB(t)
			svc := newApproveService(db, query.NewQueryRepository(db))
			if tc.prepare != nil {
				tc.prepare(t, db, svc, bankTxnID)
			}

			err := svc.ApproveTxn(testCtx(), approveRequest(bankTxnID))

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ApproveTxn: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ApproveTxn err = %v, want containing %q", err, tc.wantErr)
			}
			tc.check(t, db, bankTxnID)
		})
	}
}

// TestApproveTxn_ConcurrentApproval forces both requests to see no transaction
// event before appending, with the rival finishing first, so this request's
// Append conflicts and must reuse the rival's transaction via the recovery path.
func TestApproveTxn_ConcurrentApproval(t *testing.T) {
	db, bankTxnID := newApproveTestDB(t)
	realQ := query.NewQueryRepository(db)
	rival := newApproveService(db, realQ)

	race := &raceEventRepo{
		EventRepo: realQ.Event,
		onFirstLookup: func() {
			mustApprove(t, rival, bankTxnID)
		},
	}
	stubbedQ := *realQ
	stubbedQ.Event = race
	svc := newApproveService(db, &stubbedQ)

	mustApprove(t, svc, bankTxnID)

	// The second lookup only happens after Append failed, i.e. on the recovery path.
	if race.lookups != 2 {
		t.Errorf("event lookups: got %d, want 2", race.lookups)
	}
	expectApprovedOnce(t, db, bankTxnID)
}
