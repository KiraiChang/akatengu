package projection_repo_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/testutil"
)

const (
	ledgerCashAcct    = "1101-01" // 手頭現金（ASSET, DEBIT normal）
	ledgerExpenseAcct = "5101-01" // 房租費用（EXPENSE, DEBIT normal）
	ledgerCashId      = int64(1)
	ledgerExpenseId   = int64(2)
)

func insertLedgerAccounts(t *testing.T, db *sqlx.DB) {
	t.Helper()
	ctx := context.Background()
	for _, row := range []struct {
		id        int64
		accountId string
	}{
		{ledgerCashId, ledgerCashAcct},
		{ledgerExpenseId, ledgerExpenseAcct},
	} {
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO ledger_accounts (ledger_id, merchant_id, account_id, institution, name, type, currency, version)
			 VALUES (?, ?, ?, '測試銀行', '帳戶', 'BANK_ACCOUNT', 'TWD', 1)`,
			row.id, testMerchantID, row.accountId)
		if err != nil {
			t.Fatalf("insert ledger_account id=%d: %v", row.id, err)
		}
	}
}

// insertLedgerTxn 插入一筆有 ledger_id 的交易：DR ledger 2 (費用), CR ledger 1 (現金)。
func insertLedgerTxn(t *testing.T, db *sqlx.DB, txnID int64, date string, amount float64) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO transactions (txn_id, merchant_id, txn_date, description, total_amount, version) VALUES (?, ?, ?, '測試', ?, 1)`,
		txnID, testMerchantID, date, amount)
	if err != nil {
		t.Fatalf("insert txn id=%d: %v", txnID, err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries (txn_id, merchant_id, ledger_id, account_id, debit, credit) VALUES (?, ?, ?, ?, ?, 0)`,
		txnID, testMerchantID, ledgerExpenseId, ledgerExpenseAcct, amount)
	if err != nil {
		t.Fatalf("insert entry debit: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries (txn_id, merchant_id, ledger_id, account_id, debit, credit) VALUES (?, ?, ?, ?, 0, ?)`,
		txnID, testMerchantID, ledgerCashId, ledgerCashAcct, amount)
	if err != nil {
		t.Fatalf("insert entry credit: %v", err)
	}
}

func insertLedgerSnapshotRow(t *testing.T, db *sqlx.DB, closingID, ledgerID int64, debit, credit float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO ledger_account_balance_snapshots (merchant_id, closing_id, ledger_id, debit_total, credit_total) VALUES (?, ?, ?, ?, ?)`,
		testMerchantID, closingID, ledgerID, debit, credit)
	if err != nil {
		t.Fatalf("insertLedgerSnapshotRow closing=%d ledger=%d: %v", closingID, ledgerID, err)
	}
}

type ledgerSnapshotRow struct{ Debit, Credit float64 }

func queryLedgerSnapshots(t *testing.T, db *sqlx.DB, closingID int64) map[int64]ledgerSnapshotRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT ledger_id, debit_total, credit_total FROM ledger_account_balance_snapshots WHERE closing_id = ?`,
		closingID)
	if err != nil {
		t.Fatalf("queryLedgerSnapshots closing=%d: %v", closingID, err)
	}
	defer rows.Close()
	result := make(map[int64]ledgerSnapshotRow)
	for rows.Next() {
		var id int64
		var row ledgerSnapshotRow
		if err := rows.Scan(&id, &row.Debit, &row.Credit); err != nil {
			t.Fatalf("scan ledger snapshot: %v", err)
		}
		result[id] = row
	}
	return result
}

func assertLedgerSnapshot(t *testing.T, got map[int64]ledgerSnapshotRow, ledgerID int64, wantDebit, wantCredit float64) {
	t.Helper()
	row, ok := got[ledgerID]
	if !ok {
		t.Errorf("ledger_id %d not found in snapshots", ledgerID)
		return
	}
	if row.Debit != wantDebit {
		t.Errorf("ledger_id %d debit_total: got %.2f, want %.2f", ledgerID, row.Debit, wantDebit)
	}
	if row.Credit != wantCredit {
		t.Errorf("ledger_id %d credit_total: got %.2f, want %.2f", ledgerID, row.Credit, wantCredit)
	}
}

