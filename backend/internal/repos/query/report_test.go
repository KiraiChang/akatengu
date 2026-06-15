package query_test

import (
	"context"
	"fmt"
	"testing"

	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/report"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/testutil"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// ── 報表測試用科目常數（來自 accounts.sql seeds）─────────────────────────────

const (
	rptAcctCash     = "1101-01" // 手頭現金 (ASSET, DEBIT)
	rptAcctCash2    = "1101-02" // 活期存款 (ASSET, DEBIT)
	rptAcctCash1101 = "1101"    // 現金及約當現金 (ASSET, DEBIT, summary)
	rptAcctCash110  = "110"     // 流動資產 (ASSET, DEBIT, summary)

	rptAcctCC     = "2101-01" // 信用卡應付款 (LIABILITY, CREDIT)
	rptAcctCC2101 = "2101"    // 應付款項 (LIABILITY, CREDIT, summary)

	rptAcctSalary     = "4101-01" // 月薪 (INCOME, CREDIT)
	rptAcctBonus      = "4101-02" // 獎金及年終 (INCOME, CREDIT)
	rptAcctSalary4101 = "4101"    // 薪資所得 (INCOME, CREDIT, summary)
	rptAcctSalary410  = "410"     // 勞動所得 (INCOME, CREDIT, summary)

	rptAcctRent     = "5101-01" // 房租費用 (EXPENSE, DEBIT)
	rptAcctInterest = "5101-02" // 房屋貸款利息費用 (EXPENSE, DEBIT)
	rptAcctRent5101 = "5101"    // 居住費用 (EXPENSE, DEBIT, summary)
	rptAcctRent510  = "510"     // 生活費用 (EXPENSE, DEBIT, summary)
)

// ── 輔助函式 ──────────────────────────────────────────────────────────────────

func findBSRow(bs *report.BalanceSheet, accountId string) (decimal.Decimal, bool) {
	r, ok := findBSRowFull(bs, accountId)
	return r.Balance, ok
}

func findBSRowFull(bs *report.BalanceSheet, accountId string) (report.BalanceSheetRow, bool) {
	for _, r := range bs.Assets {
		if r.AccountId == accountId {
			return r, true
		}
	}
	for _, r := range bs.Liabilities {
		if r.AccountId == accountId {
			return r, true
		}
	}
	for _, r := range bs.Equity {
		if r.AccountId == accountId {
			return r, true
		}
	}
	return report.BalanceSheetRow{}, false
}

func assertBSBalance(t *testing.T, bs *report.BalanceSheet, accountId string, want float64) {
	t.Helper()
	balance, ok := findBSRow(bs, accountId)
	if !ok {
		t.Errorf("account %s not found in balance sheet", accountId)
		return
	}
	got, _ := balance.Float64()
	if got != want {
		t.Errorf("account %s balance: got %.2f, want %.2f", accountId, got, want)
	}
}

func findISRow(is *report.IncomeStatement, accountId string) (decimal.Decimal, bool) {
	r, ok := findISRowFull(is, accountId)
	return r.Amount, ok
}

func findISRowFull(is *report.IncomeStatement, accountId string) (report.IncomeStatementRow, bool) {
	for _, r := range is.Income {
		if r.AccountID == accountId {
			return r, true
		}
	}
	for _, r := range is.Expenses {
		if r.AccountID == accountId {
			return r, true
		}
	}
	return report.IncomeStatementRow{}, false
}

func assertISAmount(t *testing.T, is *report.IncomeStatement, accountId string, want float64) {
	t.Helper()
	amount, ok := findISRow(is, accountId)
	if !ok {
		t.Errorf("account %s not found in income statement", accountId)
		return
	}
	got, _ := amount.Float64()
	if got != want {
		t.Errorf("account %s amount: got %.2f, want %.2f", accountId, got, want)
	}
}

// ── GetBalanceSheet 測試 ──────────────────────────────────────────────────────

// TestGetBalanceSheet_NoSnapshot_LeafOnly
// 無月結，直接掃描 journal_entries；葉科目借貸正確反映 normal_balance。
func TestGetBalanceSheet_NoSnapshot_LeafOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 1101-01 (ASSET) 1000, CR 2101-01 (LIABILITY) 1000
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctCC, 1000)

	bs, err := repo.GetBalanceSheet(ctx, "2025-01-31")
	if err != nil {
		t.Fatalf("GetBalanceSheet: %v", err)
	}

	// ASSET (DEBIT normal): balance = debit - credit = 1000
	assertBSBalance(t, bs, rptAcctCash, 1000)
	// LIABILITY (CREDIT normal): balance = credit - debit = 1000
	assertBSBalance(t, bs, rptAcctCC, 1000)
}

// TestGetBalanceSheet_NoSnapshot_ParentAggregates
// 無月結，父科目餘額應等於所有子葉科目的加總。
func TestGetBalanceSheet_NoSnapshot_ParentAggregates(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 兩筆分別使用 1101 下的兩個葉科目
	sumInsertTxn(t, db, 1, "2025-01-10", rptAcctCash, rptAcctCC, 600)  // DR 1101-01 600
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctCash2, rptAcctCC, 400) // DR 1101-02 400

	bs, err := repo.GetBalanceSheet(ctx, "2025-01-31")
	if err != nil {
		t.Fatalf("GetBalanceSheet: %v", err)
	}

	assertBSBalance(t, bs, rptAcctCash, 600)
	assertBSBalance(t, bs, rptAcctCash2, 400)
	// 父科目 1101 = 600 + 400 = 1000
	assertBSBalance(t, bs, rptAcctCash1101, 1000)
	assertBSBalance(t, bs, rptAcctCash110, 1000)
	// 負債父科目 2101 = 1000
	assertBSBalance(t, bs, rptAcctCC2101, 1000)
}

// TestGetBalanceSheet_WithSnapshot_NoPostDelta
// 月結後無新交易，查詢結果應等於快照值（含父科目）。
func TestGetBalanceSheet_WithSnapshot_NoPostDelta(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapRepo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	sumInsertPeriod(t, db, 1, "MONTHLY", "2025-01-01", "2025-01-31", true)
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctCC, 1000)
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	bs, err := repo.GetBalanceSheet(ctx, "2025-01-31")
	if err != nil {
		t.Fatalf("GetBalanceSheet: %v", err)
	}

	assertBSBalance(t, bs, rptAcctCash, 1000)
	assertBSBalance(t, bs, rptAcctCash1101, 1000)
	assertBSBalance(t, bs, rptAcctCash110, 1000)
	assertBSBalance(t, bs, rptAcctCC, 1000)
	assertBSBalance(t, bs, rptAcctCC2101, 1000)
}

// TestGetBalanceSheet_WithSnapshot_WithDelta
// 月結後新增交易，結果應為快照 + delta，父科目亦同步聚合 delta。
func TestGetBalanceSheet_WithSnapshot_WithDelta(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapRepo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 一月月結：1000
	sumInsertPeriod(t, db, 1, "MONTHLY", "2025-01-01", "2025-01-31", true)
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctCC, 1000)
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	// 月結後追加 delta 500
	sumInsertTxn(t, db, 2, "2025-02-10", rptAcctCash, rptAcctCC, 500)

	bs, err := repo.GetBalanceSheet(ctx, "2025-02-28")
	if err != nil {
		t.Fatalf("GetBalanceSheet: %v", err)
	}

	// 葉科目：快照 1000 + delta 500 = 1500
	assertBSBalance(t, bs, rptAcctCash, 1500)
	assertBSBalance(t, bs, rptAcctCC, 1500)
	// 父科目同步
	assertBSBalance(t, bs, rptAcctCash1101, 1500)
	assertBSBalance(t, bs, rptAcctCash110, 1500)
	assertBSBalance(t, bs, rptAcctCC2101, 1500)
}

// ── GetIncomeStatement 測試 ───────────────────────────────────────────────────

// TestGetIncomeStatement_NoSnapshot_LeafOnly
// 無月結，直接掃描期間分錄；INCOME/EXPENSE 葉科目金額正確。
func TestGetIncomeStatement_NoSnapshot_LeafOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 5101-01 (EXPENSE) 3000, CR 4101-01 (INCOME) 3000
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctRent, rptAcctSalary, 3000)

	is, err := repo.GetIncomeStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetIncomeStatement: %v", err)
	}

	// INCOME (CREDIT normal): amount = credit - debit = 3000
	assertISAmount(t, is, rptAcctSalary, 3000)
	// EXPENSE (DEBIT normal): amount = debit - credit = 3000
	assertISAmount(t, is, rptAcctRent, 3000)
}

// TestGetIncomeStatement_NoSnapshot_ParentAggregates
// 無月結，同一父科目下多個葉科目的金額應被正確聚合。
func TestGetIncomeStatement_NoSnapshot_ParentAggregates(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 兩筆薪資（4101-01 月薪 + 4101-02 獎金）
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctRent, rptAcctSalary, 2000) // CR 4101-01 2000
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctRent, rptAcctBonus, 500)   // CR 4101-02 500
	// 兩筆居住費用（5101-01 房租 + 5101-02 利息）
	sumInsertTxn(t, db, 3, "2025-01-22", rptAcctInterest, rptAcctCash, 300) // DR 5101-02 300

	is, err := repo.GetIncomeStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetIncomeStatement: %v", err)
	}

	// 葉科目個別
	assertISAmount(t, is, rptAcctSalary, 2000)
	assertISAmount(t, is, rptAcctBonus, 500)
	assertISAmount(t, is, rptAcctRent, 2500) // 房租：debit 2000+500 (from txn1 and txn2)
	assertISAmount(t, is, rptAcctInterest, 300)
	// 父科目聚合
	assertISAmount(t, is, rptAcctSalary4101, 2500) // 4101-01 + 4101-02
	assertISAmount(t, is, rptAcctSalary410, 2500)
	assertISAmount(t, is, rptAcctRent5101, 2800) // 5101-01 + 5101-02
	assertISAmount(t, is, rptAcctRent510, 2800)
}

