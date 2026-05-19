package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=JournalEntry
//dbmap:sqlcdb=InsertJournalEntryWithIDParams
//dbmap:sqlcdb=InsertJournalEntryParams
//dbmap:sqlcdb=GetJournalEntriesRow
type Entry struct {
	MerchantID       int64                  `db:"merchant_id" json:"merchant_id"`
	EntryId          int64                  `db:"entry_id" json:"entry_id"`
	TransactionId    int64                  `db:"txn_id" json:"txn_id"`
	LedgerId         *int64                 `db:"ledger_id" json:"ledger_id"`
	AccountId        string                 `db:"account_id" json:"account_id"`
	Debit            decimal.Decimal        `db:"debit" json:"debit"`
	Credit           decimal.Decimal        `db:"credit" json:"credit"`
	Note             *string                `db:"note" json:"note"`
	CashFlowCategory enums.CashFlowCategory `db:"cash_flow_category" json:"cash_flow_category"`
	UpdatedBy        *string                `db:"updated_by" json:"updated_by"`
	UpdatedAt        *string                `db:"updated_at" json:"updated_at"`
}

//dbmap:sqlcdb=Transaction
//dbmap:sqlcdb=InsertTransactionParams
//dbmap:sqlcdb=GetTransactionPagedRow
//dbmap:sqlcdb=GetTransactionRow
type Transaction struct {
	MerchantID      int64                   `db:"merchant_id" json:"merchant_id"`
	TransactionId   int64                   `db:"txn_id" json:"txn_id"`
	TransactionDate string                  `db:"txn_date" json:"txn_date"`
	Description     string                  `db:"description" json:"description"`
	TotalAmount     decimal.Decimal         `db:"total_amount" json:"total_amount"`
	Currency        string                  `db:"currency" json:"currency"`
	Status          enums.TransactionStatus `db:"status" json:"status"`
	InstallmentId   *int64                  `db:"installment_id" json:"installment_id"`
	ReceiptNo       *string                 `db:"receipt_no" json:"receipt_no"`
	Note            *string                 `db:"note" json:"note"`
	Version         int64                   `db:"version" json:"version"`
	RefTxnId        *int64                  `db:"ref_txn_id" json:"ref_txn_id"`
	UpdatedBy       *string                 `db:"updated_by" json:"updated_by"`
	UpdatedAt       *string                 `db:"updated_at" json:"updated_at"`
}
