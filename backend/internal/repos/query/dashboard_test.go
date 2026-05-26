package query_test

import (
	"context"
	"testing"

	"akatengu/internal/repos/query"
	"akatengu/internal/testutil"
)

// TestGetDashboardSummary verifies assets, cash balance and zero income/expense when no transactions exist.
func TestGetDashboardSummary(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	// Insert running balance: 1101-01 (ASSET, cash_flow_category=CASH, DEBIT normal balance)
	// balance = 5000 - 1000 = 4000
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO account_running_balances (account_id, merchant_id, debit_total, credit_total)
		 VALUES (?, ?, ?, ?)`,
		"1101-01", testMerchantID, 5000.0, 1000.0)
	if err != nil {
		t.Fatalf("insert running balance: %v", err)
	}

	repo := query.NewDashboardRepo(db)
	today := "2026-05-26"
	summary, err := repo.GetSummary(ctx, today)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}

	if summary.AsOfDate != today {
		t.Errorf("AsOfDate: got %s, want %s", summary.AsOfDate, today)
	}
	if summary.Month != "2026-05" {
		t.Errorf("Month: got %s, want 2026-05", summary.Month)
	}

	totalAssets, _ := summary.TotalAssets.Float64()
	if totalAssets != 4000.0 {
		t.Errorf("TotalAssets: got %.2f, want 4000.00", totalAssets)
	}
	cashBalance, _ := summary.CashBalance.Float64()
	if cashBalance != 4000.0 {
		t.Errorf("CashBalance: got %.2f, want 4000.00", cashBalance)
	}

	// No transactions this month
	if !summary.MonthIncome.IsZero() {
		t.Errorf("MonthIncome: expected zero, got %s", summary.MonthIncome)
	}
	if !summary.MonthExpense.IsZero() {
		t.Errorf("MonthExpense: expected zero, got %s", summary.MonthExpense)
	}
}

// TestGetDashboardMonthlyTrend verifies monthly income/expense grouping and net calculation.
func TestGetDashboardMonthlyTrend(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	// Jan: income 3000 (credit 4101-01), expense 1000 (debit 5101-01)
	insertActiveTxnForAnalysis(t, db, 5001, "2026-01-15", "1101-01", "4101-01", 3000.0)
	insertActiveTxnForAnalysis(t, db, 5002, "2026-01-20", "5101-01", "1101-01", 1000.0)

	// Feb: income 2000 only
	insertActiveTxnForAnalysis(t, db, 5003, "2026-02-10", "1101-01", "4101-01", 2000.0)

	repo := query.NewDashboardRepo(db)
	items, err := repo.GetMonthlyTrend(ctx, "2026-01", "2026-02")
	if err != nil {
		t.Fatalf("GetMonthlyTrend: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 months, got %d", len(items))
	}

	jan := items[0]
	if jan.Month != "2026-01" {
		t.Errorf("month[0]: got %s, want 2026-01", jan.Month)
	}
	janIncome, _ := jan.Income.Float64()
	if janIncome != 3000.0 {
		t.Errorf("Jan income: got %.2f, want 3000.00", janIncome)
	}
	janExpense, _ := jan.Expense.Float64()
	if janExpense != 1000.0 {
		t.Errorf("Jan expense: got %.2f, want 1000.00", janExpense)
	}
	janNet, _ := jan.Net.Float64()
	if janNet != 2000.0 {
		t.Errorf("Jan net: got %.2f, want 2000.00", janNet)
	}

	feb := items[1]
	if feb.Month != "2026-02" {
		t.Errorf("month[1]: got %s, want 2026-02", feb.Month)
	}
	febIncome, _ := feb.Income.Float64()
	if febIncome != 2000.0 {
		t.Errorf("Feb income: got %.2f, want 2000.00", febIncome)
	}
	if !feb.Expense.IsZero() {
		t.Errorf("Feb expense: expected zero, got %s", feb.Expense)
	}
	febNet, _ := feb.Net.Float64()
	if febNet != 2000.0 {
		t.Errorf("Feb net: got %.2f, want 2000.00", febNet)
	}
}

// TestGetDashboardLedgerBalances verifies ledger balance includes name/institution/type and correct balance sign.
func TestGetDashboardLedgerBalances(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	// Insert a ledger account (BANK_ACCOUNT, links to ASSET account 1101-01)
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO ledger_accounts (ledger_id, merchant_id, account_id, institution, name, type, currency, is_active, version)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		8001, testMerchantID, "1101-01", "Test Bank", "Savings Account", "BANK_ACCOUNT", "TWD", 1, 1)
	if err != nil {
		t.Fatalf("insert ledger_account: %v", err)
	}

	// Insert running balance: debit=10000, credit=3000 → DEBIT normal balance → balance=7000
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO ledger_running_balances (ledger_id, merchant_id, debit_total, credit_total)
		 VALUES (?, ?, ?, ?)`,
		8001, testMerchantID, 10000.0, 3000.0)
	if err != nil {
		t.Fatalf("insert ledger_running_balance: %v", err)
	}

	repo := query.NewDashboardRepo(db)
	balances, err := repo.GetLedgerBalances(ctx)
	if err != nil {
		t.Fatalf("GetLedgerBalances: %v", err)
	}

	var found bool
	for _, b := range balances {
		if b.LedgerID != 8001 {
			continue
		}
		found = true
		if b.Name != "Savings Account" {
			t.Errorf("Name: got %s, want Savings Account", b.Name)
		}
		if b.Institution != "Test Bank" {
			t.Errorf("Institution: got %s, want Test Bank", b.Institution)
		}
		if b.Type != "BANK_ACCOUNT" {
			t.Errorf("Type: got %s, want BANK_ACCOUNT", b.Type)
		}
		bal, _ := b.Balance.Float64()
		if bal != 7000.0 {
			t.Errorf("Balance: got %.2f, want 7000.00", bal)
		}
	}
	if !found {
		t.Error("ledger_id=8001 not found in GetLedgerBalances result")
	}
}
