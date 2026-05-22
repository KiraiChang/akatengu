package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=Prepaid
type Prepaid struct {
	ID               int64               `db:"id" json:"id"`
	MerchantID       int64               `db:"merchant_id" json:"merchant_id"`
	TxnID            *int64              `db:"txn_id" json:"txn_id"`
	AccountID        string              `db:"account_id" json:"account_id"`
	ExpenseAccountID string              `db:"expense_account_id" json:"expense_account_id"`
	Name             string              `db:"name" json:"name"`
	TotalAmount      decimal.Decimal     `db:"total_amount" json:"total_amount"`
	AmortizedAmount  decimal.Decimal     `db:"amortized_amount" json:"amortized_amount"`
	Periods          int64               `db:"periods" json:"periods"`
	AmortizedPeriods int64               `db:"amortized_periods" json:"amortized_periods"`
	StartDate        string              `db:"start_date" json:"start_date"`
	Status           enums.PrepaidStatus `db:"status" json:"status"`
	UpdatedBy        *string             `db:"updated_by" json:"updated_by"`
	UpdatedAt        *string             `db:"updated_at" json:"updated_at"`
	Version          int64               `db:"version" json:"version"`
}

//dbmap:sqlcdb=PrepaidAmortization
type PrepaidAmortization struct {
	ID         int64           `db:"id" json:"id"`
	MerchantID int64           `db:"merchant_id" json:"merchant_id"`
	PrepaidID  int64           `db:"prepaid_id" json:"prepaid_id"`
	TxnID      int64           `db:"txn_id" json:"txn_id"`
	PeriodDate string          `db:"period_date" json:"period_date"`
	Amount     decimal.Decimal `db:"amount" json:"amount"`
	UpdatedBy  *string         `db:"updated_by" json:"updated_by"`
	UpdatedAt  *string         `db:"updated_at" json:"updated_at"`
}
