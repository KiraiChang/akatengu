package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"akatengu/internal/testutil"
)

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
	t       *testing.T
	svc     *services.EventStoreService
	ctx     context.Context
	aggID   string
	version int64
}

func newAppender(t *testing.T, svc *services.EventStoreService, ctx context.Context, aggID string) *appender {
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

func insertLedger(t *testing.T, db *sqlx.DB, id int64, typ, acctID string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO ledger_accounts(ledger_id, merchant_id, account_id, institution, name, type, currency, is_active, version, uuid)
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

func insertPeriodOpen(t *testing.T, db *sqlx.DB, id int64, startDate string) {
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

func insertAllMonthsClosed(t *testing.T, db *sqlx.DB, year int, baseID int64) {
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

func insertAnnualPeriodOpen(t *testing.T, db *sqlx.DB, id int64, year int) {
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

func insertSysAccounts(t *testing.T, db *sqlx.DB) {
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

func insertAssetTypeConfig(t *testing.T, db *sqlx.DB) {
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

func insertInvestment(t *testing.T, db *sqlx.DB, id int64, acctID, assetType, costMethod, ifrsCategory string) {
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

// insertDividendAccounts inserts "4210" and "5920" which are hardcoded in the DividendReceived factory.
func insertDividendAccounts(t *testing.T, db *sqlx.DB) {
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

// insertIncomeTransaction inserts a raw income transaction directly (bypasses pipeline).
// Used for annual close tests where we need income history without needing a period open.
func insertIncomeTransaction(t *testing.T, db *sqlx.DB, txnID int64, date, creditAcct string, amount float64) {
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
	// Balancing debit entry
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

func queryLastTxn(t *testing.T, db *sqlx.DB) txnRow {
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

func countActiveTxns(t *testing.T, db *sqlx.DB) int {
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

func queryEntries(t *testing.T, db *sqlx.DB, txnID int64) []entryRow {
	t.Helper()
	var rows []entryScanRow
	err := db.SelectContext(testCtx(), &rows,
		`SELECT account_id, ledger_id, debit, credit, cash_flow_category
		 FROM journal_entries WHERE txn_id=? AND merchant_id=? ORDER BY entry_id`, txnID, testMID)
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

func queryLastPrepaidID(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	var id int64
	err := db.GetContext(testCtx(), &id,
		`SELECT id FROM prepaids WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastPrepaidID: %v", err)
	}
	return id
}

func queryLastAssetID(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	var id int64
	err := db.GetContext(testCtx(), &id,
		`SELECT id FROM fixed_assets WHERE merchant_id=? ORDER BY id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastAssetID: %v", err)
	}
	return id
}

func queryLastInstallmentID(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	var id int64
	err := db.GetContext(testCtx(), &id,
		`SELECT installment_id FROM installments WHERE merchant_id=? ORDER BY installment_id DESC LIMIT 1`, testMID)
	if err != nil {
		t.Fatalf("queryLastInstallmentID: %v", err)
	}
	return id
}

func insertInstallment(t *testing.T, db *sqlx.DB, id int64, ledgerID int64, desc string) {
	t.Helper()
	_, err := db.ExecContext(testCtx(),
		`INSERT INTO installments(installment_id, merchant_id, installment_uuid, ledger_id, ledger_uuid, description, total_amount, total_periods,
		amount_per_period, start_date, interest_rate, interest_type, status, note)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, testMID, testInstallmentUUID(id), ledgerID, testLedgerUUID(ledgerID), desc,
	)
	if err != nil {
		t.Fatalf("insertInstallment(%d): %v", id, err)
	}
}

func testInstallmentUUID(id int64) string {
	return fmt.Sprintf("test-inst-uuid-%d", id)
}

type rbScanRow struct {
	Debit  float64 `db:"debit_total"`
	Credit float64 `db:"credit_total"`
}

func assertAccountRB(t *testing.T, db *sqlx.DB, acctID string, wantDebit, wantCredit decimal.Decimal) {
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

func assertLedgerRB(t *testing.T, db *sqlx.DB, ledgerID int64, wantDebit, wantCredit decimal.Decimal) {
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

func dec(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

func lidPtr(id int64) *int64 { return &id }
func cfPtr(s string) *string { return &s }

func assertEntry(t *testing.T, entries []entryRow, idx int, acctID string, ledgerID *int64, debit, credit decimal.Decimal, cfCat *string) {
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
// Tests
// ─────────────────────────────────────────

func TestPrepaidCreated(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "prepaid-1")
	a.do(event_types.EventPrepaidCreated.Enum(), payload.PrepaidCreatedPayload{
		AccountID:        "1104-01",
		ExpenseAccountID: "5201-01",
		LedgerID:         1,
		Name:             "人壽保險費 2026",
		TotalAmount:      dec("12000"),
		Periods:          12,
		StartDate:        "2026-05-01",
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-01" {
		t.Errorf("Date: want 2026-05-01, got %s", txn.Date)
	}
	if txn.Status != "ACTIVE" {
		t.Errorf("Status: want ACTIVE, got %s", txn.Status)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 1104-01 12000 [OPERATING], Cr 1101-02 12000 (ledger=1)
	assertEntry(t, entries, 0, "1104-01", nil, dec("12000"), dec("0"), cfPtr("OPERATING"))
	assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("0"), dec("12000"), nil)

	assertAccountRB(t, db, "1104-01", dec("12000"), dec("0"))
	assertLedgerRB(t, db, 1, dec("0"), dec("12000"))
}

func TestPrepaidAmortized(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertPeriodOpen(t, db, 2, "2026-06-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "prepaid-amort-1")

	a.do(event_types.EventPrepaidCreated.Enum(), payload.PrepaidCreatedPayload{
		AccountID:        "1104-01",
		ExpenseAccountID: "5201-01",
		LedgerID:         1,
		Name:             "人壽保險費 2026",
		TotalAmount:      dec("12000"),
		Periods:          12,
		StartDate:        "2026-05-01",
	})
	prepaidID := queryLastPrepaidID(t, db)

	a.do(event_types.EventPrepaidAmortized.Enum(), payload.PrepaidAmortizedPayload{
		PrepaidID:  prepaidID,
		PeriodDate: "2026-06",
	})

	// AmortizationAmount(12000, 12, 0, 0) = 12000/12 truncate(6) = 1000
	wantAmt := dec("1000")
	txn := queryLastTxn(t, db)
	if txn.Date != "2026-06-01" {
		t.Errorf("Date: want 2026-06-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 5201-01 1000, Cr 1104-01 1000 [OPERATING]
	assertEntry(t, entries, 0, "5201-01", nil, wantAmt, dec("0"), nil)
	assertEntry(t, entries, 1, "1104-01", nil, dec("0"), wantAmt, cfPtr("OPERATING"))

	assertAccountRB(t, db, "1104-01", dec("12000"), wantAmt)
	assertAccountRB(t, db, "5201-01", wantAmt, dec("0"))
}

func TestPrepaidDisposed(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertPeriodOpen(t, db, 2, "2026-06-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "prepaid-disp-1")

	a.do(event_types.EventPrepaidCreated.Enum(), payload.PrepaidCreatedPayload{
		AccountID:        "1104-01",
		ExpenseAccountID: "5201-01",
		LedgerID:         1,
		Name:             "人壽保險費 2026",
		TotalAmount:      dec("12000"),
		Periods:          12,
		StartDate:        "2026-05-01",
	})
	prepaidID := queryLastPrepaidID(t, db)

	a.do(event_types.EventPrepaidAmortized.Enum(), payload.PrepaidAmortizedPayload{
		PrepaidID:  prepaidID,
		PeriodDate: "2026-06",
	})

	// dispose remaining 11000 (12000 - 1000)
	a.do(event_types.EventPrepaidDisposed.Enum(), payload.PrepaidDisposedPayload{
		PrepaidID:    prepaidID,
		DisposalDate: "2026-06-15",
	})

	remaining := dec("11000")
	txn := queryLastTxn(t, db)
	if txn.Date != "2026-06-15" {
		t.Errorf("Date: want 2026-06-15, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 5201-01 11000, Cr 1104-01 11000 [OPERATING]
	assertEntry(t, entries, 0, "5201-01", nil, remaining, dec("0"), nil)
	assertEntry(t, entries, 1, "1104-01", nil, dec("0"), remaining, cfPtr("OPERATING"))

	assertAccountRB(t, db, "1104-01", dec("12000"), dec("12000"))
	assertAccountRB(t, db, "5201-01", dec("12000"), dec("0"))
}

func TestAssetPurchased_Cash(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "asset-cash-1")
	ledgerID := int64(1)
	a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
		Name:                         "辦公電腦",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("100000"),
		ResidualValue:                dec("0"),
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
		LedgerID:                     &ledgerID,
		PurchaseDate:                 "2026-05-01",
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-01" {
		t.Errorf("Date: want 2026-05-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 1201-04 100000 [INVESTING], Cr 1101-02 100000 (ledger=1)
	assertEntry(t, entries, 0, "1201-04", nil, dec("100000"), dec("0"), cfPtr("INVESTING"))
	assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("0"), dec("100000"), nil)

	assertAccountRB(t, db, "1201-04", dec("100000"), dec("0"))
	assertLedgerRB(t, db, 1, dec("0"), dec("100000"))
}

func TestAssetPurchased_Lease(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "asset-lease-1")
	a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
		Name:                         "租賃住宅",
		AssetAccountID:               "1202-01",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("240000"),
		ResidualValue:                dec("0"),
		UsefulLifeMonths:             24,
		PaymentType:                  enums.AssetPaymentTypeLease.Enum(),
		LiabilityAccountID:           "2202-02",
		PurchaseDate:                 "2026-05-01",
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-01" {
		t.Errorf("Date: want 2026-05-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 1202-01 240000 (no CF), Cr 2202-02 240000 (no CF)
	assertEntry(t, entries, 0, "1202-01", nil, dec("240000"), dec("0"), nil)
	assertEntry(t, entries, 1, "2202-02", nil, dec("0"), dec("240000"), nil)

	assertAccountRB(t, db, "1202-01", dec("240000"), dec("0"))
	assertAccountRB(t, db, "2202-02", dec("0"), dec("240000"))
}

func TestAssetDepreciated(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertPeriodOpen(t, db, 2, "2026-06-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "asset-depr-1")
	ledgerID := int64(1)
	a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
		Name:                         "辦公電腦",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("120000"),
		ResidualValue:                dec("0"),
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
		LedgerID:                     &ledgerID,
		PurchaseDate:                 "2026-05-01",
	})
	assetID := queryLastAssetID(t, db)

	a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{
		AssetID:    assetID,
		PeriodDate: "2026-06",
	})

	// DepreciationAmount(120000, 0, 60, 0, 0) = 120000/60 = 2000
	wantAmt := dec("2000")
	txn := queryLastTxn(t, db)
	if txn.Date != "2026-06-01" {
		t.Errorf("Date: want 2026-06-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 5501-03, Cr 1201-99 [OPERATING]
	assertEntry(t, entries, 0, "5501-03", nil, wantAmt, dec("0"), nil)
	assertEntry(t, entries, 1, "1201-99", nil, dec("0"), wantAmt, cfPtr("OPERATING"))

	assertAccountRB(t, db, "1201-99", dec("0"), wantAmt)
	assertAccountRB(t, db, "5501-03", wantAmt, dec("0"))
}

func TestAssetDisposed_Gain(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertPeriodOpen(t, db, 2, "2026-06-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "asset-gain-1")
	ledgerID := int64(1)

	a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
		Name:                         "辦公電腦",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("120000"),
		ResidualValue:                dec("0"),
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
		LedgerID:                     &ledgerID,
		PurchaseDate:                 "2026-05-01",
	})
	assetID := queryLastAssetID(t, db)

	// 2 depreciation periods → totalDepreciated = 4000
	a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetID: assetID, PeriodDate: "2026-05"})
	a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetID: assetID, PeriodDate: "2026-06"})

	// bookValue = 120000 - 4000 = 116000; proceeds = 124000; gain = 8000
	proceeds := int64(1)
	a.do(event_types.EventAssetDisposed.Enum(), payload.AssetDisposedPayload{
		AssetID:          assetID,
		DisposalDate:     "2026-06-30",
		Proceeds:         dec("124000"),
		ProceedsLedgerID: &proceeds,
		GainAccountID:    "4205",
		LossAccountID:    "5601",
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-06-30" {
		t.Errorf("Date: want 2026-06-30, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	// Dr 1201-99 4000 [INVESTING], Cr 1201-04 120000 [INVESTING]
	// Dr 1101-02 124000 (ledger=1), Cr 4205 8000 [INVESTING]
	assertEntry(t, entries, 0, "1201-99", nil, dec("4000"), dec("0"), cfPtr("INVESTING"))
	assertEntry(t, entries, 1, "1201-04", nil, dec("0"), dec("120000"), cfPtr("INVESTING"))
	assertEntry(t, entries, 2, "1101-02", lidPtr(1), dec("124000"), dec("0"), nil)
	assertEntry(t, entries, 3, "4205", nil, dec("0"), dec("8000"), cfPtr("INVESTING"))

	assertAccountRB(t, db, "1201-04", dec("120000"), dec("120000"))
	assertAccountRB(t, db, "1201-99", dec("4000"), dec("4000"))
	assertAccountRB(t, db, "4205", dec("0"), dec("8000"))
	assertLedgerRB(t, db, 1, dec("124000"), dec("120000"))
}

func TestAssetDisposed_Loss(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertPeriodOpen(t, db, 2, "2026-06-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "asset-loss-1")
	ledgerID := int64(1)

	a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
		Name:                         "辦公電腦",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("120000"),
		ResidualValue:                dec("0"),
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
		LedgerID:                     &ledgerID,
		PurchaseDate:                 "2026-05-01",
	})
	assetID := queryLastAssetID(t, db)

	a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetID: assetID, PeriodDate: "2026-05"})
	a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetID: assetID, PeriodDate: "2026-06"})

	// bookValue = 116000; proceeds = 100000; loss = 16000
	proceeds := int64(1)
	a.do(event_types.EventAssetDisposed.Enum(), payload.AssetDisposedPayload{
		AssetID:          assetID,
		DisposalDate:     "2026-06-30",
		Proceeds:         dec("100000"),
		ProceedsLedgerID: &proceeds,
		GainAccountID:    "4205",
		LossAccountID:    "5601",
	})

	txn := queryLastTxn(t, db)
	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	// Dr 1201-99 4000 [INVESTING], Cr 1201-04 120000 [INVESTING]
	// Dr 1101-02 100000 (ledger=1), Dr 5601 16000 [INVESTING]
	assertEntry(t, entries, 0, "1201-99", nil, dec("4000"), dec("0"), cfPtr("INVESTING"))
	assertEntry(t, entries, 1, "1201-04", nil, dec("0"), dec("120000"), cfPtr("INVESTING"))
	assertEntry(t, entries, 2, "1101-02", lidPtr(1), dec("100000"), dec("0"), nil)
	assertEntry(t, entries, 3, "5601", nil, dec("16000"), dec("0"), cfPtr("INVESTING"))

	assertAccountRB(t, db, "5601", dec("16000"), dec("0"))
	assertLedgerRB(t, db, 1, dec("100000"), dec("120000"))
}

func TestInstallmentCreated_Free(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertSysAccounts(t, db)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertLedger(t, db, 1, "LOAN", "2201-01")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "install-1")
	a.do(event_types.EventInstallmentCreated.Enum(), payload.InstallmentCreatedPayload{
		Amount:           dec("10000"),
		InstallmentCount: 12,
		StartDate:        "2026-05-01",
		InterestType:     enums.InterestTypeFree.Enum(),
		AccountId:        "1201-04",
		LedgerId:         1,
		Memo:             "辦公電腦分期",
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-01" {
		t.Errorf("Date: want 2026-05-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// FREE: Dr 1201-04 10000, Cr 2201-01 10000 (ledger=1)
	assertEntry(t, entries, 0, "1201-04", nil, dec("10000"), dec("0"), nil)
	assertEntry(t, entries, 1, "2201-01", lidPtr(1), dec("0"), dec("10000"), nil)

	assertAccountRB(t, db, "1201-04", dec("10000"), dec("0"))
	assertLedgerRB(t, db, 1, dec("0"), dec("10000"))
}

func TestInstallmentPeriodPaid_Free(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertSysAccounts(t, db)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertPeriodOpen(t, db, 2, "2026-06-01")
	insertLedger(t, db, 1, "LOAN", "2201-01")
	insertLedger(t, db, 2, "BANK_ACCOUNT", "1101-02")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "install-paid-1")
	a.do(event_types.EventInstallmentCreated.Enum(), payload.InstallmentCreatedPayload{
		Amount:           dec("10000"),
		InstallmentCount: 12,
		StartDate:        "2026-05-01",
		InterestType:     enums.InterestTypeFree.Enum(),
		AccountId:        "1201-04",
		LedgerId:         1,
		Memo:             "辦公電腦分期",
	})
	installmentID := queryLastInstallmentID(t, db)

	a.do(event_types.EventInstallmentPeriodPaid.Enum(), payload.InstallmentPeriodPaidPayload{
		InstallmentId:  installmentID,
		Period:         1,
		PaidDate:       "2026-06-01",
		PaidLedgerUuid: testLedgerUUID(2),
	})

	// Period 1: 10000/12 truncate(6) = 833.333333
	wantAmt := dec("833.333333")
	txn := queryLastTxn(t, db)
	if txn.Date != "2026-06-01" {
		t.Errorf("Date: want 2026-06-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Dr 2201-01 833.333333 (ledger=1) [FINANCING], Cr 1101-02 833.333333 (ledger=2)
	assertEntry(t, entries, 0, "2201-01", lidPtr(1), wantAmt, dec("0"), cfPtr("FINANCING"))
	assertEntry(t, entries, 1, "1101-02", lidPtr(2), dec("0"), wantAmt, nil)

	assertLedgerRB(t, db, 1, wantAmt, dec("10000"))
	assertLedgerRB(t, db, 2, dec("0"), wantAmt)
}

func TestInvestmentBought(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertAssetTypeConfig(t, db)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")
	insertInvestment(t, db, 1, "1102-01", "STOCK", "AVG", "FVTPL")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "inv-buy-1")
	a.do(event_types.EventInvestmentBought.Enum(), payload.InvestmentBoughtPayload{
		InvestmentUUID: testInvestmentUUID(1),
		Date:           "2026-05-01",
		Quantity:       dec("10"),
		UnitPrice:      dec("100"),
		ExchangeRate:   dec("1"),
		Fee:            dec("5"),
		Tax:            dec("0"),
		LedgerId:       1,
	})

	// cost = 10×100×1 = 1000; totalCost = 1005 (fee=5)
	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-01" {
		t.Errorf("Date: want 2026-05-01, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	// Dr 1102-01 1000 [INVESTING], Cr 1101-02 1005 (ledger=1), Dr 5402-01 5 [INVESTING]
	assertEntry(t, entries, 0, "1102-01", nil, dec("1000"), dec("0"), cfPtr("INVESTING"))
	assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("0"), dec("1005"), nil)
	assertEntry(t, entries, 2, "5402-01", nil, dec("5"), dec("0"), cfPtr("INVESTING"))

	assertAccountRB(t, db, "1102-01", dec("1000"), dec("0"))
	assertAccountRB(t, db, "5402-01", dec("5"), dec("0"))
	assertLedgerRB(t, db, 1, dec("0"), dec("1005"))
}

func TestInvestmentSold_Gain(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertAssetTypeConfig(t, db)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")
	insertInvestment(t, db, 1, "1102-01", "STOCK", "AVG", "FVTPL")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "inv-sell-1")

	// Buy 10 @ 100 (no fee to keep costBasis simple)
	a.do(event_types.EventInvestmentBought.Enum(), payload.InvestmentBoughtPayload{
		InvestmentUUID: testInvestmentUUID(1),
		Date:           "2026-05-01",
		Quantity:       dec("10"),
		UnitPrice:      dec("100"),
		ExchangeRate:   dec("1"),
		Fee:            dec("0"),
		Tax:            dec("0"),
		LedgerId:       1,
	})

	// Sell 10 @ 120; fee=5; netProceeds = 10×120-5 = 1195; costBasis = 1000; gain = 200
	a.do(event_types.EventInvestmentSold.Enum(), payload.InvestmentSoldPayload{
		InvestmentUUID: testInvestmentUUID(1),
		Date:           "2026-05-15",
		Quantity:       dec("10"),
		UnitPrice:      dec("120"),
		ExchangeRate:   dec("1"),
		Fee:            dec("5"),
		Tax:            dec("0"),
		LedgerId:       1,
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-15" {
		t.Errorf("Date: want 2026-05-15, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	// Dr 1101-02 1195 (ledger=1), Cr 1102-01 1000 [INVESTING]
	// Dr 5402-01 5 [INVESTING], Cr 4203-01 200 [INVESTING]
	assertEntry(t, entries, 0, "1101-02", lidPtr(1), dec("1195"), dec("0"), nil)
	assertEntry(t, entries, 1, "1102-01", nil, dec("0"), dec("1000"), cfPtr("INVESTING"))
	assertEntry(t, entries, 2, "5402-01", nil, dec("5"), dec("0"), cfPtr("INVESTING"))
	assertEntry(t, entries, 3, "4203-01", nil, dec("0"), dec("200"), cfPtr("INVESTING"))

	assertAccountRB(t, db, "1102-01", dec("1000"), dec("1000"))
	assertAccountRB(t, db, "4203-01", dec("0"), dec("200"))
	assertLedgerRB(t, db, 1, dec("1195"), dec("1000"))
}

func TestDividendReceived(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertDividendAccounts(t, db)
	insertAssetTypeConfig(t, db)
	insertPeriodOpen(t, db, 1, "2026-05-01")
	insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")
	insertInvestment(t, db, 1, "1102-01", "STOCK", "AVG", "FVTPL")

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "div-1")

	// Must have a position before receiving dividends
	a.do(event_types.EventInvestmentBought.Enum(), payload.InvestmentBoughtPayload{
		InvestmentUUID: testInvestmentUUID(1),
		Date:           "2026-05-01",
		Quantity:       dec("10"),
		UnitPrice:      dec("100"),
		ExchangeRate:   dec("1"),
		Fee:            dec("0"),
		Tax:            dec("0"),
		LedgerId:       1,
	})

	a.do(event_types.EventDividendReceived.Enum(), payload.DividendReceivedPayload{
		InvestmentUUID: testInvestmentUUID(1),
		Date:           "2026-05-15",
		Amount:         dec("1000"),
		ExchangeRate:   dec("1"),
		WithholdingTax: dec("0"),
		Ratio:          dec("0"),
		LedgerId:       1,
	})

	txn := queryLastTxn(t, db)
	if txn.Date != "2026-05-15" {
		t.Errorf("Date: want 2026-05-15, got %s", txn.Date)
	}

	entries := queryEntries(t, db, txn.TxnID)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Cr 4210 1000; Dr 1101-02 1000 (netAmountTWD = 1000-0 = 1000)
	assertEntry(t, entries, 0, "4210", nil, dec("0"), dec("1000"), nil)
	assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("1000"), dec("0"), nil)

	assertAccountRB(t, db, "4210", dec("0"), dec("1000"))
	assertLedgerRB(t, db, 1, dec("1000"), dec("1000"))
}

func TestPeriodAnnualClosed(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertSysAccounts(t, db)

	// Raw income transaction in 2025 (bypasses pipeline — no open period needed)
	insertIncomeTransaction(t, db, 9001, "2025-06-15", "4101-01", 50000)

	// All 12 months of 2025 CLOSED (IDs 1–12), annual period OPEN (ID 13)
	insertAllMonthsClosed(t, db, 2025, 1)
	insertAnnualPeriodOpen(t, db, 13, 2025)

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "period-annual-1")
	a.do(event_types.EventPeriodAnnualClosed.Enum(), payload.PeriodAnnualClosedPayload{
		ClosingId: 13,
		ClosedAt:  "2025-12-31",
	})

	// Two new transactions should have been created
	var txns []struct {
		TxnID   int64  `db:"txn_id"`
		TxnDate string `db:"txn_date"`
	}
	err := db.SelectContext(testCtx(), &txns,
		`SELECT txn_id, txn_date FROM transactions
		 WHERE merchant_id=? AND status='ACTIVE' AND txn_id != 9001
		 ORDER BY txn_id`,
		testMID)
	if err != nil {
		t.Fatalf("query annual close txns: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 generated txns, got %d", len(txns))
	}

	closingTxnID := txns[0].TxnID
	openingTxnID := txns[1].TxnID

	if txns[0].TxnDate != "2025-12-31" {
		t.Errorf("closing txn date: want 2025-12-31, got %s", txns[0].TxnDate)
	}
	if txns[1].TxnDate != "2026-01-01" {
		t.Errorf("opening txn date: want 2026-01-01, got %s", txns[1].TxnDate)
	}

	// Closing txn: Dr 4101-01 50000 (close income), Cr 3102-01 50000 (net income)
	closingEntries := queryEntries(t, db, closingTxnID)
	incomeDrFound, netIncomeCrFound := false, false
	for _, e := range closingEntries {
		if e.AccountID == "4101-01" && e.Debit.Equal(dec("50000")) {
			incomeDrFound = true
		}
		if e.AccountID == "3102-01" && e.Credit.Equal(dec("50000")) {
			netIncomeCrFound = true
		}
	}
	if !incomeDrFound {
		t.Error("closing txn: missing Dr 4101-01 50000")
	}
	if !netIncomeCrFound {
		t.Error("closing txn: missing Cr 3102-01 50000")
	}

	// Opening txn: Dr 3102-01 50000, Cr 3101-01 50000
	openingEntries := queryEntries(t, db, openingTxnID)
	netIncomeDrFound, openingCrFound := false, false
	for _, e := range openingEntries {
		if e.AccountID == "3102-01" && e.Debit.Equal(dec("50000")) {
			netIncomeDrFound = true
		}
		if e.AccountID == "3101-01" && e.Credit.Equal(dec("50000")) {
			openingCrFound = true
		}
	}
	if !netIncomeDrFound {
		t.Error("opening txn: missing Dr 3102-01 50000")
	}
	if !openingCrFound {
		t.Error("opening txn: missing Cr 3101-01 50000")
	}

	// RB: 3101-01 C=50000 from the opening transfer
	assertAccountRB(t, db, "3101-01", dec("0"), dec("50000"))
}

func TestPeriodAnnualReopened(t *testing.T) {
	db := testutil.NewTestDB(t)
	insertSysAccounts(t, db)

	insertIncomeTransaction(t, db, 9001, "2025-06-15", "4101-01", 50000)
	insertAllMonthsClosed(t, db, 2025, 1)
	insertAnnualPeriodOpen(t, db, 13, 2025)

	svc := newSvc(db)
	a := newAppender(t, svc, testCtx(), "period-reopen-1")

	a.do(event_types.EventPeriodAnnualClosed.Enum(), payload.PeriodAnnualClosedPayload{
		ClosingId: 13,
		ClosedAt:  "2025-12-31",
	})

	a.do(event_types.EventPeriodAnnualReopened.Enum(), payload.PeriodAnnualReopenedPayload{
		ClosingId:  13,
		Reason:     "correction",
		ReopenedAt: "2026-01-05",
	})

	// Expect 4 transactions generated: 2 from close + 2 reverse from reopen
	var txns []struct {
		TxnID int64  `db:"txn_id"`
		RefID *int64 `db:"ref_txn_id"`
	}
	err := db.SelectContext(testCtx(), &txns,
		`SELECT txn_id, ref_txn_id FROM transactions
		 WHERE merchant_id=? AND txn_id != 9001 ORDER BY txn_id`,
		testMID)
	if err != nil {
		t.Fatalf("query reopen txns: %v", err)
	}
	if len(txns) < 4 {
		t.Fatalf("expected at least 4 generated txns, got %d", len(txns))
	}

	// The last 2 txns (from reopen) must have ref_txn_id pointing back to the close txns
	reverses := txns[len(txns)-2:]
	for _, r := range reverses {
		if r.RefID == nil {
			t.Errorf("reopen txn %d: expected ref_txn_id, got nil", r.TxnID)
		}
	}
}
