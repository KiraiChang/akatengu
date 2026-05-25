package query_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/handler/response/model"
	"akatengu/internal/repos/query"
	"akatengu/internal/testutil"
)

// insertActiveTxnForAnalysis inserts a transaction (status=ACTIVE) with two journal entries.
func insertActiveTxnForAnalysis(t *testing.T, db *sqlx.DB, txnID int64, date, debitAcct, creditAcct string, amount float64) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO transactions (txn_id, merchant_id, txn_date, description, total_amount, status, version)
		 VALUES (?, ?, ?, 'test', ?, 'ACTIVE', 1)`,
		txnID, testMerchantID, date, amount)
	if err != nil {
		t.Fatalf("insertActiveTxnForAnalysis txn_id=%d: %v", txnID, err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries (txn_id, merchant_id, account_id, debit, credit) VALUES (?, ?, ?, ?, 0)`,
		txnID, testMerchantID, debitAcct, amount)
	if err != nil {
		t.Fatalf("insertActiveTxnForAnalysis debit: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries (txn_id, merchant_id, account_id, debit, credit) VALUES (?, ?, ?, 0, ?)`,
		txnID, testMerchantID, creditAcct, amount)
	if err != nil {
		t.Fatalf("insertActiveTxnForAnalysis credit: %v", err)
	}
}

// TestGetAccountDirectChildrenWithBalance verifies children are returned with running balances.
func TestGetAccountDirectChildrenWithBalance(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	// Insert running balances for children of 1101.
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO account_running_balances (account_id, merchant_id, debit_total, credit_total)
		 VALUES (?, ?, ?, ?)`,
		"1101-01", testMerchantID, 3000.0, 0.0)
	if err != nil {
		t.Fatalf("insert running balance 1101-01: %v", err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO account_running_balances (account_id, merchant_id, debit_total, credit_total)
		 VALUES (?, ?, ?, ?)`,
		"1101-02", testMerchantID, 5000.0, 500.0)
	if err != nil {
		t.Fatalf("insert running balance 1101-02: %v", err)
	}

	repo := query.NewAccountAnalysisRepo(db)
	children, err := repo.GetAccountDirectChildrenWithBalance(ctx, "1101")
	if err != nil {
		t.Fatalf("GetAccountDirectChildrenWithBalance: %v", err)
	}
	if len(children) == 0 {
		t.Fatal("expected children, got none")
	}

	found0101, found0102 := false, false
	for _, c := range children {
		switch c.AccountID {
		case "1101-01":
			found0101 = true
			d, _ := c.DebitTotal.Float64()
			cr, _ := c.CreditTotal.Float64()
			if d != 3000.0 || cr != 0.0 {
				t.Errorf("1101-01: debit=%.2f credit=%.2f, want 3000.00/0.00", d, cr)
			}
		case "1101-02":
			found0102 = true
			d, _ := c.DebitTotal.Float64()
			cr, _ := c.CreditTotal.Float64()
			if d != 5000.0 || cr != 500.0 {
				t.Errorf("1101-02: debit=%.2f credit=%.2f, want 5000.00/500.00", d, cr)
			}
		}
	}
	if !found0101 {
		t.Error("1101-01 not found in children")
	}
	if !found0102 {
		t.Error("1101-02 not found in children")
	}
}

// TestGetAccountDirectChildrenWithBalance_NoBalance verifies zero balance when no running balance exists.
func TestGetAccountDirectChildrenWithBalance_NoBalance(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	repo := query.NewAccountAnalysisRepo(db)
	children, err := repo.GetAccountDirectChildrenWithBalance(ctx, "1101")
	if err != nil {
		t.Fatalf("GetAccountDirectChildrenWithBalance: %v", err)
	}
	for _, c := range children {
		if !c.DebitTotal.IsZero() || !c.CreditTotal.IsZero() {
			t.Errorf("%s: expected zero balance, got debit=%s credit=%s", c.AccountID, c.DebitTotal, c.CreditTotal)
		}
	}
}

