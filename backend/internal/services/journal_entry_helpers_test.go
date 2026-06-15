package services_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
)

// tHelper 是 helper 函數所需的最小測試接口。
// 同時被 *testing.T 與 GinkgoT() 滿足，不依賴 tHelper 的私有方法。
type tHelper interface {
	Helper()
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
}

const testMID int64 = 1

// ─────────────────────────────────────────
// Service / context helpers
// ─────────────────────────────────────────

func newSvc(db *sqlx.DB) *services.EventStoreService {
	uow := event_store.NewUnitOfWork(db)
	q := query.NewQueryRepository(db)
	return services.NewEventStoreService(uow, q)
}

func testCtx() context.Context {
	return context.WithValue(context.Background(), ctxkey.MerchantID, testMID)
}

// ─────────────────────────────────────────
// Appender: tracks aggregate version
// ─────────────────────────────────────────

type appender struct {
	t       tHelper
	svc     *services.EventStoreService
	ctx     context.Context
	aggID   string
	version int64
}

func newAppender(t tHelper, svc *services.EventStoreService, ctx context.Context, aggID string) *appender {
	t.Helper()
	return &appender{t: t, svc: svc, ctx: ctx, aggID: aggID}
}

func (a *appender) do(et event_types.EventType, p any) {
	a.t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		a.t.Fatalf("marshal payload: %v", err)
	}
	_, err = a.svc.Append(a.ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateTransaction.Enum(),
		AggregateID:     a.aggID,
		ExpectedVersion: a.version,
		EventType:       et,
		Payload:         raw,
	})
	if err != nil {
		a.t.Fatalf("Append %s: %v", et, err)
	}
	a.version++
}

// ─────────────────────────────────────────
// DB setup helpers
// ─────────────────────────────────────────

func insertLedger(t tHelper, db *sqlx.DB, id int64, typ, acctID string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO ledger_accounts(ledger_id, merchant_id, account_id, institution, name, type, currency, is_active, version, ledger_uuid)
		 VALUES(?, ?, ?, 'TEST', 'test ledger', ?, 'TWD', 1, 1, ?)`,
		id, testMID, acctID, typ, testLedgerUUID(id),
	)
	if err != nil {
		t.Fatalf("insertLedger(%d): %v", id, err)
	}
}

func testLedgerUUID(id int64) string {
	return fmt.Sprintf("test-ledger-uuid-%d", id)
}

func insertPeriodOpen(t tHelper, db *sqlx.DB, id int64, startDate string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO period_closings(closing_id, merchant_id, period_type, period_start, period_end, status, closed_at)
		 VALUES(?, ?, 'MONTHLY', ?, date(?, '+1 month', '-1 day'), 'OPEN', NULL)`,
		id, testMID, startDate, startDate,
	)
	if err != nil {
		t.Fatalf("insertPeriodOpen(%d, %s): %v", id, startDate, err)
	}
}

func insertAllMonthsClosed(t tHelper, db *sqlx.DB, year int, baseID int64) {
	t.Helper()
	ctx := testCtx()
	for m := 1; m <= 12; m++ {
		startDate := fmt.Sprintf("%04d-%02d-01", year, m)
		id := baseID + int64(m) - 1
		_, err := db.ExecContext(ctx,
			`INSERT INTO period_closings(closing_id, merchant_id, period_type, period_start, period_end, status, closed_at)
			 VALUES(?, ?, 'MONTHLY', ?, date(?, '+1 month', '-1 day'), 'CLOSED', datetime('now'))`,
			id, testMID, startDate, startDate,
		)
		if err != nil {
			t.Fatalf("insertAllMonthsClosed month %d: %v", m, err)
		}
	}
}

