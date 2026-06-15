package query_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
)

// ── 科目常數（來自 accounts.sql seeds）────────────────────────────────────────

const (
	testMerchantID int64 = 1 // 與 seeds 的 merchant_id 一致

	// 葉科目
	sumAcctCash    = "1101-01" // 手頭現金（ASSET, DEBIT normal, is_summary=0）
	sumAcctExpense = "5101-01" // 房租費用（EXPENSE, DEBIT normal, is_summary=0）
	// 直接父科目
	sumAcctCash1101 = "1101" // 現金及約當現金（is_summary=1）
	sumAcctExp5101  = "5101" // 居住費用（is_summary=1）
	// 祖父科目
	sumAcctCash110 = "110" // 流動資產（is_summary=1）
	sumAcctExp510  = "510" // 生活費用（is_summary=1）
)

// testCtx returns a context pre-loaded with testMerchantID.
func testCtx() context.Context {
	return context.WithValue(context.Background(), ctxkey.MerchantID, testMerchantID)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func sumInsertPeriod(t *testing.T, db *sqlx.DB, id int64, periodType, start, end string, closed bool) {
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
		t.Fatalf("sumInsertPeriod id=%d: %v", id, err)
	}
}

// sumInsertTxn 插入借貸分錄（以 account_id 分錄，無 ledger_id）。
func sumInsertTxn(t *testing.T, db *sqlx.DB, txnID int64, date, debitAcct, creditAcct string, amount float64) {
	t.Helper()
	sumInsertTxnWithCF(t, db, txnID, date, debitAcct, creditAcct, amount, nil, nil)
}

// sumInsertTxnWithCF inserts a transaction with two journal entries and
// populates entry_cf_categories for non-nil CF values so report queries can JOIN it.
func sumInsertTxnWithCF(t *testing.T, db *sqlx.DB, txnID int64, date, debitAcct, creditAcct string, amount float64, debitCF, creditCF *string) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO transactions (txn_id, merchant_id, txn_date, description, total_amount, version) VALUES (?, ?, ?, '測試', ?, 1)`,
		txnID, testMerchantID, date, amount)
	if err != nil {
		t.Fatalf("sumInsertTxnWithCF id=%d: %v", txnID, err)
	}
	debitUUID := fmt.Sprintf("test-entry-%d-0", txnID)
	creditUUID := fmt.Sprintf("test-entry-%d-1", txnID)
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries (txn_id, merchant_id, account_id, entry_uuid, debit, credit) VALUES (?, ?, ?, ?, ?, 0)`,
		txnID, testMerchantID, debitAcct, debitUUID, amount)
	if err != nil {
		t.Fatalf("sumInsertTxnWithCF debit: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO journal_entries (txn_id, merchant_id, account_id, entry_uuid, debit, credit) VALUES (?, ?, ?, ?, 0, ?)`,
		txnID, testMerchantID, creditAcct, creditUUID, amount)
	if err != nil {
		t.Fatalf("sumInsertTxnWithCF credit: %v", err)
	}
	if debitCF != nil {
		_, err = db.ExecContext(ctx,
			`INSERT INTO entry_cf_categories (entry_uuid, merchant_id, cf_category, is_confirmed) VALUES (?, ?, ?, 1)`,
			debitUUID, testMerchantID, *debitCF)
		if err != nil {
			t.Fatalf("sumInsertTxnWithCF debit ecc: %v", err)
		}
	}
	if creditCF != nil {
		_, err = db.ExecContext(ctx,
			`INSERT INTO entry_cf_categories (entry_uuid, merchant_id, cf_category, is_confirmed) VALUES (?, ?, ?, 1)`,
			creditUUID, testMerchantID, *creditCF)
		if err != nil {
			t.Fatalf("sumInsertTxnWithCF credit ecc: %v", err)
		}
	}
}

// findBalance 在結果中找指定科目的餘額記錄。
func findBalance(t *testing.T, rows []projection.AccountBalance, accountId string) projection.AccountBalance {
	t.Helper()
	for _, r := range rows {
		if r.AccountId == accountId {
			return r
		}
	}
	t.Fatalf("account %s not found in result (%d rows)", accountId, len(rows))
	return projection.AccountBalance{}
}

// assertBalance 驗證指定科目的借貸加總。
func assertBalance(t *testing.T, rows []projection.AccountBalance, accountId string, wantDebit, wantCredit float64) {
	t.Helper()
	r := findBalance(t, rows, accountId)
	gotDebit, _ := r.DebitTotal.Float64()
	gotCredit, _ := r.CreditTotal.Float64()
	if gotDebit != wantDebit {
		t.Errorf("account %s debit_total: got %.2f, want %.2f", accountId, gotDebit, wantDebit)
	}
	if gotCredit != wantCredit {
		t.Errorf("account %s credit_total: got %.2f, want %.2f", accountId, gotCredit, wantCredit)
	}
}
