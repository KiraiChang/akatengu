package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=EntryCfCategory
type EntryCFCategory struct {
	EntryUuid   string                 `db:"entry_uuid"   json:"entry_uuid"`
	MerchantID  int64                  `db:"merchant_id"  json:"merchant_id"`
	CfCategory  enums.CashFlowCategory `db:"cf_category"  json:"cf_category"`
	IsConfirmed bool                   `db:"is_confirmed" json:"is_confirmed"`
	UpdatedAt   *string                `db:"updated_at"   json:"updated_at"`
	UpdatedBy   *string                `db:"updated_by"   json:"updated_by"`
}

//dbmap:sqlcdb=ListTransactionCFReviewRow
type TransactionCFReview struct {
	TxnID       int64                  `db:"txn_id"  json:"txn_id"`
	TxnUuid     string                 `db:"txn_uuid" json:"txn_uuid"`
	TxnDate     string                 `db:"txn_date" json:"txn_date"`
	Description string                 `db:"description" json:"description"`
	TotalAmount decimal.Decimal        `db:"total_amount"  json:"total_amount"`
	Currency    string                 `db:"currency"  json:"currency"`
	EntryID     int64                  `db:"entry_id"  json:"entry_id"`
	EntryUuid   string                 `db:"entry_uuid" json:"entry_uuid"`
	AccountID   string                 `db:"account_id" json:"account_id"`
	LedgerID    *int64                 `db:"ledger_id"  json:"ledger_id"`
	Debit       decimal.Decimal        `db:"debit"  json:"debit"`
	Credit      decimal.Decimal        `db:"credit"  json:"credit"`
	CfCategory  enums.CashFlowCategory `db:"cf_category"  json:"cf_category"`
	IsConfirmed bool                   `db:"is_confirmed" json:"is_confirmed"`
}