func insertAnnualPeriodOpen(t tHelper, db *sqlx.DB, id int64, year int) {
	t.Helper()
	startDate := fmt.Sprintf("%04d-01-01", year)
	endDate := fmt.Sprintf("%04d-12-31", year)
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO period_closings(closing_id, merchant_id, period_type, period_start, period_end, status, closed_at)
		 VALUES(?, ?, 'ANNUAL', ?, ?, 'OPEN', NULL)`,
		id, testMID, startDate, endDate,
	)
	if err != nil {
		t.Fatalf("insertAnnualPeriodOpen: %v", err)
	}
}

func insertFixedAssetCategory(t tHelper, db *sqlx.DB, uuid, name, assetAcct, accumAcct, deprAcct string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO fixed_asset_categories (category_uuid, merchant_id, name, asset_account_id, accum_depreciation_account_id, depreciation_expense_account_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		uuid, testMID, name, assetAcct, accumAcct, deprAcct,
	)
	if err != nil {
		t.Fatalf("insertFixedAssetCategory: %v", err)
	}
}

func insertPrepaidCategory(t tHelper, db *sqlx.DB, uuid, name, accountID, expenseAccountID string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO prepaid_categories (category_uuid, merchant_id, name, account_id, expense_account_id)
		 VALUES (?, ?, ?, ?, ?)`,
		uuid, testMID, name, accountID, expenseAccountID,
	)
	if err != nil {
		t.Fatalf("insertPrepaidCategory: %v", err)
	}
}

func insertSysAccounts(t tHelper, db *sqlx.DB) {
	t.Helper()
	ctx := testCtx()
	rows := []struct{ code, desc, acct string }{
		{"SYS:ASSET:PREPAID_INTEREST", "預付利息（資產）", "1104-04"},
		{"SYS:EXPENSE:LOAN:INTEREST", "貸款利息費用", "5401-01"},
		{"SYS:EXPENSE:CREDIT_CARD:INTEREST", "信用卡利息費用", "5401-02"},
		{"SYS:EQUITY:CLOSE_NET_INCOME", "結清淨收益（權益）", "3102-01"},
		{"SYS:EQUITY:EQUITY_OPENING", "期初權益", "3101-01"},
	}
	for _, r := range rows {
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO sys_accounts(merchant_id, sys_code, description, account_id)
			 VALUES(?, ?, ?, ?)`,
			testMID, r.code, r.desc, r.acct,
		)
		if err != nil {
			t.Fatalf("insertSysAccounts(%s): %v", r.code, err)
		}
	}
}

func insertAssetTypeConfig(t tHelper, db *sqlx.DB) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT OR IGNORE INTO asset_type_account_config(
		   merchant_id, asset_type,
		   realized_gain_account_id, realized_loss_account_id,
		   unrealized_gain_account_id, unrealized_loss_account_id,
		   oci_account_id, fee_account_id, tax_account_id, account_id
		 ) VALUES(?, 'STOCK', '4203-01', '5402-11', '4204-01', '5402-21', '3102-02', '5402-01', '5402-05', '1102-01')`,
		testMID,
	)
	if err != nil {
		t.Fatalf("insertAssetTypeConfig: %v", err)
	}
}

func testInvestmentUUID(id int64) string {
	return fmt.Sprintf("test-inv-uuid-%d", id)
}

func insertInvestment(t tHelper, db *sqlx.DB, id int64, acctID, assetType, costMethod, ifrsCategory string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO investments(investment_id, merchant_id, account_id, asset_type, currency, symbol, name, cost_method, ifrs_category, is_active, version, uuid)
		 VALUES(?, ?, ?, ?, 'TWD', 'TEST', 'TestStock', ?, ?, 1, 1, ?)`,
		id, testMID, acctID, assetType, costMethod, ifrsCategory, testInvestmentUUID(id),
	)
	if err != nil {
		t.Fatalf("insertInvestment(%d): %v", id, err)
	}
}

func insertDividendAccounts(t tHelper, db *sqlx.DB) {
	t.Helper()
	ctx := testCtx()
	rows := []struct{ id, typ, nb string }{
		{"4210", "INCOME", "CREDIT"},
		{"5920", "EXPENSE", "DEBIT"},
	}
	for _, a := range rows {
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO accounts(account_id, merchant_id, parent_id, name, type, normal_balance, is_summary, is_active, version)
			 VALUES(?, ?, NULL, ?, ?, ?, 0, 1, 1)`,
			a.id, testMID, a.id, a.typ, a.nb,
		)
		if err != nil {
			t.Fatalf("insertDividendAccounts(%s): %v", a.id, err)
		}
	}
}

