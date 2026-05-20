package projection

import "akatengu/internal/enums"

//dbmap:sqlcdb=LedgerAccountTypeConfig
type LedgerAccountTypeConfig struct {
	ID         int64                   `db:"id"          json:"id"`
	MerchantID int64                   `db:"merchant_id" json:"merchant_id"`
	Type       enums.LedgerAccountType `db:"type"        json:"type"`
	AccountID  string                  `db:"account_id"  json:"account_id"`
	UpdatedBy  *string                 `db:"updated_by"  json:"updated_by"`
	UpdatedAt  *string                 `db:"updated_at"  json:"updated_at"`
	Version    int64                   `db:"version"     json:"version"`
}

type LedgerAccountTypeConfigResult struct {
	LedgerAccountTypeConfig
	Descendants []Account `json:"descendants"`
}

//dbmap:sqlcdb=AssetTypeAccountConfig
type AssetTypeAccountConfig struct {
	ID                      int64           `db:"id"                          json:"id"`
	MerchantID              int64           `db:"merchant_id"                 json:"merchant_id"`
	AssetType               enums.AssetType `db:"asset_type"                  json:"asset_type"`
	RealizedGainAccountID   string          `db:"realized_gain_account_id"    json:"realized_gain_account_id"`
	RealizedLossAccountID   string          `db:"realized_loss_account_id"    json:"realized_loss_account_id"`
	UnrealizedGainAccountID string          `db:"unrealized_gain_account_id"  json:"unrealized_gain_account_id"`
	UnrealizedLossAccountID string          `db:"unrealized_loss_account_id"  json:"unrealized_loss_account_id"`
	OciAccountID            *string         `db:"oci_account_id"              json:"oci_account_id"`
	FeeAccountID            string          `db:"fee_account_id"              json:"fee_account_id"`
	TaxAccountID            string          `db:"tax_account_id"              json:"tax_account_id"`
	UpdatedBy               *string         `db:"updated_by"                  json:"updated_by"`
	UpdatedAt               *string         `db:"updated_at"                  json:"updated_at"`
	Version                 int64           `db:"version"                     json:"version"`
}
