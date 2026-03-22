package payload

import (
	"github.com/shopspring/decimal"
)

// TransactionEntryPayload 範例
// { "account_id": "5220",    "ledger_id": null, "debit": 1250, "credit": 0 },
type TransactionEntryPayload struct {
	AccountId string          `json:"account_id"`
	LedgerId  *int64          `json:"ledger_id"`
	Debit     decimal.Decimal `json:"debit"`
	Credit    decimal.Decimal `json:"credit"`
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
	TransactionDate string          `json:"transaction_date"`
	Description     string          `json:"description"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Currency        string          `json:"currency"`
	ReceiptNo       *string         `json:"receipt_no"`
	Note            *string         `json:"note"`

	Entries []TransactionEntryPayload `json:"entries"`
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