// TestGetIncomeStatement_BothSnaps_MonthlyAligned
// snap_pre.period_end = startDate-1（月底對齊），delta_pre 為空；
// period_income = snap_end - snap_pre。
func TestGetIncomeStatement_BothSnaps_MonthlyAligned(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapRepo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 十二月交易 → snap_pre
	sumInsertPeriod(t, db, 1, "MONTHLY", "2024-12-01", "2024-12-31", true)
	sumInsertTxn(t, db, 1, "2024-12-15", rptAcctRent, rptAcctSalary, 1000) // CR 薪資 1000
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert snap_pre: %v", err)
	}

	// 一月交易 → snap_end
	sumInsertPeriod(t, db, 2, "MONTHLY", "2025-01-01", "2025-01-31", true)
	sumInsertTxn(t, db, 2, "2025-01-15", rptAcctRent, rptAcctSalary, 2000) // CR 薪資 2000
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert snap_end: %v", err)
	}

	// 查詢一月損益 [2025-01-01, 2025-01-31]
	// snap_end credit=3000, snap_pre credit=1000, delta_pre=0, delta_end=0
	// period = (3000+0) - (1000+0) = 2000
	is, err := repo.GetIncomeStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetIncomeStatement: %v", err)
	}

	// salary: (3000+0) - (1000+0) = 2000（Jan only）
	assertISAmount(t, is, rptAcctSalary, 2000)
	assertISAmount(t, is, rptAcctSalary4101, 2000)
	assertISAmount(t, is, rptAcctSalary410, 2000)
	// rent: (3000+0) - (1000+0) = 2000（Jan only，快照公式同樣消去 Dec）
	assertISAmount(t, is, rptAcctRent, 2000)
	assertISAmount(t, is, rptAcctRent5101, 2000)
}

// TestGetIncomeStatement_BothSnaps_WithDeltaPre
// 三個連續月結（Nov→Dec→Jan），查詢範圍 [2024-12-16, 2025-01-31]，
// delta_pre 非空（Dec 前半段 Dec10 交易），驗證兩快照公式正確。
func TestGetIncomeStatement_BothSnaps_ThreeCloses(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapRepo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 十一月月結 (closing_id=1, period_end=2024-11-30)—— snap_pre for [Dec16, Jan31]
	sumInsertPeriod(t, db, 1, "MONTHLY", "2024-11-01", "2024-11-30", true)
	// 無十一月交易，snap_pre.credit = 0
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert snap_pre: %v", err)
	}

	// 十二月月結 (closing_id=2, period_end=2024-12-31)
	sumInsertPeriod(t, db, 2, "MONTHLY", "2024-12-01", "2024-12-31", true)
	sumInsertTxn(t, db, 1, "2024-12-10", rptAcctRent, rptAcctSalary, 600) // Dec 前段（落入 delta_pre）
	sumInsertTxn(t, db, 2, "2024-12-20", rptAcctRent, rptAcctSalary, 400) // Dec 後段（含入 snap_end）
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 2); err != nil {
		t.Fatalf("BulkInsert dec: %v", err)
	}

	// 一月月結 (closing_id=3, period_end=2025-01-31) —— snap_end for [Dec16, Jan31]
	sumInsertPeriod(t, db, 3, "MONTHLY", "2025-01-01", "2025-01-31", true)
	sumInsertTxn(t, db, 3, "2025-01-15", rptAcctRent, rptAcctSalary, 2000) // Jan 薪資
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 3); err != nil {
		t.Fatalf("BulkInsert snap_end: %v", err)
	}

	// 查詢 [2024-12-16, 2025-01-31]
	// GetLatestClosedPeriodBetween("2024-12-16","2025-01-31") → closing_id=3, period_end=2025-01-31
	// GetLatestClosedPeriodBefore("2024-12-16") → closing_id=1, period_end=2024-11-30
	//   (Dec close 2024-12-31 >= 2024-12-16, so NOT < startDate)
	//
	// snap_end (closing_id=3): credit = snap_dec(0+600+400) + jan(2000) = 3000
	// snap_pre (closing_id=1): credit = 0
	// delta_end: txn_date > 2025-01-31 AND <= 2025-01-31 → 0
	// delta_pre: txn_date > 2024-11-30 AND < 2024-12-16 → Dec10 txn (600)
	// period_credit = (3000+0) - (0+600) = 2400 = Dec20(400) + Jan(2000) ✓
	is, err := repo.GetIncomeStatement(ctx, "2024-12-16", "2025-01-31")
	if err != nil {
		t.Fatalf("GetIncomeStatement: %v", err)
	}

	assertISAmount(t, is, rptAcctSalary, 2400)
	assertISAmount(t, is, rptAcctSalary4101, 2400)
	assertISAmount(t, is, rptAcctSalary410, 2400)
}

// TestGetBalanceSheet_HasChildAndDepth
// has_child 必須根據「實際有子科目指向自己」判斷，而非 is_summary 標記；
// 葉科目 has_child=false，父科目 has_child=true，且 depth 關係正確。
func TestGetBalanceSheet_HasChildAndDepth(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 1101-01 (ASSET 葉) 1000, CR 2101-01 (LIABILITY 葉) 1000
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctCC, 1000)

	bs, err := repo.GetBalanceSheet(ctx, "2025-01-31")
	if err != nil {
		t.Fatalf("GetBalanceSheet: %v", err)
	}

	// 葉科目 1101-01：has_child=false，depth > 0
	leaf, ok := findBSRowFull(bs, rptAcctCash)
	if !ok {
		t.Fatalf("account %s not found", rptAcctCash)
	}
	if leaf.HasChild {
		t.Errorf("account %s: HasChild got true, want false", rptAcctCash)
	}
	if leaf.Depth == 0 {
		t.Errorf("account %s: Depth got 0, expected > 0 for non-root leaf", rptAcctCash)
	}

	// 父科目 1101：has_child=true，depth 比子科目淺
	parent1101, ok := findBSRowFull(bs, rptAcctCash1101)
	if !ok {
		t.Fatalf("account %s not found", rptAcctCash1101)
	}
	if !parent1101.HasChild {
		t.Errorf("account %s: HasChild got false, want true", rptAcctCash1101)
	}
	if parent1101.Depth >= leaf.Depth {
		t.Errorf("account %s depth %d should be shallower than leaf %s depth %d",
			rptAcctCash1101, parent1101.Depth, rptAcctCash, leaf.Depth)
	}

	// 父科目 110：has_child=true（因為 1101 的 parent_id = 110）
	parent110, ok := findBSRowFull(bs, rptAcctCash110)
	if !ok {
		t.Fatalf("account %s not found", rptAcctCash110)
	}
	if !parent110.HasChild {
		t.Errorf("account %s: HasChild got false, want true", rptAcctCash110)
	}
	if parent110.Depth >= parent1101.Depth {
		t.Errorf("account %s depth %d should be shallower than %s depth %d",
			rptAcctCash110, parent110.Depth, rptAcctCash1101, parent1101.Depth)
	}
}

// TestGetIncomeStatement_HasChildAndDepth
// 損益表中葉科目 has_child=false，父科目 has_child=true，depth 層級關係正確。
func TestGetIncomeStatement_HasChildAndDepth(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 5101-01 (EXPENSE 葉) 3000, CR 4101-01 (INCOME 葉) 3000
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctRent, rptAcctSalary, 3000)

	is, err := repo.GetIncomeStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetIncomeStatement: %v", err)
	}

	// 葉科目 4101-01：has_child=false，depth > 0
	incLeaf, ok := findISRowFull(is, rptAcctSalary)
	if !ok {
		t.Fatalf("account %s not found", rptAcctSalary)
	}
	if incLeaf.HasChild {
		t.Errorf("account %s: HasChild got true, want false", rptAcctSalary)
	}
	if incLeaf.Depth == 0 {
		t.Errorf("account %s: Depth got 0, expected > 0", rptAcctSalary)
	}

	// 父科目 4101：has_child=true
	incParent, ok := findISRowFull(is, rptAcctSalary4101)
	if !ok {
		t.Fatalf("account %s not found", rptAcctSalary4101)
	}
	if !incParent.HasChild {
		t.Errorf("account %s: HasChild got false, want true", rptAcctSalary4101)
	}
	if incParent.Depth >= incLeaf.Depth {
		t.Errorf("account %s depth %d should be shallower than leaf %s depth %d",
			rptAcctSalary4101, incParent.Depth, rptAcctSalary, incLeaf.Depth)
	}

	// 葉科目 5101-01：has_child=false
	expLeaf, ok := findISRowFull(is, rptAcctRent)
	if !ok {
		t.Fatalf("account %s not found", rptAcctRent)
	}
	if expLeaf.HasChild {
		t.Errorf("account %s: HasChild got true, want false", rptAcctRent)
	}

	// 父科目 5101：has_child=true
	expParent, ok := findISRowFull(is, rptAcctRent5101)
	if !ok {
		t.Fatalf("account %s not found", rptAcctRent5101)
	}
	if !expParent.HasChild {
		t.Errorf("account %s: HasChild got false, want true", rptAcctRent5101)
	}
}

// ── 股東權益變動表 & 現金流量表 測試常數 ─────────────────────────────────────────

