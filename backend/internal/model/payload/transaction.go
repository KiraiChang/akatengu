package payload

import (
	"akatengu/internal/enums"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// TransactionEntryPayload 範例
// { "account_id": "5220", "ledger_id": null, "debit": 1250, "credit": 0, "cash_flow_category": "OPERATING" }
type TransactionEntryPayload struct {
	AccountId        string                  `json:"account_id"`
	LedgerId         *int64                  `json:"ledger_id"`
	Debit            decimal.Decimal         `json:"debit"`
	Credit           decimal.Decimal         `json:"credit"`
	CashFlowCategory *enums.CashFlowCategory `json:"cash_flow_category"`
}

// TransactionCreatedPayload 範例
// transaction.created
//
//	{
//	  "txn_id": 42,
//	  "txn_date": "2026-03-09",
//	  "description": "家樂福採購",
//	  "total_amount": 1250,
//	  "entries": [
//	    { "account_id": "5220",    "ledger_id": null, "debit": 1250, "credit": 0 },
//	    { "account_id": "2100-01", "ledger_id": 2,    "debit": 0,    "credit": 1250 }
//	  ]
//	}
type TransactionCreatedPayload struct {
	TransactionDate string                    `json:"transaction_date"`
	Description     string                    `json:"description"`
	TotalAmount     decimal.Decimal           `json:"total_amount"`
	Currency        string                    `json:"currency"`
	ReceiptNo       *string                   `json:"receipt_no"`
	Note            *string                   `json:"note"`
	RefTxnId        *int64                    `json:"ref_txn_id"`
	Entries         []TransactionEntryPayload `json:"entries"`
}

func (p TransactionCreatedPayload) Validate() error {
	var errs []string

	if p.TransactionDate == "" {
		errs = append(errs, "txn_date is required")
	} else if _, err := time.Parse("2006-01-02", p.TransactionDate); err != nil {
		errs = append(errs, "txn_date must be YYYY-MM-DD")
	}
	if p.Description == "" {
		errs = append(errs, "description is required")
	}
	if !p.TotalAmount.IsPositive() {
		errs = append(errs, "total_amount must be > 0")
	}
	if len(p.Entries) < 2 {
		errs = append(errs, "entries must have at least 2 items")
	}

	// 借貸必須平衡
	var totalDebit, totalCredit decimal.Decimal
	for i, e := range p.Entries {
		if e.AccountId == "" {
			errs = append(errs, fmt.Sprintf("entries[%d].account_id is required", i))
		}
		totalDebit = totalDebit.Add(e.Debit)
		totalCredit = totalCredit.Add(e.Credit)
	}
	if !totalDebit.Sub(totalCredit).IsZero() {
		errs = append(errs, fmt.Sprintf(
			"entries not balanced: debit=%s credit=%s", totalDebit.StringFixed(2), totalCredit.StringFixed(2),
		))
	}

	return joinErrors(errs)
}

// TransactionCorrectedPayload 範例
// transaction.corrected
//
//	{
//	  "original_txn_id": 42,
//	  "void_txn_id": 43,
//	  "correction_txn_id": 44,
//	  "reason": "金額輸入錯誤，應為 1520"
//	}
type TransactionCorrectedPayload struct {
	OriginalTransactionId  int64  `json:"original_txn_id"`
	VoidTransactionId      int64  `json:"void_txn_id"`
	CorrectedTransactionId int64  `json:"corrected_txn_id"`
	Reason                 string `json:"reason"`
}

func (p TransactionCorrectedPayload) Validate() error {
	var errs []string
	if p.OriginalTransactionId <= 0 {
		errs = append(errs, "original_txn_id is required")
	}
	if p.VoidTransactionId <= 0 {
		errs = append(errs, "void_txn_id is required")
	}
	if p.CorrectedTransactionId <= 0 {
		errs = append(errs, "correction_txn_id is required")
	}
	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}

	return joinErrors(errs)
}

// TransactionVoidedPayload 範例
// transaction.voided
//
//	{
//	  "txn_id": 42,
//	  "void_txn_id": 43,
//	  "reason": "重複記帳"
//	}
type TransactionVoidedPayload struct {
	TransactionId     int64  `json:"txn_id"`
	VoidTransactionId int64  `json:"void_txn_id"`
	Reason            string `json:"reason"`
}

func (p TransactionVoidedPayload) Validate() error {
	var errs []string
	if p.TransactionId <= 0 {
		errs = append(errs, "txn_id is required")
	}
	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}

	return joinErrors(errs)
}