func insertIncomeTransaction(t tHelper, db *sqlx.DB, txnID int64, date, creditAcct string, amount float64) {
	t.Helper()
	ctx := testCtx()
	_, err := db.ExecContext(ctx,
		`INSERT INTO transactions(txn_id, merchant_id, txn_date, description, total_amount, currency, status, version)
		 VALUES(?, ?, ?, 'test income', ?, 'TWD', 'ACTIVE', 1)`,
		txnID, testMID, date, amount,
	)
	if err != nil {
		t.Fatalf("insertIncomeTransaction: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries(txn_id, merchant_id, account_id, debit, credit)
		 VALUES(?, ?, ?, 0, ?)`,
		txnID, testMID, creditAcct, amount,
	)
	if err != nil {
		t.Fatalf("insertIncomeTransaction credit entry: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries(txn_id, merchant_id, account_id, debit, credit)
		 VALUES(?, ?, '1101-02', ?, 0)`,
		txnID, testMID, amount,
	)
	if err != nil {
		t.Fatalf("insertIncomeTransaction debit entry: %v", err)
	}
}

// ─────────────────────────────────────────
// Category query helpers
// ─────────────────────────────────────────

type fixedAssetCategoryRow struct {
	CategoryUUID                 string `db:"category_uuid"`
	Name                         string `db:"name"`
	AssetAccountID               string `db:"asset_account_id"`
	AccumDepreciationAccountID   string `db:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string `db:"depreciation_expense_account_id"`
	IsActive                     bool   `db:"is_active"`
}

func queryFixedAssetCategory(t tHelper, db *sqlx.DB, uuid string) fixedAssetCategoryRow {
	t.Helper()
	var row fixedAssetCategoryRow
	err := db.GetContext(testCtx(), &row,
		`SELECT category_uuid, name, asset_account_id, accum_depreciation_account_id, depreciation_expense_account_id, is_active
		 FROM fixed_asset_categories WHERE category_uuid=? AND merchant_id=?`, uuid, testMID)
	if err != nil {
		t.Fatalf("queryFixedAssetCategory(%s): %v", uuid, err)
	}
	return row
}

func queryLastFixedAssetCategoryUUID(t tHelper, db *sqlx.DB) string {
	t.Helper()
	var id string
	err := db.GetContext(testCtx(), &id,
		`SELECT category_uuid FROM fixed_asset_categories WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastFixedAssetCategoryUUID: %v", err)
	}
	return id
}

func queryLastPrepaidCategoryUUID(t tHelper, db *sqlx.DB) string {
	t.Helper()
	var id string
	err := db.GetContext(testCtx(), &id,
		`SELECT category_uuid FROM prepaid_categories WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastPrepaidCategoryUUID: %v", err)
	}
	return id
}

type prepaidCategoryRow struct {
	CategoryUUID     string `db:"category_uuid"`
	Name             string `db:"name"`
	AccountID        string `db:"account_id"`
	ExpenseAccountID string `db:"expense_account_id"`
	IsActive         bool   `db:"is_active"`
}

func queryPrepaidCategory(t tHelper, db *sqlx.DB, uuid string) prepaidCategoryRow {
	t.Helper()
	var row prepaidCategoryRow
	err := db.GetContext(testCtx(), &row,
		`SELECT category_uuid, name, account_id, expense_account_id, is_active
		 FROM prepaid_categories WHERE category_uuid=? AND merchant_id=?`, uuid, testMID)
	if err != nil {
		t.Fatalf("queryPrepaidCategory(%s): %v", uuid, err)
	}
	return row
}

// ─────────────────────────────────────────
// Query helpers
// ─────────────────────────────────────────

type txnRow struct {
	TxnID       int64
	Date        string
	Description string
	TotalAmount decimal.Decimal
	Status      string
}

type txnScanRow struct {
	TxnID       int64   `db:"txn_id"`
	Date        string  `db:"txn_date"`
	Description string  `db:"description"`
	TotalAmount float64 `db:"total_amount"`
	Status      string  `db:"status"`
}

func toTxnRow(r txnScanRow) txnRow {
	return txnRow{
		TxnID:       r.TxnID,
		Date:        r.Date,
		Description: r.Description,
		TotalAmount: decimal.NewFromFloat(r.TotalAmount),
		Status:      r.Status,
	}
}

func queryLastTxn(t tHelper, db *sqlx.DB) txnRow {
	t.Helper()
	var row txnScanRow
	err := db.GetContext(testCtx(), &row,
		`SELECT txn_id, txn_date, description, total_amount, status
		 FROM transactions WHERE merchant_id=? AND status='ACTIVE' ORDER BY txn_id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastTxn: %v", err)
	}
	return toTxnRow(row)
}