const (
	// Equity leaf accounts (from seeds)
	rptAcctEqLeaf1 = "3101-01" // 期初結餘 (EQUITY, CREDIT normal, leaf)
	rptAcctEqLeaf2 = "3103-02" // 個人增資 (EQUITY, CREDIT normal, leaf, FINANCING)
	rptAcctEqLeaf3 = "3103-01" // 個人提領 (EQUITY, DEBIT normal, leaf, FINANCING)
	rptAcctEqSub   = "3103"    // 提撥與提領 (EQUITY, summary, parent of 3103-01/3103-02)
	rptAcctEqTop   = "310"     // 個人淨資產 (EQUITY, top-level summary, parent_id=NULL)

	// Cash-flow category accounts (from seeds)
	rptAcctOperating = "1103-01" // 應收薪資款 (ASSET, DEBIT, OPERATING)
	rptAcctInvesting = "1102-01" // 投資 FVTPL (ASSET, DEBIT, INVESTING)
	// FINANCING: rptAcctEqLeaf2 (3103-02) 已在上方定義
)

// ── 輔助函式 ──────────────────────────────────────────────────────────────────

func findEqItem(stmt *report.EquityStatement, accountId string) (report.EquityItem, bool) {
	for _, item := range stmt.Items {
		if item.AccountId == accountId {
			return item, true
		}
	}
	return report.EquityItem{}, false
}

func assertEqItem(t *testing.T, stmt *report.EquityStatement, accountId string, wantBegin, wantChange, wantEnd float64) {
	t.Helper()
	item, ok := findEqItem(stmt, accountId)
	if !ok {
		t.Errorf("equity account %s not found in statement (%d items)", accountId, len(stmt.Items))
		return
	}
	gotBegin, _ := item.BeginBalance.Float64()
	gotChange, _ := item.PeriodChange.Float64()
	gotEnd, _ := item.EndBalance.Float64()
	if gotBegin != wantBegin {
		t.Errorf("equity %s begin_balance: got %.2f, want %.2f", accountId, gotBegin, wantBegin)
	}
	if gotChange != wantChange {
		t.Errorf("equity %s period_change: got %.2f, want %.2f", accountId, gotChange, wantChange)
	}
	if gotEnd != wantEnd {
		t.Errorf("equity %s end_balance: got %.2f, want %.2f", accountId, gotEnd, wantEnd)
	}
}

func assertCFItem(t *testing.T, items []report.CashFlowItem, accountId string, wantAmount float64) {
	t.Helper()
	for _, item := range items {
		if item.AccountId == accountId {
			got, _ := item.Amount.Float64()
			if got != wantAmount {
				t.Errorf("CF item %s amount: got %.2f, want %.2f", accountId, got, wantAmount)
			}
			return
		}
	}
	t.Errorf("CF item %s not found in section", accountId)
}

// ── GetEquityStatement 測試 ───────────────────────────────────────────────────

// TestGetEquityStatement_BeginAndPeriodDistinction
// start_date 前的分錄計入 begin_balance，期間內計入 period_change。
func TestGetEquityStatement_BeginAndPeriodDistinction(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 期初前：DR 1101-01, CR 3101-01 (equity) 5000
	sumInsertTxn(t, db, 1, "2025-01-31", rptAcctCash, rptAcctEqLeaf1, 5000)
	// 期間內：DR 1101-01, CR 3103-02 (equity) 2000
	sumInsertTxn(t, db, 2, "2025-02-10", rptAcctCash, rptAcctEqLeaf2, 2000)

	stmt, err := repo.GetEquityStatement(ctx, "2025-02-01", "2025-02-28")
	if err != nil {
		t.Fatalf("GetEquityStatement: %v", err)
	}

	// 3101-01: begin_credit=5000 → begin_balance=5000，期間無異動
	assertEqItem(t, stmt, rptAcctEqLeaf1, 5000, 0, 5000)
	// 3103-02: 無期初，period_credit=2000 → period_change=2000
	assertEqItem(t, stmt, rptAcctEqLeaf2, 0, 2000, 2000)
}

// TestGetEquityStatement_SummaryAggregates
// 彙總科目 3103 的 period_change 應等於子葉科目 3103-01 + 3103-02 的加總。
func TestGetEquityStatement_SummaryAggregates(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 3103-01 (EQUITY DEBIT, 提領) 1000, CR 1101-01 → period_change = 0 - 1000 = -1000
	sumInsertTxn(t, db, 1, "2025-01-10", rptAcctEqLeaf3, rptAcctCash, 1000)
	// DR 1101-01, CR 3103-02 (EQUITY CREDIT, 增資) 3000 → period_change = 3000 - 0 = 3000
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctCash, rptAcctEqLeaf2, 3000)

	stmt, err := repo.GetEquityStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetEquityStatement: %v", err)
	}

	assertEqItem(t, stmt, rptAcctEqLeaf3, 0, -1000, -1000)
	assertEqItem(t, stmt, rptAcctEqLeaf2, 0, 3000, 3000)
	// 父科目 3103 = -1000 + 3000 = 2000
	assertEqItem(t, stmt, rptAcctEqSub, 0, 2000, 2000)
}

// TestGetEquityStatement_NetIncomeVirtualRow
// 有收入 / 費用時，虛擬本期淨利行必須出現且金額正確。
func TestGetEquityStatement_NetIncomeVirtualRow(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 收入 4000：DR 1101-01, CR 4101-01
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctSalary, 4000)
	// 費用 1500：DR 5101-01, CR 1101-01
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctRent, rptAcctCash, 1500)

	stmt, err := repo.GetEquityStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetEquityStatement: %v", err)
	}

	netIncome, _ := stmt.NetIncome.Float64()
	if netIncome != 2500 {
		t.Errorf("NetIncome: got %.2f, want 2500", netIncome)
	}

	// 找虛擬行
	found := false
	for _, item := range stmt.Items {
		if item.IsVirtual {
			found = true
			period, _ := item.PeriodChange.Float64()
			if period != 2500 {
				t.Errorf("virtual row period_change: got %.2f, want 2500", period)
			}
			endBal, _ := item.EndBalance.Float64()
			if endBal != 2500 {
				t.Errorf("virtual row end_balance: got %.2f, want 2500", endBal)
			}
			break
		}
	}
	if !found {
		t.Error("virtual net income item not found in equity statement items")
	}
}

// TestGetEquityStatement_TotalIncludesNetIncome
// total_begin 僅含頂層科目 (310)，total_period 含權益變動 + 本期淨利，
// total_end = total_begin + total_period。
func TestGetEquityStatement_TotalIncludesNetIncome(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 期初：DR 1101-01, CR 3101-01 3000 → begin_balance 3000
	sumInsertTxn(t, db, 1, "2024-12-31", rptAcctCash, rptAcctEqLeaf1, 3000)
	// 期間權益增加：DR 1101-01, CR 3103-02 1000
	sumInsertTxn(t, db, 2, "2025-01-15", rptAcctCash, rptAcctEqLeaf2, 1000)
	// 期間收入：DR 1101-01, CR 4101-01 500 → net_income = 500
	sumInsertTxn(t, db, 3, "2025-01-20", rptAcctCash, rptAcctSalary, 500)

	stmt, err := repo.GetEquityStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetEquityStatement: %v", err)
	}

	// 頂層科目 310 彙總 begin = 3000，period_change = 1000
	assertEqItem(t, stmt, rptAcctEqTop, 3000, 1000, 4000)

	totalBegin, _ := stmt.TotalBeginBalance.Float64()
	totalChange, _ := stmt.TotalPeriodChange.Float64()
	totalEnd, _ := stmt.TotalEndBalance.Float64()

	if totalBegin != 3000 {
		t.Errorf("TotalBeginBalance: got %.2f, want 3000", totalBegin)
	}
	// total_period = 1000 (equity) + 500 (net_income)
	if totalChange != 1500 {
		t.Errorf("TotalPeriodChange: got %.2f, want 1500 (equity 1000 + net_income 500)", totalChange)
	}
	if totalEnd != 4500 {
		t.Errorf("TotalEndBalance: got %.2f, want 4500", totalEnd)
	}
}

// ── GetCashFlowStatement 測試 ─────────────────────────────────────────────────

// TestGetCashFlowStatement_BeginningAndEndingCash
// CASH 類別科目的期初/期末餘額計算正確。
func TestGetCashFlowStatement_BeginningAndEndingCash(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 期初前：DR 1101-01 (CASH) 5000，CR 3101-01 (equity)
	sumInsertTxn(t, db, 1, "2024-12-31", rptAcctCash, rptAcctEqLeaf1, 5000)
	// 期間內：DR 1101-01 2000，CR 3103-02
	sumInsertTxn(t, db, 2, "2025-01-15", rptAcctCash, rptAcctEqLeaf2, 2000)
	// 期間內：DR 5101-01, CR 1101-01 500（現金流出）
	sumInsertTxn(t, db, 3, "2025-01-20", rptAcctRent, rptAcctCash, 500)

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// beginning_cash = CASH 科目累計 debit - credit < 2025-01-01 = 5000 - 0 = 5000
	begin, _ := cf.BeginningCash.Float64()
	if begin != 5000 {
		t.Errorf("BeginningCash: got %.2f, want 5000", begin)
	}
	// ending_cash = CASH 科目累計 debit - credit <= 2025-01-31 = 7000 - 500 = 6500
	end, _ := cf.EndingCash.Float64()
	if end != 6500 {
		t.Errorf("EndingCash: got %.2f, want 6500", end)
	}
}

// cfPtr converts a string literal to a pointer, used for sumInsertTxnWithCF CF params.
func cfPtr(s string) *string { return &s }

