package projection

//dbmap:sqlcdb=FixedAssetCategory
type FixedAssetCategory struct {
	ID                           int64   `db:"id" json:"id"`
	CategoryUUID                 string  `db:"category_uuid" json:"category_uuid"`
	MerchantID                   int64   `db:"merchant_id" json:"merchant_id"`
	Name                         string  `db:"name" json:"name"`
	AssetAccountID               string  `db:"asset_account_id" json:"asset_account_id"`
	AccumDepreciationAccountID   string  `db:"accum_depreciation_account_id" json:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string  `db:"depreciation_expense_account_id" json:"depreciation_expense_account_id"`
	IsActive                     bool    `db:"is_active" json:"is_active"`
	UpdatedBy                    *string `db:"updated_by" json:"updated_by"`
	UpdatedAt                    string  `db:"updated_at" json:"updated_at"`
	Version                      int64   `db:"version" json:"version"`
}