func countActiveTxns(t tHelper, db *sqlx.DB) int {
	t.Helper()
	var n int
	err := db.GetContext(testCtx(), &n,
		`SELECT COUNT(*) FROM transactions WHERE merchant_id=? AND status='ACTIVE'`, testMID)
	if err != nil {
		t.Fatalf("countActiveTxns: %v", err)
	}
	return n
}

type entryScanRow struct {
	AccountID        string  `db:"account_id"`
	LedgerID         *int64  `db:"ledger_id"`
	Debit            float64 `db:"debit"`
	Credit           float64 `db:"credit"`
	CashFlowCategory *string `db:"cash_flow_category"`
}

type entryRow struct {
	AccountID        string
	LedgerID         *int64
	Debit            decimal.Decimal
	Credit           decimal.Decimal
	CashFlowCategory *string
}

func queryEntries(t tHelper, db *sqlx.DB, txnID int64) []entryRow {
	t.Helper()
	var rows []entryScanRow
	err := db.SelectContext(testCtx(), &rows,
		`SELECT je.account_id, je.ledger_id, je.debit, je.credit, ecc.cf_category AS cash_flow_category
		 FROM journal_entries je
		 LEFT JOIN entry_cf_categories ecc ON ecc.entry_uuid = je.entry_uuid AND ecc.merchant_id = je.merchant_id
		 WHERE je.txn_id=? AND je.merchant_id=? ORDER BY je.entry_id`, txnID, testMID)
	if err != nil {
		t.Fatalf("queryEntries(%d): %v", txnID, err)
	}
	result := make([]entryRow, len(rows))
	for i, r := range rows {
		result[i] = entryRow{
			AccountID:        r.AccountID,
			LedgerID:         r.LedgerID,
			Debit:            decimal.NewFromFloat(r.Debit),
			Credit:           decimal.NewFromFloat(r.Credit),
			CashFlowCategory: r.CashFlowCategory,
		}
	}
	return result
}

func queryLastPrepaidID(t tHelper, db *sqlx.DB) int64 {
	t.Helper()
	var id int64
	err := db.GetContext(testCtx(), &id,
		`SELECT id FROM prepaids WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastPrepaidID: %v", err)
	}
	return id
}

func queryLastAssetID(t tHelper, db *sqlx.DB) int64 {
	t.Helper()
	var id int64
	err := db.GetContext(testCtx(), &id,
		`SELECT id FROM fixed_assets WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastAssetID: %v", err)
	}
	return id
}

func queryLastInstallmentID(t tHelper, db *sqlx.DB) int64 {
	t.Helper()
	var id int64
	err := db.GetContext(testCtx(), &id,
		`SELECT installment_id FROM installments WHERE merchant_id=? ORDER BY installment_id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastInstallmentID: %v", err)
	}
	return id
}

func queryLastInstallmentUUID(t tHelper, db *sqlx.DB) string {
	t.Helper()
	var uuid string
	err := db.GetContext(testCtx(), &uuid,
		`SELECT installment_uuid FROM installments WHERE merchant_id=? ORDER BY installment_id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastInstallmentUUID: %v", err)
	}
	return uuid
}

func queryLastPrepaidUUID(t tHelper, db *sqlx.DB) string {
	t.Helper()
	var uuid string
	err := db.GetContext(testCtx(), &uuid,
		`SELECT prepaid_uuid FROM prepaids WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastPrepaidUUID: %v", err)
	}
	return uuid
}

func queryLastAssetUUID(t tHelper, db *sqlx.DB) string {
	t.Helper()
	var uuid string
	err := db.GetContext(testCtx(), &uuid,
		`SELECT asset_uuid FROM fixed_assets WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastAssetUUID: %v", err)
	}
	return uuid
}

