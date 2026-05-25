package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=TransactionTemplate
type TransactionTemplate struct {
	ID          int64   `db:"id"          json:"id"`
	MerchantID  int64   `db:"merchant_id" json:"merchant_id"`
	Name        string  `db:"name"        json:"name"`
	Description *string `db:"description" json:"description"`
	Tag         *string `db:"tag"         json:"tag"`
	UpdatedBy   *string `db:"updated_by"  json:"updated_by"`
	UpdatedAt   *string `db:"updated_at"  json:"updated_at"`
	Version     int64   `db:"version"     json:"version"`
}

//dbmap:sqlcdb=TransactionTemplateEntry
type TransactionTemplateEntry struct {
	ID               int64                    `db:"id"                 json:"id"`
	MerchantID       int64                    `db:"merchant_id"        json:"merchant_id"`
	TemplateID       int64                    `db:"template_id"        json:"template_id"`
	SortOrder        int64                    `db:"sort_order"         json:"sort_order"`
	AccountID        string                   `db:"account_id"         json:"account_id"`
	LedgerID         *int64                   `db:"ledger_id"          json:"ledger_id"`
	Debit            decimal.Decimal          `db:"debit"              json:"debit"`
	Credit           decimal.Decimal          `db:"credit"             json:"credit"`
	Note             *string                  `db:"note"               json:"note"`
	CashFlowCategory *enums.CashFlowCategory  `db:"cash_flow_category" json:"cash_flow_category"`
}

type TransactionTemplateDetail struct {
	TransactionTemplate
	Entries []TransactionTemplateEntry `json:"entries"`
}