// TestGetCashFlowStatement_ThreeSections
// 三大活動分類各自出現在對應 section，金額符合 credit−debit 公式。
// 分類由 entry_cf_categories.cf_category 決定。
func TestGetCashFlowStatement_ThreeSections(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// OPERATING: DR 1101-01 (CASH), CR 1103-01 800 → 1103-01 entry 標記 OPERATING
	sumInsertTxnWithCF(t, db, 1, "2025-01-10", rptAcctCash, rptAcctOperating, 800, nil, cfPtr("OPERATING"))
	// INVESTING: DR 1102-01 2000 → 1102-01 entry 標記 INVESTING，CR 1101-01 (CASH)
	sumInsertTxnWithCF(t, db, 2, "2025-01-15", rptAcctInvesting, rptAcctCash, 2000, cfPtr("INVESTING"), nil)
	// FINANCING: DR 1101-01 (CASH), CR 3103-02 3000 → 3103-02 entry 標記 FINANCING
	sumInsertTxnWithCF(t, db, 3, "2025-01-20", rptAcctCash, rptAcctEqLeaf2, 3000, nil, cfPtr("FINANCING"))

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// 1103-01 credit 800 → amount = 800 - 0 = 800
	assertCFItem(t, cf.OperatingActivities.Adjustments, rptAcctOperating, 800)
	// 1102-01 debit 2000 → amount = 0 - 2000 = -2000
	assertCFItem(t, cf.InvestingActivities.Items, rptAcctInvesting, -2000)
	// 3103-02 credit 3000 → amount = 3000 - 0 = 3000
	assertCFItem(t, cf.FinancingActivities.Items, rptAcctEqLeaf2, 3000)
}

// TestGetCashFlowStatement_NetIncomeInOperating
// 本期淨利計入營業活動小計。
func TestGetCashFlowStatement_NetIncomeInOperating(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 收入 4000
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctSalary, 4000)
	// 費用 1500
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctRent, rptAcctCash, 1500)

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	netIncome, _ := cf.OperatingActivities.NetIncome.Float64()
	if netIncome != 2500 {
		t.Errorf("OperatingActivities.NetIncome: got %.2f, want 2500", netIncome)
	}
	// 無其他 OPERATING 調整項，總計 = 淨利
	opTotal, _ := cf.OperatingActivities.Total.Float64()
	if opTotal != 2500 {
		t.Errorf("OperatingActivities.Total: got %.2f, want 2500", opTotal)
	}
}

// TestGetCashFlowStatement_WithSnapshot
// 月結快照存在時，全量掃描查詢仍能正確按日期邊界篩選：
// 期初現金包含快照前交易，調整項僅計入查詢期間內的交易。
func TestGetCashFlowStatement_WithSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapRepo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// M-1（一月）：收入薪資 3000，現金流入 3000
	sumInsertPeriod(t, db, 1, "MONTHLY", "2025-01-01", "2025-01-31", true)
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctSalary, 3000) // CR income, DR CASH（無 CF 分類）
	// 有 OPERATING 帳：應收薪資 800（M-1 期間），1103-01 entry 標記 OPERATING
	sumInsertTxnWithCF(t, db, 2, "2025-01-20", rptAcctCash, rptAcctOperating, 800, nil, cfPtr("OPERATING"))
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	// M（二月，查詢期間）：OPERATING 應收帳 500，INVESTING 買入 1000
	sumInsertTxnWithCF(t, db, 3, "2025-02-10", rptAcctCash, rptAcctOperating, 500, nil, cfPtr("OPERATING"))   // CR 1103-01（應收收回，流入）
	sumInsertTxnWithCF(t, db, 4, "2025-02-20", rptAcctInvesting, rptAcctCash, 1000, cfPtr("INVESTING"), nil) // DR 1102-01（買入投資，流出）

	cf, err := repo.GetCashFlowStatement(ctx, "2025-02-01", "2025-02-28")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// beginning_cash = CASH 累計 debit - credit < 2025-02-01 = (3000+800) - 0 = 3800
	begin, _ := cf.BeginningCash.Float64()
	if begin != 3800 {
		t.Errorf("BeginningCash: got %.2f, want 3800", begin)
	}

	// ending_cash = CASH 累計 debit - credit <= 2025-02-28 = (3000+800+500) - 1000 = 3300
	end, _ := cf.EndingCash.Float64()
	if end != 3300 {
		t.Errorf("EndingCash: got %.2f, want 3300", end)
	}

	// OPERATING 調整項只計二月的 500，不計一月的 800
	assertCFItem(t, cf.OperatingActivities.Adjustments, rptAcctOperating, 500)

	// INVESTING 只計二月的 -1000
	assertCFItem(t, cf.InvestingActivities.Items, rptAcctInvesting, -1000)
}

// TestGetCashFlowStatement_NetChange
// NetChange = 營業活動 + 投資活動 + 籌資活動。
func TestGetCashFlowStatement_NetChange(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 收入 3000 → operating net_income = 3000（無 CF 分類，透過淨利計算）
	sumInsertTxn(t, db, 1, "2025-01-05", rptAcctCash, rptAcctSalary, 3000)
	// 買入投資 1000 → investing = -1000（1102-01 entry 標記 INVESTING）
	sumInsertTxnWithCF(t, db, 2, "2025-01-10", rptAcctInvesting, rptAcctCash, 1000, cfPtr("INVESTING"), nil)
	// 增資 500 → financing = +500（3103-02 entry 標記 FINANCING）
	sumInsertTxnWithCF(t, db, 3, "2025-01-15", rptAcctCash, rptAcctEqLeaf2, 500, nil, cfPtr("FINANCING"))

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// net_change = 3000 (op) + (-1000) (inv) + 500 (fin) = 2500
	netChange, _ := cf.NetChange.Float64()
	if netChange != 2500 {
		t.Errorf("NetChange: got %.2f, want 2500 (op 3000 + inv -1000 + fin 500)", netChange)
	}
}

// TestGetEquityStatement_WithSnapshot
// 月結快照存在時，全量掃描查詢仍能正確分離期初餘額與期間變動。
func TestGetEquityStatement_WithSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapRepo := projection_repo.NewAccountBalanceSnapshotRepo(sqlcdb.New(db))
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// M-1（一月）：期初權益 5000
	sumInsertPeriod(t, db, 1, "MONTHLY", "2025-01-01", "2025-01-31", true)
	sumInsertTxn(t, db, 1, "2025-01-20", rptAcctCash, rptAcctEqLeaf1, 5000) // CR 3101-01
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	// M（二月，查詢期間）：增資 2000、提領 800
	sumInsertTxn(t, db, 2, "2025-02-10", rptAcctCash, rptAcctEqLeaf2, 2000) // CR 3103-02（增資）
	sumInsertTxn(t, db, 3, "2025-02-15", rptAcctEqLeaf3, rptAcctCash, 800)  // DR 3103-01（提領）

	stmt, err := repo.GetEquityStatement(ctx, "2025-02-01", "2025-02-28")
	if err != nil {
		t.Fatalf("GetEquityStatement: %v", err)
	}

	// 3101-01：begin_credit=5000（一月），期間無異動
	assertEqItem(t, stmt, rptAcctEqLeaf1, 5000, 0, 5000)
	// 3103-02：無期初，period_credit=2000
	assertEqItem(t, stmt, rptAcctEqLeaf2, 0, 2000, 2000)
	// 3103-01（DEBIT normal）：無期初，period_debit=800 → period_change = credit-debit = -800
	assertEqItem(t, stmt, rptAcctEqLeaf3, 0, -800, -800)
	// 彙總 3103 = 0 begin，2000-800=1200 period
	assertEqItem(t, stmt, rptAcctEqSub, 0, 1200, 1200)

	// total_begin 只取頂層 310 = 5000（3101-01 屬 310 子孫）
	totalBegin, _ := stmt.TotalBeginBalance.Float64()
	if totalBegin != 5000 {
		t.Errorf("TotalBeginBalance: got %.2f, want 5000", totalBegin)
	}
}

// ── GetCashFlowStatement 科目階層彙總測試 ────────────────────────────────────────

// rptAcctOperatingSummary / rptAcctOperatingLeafA / rptAcctOperatingLeafB
// 用於階層彙總測試的科目：1103 彙總、1103-01、1103-02
const (
	rptAcctCFSummary = "1103"    // 應收款項 (ASSET, summary, currently NULL category)
	rptAcctCFLeafA   = "1103-01" // 應收薪資款 (ASSET, DEBIT, OPERATING) — 測試時改為 NULL
	rptAcctCFLeafB   = "1103-02" // 應收租金 (ASSET, DEBIT, OPERATING)   — 測試時改為 NULL
)

