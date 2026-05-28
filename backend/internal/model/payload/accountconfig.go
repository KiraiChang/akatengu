package payload

import "akatengu/internal/enums"

// LedgerAccountTypeConfigUpdatedPayload 更新帳戶類型的錨定科目設定。
type LedgerAccountTypeConfigUpdatedPayload struct {
	Type      enums.LedgerAccountType `json:"type"`
	AccountID string                  `json:"account_id"`
}

func (p LedgerAccountTypeConfigUpdatedPayload) Validate() error {
	var errs []string
	if p.Type.IsZero() {
		errs = append(errs, "type is required")
	}
	if p.AccountID == "" {
		errs = append(errs, "account_id is required")
	}
	return joinErrors(errs)
}

// AssetTypeAccountConfigUpdatedPayload 更新資產類型的科目設定。
type AssetTypeAccountConfigUpdatedPayload struct {
	AssetType               enums.AssetType `json:"asset_type"`
	RealizedGainAccountID   string          `json:"realized_gain_account_id"`
	RealizedLossAccountID   string          `json:"realized_loss_account_id"`
	UnrealizedGainAccountID string          `json:"unrealized_gain_account_id"`
	UnrealizedLossAccountID string          `json:"unrealized_loss_account_id"`
	OCIAccountID            *string         `json:"oci_account_id"`
	FeeAccountID            string          `json:"fee_account_id"`
	TaxAccountID            string          `json:"tax_account_id"`
	AccountID               *string         `json:"account_id,omitempty"`
}

func (p AssetTypeAccountConfigUpdatedPayload) Validate() error {
	var errs []string
	if p.AssetType.IsZero() {
		errs = append(errs, "asset_type is required")
	}
	if p.RealizedGainAccountID == "" {
		errs = append(errs, "realized_gain_account_id is required")
	}
	if p.RealizedLossAccountID == "" {
		errs = append(errs, "realized_loss_account_id is required")
	}
	if p.UnrealizedGainAccountID == "" {
		errs = append(errs, "unrealized_gain_account_id is required")
	}
	if p.UnrealizedLossAccountID == "" {
		errs = append(errs, "unrealized_loss_account_id is required")
	}
	if p.FeeAccountID == "" {
		errs = append(errs, "fee_account_id is required")
	}
	if p.TaxAccountID == "" {
		errs = append(errs, "tax_account_id is required")
	}
	return joinErrors(errs)
}