// TestGetAccountJournalEntriesPaged verifies pagination and date range filtering.
func TestGetAccountJournalEntriesPaged(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	insertActiveTxnForAnalysis(t, db, 1001, "2026-01-10", "1101-01", "5101-01", 1000.0)
	insertActiveTxnForAnalysis(t, db, 1002, "2026-02-15", "1101-01", "5101-01", 2000.0)
	insertActiveTxnForAnalysis(t, db, 1003, "2026-03-20", "1101-01", "5101-01", 500.0)

	repo := query.NewAccountAnalysisRepo(db)

	// Full range: all 3 entries.
	req := model.PaginationParams{Page: 1, PageSize: 50}
	req.SetDefaults()
	entries, total, err := repo.GetAccountJournalEntriesPaged(ctx, "1101-01", "2026-01-01", "2026-03-31", req)
	if err != nil {
		t.Fatalf("GetAccountJournalEntriesPaged: %v", err)
	}
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(entries) != 3 {
		t.Errorf("entries len: got %d, want 3", len(entries))
	}

	// Restrict to January only.
	req2 := model.PaginationParams{Page: 1, PageSize: 50}
	req2.SetDefaults()
	entries2, total2, err := repo.GetAccountJournalEntriesPaged(ctx, "1101-01", "2026-01-01", "2026-01-31", req2)
	if err != nil {
		t.Fatalf("GetAccountJournalEntriesPaged (Jan): %v", err)
	}
	if total2 != 1 {
		t.Errorf("Jan total: got %d, want 1", total2)
	}
	got, _ := entries2[0].Debit.Float64()
	if got != 1000.0 {
		t.Errorf("Jan debit: got %.2f, want 1000.00", got)
	}

	// Page 2, page_size 2 → 1 entry (oldest entry).
	req3 := model.PaginationParams{Page: 2, PageSize: 2}
	req3.SetDefaults()
	entries3, total3, err := repo.GetAccountJournalEntriesPaged(ctx, "1101-01", "2026-01-01", "2026-03-31", req3)
	if err != nil {
		t.Fatalf("GetAccountJournalEntriesPaged (page2): %v", err)
	}
	if total3 != 3 {
		t.Errorf("page2 total: got %d, want 3", total3)
	}
	if len(entries3) != 1 {
		t.Errorf("page2 entries len: got %d, want 1", len(entries3))
	}
}

// TestGetAccountMonthlyBalances verifies monthly grouping and has_snapshot flag.
func TestGetAccountMonthlyBalances(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	// January: two transactions (total debit 1500).
	insertActiveTxnForAnalysis(t, db, 2001, "2026-01-10", "1101-01", "5101-01", 1000.0)
	insertActiveTxnForAnalysis(t, db, 2002, "2026-01-25", "1101-01", "5101-01", 500.0)
	// February: one transaction.
	insertActiveTxnForAnalysis(t, db, 2003, "2026-02-14", "1101-01", "5101-01", 2000.0)
	// March: no transactions for 1101-01.

	// Close January.
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO period_closings (closing_id, merchant_id, period_type, period_start, period_end, status, closed_at)
		 VALUES (?, ?, 'MONTHLY', ?, ?, 'CLOSED', ?)`,
		9001, testMerchantID, "2026-01-01", "2026-01-31", "2026-02-01")
	if err != nil {
		t.Fatalf("insert period_closing: %v", err)
	}

	repo := query.NewAccountAnalysisRepo(db)
	balances, err := repo.GetAccountMonthlyBalances(ctx, "1101-01", "2026-01", "2026-03")
	if err != nil {
		t.Fatalf("GetAccountMonthlyBalances: %v", err)
	}
	if len(balances) != 3 {
		t.Fatalf("expected 3 months, got %d", len(balances))
	}

	jan := balances[0]
	if jan.Month != "2026-01" {
		t.Errorf("month[0]: got %s, want 2026-01", jan.Month)
	}
	janDebit, _ := jan.DebitTotal.Float64()
	if janDebit != 1500.0 {
		t.Errorf("Jan debit: got %.2f, want 1500.00", janDebit)
	}
	if !jan.HasSnapshot {
		t.Error("Jan should have HasSnapshot=true (closed period)")
	}

	feb := balances[1]
	if feb.Month != "2026-02" {
		t.Errorf("month[1]: got %s, want 2026-02", feb.Month)
	}
	febDebit, _ := feb.DebitTotal.Float64()
	if febDebit != 2000.0 {
		t.Errorf("Feb debit: got %.2f, want 2000.00", febDebit)
	}
	if feb.HasSnapshot {
		t.Error("Feb should have HasSnapshot=false (open period)")
	}

	mar := balances[2]
	if mar.Month != "2026-03" {
		t.Errorf("month[2]: got %s, want 2026-03", mar.Month)
	}
	if !mar.DebitTotal.IsZero() || !mar.CreditTotal.IsZero() {
		t.Errorf("Mar: expected zero, got debit=%s credit=%s", mar.DebitTotal, mar.CreditTotal)
	}
	if mar.HasSnapshot {
		t.Error("Mar should have HasSnapshot=false")
	}
}

// TestGetAccountMonthlyBalances_InvalidMonth verifies error on bad month format.
func TestGetAccountMonthlyBalances_InvalidMonth(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := testCtx()

	repo := query.NewAccountAnalysisRepo(db)
	_, err := repo.GetAccountMonthlyBalances(ctx, "1101-01", "2026/01", "2026-03")
	if err == nil {
		t.Error("expected error for invalid from month format")
	}
}