// TestLedger_BulkInsert_FirstPeriod_NoHistory 第一次月結無前期快照，快照 = 本期交易加總。
func TestLedger_BulkInsert_FirstPeriod_NoHistory(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewLedgerAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertLedgerAccounts(t, db)
	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertLedgerTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	snap := queryLedgerSnapshots(t, db, 1)
	assertLedgerSnapshot(t, snap, ledgerExpenseId, 100, 0)
	assertLedgerSnapshot(t, snap, ledgerCashId, 0, 100)
}

// TestLedger_BulkInsert_SecondPeriod_Incremental 第二次月結快照 = 前期快照 + 本期增量。
func TestLedger_BulkInsert_SecondPeriod_Incremental(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewLedgerAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertLedgerAccounts(t, db)
	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertLedgerTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	insertLedgerTxn(t, db, 2, "2025-02-10", 50)
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert Feb: %v", err)
	}

	snap := queryLedgerSnapshots(t, db, 2)
	assertLedgerSnapshot(t, snap, ledgerExpenseId, 150, 0) // 100 + 50
	assertLedgerSnapshot(t, snap, ledgerCashId, 0, 150)    // 100 + 50
}

// TestLedger_BulkInsert_EmptyPeriod_CopiesPrevSnapshot 本期無交易，快照沿用前期數值。
func TestLedger_BulkInsert_EmptyPeriod_CopiesPrevSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewLedgerAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertLedgerAccounts(t, db)
	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertLedgerTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert Feb (empty): %v", err)
	}

	snap := queryLedgerSnapshots(t, db, 2)
	assertLedgerSnapshot(t, snap, ledgerExpenseId, 100, 0)
	assertLedgerSnapshot(t, snap, ledgerCashId, 0, 100)
}

// TestLedger_BulkInsert_Annual_UsesPrevAnnualSnapshot 年結以前年年結快照為基準，不使用月結快照。
func TestLedger_BulkInsert_Annual_UsesPrevAnnualSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewLedgerAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertLedgerAccounts(t, db)

	// 年結 2024（id=10）作為基準，直接插入快照
	insertPeriod(t, db, 10, string(enums.PeriodAnnual), "2024-01-01", "2024-12-31", true)
	insertLedgerSnapshotRow(t, db, 10, ledgerExpenseId, 1000, 0)
	insertLedgerSnapshotRow(t, db, 10, ledgerCashId, 0, 1000)

	// 月結 2025-01（id=11）：period_end='2025-01-31' >= '2025-01-01'，不應被選為年結基準
	insertPeriod(t, db, 11, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertLedgerSnapshotRow(t, db, 11, ledgerExpenseId, 500, 0)
	insertLedgerSnapshotRow(t, db, 11, ledgerCashId, 0, 500)

	// 年結 2025（id=12），本年唯一交易
	insertPeriod(t, db, 12, string(enums.PeriodAnnual), "2025-01-01", "2025-12-31", true)
	insertLedgerTxn(t, db, 1, "2025-06-15", 200)

	if err := repo.BulkInsert(ctx, testMerchantID, 12); err != nil {
		t.Fatalf("BulkInsert Annual 2025: %v", err)
	}

	snap := queryLedgerSnapshots(t, db, 12)
	// 正確路徑：年結 2024（1000）+ 2025 delta（200）= 1200
	assertLedgerSnapshot(t, snap, ledgerExpenseId, 1200, 0)
	assertLedgerSnapshot(t, snap, ledgerCashId, 0, 1200)
}

// TestLedger_DeleteByClosingId_RemovesOnlyTarget 只刪除指定期間快照，不影響其他期間。
func TestLedger_DeleteByClosingId_RemovesOnlyTarget(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewLedgerAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertLedgerAccounts(t, db)
	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	insertLedgerTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert 1: %v", err)
	}
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert 2: %v", err)
	}
	if err := repo.DeleteByClosingId(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("DeleteByClosingId: %v", err)
	}

	if snap := queryLedgerSnapshots(t, db, 1); len(snap) != 0 {
		t.Errorf("period 1 snapshots should be deleted, got %d rows", len(snap))
	}
	if snap := queryLedgerSnapshots(t, db, 2); len(snap) == 0 {
		t.Error("period 2 snapshots should remain after deleting period 1")
	}
}