// TestGetCashFlowStatement_HierarchyAggregation
// 葉節點 entry 標記 OPERATING，父科目應以彙總行（is_summary=true）出現，
// 金額為所有後裔 OPERATING entry 的加總；葉節點行也同步出現。
func TestGetCashFlowStatement_HierarchyAggregation(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 期間交易：1103-01 credit 800、1103-02 credit 500，entry 均標記 OPERATING
	sumInsertTxnWithCF(t, db, 1, "2025-01-10", rptAcctCash, rptAcctCFLeafA, 800, nil, cfPtr("OPERATING"))
	sumInsertTxnWithCF(t, db, 2, "2025-01-20", rptAcctCash, rptAcctCFLeafB, 500, nil, cfPtr("OPERATING"))

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// 彙總行 1103 應出現在 OPERATING，金額 = (800+500) credit - 0 debit = 1300
	var summaryFound *report.CashFlowItem
	for i, item := range cf.OperatingActivities.Adjustments {
		if item.AccountId == rptAcctCFSummary {
			summaryFound = &cf.OperatingActivities.Adjustments[i]
			break
		}
	}
	if summaryFound == nil {
		t.Fatalf("summary account %s not found in OPERATING adjustments", rptAcctCFSummary)
	}
	if !summaryFound.IsSummary {
		t.Errorf("account %s: IsSummary got false, want true", rptAcctCFSummary)
	}
	got, _ := summaryFound.Amount.Float64()
	if got != 1300 {
		t.Errorf("account %s amount: got %.2f, want 1300", rptAcctCFSummary, got)
	}

	// 葉節點也應出現（entry 帶有 OPERATING），is_summary=false
	for _, leafID := range []string{rptAcctCFLeafA, rptAcctCFLeafB} {
		var leafFound *report.CashFlowItem
		for i, item := range cf.OperatingActivities.Adjustments {
			if item.AccountId == leafID {
				leafFound = &cf.OperatingActivities.Adjustments[i]
				break
			}
		}
		if leafFound == nil {
			t.Errorf("leaf account %s not found in OPERATING adjustments", leafID)
			continue
		}
		if leafFound.IsSummary {
			t.Errorf("leaf account %s: IsSummary got true, want false", leafID)
		}
	}
}

// TestGetCashFlowStatement_HierarchyAggregation_LeafIsSummaryFalse
// 葉節點 entry 標記 OPERATING，該葉科目行的 is_summary 應為 false。
func TestGetCashFlowStatement_HierarchyAggregation_LeafIsSummaryFalse(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 1103-01 entry 標記 OPERATING
	sumInsertTxnWithCF(t, db, 1, "2025-01-10", rptAcctCash, rptAcctCFLeafA, 600, nil, cfPtr("OPERATING"))

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	var found *report.CashFlowItem
	for i, item := range cf.OperatingActivities.Adjustments {
		if item.AccountId == rptAcctCFLeafA {
			found = &cf.OperatingActivities.Adjustments[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("leaf account %s not found in OPERATING adjustments", rptAcctCFLeafA)
	}
	if found.IsSummary {
		t.Errorf("account %s: IsSummary got true, want false for leaf", rptAcctCFLeafA)
	}
}

// ── GetDirectCashFlowStatement 測試 ──────────────────────────────────────────

// TestGetDirectCashFlowStatement_OperatingReceiptsAndPayments
// 有收入與費用的現金交易時，CashReceived / CashPaid / Total 計算正確。
func TestGetDirectCashFlowStatement_OperatingReceiptsAndPayments(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 收入：DR 1101-01 (CASH) 3000, CR 4101-01 (INCOME) → operating（有 INCOME 科目）
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctSalary, 3000)
	// 費用：DR 5101-01 (EXPENSE), CR 1101-01 (CASH) 1500 → operating（有 EXPENSE 科目）
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctRent, rptAcctCash, 1500)

	dcf, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	received, _ := dcf.OperatingActivities.CashReceived.Float64()
	paid, _ := dcf.OperatingActivities.CashPaid.Float64()
	total, _ := dcf.OperatingActivities.Total.Float64()

	if received != 3000 {
		t.Errorf("CashReceived: got %.2f, want 3000", received)
	}
	if paid != 1500 {
		t.Errorf("CashPaid: got %.2f, want 1500", paid)
	}
	if total != 1500 {
		t.Errorf("Operating Total: got %.2f, want 1500 (3000-1500)", total)
	}
}

// TestGetDirectCashFlowStatement_MatchesIndirectOperatingTotal
// 直接法的 OperatingActivities.Total 必須等於間接法的 OperatingActivities.Total。
func TestGetDirectCashFlowStatement_MatchesIndirectOperatingTotal(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 現金收入：DR CASH 3000, CR INCOME（direct CashReceived += 3000；indirect NetIncome += 3000）
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctSalary, 3000)
	// 現金費用：DR EXPENSE, CR CASH 1000（direct CashPaid += 1000；indirect NetIncome -= 1000）
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctRent, rptAcctCash, 1000)
	// OPERATING 調整：DR CASH 200, CR 1103-01 entry OPERATING
	// （direct CashReceived += 200；indirect OPERATING Adjustments += 200）
	sumInsertTxnWithCF(t, db, 3, "2025-01-25", rptAcctCash, rptAcctOperating, 200, nil, cfPtr("OPERATING"))

	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}
	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	if !direct.OperatingActivities.Total.Equal(indirect.OperatingActivities.Total) {
		t.Errorf("Operating Total mismatch: direct=%.2f, indirect=%.2f",
			direct.OperatingActivities.Total.InexactFloat64(),
			indirect.OperatingActivities.Total.InexactFloat64())
	}
}

// TestGetDirectCashFlowStatement_InvestingMatchesIndirect
// 投資活動的 Items 與 Total 與間接法相同。
func TestGetDirectCashFlowStatement_InvestingMatchesIndirect(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 買入投資：DR 1102-01 (entry INVESTING), CR 1101-01 (CASH) 1500
	sumInsertTxnWithCF(t, db, 1, "2025-01-15", rptAcctInvesting, rptAcctCash, 1500, cfPtr("INVESTING"), nil)

	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}
	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	if !direct.InvestingActivities.Total.Equal(indirect.InvestingActivities.Total) {
		t.Errorf("Investing Total mismatch: direct=%.2f, indirect=%.2f",
			direct.InvestingActivities.Total.InexactFloat64(),
			indirect.InvestingActivities.Total.InexactFloat64())
	}
	if len(direct.InvestingActivities.Items) != len(indirect.InvestingActivities.Items) {
		t.Errorf("Investing Items count: direct=%d, indirect=%d",
			len(direct.InvestingActivities.Items), len(indirect.InvestingActivities.Items))
	}
	// 1102-01 debit 1500 → amount = 0 - 1500 = -1500
	assertCFItem(t, direct.InvestingActivities.Items, rptAcctInvesting, -1500)
}

// TestGetDirectCashFlowStatement_FinancingMatchesIndirect
// 籌資活動的 Items 與 Total 與間接法相同。
func TestGetDirectCashFlowStatement_FinancingMatchesIndirect(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 增資：DR 1101-01 (CASH), CR 3103-02 (entry FINANCING) 2000
	sumInsertTxnWithCF(t, db, 1, "2025-01-15", rptAcctCash, rptAcctEqLeaf2, 2000, nil, cfPtr("FINANCING"))

	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}
	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	if !direct.FinancingActivities.Total.Equal(indirect.FinancingActivities.Total) {
		t.Errorf("Financing Total mismatch: direct=%.2f, indirect=%.2f",
			direct.FinancingActivities.Total.InexactFloat64(),
			indirect.FinancingActivities.Total.InexactFloat64())
	}
	// 3103-02 credit 2000 → amount = 2000 - 0 = 2000
	assertCFItem(t, direct.FinancingActivities.Items, rptAcctEqLeaf2, 2000)
}

// TestGetDirectCashFlowStatement_NetChange
// NetChange = 三大活動之和；EndingCash - BeginningCash = NetChange。
func TestGetDirectCashFlowStatement_NetChange(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 期初前：DR CASH 1000, CR 3101-01（令 BeginningCash = 1000）
	sumInsertTxn(t, db, 1, "2024-12-31", rptAcctCash, rptAcctEqLeaf1, 1000)
	// 收入 3000 → operating CashReceived = 3000；OperatingTotal = 3000
	sumInsertTxn(t, db, 2, "2025-01-05", rptAcctCash, rptAcctSalary, 3000)
	// 買入投資 1000 → InvestingTotal = -1000
	sumInsertTxnWithCF(t, db, 3, "2025-01-10", rptAcctInvesting, rptAcctCash, 1000, cfPtr("INVESTING"), nil)
	// 增資 500 → FinancingTotal = +500
	sumInsertTxnWithCF(t, db, 4, "2025-01-15", rptAcctCash, rptAcctEqLeaf2, 500, nil, cfPtr("FINANCING"))

	dcf, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// NetChange = 3000 (op) + (-1000) (inv) + 500 (fin) = 2500
	netChange, _ := dcf.NetChange.Float64()
	if netChange != 2500 {
		t.Errorf("NetChange: got %.2f, want 2500 (op 3000 + inv -1000 + fin 500)", netChange)
	}

	// EndingCash(3500) - BeginningCash(1000) = NetChange(2500)
	cashDiff := dcf.EndingCash.Sub(dcf.BeginningCash)
	if !cashDiff.Equal(dcf.NetChange) {
		t.Errorf("EndingCash - BeginningCash = %.2f, NetChange = %.2f; should be equal",
			cashDiff.InexactFloat64(), dcf.NetChange.InexactFloat64())
	}
}

// TestGetDirectCashFlowStatement_WithOperatingTaggedEntry
// 僅含 OPERATING-tagged 分錄（無 INCOME/EXPENSE 科目）時，
// 現金科目的借貸正確納入 CashReceived / CashPaid。
func TestGetDirectCashFlowStatement_WithOperatingTaggedEntry(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 1101-01 (CASH) 400, CR 1103-01（entry 標記 OPERATING）；無 INCOME/EXPENSE 科目
	sumInsertTxnWithCF(t, db, 1, "2025-01-10", rptAcctCash, rptAcctOperating, 400, nil, cfPtr("OPERATING"))

	dcf, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	received, _ := dcf.OperatingActivities.CashReceived.Float64()
	paid, _ := dcf.OperatingActivities.CashPaid.Float64()
	total, _ := dcf.OperatingActivities.Total.Float64()

	if received != 400 {
		t.Errorf("CashReceived: got %.2f, want 400", received)
	}
	if paid != 0 {
		t.Errorf("CashPaid: got %.2f, want 0", paid)
	}
	if total != 400 {
		t.Errorf("Operating Total: got %.2f, want 400", total)
	}
}