// ─────────────────────────────────────────
// Assertion helpers
// ─────────────────────────────────────────

type rbScanRow struct {
	Debit  float64 `db:"debit_total"`
	Credit float64 `db:"credit_total"`
}

func assertAccountRB(t tHelper, db *sqlx.DB, acctID string, wantDebit, wantCredit decimal.Decimal) {
	t.Helper()
	var row rbScanRow
	err := db.GetContext(testCtx(), &row,
		`SELECT debit_total, credit_total FROM account_running_balances
		 WHERE account_id=? AND merchant_id=?`, acctID, testMID)
	if err != nil {
		t.Fatalf("assertAccountRB(%s): %v", acctID, err)
	}
	gotD := decimal.NewFromFloat(row.Debit)
	gotC := decimal.NewFromFloat(row.Credit)
	if !gotD.Equal(wantDebit) || !gotC.Equal(wantCredit) {
		t.Errorf("account_running_balance[%s]: want D=%s C=%s, got D=%s C=%s",
			acctID, wantDebit, wantCredit, gotD, gotC)
	}
}

func assertLedgerRB(t tHelper, db *sqlx.DB, ledgerID int64, wantDebit, wantCredit decimal.Decimal) {
	t.Helper()
	var row rbScanRow
	err := db.GetContext(testCtx(), &row,
		`SELECT debit_total, credit_total FROM ledger_running_balances
		 WHERE ledger_id=? AND merchant_id=?`, ledgerID, testMID)
	if err != nil {
		t.Fatalf("assertLedgerRB(%d): %v", ledgerID, err)
	}
	gotD := decimal.NewFromFloat(row.Debit)
	gotC := decimal.NewFromFloat(row.Credit)
	if !gotD.Equal(wantDebit) || !gotC.Equal(wantCredit) {
		t.Errorf("ledger_running_balance[%d]: want D=%s C=%s, got D=%s C=%s",
			ledgerID, wantDebit, wantCredit, gotD, gotC)
	}
}

func assertEntry(t tHelper, entries []entryRow, idx int, acctID string, ledgerID *int64, debit, credit decimal.Decimal, cfCat *string) {
	t.Helper()
	if idx >= len(entries) {
		t.Fatalf("entry[%d] not found (total %d entries)", idx, len(entries))
	}
	e := entries[idx]
	if e.AccountID != acctID {
		t.Errorf("entry[%d].AccountID: want %q, got %q", idx, acctID, e.AccountID)
	}
	switch {
	case ledgerID == nil && e.LedgerID != nil:
		t.Errorf("entry[%d].LedgerID: want nil, got %d", idx, *e.LedgerID)
	case ledgerID != nil && (e.LedgerID == nil || *e.LedgerID != *ledgerID):
		got := "<nil>"
		if e.LedgerID != nil {
			got = fmt.Sprintf("%d", *e.LedgerID)
		}
		t.Errorf("entry[%d].LedgerID: want %d, got %s", idx, *ledgerID, got)
	}
	if !e.Debit.Equal(debit) {
		t.Errorf("entry[%d][%s].Debit: want %s, got %s", idx, acctID, debit, e.Debit)
	}
	if !e.Credit.Equal(credit) {
		t.Errorf("entry[%d][%s].Credit: want %s, got %s", idx, acctID, credit, e.Credit)
	}
	switch {
	case cfCat == nil && e.CashFlowCategory != nil:
		t.Errorf("entry[%d][%s].CashFlowCategory: want nil, got %q", idx, acctID, *e.CashFlowCategory)
	case cfCat != nil && e.CashFlowCategory == nil:
		t.Errorf("entry[%d][%s].CashFlowCategory: want %q, got nil", idx, acctID, *cfCat)
	case cfCat != nil && e.CashFlowCategory != nil && *cfCat != *e.CashFlowCategory:
		t.Errorf("entry[%d][%s].CashFlowCategory: want %q, got %q", idx, acctID, *cfCat, *e.CashFlowCategory)
	}
}

// ─────────────────────────────────────────
// Utility
// ─────────────────────────────────────────

func dec(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

func lidPtr(id int64) *int64 { return &id }
func cfPtr(s string) *string { return &s }
