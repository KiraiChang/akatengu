package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=FixedAsset
type FixedAsset struct {
	ID                           int64                    `db:"id" json:"id"`
	MerchantID                   int64                    `db:"merchant_id" json:"merchant_id"`
	TxnID                        *int64                   `db:"txn_id" json:"txn_id"`
	Name                         string                   `db:"name" json:"name"`
	AssetAccountID               string                   `db:"asset_account_id" json:"asset_account_id"`
	AccumDepreciationAccountID   string                   `db:"accum_depreciation_account_id" json:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string                   `db:"depreciation_expense_account_id" json:"depreciation_expense_account_id"`
	Cost                         decimal.Decimal          `db:"cost" json:"cost"`
	ResidualValue                decimal.Decimal          `db:"residual_value" json:"residual_value"`
	UsefulLifeMonths             int64                    `db:"useful_life_months" json:"useful_life_months"`
	DepreciationMethod           enums.DepreciationMethod `db:"depreciation_method" json:"depreciation_method"`
	PaymentType                  enums.AssetPaymentType   `db:"payment_type" json:"payment_type"`
	TotalDepreciated             decimal.Decimal          `db:"total_depreciated" json:"total_depreciated"`
	DepreciatedPeriods           int64                    `db:"depreciated_periods" json:"depreciated_periods"`
	PurchaseDate                 string                   `db:"purchase_date" json:"purchase_date"`
	DisposalDate                 *string                  `db:"disposal_date" json:"disposal_date"`
	Status                       enums.FixedAssetStatus   `db:"status" json:"status"`
	UpdatedBy                    *string                  `db:"updated_by" json:"updated_by"`
	UpdatedAt                    *string                  `db:"updated_at" json:"updated_at"`
	Version                      int64                    `db:"version" json:"version"`
}

//dbmap:sqlcdb=FixedAssetDepreciation
type FixedAssetDepreciation struct {
	ID         int64           `db:"id" json:"id"`
	MerchantID int64           `db:"merchant_id" json:"merchant_id"`
	AssetID    int64           `db:"asset_id" json:"asset_id"`
	TxnID      int64           `db:"txn_id" json:"txn_id"`
	PeriodDate string          `db:"period_date" json:"period_date"`
	Amount     decimal.Decimal `db:"amount" json:"amount"`
	UpdatedBy  *string         `db:"updated_by" json:"updated_by"`
	UpdatedAt  *string         `db:"updated_at" json:"updated_at"`
}
