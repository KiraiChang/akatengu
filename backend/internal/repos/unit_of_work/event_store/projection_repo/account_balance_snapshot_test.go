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

// 使用 seeds 中真實的科目（seeds 以 merchant_id=1 插入）
const (
	testMerchantID int64 = 1

	// 葉科目（is_summary=0）
	acctCash    = "1101-01" // 手頭現金（ASSET, DEBIT normal）
	acctCash2   = "1101-02" // 銀行活期存款（ASSET, DEBIT normal）
	acctExpense = "5101-01" // 房租費用（EXPENSE, DEBIT normal）
	// 匯總科目（is_summary=1）
	acctCash1101Parent = "1101" // 現金及約當現金（1101-01 / 1101-02 的父科目）
	acctCash110Parent  = "110"  // 流動資產（1101 的父科目）
	acctExp5101Parent  = "5101" // 居住費用（5101-01 的父科目）
	acctExp510Parent   = "510"  // 生活費用（5101 的父科目）
)

func insertPeriod(t *testing.T, db *sqlx.DB, id int64, periodType string, start, end string, closed bool) {
	t.Helper()
	status := string(enums.PeriodTypeStatusOpen)
	var closedAt *string
	if closed {
		status = string(enums.PeriodTypeStatusClosed)
		closedAt = &end
	}
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO period_closings (closing_id, merchant_id, period_type, period_start, period_end, status, closed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, testMerchantID, periodType, start, end, status, closedAt)
	if err != nil {
		t.Fatalf("insertPeriod id=%d type=%s: %v", id, periodType, err)
	}
}

// insertTxn 插入一筆借貸平衡的交易（使用 testMerchantID）。
func insertTxn(t *testing.T, db *sqlx.DB, txnID int64, date, debitAcct, creditAcct string, amount float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO transactions (txn_id, merchant_id, txn_date, description, total_amount, version)
		 VALUES (?, ?, ?, '測試交易', ?, 1)`,
		txnID, testMerchantID, date, amount)
	if err != nil {
		t.Fatalf("insertTxn id=%d: %v", txnID, err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO journal_entries (txn_id, merchant_id, account_id, debit, credit) VALUES (?, ?, ?, ?, 0)`,
		txnID, testMerchantID, debitAcct, amount)
	if err != nil {
		t.Fatalf("insertTxn debit: %v", err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO journal_entries (txn_id, merchant_id, account_id, debit, credit) VALUES (?, ?, ?, 0, ?)`,
		txnID, testMerchantID, creditAcct, amount)
	if err != nil {
		t.Fatalf("insertTxn credit: %v", err)
	}
}

// insertExpenseTxn 插入一筆交易：借 acctExpense amount，貸 acctCash amount。
func insertExpenseTxn(t *testing.T, db *sqlx.DB, txnID int64, date string, amount float64) {
	t.Helper()
	insertTxn(t, db, txnID, date, acctExpense, acctCash, amount)
}

// insertSnapshotRow 直接寫入快照資料，用於測試情境的前置準備。
func insertSnapshotRow(t *testing.T, db *sqlx.DB, closingID int64, accountID string, debit, credit float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO account_balance_snapshots (merchant_id, closing_id, account_id, debit_total, credit_total)
		 VALUES (?, ?, ?, ?, ?)`,
		testMerchantID, closingID, accountID, debit, credit)
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

	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
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

	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	insertExpenseTxn(t, db, 2, "2025-02-10", 50)
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
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

	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
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
	insertPeriod(t, db, 10, string(enums.PeriodAnnual), "2024-01-01", "2024-12-31", true)
	insertSnapshotRow(t, db, 10, acctExpense, 1000, 0)
	insertSnapshotRow(t, db, 10, acctCash, 0, 1000)

	// 月結 2025-01（id=11）已關帳，period_end='2025-01-31' >= '2025-01-01'
	// 不符合 period_end < annual.period_start，不應被選為年結基準
	insertPeriod(t, db, 11, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertSnapshotRow(t, db, 11, acctExpense, 500, 0) // 若誤用此值 → 500+200=700（錯誤）
	insertSnapshotRow(t, db, 11, acctCash, 0, 500)

	// 年結 2025（id=12），period_start='2025-01-01'
	insertPeriod(t, db, 12, string(enums.PeriodAnnual), "2025-01-01", "2025-12-31", true)
	insertExpenseTxn(t, db, 1, "2025-06-15", 200) // 2025 全年唯一交易

	if err := repo.BulkInsert(ctx, testMerchantID, 12); err != nil {
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

	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert 1: %v", err)
	}
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert 2: %v", err)
	}

	if err := repo.DeleteByClosingId(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("DeleteByClosingId: %v", err)
	}

	if snap := querySnapshots(t, db, 1); len(snap) != 0 {
		t.Errorf("period 1 snapshots should be deleted, got %d rows", len(snap))
	}
	if snap := querySnapshots(t, db, 2); len(snap) == 0 {
		t.Error("period 2 snapshots should remain after deleting period 1")
	}
}

