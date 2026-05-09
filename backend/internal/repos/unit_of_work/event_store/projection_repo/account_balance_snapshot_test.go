package projection_repo_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/testutil"
)

// 使用 seeds 中真實的非匯總科目
const (
	acctCash    = "1101-01" // 手頭現金（ASSET, DEBIT normal）
	acctExpense = "5101-01" // 房租費用（EXPENSE, DEBIT normal）
)

func insertMonthly(t *testing.T, db *sqlx.DB, id int64, start, end string, closed bool) {
	t.Helper()
	status := "OPEN"
	var closedAt *string
	if closed {
		status = "CLOSED"
		closedAt = &end
	}
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO period_closings (closing_id, period_type, period_start, period_end, status, closed_at)
		 VALUES (?, 'MONTHLY', ?, ?, ?, ?)`,
		id, start, end, status, closedAt)
	if err != nil {
		t.Fatalf("insertMonthly id=%d: %v", id, err)
	}
}

func insertAnnual(t *testing.T, db *sqlx.DB, id int64, start, end string, closed bool) {
	t.Helper()
	status := "OPEN"
	var closedAt *string
	if closed {
		status = "CLOSED"
		closedAt = &end
	}
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO period_closings (closing_id, period_type, period_start, period_end, status, closed_at)
		 VALUES (?, 'ANNUAL', ?, ?, ?, ?)`,
		id, start, end, status, closedAt)
	if err != nil {
		t.Fatalf("insertAnnual id=%d: %v", id, err)
	}
}

// insertExpenseTxn 插入一筆交易：借 acctExpense amount，貸 acctCash amount。
func insertExpenseTxn(t *testing.T, db *sqlx.DB, txnID int64, date string, amount float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO transactions (txn_id, txn_date, description, total_amount, version)
		 VALUES (?, ?, '測試交易', ?, 1)`,
		txnID, date, amount)
	if err != nil {
		t.Fatalf("insertTxn id=%d: %v", txnID, err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO journal_entries (txn_id, account_id, debit, credit) VALUES (?, ?, ?, 0)`,
		txnID, acctExpense, amount)
	if err != nil {
		t.Fatalf("insertEntry debit: %v", err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO journal_entries (txn_id, account_id, debit, credit) VALUES (?, ?, 0, ?)`,
		txnID, acctCash, amount)
	if err != nil {
		t.Fatalf("insertEntry credit: %v", err)
	}
}

// insertSnapshotRow 直接寫入快照資料，用於測試情境的前置準備。
func insertSnapshotRow(t *testing.T, db *sqlx.DB, closingID int64, accountID string, debit, credit float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO account_balance_snapshots (closing_id, account_id, debit_total, credit_total)
		 VALUES (?, ?, ?, ?)`,
		closingID, accountID, debit, credit)
	if err != nil {
		t.Fatalf("insertSnapshotRow closing=%d account=%s: %v", closingID, accountID, err)
	}
}

type snapshotRow struct{ Debit, Credit float64 }