// ── 投資 & 分期 CF 測試用科目常數（來自 accounts.sql seeds）─────────────────────

const (
	rptAcctInvestFVTPL = "1102-01" // 透過損益按公允價值衡量之投資 (ASSET, DEBIT, INVESTING)
	rptAcctStockFee    = "5402-01" // 股票交易手續費 (EXPENSE, DEBIT)
	rptAcctStockTax    = "5402-05" // 股票交易證交稅 (EXPENSE, DEBIT)
	rptAcctRealGain    = "4203-01" // 股票處分利得 (INCOME, CREDIT)
	rptAcctUnrealGain  = "4204-01" // 股票未實現評價利益 (INCOME, CREDIT)
	rptAcctLoan        = "2201-01" // 房屋貸款 (LIABILITY, CREDIT, account FINANCING)
)

// cfEntry 用於 sumInsertTxnMulti 的單筆分錄描述。
type cfEntry struct {
	accountID string
	debit     float64
	credit    float64
	cf        *string
}

// sumInsertTxnMulti 插入含任意數量分錄的交易，供需要 3+ 分錄的測試情境使用。
// Also populates entry_cf_categories for entries with non-nil cf so report queries can JOIN it.
func sumInsertTxnMulti(t *testing.T, db *sqlx.DB, txnID int64, date string, totalAmount float64, entries []cfEntry) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO transactions (txn_id, merchant_id, txn_date, description, total_amount, version) VALUES (?, ?, ?, '測試', ?, 1)`,
		txnID, testMerchantID, date, totalAmount)
	if err != nil {
		t.Fatalf("sumInsertTxnMulti id=%d: %v", txnID, err)
	}
	for i, e := range entries {
		entryUUID := fmt.Sprintf("test-entry-%d-%d", txnID, i)
		_, err = db.ExecContext(ctx,
			`INSERT INTO journal_entries (txn_id, merchant_id, account_id, entry_uuid, debit, credit) VALUES (?, ?, ?, ?, ?, ?)`,
			txnID, testMerchantID, e.accountID, entryUUID, e.debit, e.credit)
		if err != nil {
			t.Fatalf("sumInsertTxnMulti entry %s: %v", e.accountID, err)
		}
		if e.cf != nil {
			_, err = db.ExecContext(ctx,
				`INSERT INTO entry_cf_categories (entry_uuid, merchant_id, cf_category, is_confirmed) VALUES (?, ?, ?, 1)`,
				entryUUID, testMerchantID, *e.cf)
			if err != nil {
				t.Fatalf("sumInsertTxnMulti entry_cf_category %s: %v", e.accountID, err)
			}
		}
	}
}

// ── 投資買入 CF 測試 ──────────────────────────────────────────────────────────

// TestGetCashFlowStatement_InvestmentBuy_InvestingOnly
// 買入投資（無費用）：全額歸投資活動，NI=0，間接法營業=0，NetChange=-price。
func TestGetCashFlowStatement_InvestmentBuy_InvestingOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 1102-01 [INVESTING] 10000, CR 1101-01 (Cash) 10000
	sumInsertTxnWithCF(t, db, 1, "2025-01-15", rptAcctInvestFVTPL, rptAcctCash, 10000, cfPtr("INVESTING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// NI = 0（無損益科目）；買入投資費用已從 NI 排除
	niVal, _ := indirect.OperatingActivities.NetIncome.Float64()
	if niVal != 0 {
		t.Errorf("OperatingActivities.NetIncome: got %.2f, want 0", niVal)
	}
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("OperatingActivities.Total: got %.2f, want 0", opTotal)
	}

	// 投資活動：1102-01 amount = credit-debit = 0-10000 = -10000
	assertCFItem(t, indirect.InvestingActivities.Items, rptAcctInvestFVTPL, -10000)
	invTotal, _ := indirect.InvestingActivities.Total.Float64()
	if invTotal != -10000 {
		t.Errorf("InvestingActivities.Total: got %.2f, want -10000", invTotal)
	}

	// NetChange = -10000（現金淨減少）
	netChange, _ := indirect.NetChange.Float64()
	if netChange != -10000 {
		t.Errorf("NetChange: got %.2f, want -10000", netChange)
	}
}

// TestGetCashFlowStatement_InvestmentBuyWithFee_NIReclassification
// 買入含手續費：費用科目標記 INVESTING，NI 中的費用必須從營業 NI 移除（重分類）。
func TestGetCashFlowStatement_InvestmentBuyWithFee_NIReclassification(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 買入 10000：DR 1102-01 [INVESTING], CR 1101-01
	sumInsertTxnWithCF(t, db, 1, "2025-01-10", rptAcctInvestFVTPL, rptAcctCash, 10000, cfPtr("INVESTING"), nil)
	// 手續費 30：DR 5402-01 [INVESTING], CR 1101-01
	sumInsertTxnWithCF(t, db, 2, "2025-01-10", rptAcctStockFee, rptAcctCash, 30, cfPtr("INVESTING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// IS: NetIncome = -30（費用）；但費用標 INVESTING → 從 NI 移除
	// OperatingActivities.NetIncome = -30 - (-30) = 0
	opNI, _ := indirect.OperatingActivities.NetIncome.Float64()
	if opNI != 0 {
		t.Errorf("OperatingActivities.NetIncome after reclassification: got %.2f, want 0", opNI)
	}
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("OperatingActivities.Total: got %.2f, want 0", opTotal)
	}

	// InvestingTotal = -10000 (資產) + -30 (手續費) = -10030
	invTotal, _ := indirect.InvestingActivities.Total.Float64()
	if invTotal != -10030 {
		t.Errorf("InvestingActivities.Total: got %.2f, want -10030", invTotal)
	}

	netChange, _ := indirect.NetChange.Float64()
	if netChange != -10030 {
		t.Errorf("NetChange: got %.2f, want -10030", netChange)
	}
}

// TestGetCashFlowStatement_InvestmentSell_DirectIndirectConsistency
// 出售含手續費/交易稅：已實現損益、費用均標 INVESTING，
// 間接法 OperatingNI=0，直接法 OperatingTotal=0，兩者一致。
func TestGetCashFlowStatement_InvestmentSell_DirectIndirectConsistency(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 先買入（成本 10000）
	sumInsertTxnWithCF(t, db, 1, "2025-01-05", rptAcctInvestFVTPL, rptAcctCash, 10000, cfPtr("INVESTING"), nil)

	// 出售：netProceeds=11920, cost=10000, gain=2000, fee=50, tax=30
	// DR Cash 11920 [NULL], CR Investment 10000 [INVESTING], CR Gain 2000 [INVESTING],
	// DR Fee 50 [INVESTING], DR Tax 30 [INVESTING]
	// 借貸驗算：DR=11920+50+30=12000, CR=10000+2000=12000 ✓
	sumInsertTxnMulti(t, db, 2, "2025-01-20", 11920, []cfEntry{
		{accountID: rptAcctCash, debit: 11920, credit: 0, cf: nil},
		{accountID: rptAcctInvestFVTPL, debit: 0, credit: 10000, cf: cfPtr("INVESTING")},
		{accountID: rptAcctRealGain, debit: 0, credit: 2000, cf: cfPtr("INVESTING")},
		{accountID: rptAcctStockFee, debit: 50, credit: 0, cf: cfPtr("INVESTING")},
		{accountID: rptAcctStockTax, debit: 30, credit: 0, cf: cfPtr("INVESTING")},
	})

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：NI=2000-80=1920，niReclassification=2000-50-30=1920 → OperatingNI=0
	opNI, _ := indirect.OperatingActivities.NetIncome.Float64()
	if opNI != 0 {
		t.Errorf("indirect OperatingActivities.NetIncome: got %.2f, want 0", opNI)
	}
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingActivities.Total: got %.2f, want 0", opTotal)
	}

	// 直接法：出售交易無 INCOME/EXPENSE 分錄未標 INVESTING → 不進 operating_txns
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingActivities.Total: got %.2f, want 0 (sell must not be operating)", directOpTotal)
	}

	// InvestingTotal 兩者一致：buy -10000 + sell(10000+2000-50-30) = -10000 + 11920 = 1920
	if !indirect.InvestingActivities.Total.Equal(direct.InvestingActivities.Total) {
		t.Errorf("InvestingActivities.Total mismatch: indirect=%.2f, direct=%.2f",
			indirect.InvestingActivities.Total.InexactFloat64(),
			direct.InvestingActivities.Total.InexactFloat64())
	}

	// 整體 NetChange 一致
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// TestGetCashFlowStatement_FVTPLMark_NonCashZeroNetChange
// FVTPL 公允價值調整（非現金）：投資資產標 OPERATING，沖銷 NI 影響，
// 間接法 OperatingTotal=0，NetChange=0。
func TestGetCashFlowStatement_FVTPLMark_NonCashZeroNetChange(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 1102-01 [OPERATING] 500（非現金：投資帳面增加）, CR 4204-01 [NULL] 500
	sumInsertTxnWithCF(t, db, 1, "2025-01-15", rptAcctInvestFVTPL, rptAcctUnrealGain, 500, cfPtr("OPERATING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：NI=500（未實現利益入 NI）；OPERATING 調整項=-500（沖銷非現金）→ Total=0
	opNI, _ := indirect.OperatingActivities.NetIncome.Float64()
	if opNI != 500 {
		t.Errorf("indirect OperatingActivities.NetIncome: got %.2f, want 500", opNI)
	}
	assertCFItem(t, indirect.OperatingActivities.Adjustments, rptAcctInvestFVTPL, -500)
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingActivities.Total: got %.2f, want 0 (non-cash)", opTotal)
	}

	// NetChange = 0（無現金移動）
	netChange, _ := indirect.NetChange.Float64()
	if netChange != 0 {
		t.Errorf("indirect NetChange: got %.2f, want 0", netChange)
	}

	// 直接法：4204-01 是 INCOME/NULL → 該交易進 operating_txns；
	// 但兩個分錄帳戶都非 CASH（1102-01 INVESTING, 4204-01 INCOME）→ CashReceived=0, CashPaid=0
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingActivities.Total: got %.2f, want 0 (no cash in FVTPL mark)", directOpTotal)
	}
	directNetChange, _ := direct.NetChange.Float64()
	if directNetChange != 0 {
		t.Errorf("direct NetChange: got %.2f, want 0", directNetChange)
	}
}

// ── 分期還款 CF 測試 ──────────────────────────────────────────────────────────

// TestGetCashFlowStatement_InstallmentFree_FinancingTag
// 免息分期每期還款：DR 貸款負債 [FINANCING]，CR 銀行；
// 融資活動現金流出，間接法與直接法一致。
func TestGetCashFlowStatement_InstallmentFree_FinancingTag(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 2201-01 (房屋貸款負債) [FINANCING] 1000, CR 1101-01 (Cash)
	sumInsertTxnWithCF(t, db, 1, "2025-01-15", rptAcctLoan, rptAcctCash, 1000, cfPtr("FINANCING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：NI=0，FinancingTotal=-1000（負債減少 credit-debit=0-1000=-1000）
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingActivities.Total: got %.2f, want 0", opTotal)
	}
	assertCFItem(t, indirect.FinancingActivities.Items, rptAcctLoan, -1000)
	finTotal, _ := indirect.FinancingActivities.Total.Float64()
	if finTotal != -1000 {
		t.Errorf("indirect FinancingActivities.Total: got %.2f, want -1000", finTotal)
	}

	// 直接法：還款交易無 INCOME/EXPENSE 分錄，不進 operating_txns → Operating=0
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingActivities.Total: got %.2f, want 0", directOpTotal)
	}

	// 兩者 FinancingTotal 與 NetChange 一致
	if !indirect.FinancingActivities.Total.Equal(direct.FinancingActivities.Total) {
		t.Errorf("FinancingActivities.Total mismatch: indirect=%.2f, direct=%.2f",
			indirect.FinancingActivities.Total.InexactFloat64(),
			direct.FinancingActivities.Total.InexactFloat64())
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// TestGetCashFlowStatement_InvestmentAndIncome_DirectIndirectConsistency
// 同期含一般薪資收入 + 投資買入（含手續費）：
// 直接法僅薪資交易進 operating_txns，買入與手續費不進；
// 間接法透過 NI 重分類排除手續費；兩者 OperatingTotal 相同。
func TestGetCashFlowStatement_InvestmentAndIncome_DirectIndirectConsistency(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 薪資收入：DR 1101-01 3000, CR 4101-01
	sumInsertTxn(t, db, 1, "2025-01-05", rptAcctCash, rptAcctSalary, 3000)
	// 買入投資：DR 1102-01 [INVESTING], CR 1101-01
	sumInsertTxnWithCF(t, db, 2, "2025-01-10", rptAcctInvestFVTPL, rptAcctCash, 5000, cfPtr("INVESTING"), nil)
	// 手續費：DR 5402-01 [INVESTING], CR 1101-01
	sumInsertTxnWithCF(t, db, 3, "2025-01-10", rptAcctStockFee, rptAcctCash, 20, cfPtr("INVESTING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：IS NI = 3000(income) - 20(expense) = 2980
	//   niReclassification = -20（手續費 EXPENSE INVESTING → credit-debit = 0-20 = -20）
	//   OperatingNI = 2980 - (-20) = 3000
	opNI, _ := indirect.OperatingActivities.NetIncome.Float64()
	if opNI != 3000 {
		t.Errorf("indirect OperatingActivities.NetIncome: got %.2f, want 3000", opNI)
	}
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 3000 {
		t.Errorf("indirect OperatingActivities.Total: got %.2f, want 3000", opTotal)
	}

	// 直接法：薪資交易（INCOME/NULL）→ operating_txns，CashReceived=3000
	//   買入、手續費（INVESTING tag）→ 不進 operating_txns
	directReceived, _ := direct.OperatingActivities.CashReceived.Float64()
	if directReceived != 3000 {
		t.Errorf("direct CashReceived: got %.2f, want 3000", directReceived)
	}
	directPaid, _ := direct.OperatingActivities.CashPaid.Float64()
	if directPaid != 0 {
		t.Errorf("direct CashPaid: got %.2f, want 0 (fee excluded from operating)", directPaid)
	}

	// 兩者 OperatingTotal 一致
	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingActivities.Total mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}

	// InvestingTotal：1102-01=-5000, 5402-01=-20 → -5020
	invTotal, _ := indirect.InvestingActivities.Total.Float64()
	if invTotal != -5020 {
		t.Errorf("indirect InvestingActivities.Total: got %.2f, want -5020", invTotal)
	}
	if !indirect.InvestingActivities.Total.Equal(direct.InvestingActivities.Total) {
		t.Errorf("InvestingActivities.Total mismatch: indirect=%.2f, direct=%.2f",
			indirect.InvestingActivities.Total.InexactFloat64(),
			direct.InvestingActivities.Total.InexactFloat64())
	}

	// NetChange = 3000 (op) + (-5020) (inv) = -2020
	netChange, _ := indirect.NetChange.Float64()
	if netChange != -2020 {
		t.Errorf("indirect NetChange: got %.2f, want -2020", netChange)
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// ── 預付費用 CF 測試 ──────────────────────────────────────────────────────────

// 預付費用帳號常數（來自 accounts.sql seeds）
const (
	rptAcctPrepaidAsset    = "1103-01" // 應收薪資款 (ASSET, DEBIT) — 作為預付科目的代替
	rptAcctPrepaidExpense  = "5101-01" // 房租費用 (EXPENSE, DEBIT)
	rptAcctAccumDepr       = "1103-02" // 應收租金 (ASSET, DEBIT) — 作為累計折舊的代替
	rptAcctAsset           = "1102-01" // 投資 FVTPL (ASSET, DEBIT, INVESTING) — 作為固定資產的代替
	rptAcctDisposalGain    = "4205" // 處分固定資產利得 (INCOME, CREDIT, INVESTING)
	rptAcctDisposalLoss    = "5601" // 固定資產處分損失 (EXPENSE, DEBIT, INVESTING)
)

// TestGetCashFlowStatement_PrepaidCreated_OperatingOutflow
// 預付費用付款：DR 預付科目 [OPERATING], CR 現金。
// 間接法：NI=0，OPERATING 調整=-12000（預付資產增加）→ OperatingTotal=-12000。
// 直接法：OPERATING-tagged entry 觸發 operating_txns，CR 現金→ CashPaid=12000 → OperatingTotal=-12000。
// 兩法一致。
func TestGetCashFlowStatement_PrepaidCreated_OperatingOutflow(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 預付科目 [OPERATING] 12000, CR 現金 [NULL] 12000
	sumInsertTxnWithCF(t, db, 1, "2025-01-15", rptAcctPrepaidAsset, rptAcctCash, 12000, cfPtr("OPERATING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：NI=0，OPERATING 調整：DR 預付 → credit-debit = 0-12000 = -12000
	niVal, _ := indirect.OperatingActivities.NetIncome.Float64()
	if niVal != 0 {
		t.Errorf("indirect NetIncome: got %.2f, want 0", niVal)
	}
	assertCFItem(t, indirect.OperatingActivities.Adjustments, rptAcctPrepaidAsset, -12000)
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != -12000 {
		t.Errorf("indirect OperatingTotal: got %.2f, want -12000", opTotal)
	}

	// 直接法：OPERATING 觸發 operating_txns，CR 現金 → CashPaid=12000
	directPaid, _ := direct.OperatingActivities.CashPaid.Float64()
	if directPaid != 12000 {
		t.Errorf("direct CashPaid: got %.2f, want 12000", directPaid)
	}

	// 兩法 OperatingTotal 一致
	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// TestGetCashFlowStatement_PrepaidAmortized_NonCash
// 預付費用攤提（非現金）：DR 費用, CR 預付 [OPERATING]。
// 間接法：NI=-3000，OPERATING add-back=+3000 → OperatingTotal=0。
// 直接法：費用科目→ operating_txns，但無現金科目 → CashPaid=0 → OperatingTotal=0。
// 兩法一致（均為 0）。
func TestGetCashFlowStatement_PrepaidAmortized_NonCash(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 費用 [NULL] 3000, CR 預付科目 [OPERATING] 3000（攤提分錄）
	sumInsertTxnWithCF(t, db, 1, "2025-01-31", rptAcctPrepaidExpense, rptAcctPrepaidAsset, 3000, nil, cfPtr("OPERATING"))

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：NI=-3000（費用），OPERATING CR 預付 → credit-debit=+3000（加回）→ Total=0
	niVal, _ := indirect.OperatingActivities.NetIncome.Float64()
	if niVal != -3000 {
		t.Errorf("indirect NetIncome: got %.2f, want -3000", niVal)
	}
	assertCFItem(t, indirect.OperatingActivities.Adjustments, rptAcctPrepaidAsset, 3000)
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingTotal: got %.2f, want 0 (non-cash amortization)", opTotal)
	}

	// 直接法：費用帳→ operating_txns，無現金科目 → OperatingTotal=0
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingTotal: got %.2f, want 0", directOpTotal)
	}

	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// ── 固定資產 CF 測試 ──────────────────────────────────────────────────────────

// TestGetCashFlowStatement_AssetPurchase_InvestingOutflow
// 現金購買固定資產：DR 資產 [INVESTING], CR 現金。
// 間接法：NI=0，投資活動=-30000 → NetChange=-30000。
// 直接法：無 INCOME/EXPENSE → 不進 operating；InvestingTotal=-30000。
// 兩法 InvestingTotal、NetChange 一致。
func TestGetCashFlowStatement_AssetPurchase_InvestingOutflow(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 資產 [INVESTING] 30000, CR 現金 [NULL] 30000
	sumInsertTxnWithCF(t, db, 1, "2025-01-10", rptAcctAsset, rptAcctCash, 30000, cfPtr("INVESTING"), nil)

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 無 NI、無 OPERATING 調整
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingTotal: got %.2f, want 0", opTotal)
	}

	// 投資活動：DR 資產 30000 → amount = 0-30000 = -30000
	assertCFItem(t, indirect.InvestingActivities.Items, rptAcctAsset, -30000)
	invTotal, _ := indirect.InvestingActivities.Total.Float64()
	if invTotal != -30000 {
		t.Errorf("indirect InvestingTotal: got %.2f, want -30000", invTotal)
	}

	// 直接法一致
	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}
	if !indirect.InvestingActivities.Total.Equal(direct.InvestingActivities.Total) {
		t.Errorf("InvestingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.InvestingActivities.Total.InexactFloat64(),
			direct.InvestingActivities.Total.InexactFloat64())
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// TestGetCashFlowStatement_AssetDepreciation_NonCash
// 固定資產折舊（非現金）：DR 折舊費用, CR 累計折舊 [OPERATING]。
// 間接法：NI=-5000，OPERATING add-back=+5000 → OperatingTotal=0。
// 直接法：費用科目→ operating_txns，無現金科目 → OperatingTotal=0。
// 兩法均為 0，NetChange=0。
func TestGetCashFlowStatement_AssetDepreciation_NonCash(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// DR 折舊費用 [NULL] 5000, CR 累計折舊 [OPERATING] 5000
	sumInsertTxnWithCF(t, db, 1, "2025-01-31", rptAcctPrepaidExpense, rptAcctAccumDepr, 5000, nil, cfPtr("OPERATING"))

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：NI=-5000，OPERATING CR 累計折舊 → credit-debit=+5000 → Total=0
	niVal, _ := indirect.OperatingActivities.NetIncome.Float64()
	if niVal != -5000 {
		t.Errorf("indirect NetIncome: got %.2f, want -5000", niVal)
	}
	assertCFItem(t, indirect.OperatingActivities.Adjustments, rptAcctAccumDepr, 5000)
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingTotal: got %.2f, want 0 (non-cash depreciation)", opTotal)
	}

	// 直接法：費用帳→ operating_txns，無現金科目 → OperatingTotal=0
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingTotal: got %.2f, want 0", directOpTotal)
	}

	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}
	netChange, _ := indirect.NetChange.Float64()
	if netChange != 0 {
		t.Errorf("indirect NetChange: got %.2f, want 0 (no cash movement)", netChange)
	}
}

// TestGetCashFlowStatement_AssetDisposal_WithGain_DirectIndirectConsistency
// 固定資產處分（有利得）：
// DR 累計折舊 [INVESTING], CR 資產 [INVESTING], DR 現金 [NULL], CR 利得 [INVESTING]。
// 間接法：NI=利得，重分類至 INVESTING → OperatingNI=0；InvestingTotal=收款金額。
// 直接法：利得 [INVESTING] 排除 → 不進 operating；InvestingTotal 一致。
func TestGetCashFlowStatement_AssetDisposal_WithGain_DirectIndirectConsistency(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 資產資料：原價 50000，累計折舊 20000，帳面值 30000，售價 35000，利得 5000
	// DR 累計折舊 [INVESTING] 20000, CR 資產 [INVESTING] 50000
	// DR 現金 [NULL] 35000, CR 利得 [INVESTING] 5000
	// 借：20000+35000=55000；貸：50000+5000=55000 ✓
	sumInsertTxnMulti(t, db, 1, "2025-01-20", 35000, []cfEntry{
		{accountID: rptAcctAccumDepr, debit: 20000, credit: 0, cf: cfPtr("INVESTING")},
		{accountID: rptAcctAsset, debit: 0, credit: 50000, cf: cfPtr("INVESTING")},
		{accountID: rptAcctCash, debit: 35000, credit: 0, cf: nil},
		{accountID: rptAcctDisposalGain, debit: 0, credit: 5000, cf: cfPtr("INVESTING")},
	})

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：IS NI = 5000（利得 INCOME），但利得 [INVESTING] → 重分類移出 NI
	// OperatingNI = 5000 - 5000 = 0
	opNI, _ := indirect.OperatingActivities.NetIncome.Float64()
	if opNI != 0 {
		t.Errorf("indirect OperatingNI: got %.2f, want 0 (gain reclassified to INVESTING)", opNI)
	}
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingTotal: got %.2f, want 0", opTotal)
	}

	// InvestingTotal = (-20000) + 50000 + 5000 = 35000（等於現金收款）
	invTotal, _ := indirect.InvestingActivities.Total.Float64()
	if invTotal != 35000 {
		t.Errorf("indirect InvestingTotal: got %.2f, want 35000 (= cash proceeds)", invTotal)
	}

	// 直接法：利得 [INVESTING] 排除 → 整筆交易不進 operating；OperatingTotal=0
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingTotal: got %.2f, want 0", directOpTotal)
	}

	// 兩法一致
	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}
	if !indirect.InvestingActivities.Total.Equal(direct.InvestingActivities.Total) {
		t.Errorf("InvestingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.InvestingActivities.Total.InexactFloat64(),
			direct.InvestingActivities.Total.InexactFloat64())
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}

// TestGetCashFlowStatement_AssetDisposal_WithLoss_DirectIndirectConsistency
// 固定資產處分（有損失）：
// DR 累計折舊 [INVESTING], CR 資產 [INVESTING], DR 現金 [NULL], DR 損失 [INVESTING]。
// 間接法：NI=-損失，重分類至 INVESTING → OperatingNI=0；InvestingTotal=現金收款。
func TestGetCashFlowStatement_AssetDisposal_WithLoss_DirectIndirectConsistency(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 資產資料：原價 50000，累計折舊 20000，帳面值 30000，售價 25000，損失 5000
	// DR 累計折舊 [INVESTING] 20000, CR 資產 [INVESTING] 50000
	// DR 現金 [NULL] 25000, DR 損失 [INVESTING] 5000
	// 借：20000+25000+5000=50000；貸：50000 ✓
	sumInsertTxnMulti(t, db, 1, "2025-01-20", 25000, []cfEntry{
		{accountID: rptAcctAccumDepr, debit: 20000, credit: 0, cf: cfPtr("INVESTING")},
		{accountID: rptAcctAsset, debit: 0, credit: 50000, cf: cfPtr("INVESTING")},
		{accountID: rptAcctCash, debit: 25000, credit: 0, cf: nil},
		{accountID: rptAcctDisposalLoss, debit: 5000, credit: 0, cf: cfPtr("INVESTING")},
	})

	indirect, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}
	direct, err := repo.GetDirectCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetDirectCashFlowStatement: %v", err)
	}

	// 間接法：IS NI = -5000（損失 EXPENSE），損失 [INVESTING] → 重分類移出 NI
	// OperatingNI = -5000 - (-5000) = 0
	opNI, _ := indirect.OperatingActivities.NetIncome.Float64()
	if opNI != 0 {
		t.Errorf("indirect OperatingNI: got %.2f, want 0 (loss reclassified to INVESTING)", opNI)
	}
	opTotal, _ := indirect.OperatingActivities.Total.Float64()
	if opTotal != 0 {
		t.Errorf("indirect OperatingTotal: got %.2f, want 0", opTotal)
	}

	// InvestingTotal = (-20000) + 50000 + (-5000) = 25000（等於現金收款）
	invTotal, _ := indirect.InvestingActivities.Total.Float64()
	if invTotal != 25000 {
		t.Errorf("indirect InvestingTotal: got %.2f, want 25000 (= cash proceeds)", invTotal)
	}

	// 直接法：損失 [INVESTING] 排除 → 不進 operating；OperatingTotal=0
	directOpTotal, _ := direct.OperatingActivities.Total.Float64()
	if directOpTotal != 0 {
		t.Errorf("direct OperatingTotal: got %.2f, want 0", directOpTotal)
	}

	// 兩法一致
	if !indirect.OperatingActivities.Total.Equal(direct.OperatingActivities.Total) {
		t.Errorf("OperatingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.OperatingActivities.Total.InexactFloat64(),
			direct.OperatingActivities.Total.InexactFloat64())
	}
	if !indirect.InvestingActivities.Total.Equal(direct.InvestingActivities.Total) {
		t.Errorf("InvestingTotal mismatch: indirect=%.2f, direct=%.2f",
			indirect.InvestingActivities.Total.InexactFloat64(),
			direct.InvestingActivities.Total.InexactFloat64())
	}
	if !indirect.NetChange.Equal(direct.NetChange) {
		t.Errorf("NetChange mismatch: indirect=%.2f, direct=%.2f",
			indirect.NetChange.InexactFloat64(), direct.NetChange.InexactFloat64())
	}
}
