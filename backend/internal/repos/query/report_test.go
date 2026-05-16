package query_test

import (
	"testing"

	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/report"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/testutil"

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

// TestGetCashFlowStatement_ThreeSections
// 三大活動分類各自出現在對應 section，金額符合 credit−debit 公式。
func TestGetCashFlowStatement_ThreeSections(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// OPERATING: DR 1101-01, CR 1103-01 800（收回應收款 → 現金流入 +800）
	sumInsertTxn(t, db, 1, "2025-01-10", rptAcctCash, rptAcctOperating, 800)
	// INVESTING: DR 1102-01, CR 1101-01 2000（買入投資 → 現金流出 -2000）
	sumInsertTxn(t, db, 2, "2025-01-15", rptAcctInvesting, rptAcctCash, 2000)
	// FINANCING: DR 1101-01, CR 3103-02 3000（增資 → 現金流入 +3000）
	sumInsertTxn(t, db, 3, "2025-01-20", rptAcctCash, rptAcctEqLeaf2, 3000)

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
	sumInsertTxn(t, db, 1, "2025-01-15", rptAcctCash, rptAcctSalary, 3000) // CR income, DR CASH
	// 有 OPERATING 帳：應收薪資 800（M-1 期間）
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctCash, rptAcctOperating, 800) // CR 1103-01
	if err := snapRepo.BulkInsert(ctx, testMerchantID, 1); err != nil {
		t.Fatalf("BulkInsert: %v", err)
	}

	// M（二月，查詢期間）：OPERATING 應收帳 500，INVESTING 買入 1000
	sumInsertTxn(t, db, 3, "2025-02-10", rptAcctCash, rptAcctOperating, 500)  // CR 1103-01（應收收回，流入）
	sumInsertTxn(t, db, 4, "2025-02-20", rptAcctInvesting, rptAcctCash, 1000) // DR 1102-01（買入投資，流出）

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

	// 收入 3000 → operating net_income = 3000（無其他調整）
	sumInsertTxn(t, db, 1, "2025-01-05", rptAcctCash, rptAcctSalary, 3000)
	// 買入投資 1000 → investing = -1000
	sumInsertTxn(t, db, 2, "2025-01-10", rptAcctInvesting, rptAcctCash, 1000)
	// 增資 500 → financing = +500
	sumInsertTxn(t, db, 3, "2025-01-15", rptAcctCash, rptAcctEqLeaf2, 500)

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
// 將彙總科目設定 cash_flow_category，其後裔葉節點設為 NULL；
// 現金流量表應出現彙總科目行（is_summary=true），金額為所有後裔期間變動的加總。
func TestGetCashFlowStatement_HierarchyAggregation(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 將彙總節點 1103 設為 OPERATING，葉節點清為 NULL（只顯示彙總行）
	_, err := db.ExecContext(ctx,
		`UPDATE accounts SET cash_flow_category = 'OPERATING' WHERE account_id = ? AND merchant_id = ?`,
		rptAcctCFSummary, testMerchantID)
	if err != nil {
		t.Fatalf("update summary category: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`UPDATE accounts SET cash_flow_category = NULL WHERE account_id IN (?, ?) AND merchant_id = ?`,
		rptAcctCFLeafA, rptAcctCFLeafB, testMerchantID)
	if err != nil {
		t.Fatalf("update leaf category: %v", err)
	}

	// 期間交易：1103-01 credit 800（現金收回應收薪資），1103-02 credit 500（收回應收租金）
	sumInsertTxn(t, db, 1, "2025-01-10", rptAcctCash, rptAcctCFLeafA, 800)
	sumInsertTxn(t, db, 2, "2025-01-20", rptAcctCash, rptAcctCFLeafB, 500)

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	// 彙總行 1103 應出現在 OPERATING，金額 = (800+500) credit - 0 debit = 1300
	var found *report.CashFlowItem
	for i, item := range cf.OperatingActivities.Adjustments {
		if item.AccountId == rptAcctCFSummary {
			found = &cf.OperatingActivities.Adjustments[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("summary account %s not found in OPERATING adjustments", rptAcctCFSummary)
	}
	if !found.IsSummary {
		t.Errorf("account %s: IsSummary got false, want true", rptAcctCFSummary)
	}
	got, _ := found.Amount.Float64()
	if got != 1300 {
		t.Errorf("account %s amount: got %.2f, want 1300", rptAcctCFSummary, got)
	}

	// 葉節點（category=NULL）不應出現在結果中
	for _, item := range cf.OperatingActivities.Adjustments {
		if item.AccountId == rptAcctCFLeafA || item.AccountId == rptAcctCFLeafB {
			t.Errorf("leaf account %s should not appear in CF when category is NULL", item.AccountId)
		}
	}
}

// TestGetCashFlowStatement_HierarchyAggregation_LeafIsSummaryFalse
// 一般葉節點（非彙總）的 is_summary 欄位應為 false。
func TestGetCashFlowStatement_HierarchyAggregation_LeafIsSummaryFalse(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := query.NewReportRepo(db)
	ctx := testCtx()

	// 使用既有葉節點 1103-01（OPERATING），不修改 category
	sumInsertTxn(t, db, 1, "2025-01-10", rptAcctCash, rptAcctOperating, 600)

	cf, err := repo.GetCashFlowStatement(ctx, "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("GetCashFlowStatement: %v", err)
	}

	var found *report.CashFlowItem
	for i, item := range cf.OperatingActivities.Adjustments {
		if item.AccountId == rptAcctOperating {
			found = &cf.OperatingActivities.Adjustments[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("leaf account %s not found in OPERATING adjustments", rptAcctOperating)
	}
	if found.IsSummary {
		t.Errorf("account %s: IsSummary got true, want false for leaf", rptAcctOperating)
	}
}