func querySnapshots(t *testing.T, db *sqlx.DB, closingID int64) map[string]snapshotRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT account_id, debit_total, credit_total
		 FROM account_balance_snapshots WHERE closing_id = ?`,
		closingID)
	if err != nil {
		t.Fatalf("querySnapshots closing=%d: %v", closingID, err)
	}
	defer rows.Close()
	result := make(map[string]snapshotRow)
	for rows.Next() {
		var acct string
		var row snapshotRow
		if err := rows.Scan(&acct, &row.Debit, &row.Credit); err != nil {
			t.Fatalf("scan snapshot: %v", err)
		}
		result[acct] = row
	}
	return result
}

func assertSnapshot(t *testing.T, got map[string]snapshotRow, acct string, wantDebit, wantCredit float64) {
	t.Helper()
	row, ok := got[acct]
	if !ok {
		t.Errorf("account %s not found in snapshots", acct)
		return
	}
	if row.Debit != wantDebit {
		t.Errorf("account %s debit_total: got %.2f, want %.2f", acct, row.Debit, wantDebit)
	}
	if row.Credit != wantCredit {
		t.Errorf("account %s credit_total: got %.2f, want %.2f", acct, row.Credit, wantCredit)
	}
}

// TestBulkInsert_FirstPeriod_NoHistory 第一次月結無前期快照，快照 = 本期交易加總。
func TestBulkInsert_FirstPeriod_NoHistory(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertMonthly(t, db, 1, "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	snap := querySnapshots(t, db, 1)
	assertSnapshot(t, snap, acctExpense, 100, 0)
	assertSnapshot(t, snap, acctCash, 0, 100)
}

// TestBulkInsert_SecondPeriod_Incremental 第二次月結快照 = 前期快照 + 本期增量。
func TestBulkInsert_SecondPeriod_Incremental(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertMonthly(t, db, 1, "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	insertMonthly(t, db, 2, "2025-02-01", "2025-02-28", true)
	insertExpenseTxn(t, db, 2, "2025-02-10", 50)
	if err := repo.BulkInsert(ctx, 2); err != nil {
		t.Fatalf("BulkInsert Feb: %v", err)
	}

	snap := querySnapshots(t, db, 2)
	assertSnapshot(t, snap, acctExpense, 150, 0) // 100 + 50
	assertSnapshot(t, snap, acctCash, 0, 150)    // 100 + 50
}

// TestBulkInsert_EmptyPeriod_CopiesPrevSnapshot 空白期間無交易，快照沿用前期數值。
func TestBulkInsert_EmptyPeriod_CopiesPrevSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertMonthly(t, db, 1, "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	// 二月無任何交易
	insertMonthly(t, db, 2, "2025-02-01", "2025-02-28", true)
	if err := repo.BulkInsert(ctx, 2); err != nil {
		t.Fatalf("BulkInsert Feb (empty): %v", err)
	}

	snap := querySnapshots(t, db, 2)
	assertSnapshot(t, snap, acctExpense, 100, 0) // delta=0，等於一月快照
	assertSnapshot(t, snap, acctCash, 0, 100)
}

// TestBulkInsert_Annual_UsesPrevAnnualSnapshot 年結以上一年年結快照為基準，
// 不使用月結快照鏈；period_end < annual.period_start 過濾確保月結快照不被誤選。
func TestBulkInsert_Annual_UsesPrevAnnualSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	// 年結 2024（id=10）作為基準，直接插入快照
	insertAnnual(t, db, 10, "2024-01-01", "2024-12-31", true)
	insertSnapshotRow(t, db, 10, acctExpense, 1000, 0)
	insertSnapshotRow(t, db, 10, acctCash, 0, 1000)

	// 月結 2025-01（id=11）已關帳，period_end='2025-01-31' >= '2025-01-01'
	// 不符合 period_end < annual.period_start，不應被選為年結基準
	insertMonthly(t, db, 11, "2025-01-01", "2025-01-31", true)
	insertSnapshotRow(t, db, 11, acctExpense, 500, 0) // 若誤用此值 → 500+200=700（錯誤）
	insertSnapshotRow(t, db, 11, acctCash, 0, 500)

	// 年結 2025（id=12），period_start='2025-01-01'
	insertAnnual(t, db, 12, "2025-01-01", "2025-12-31", true)
	insertExpenseTxn(t, db, 1, "2025-06-15", 200) // 2025 全年唯一交易

	if err := repo.BulkInsert(ctx, 12); err != nil {
		t.Fatalf("BulkInsert Annual 2025: %v", err)
	}

	snap := querySnapshots(t, db, 12)
	// 正確路徑：年結 2024（1000）+ 2025 delta（200）= 1200
	// 誤用月結（500）+ 200 = 700；無基準（0）+ 200 = 200
	assertSnapshot(t, snap, acctExpense, 1200, 0)
	assertSnapshot(t, snap, acctCash, 0, 1200)
}

// TestDeleteByClosingId_RemovesOnlyTargetPeriod 只刪除指定期間的快照，其他期間不受影響。
func TestDeleteByClosingId_RemovesOnlyTargetPeriod(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertMonthly(t, db, 1, "2025-01-01", "2025-01-31", true)
	insertMonthly(t, db, 2, "2025-02-01", "2025-02-28", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, 1); err != nil {
		t.Fatalf("BulkInsert 1: %v", err)
	}
	if err := repo.BulkInsert(ctx, 2); err != nil {
		t.Fatalf("BulkInsert 2: %v", err)
	}

	if err := repo.DeleteByClosingId(ctx, 1); err != nil {
		t.Fatalf("DeleteByClosingId: %v", err)
	}

	if snap := querySnapshots(t, db, 1); len(snap) != 0 {
		t.Errorf("period 1 snapshots should be deleted, got %d rows", len(snap))
	}
	if snap := querySnapshots(t, db, 2); len(snap) == 0 {
		t.Error("period 2 snapshots should remain after deleting period 1")
	}
}
