package projection

import (
	"akatengu/internal/model/enums"

	"github.com/shopspring/decimal"
)

type Entry struct {
	EntryId       int64           `db:"entry_id"`
	TransactionId int64           `db:"txn_id"`
	LedgerId      *int64          `db:"ledger_id"`
	AccountId     string          `db:"account_id"`
	Debit         decimal.Decimal `db:"debit"`
	Credit        decimal.Decimal `db:"credit"`
	Note          *string         `db:"note"`
}

type Transaction struct {
	TransactionId   int64                   `db:"txn_id"`
	TransactionDate string                  `db:"txn_date"`
	Description     string                  `db:"description"`
	TotalAmount     decimal.Decimal         `db:"total_amount"`
	Currency        string                  `db:"currency"`
	Status          enums.TransactionStatus `db:"status"`
	InstallmentId   *int64                  `db:"installment_id"`
	ReceiptNo       *string                 `db:"receipt_no"`
	Note            *string                 `db:"note"`
	Version         int64                   `db:"version"`
	RefTxnId        *int64                  `db:"ref_txn_id"`
}