// TestBulkInsert_ParentAggregatesLeaf 父科目快照應等於其子葉科目快照的加總（兩層）。
func TestBulkInsert_ParentAggregatesLeaf(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	// DR 5101-01 100, CR 1101-01 100
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)

	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	snap := querySnapshots(t, db, 1)
	// 葉科目
	assertSnapshot(t, snap, acctExpense, 100, 0)
	assertSnapshot(t, snap, acctCash, 0, 100)
	// 直接父科目
	assertSnapshot(t, snap, acctExp5101Parent, 100, 0)
	assertSnapshot(t, snap, acctCash1101Parent, 0, 100)
	// 祖父科目
	assertSnapshot(t, snap, acctExp510Parent, 100, 0)
	assertSnapshot(t, snap, acctCash110Parent, 0, 100)
}

// TestBulkInsert_ParentAggregatesTwoLeaves 父科目應加總同層的兩個葉科目。
func TestBulkInsert_ParentAggregatesTwoLeaves(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	// DR 5101-01 100, CR 1101-01 100
	insertTxn(t, db, 1, "2025-01-10", acctExpense, acctCash, 100)
	// DR 5101-01 50, CR 1101-02 50（同一父科目 1101 下的另一葉科目）
	insertTxn(t, db, 2, "2025-01-20", acctExpense, acctCash2, 50)

	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	snap := querySnapshots(t, db, 1)
	// 葉科目各自累積
	assertSnapshot(t, snap, acctCash, 0, 100)
	assertSnapshot(t, snap, acctCash2, 0, 50)
	assertSnapshot(t, snap, acctExpense, 150, 0)
	// 父科目 1101 應加總 1101-01 + 1101-02
	assertSnapshot(t, snap, acctCash1101Parent, 0, 150)
	// 父科目 5101 僅有 5101-01
	assertSnapshot(t, snap, acctExp5101Parent, 150, 0)
}

// TestBulkInsert_ParentIncrementalAggregation 父科目快照在連續月結中應累積正確數值。
func TestBulkInsert_ParentIncrementalAggregation(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	ctx := context.Background()

	// 一月：100
	insertPeriod(t, db, 1, string(enums.PeriodMonthly), "2025-01-01", "2025-01-31", true)
	insertExpenseTxn(t, db, 1, "2025-01-15", 100)
	if err := repo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert Jan: %v", err)
	}

	snap1 := querySnapshots(t, db, 1)
	assertSnapshot(t, snap1, acctExp5101Parent, 100, 0)
	assertSnapshot(t, snap1, acctCash1101Parent, 0, 100)

	// 二月：再加 50，累積應為 150
	insertPeriod(t, db, 2, string(enums.PeriodMonthly), "2025-02-01", "2025-02-28", true)
	insertExpenseTxn(t, db, 2, "2025-02-10", 50)
	if err := repo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert Feb: %v", err)
	}

	snap2 := querySnapshots(t, db, 2)
	// 葉科目累積：100 + 50 = 150
	assertSnapshot(t, snap2, acctExpense, 150, 0)
	assertSnapshot(t, snap2, acctCash, 0, 150)
	// 父科目由葉科目快照加總得出，同樣應為 150
	assertSnapshot(t, snap2, acctExp5101Parent, 150, 0)
	assertSnapshot(t, snap2, acctCash1101Parent, 0, 150)
	assertSnapshot(t, snap2, acctExp510Parent, 150, 0)
	assertSnapshot(t, snap2, acctCash110Parent, 0, 150)
}
