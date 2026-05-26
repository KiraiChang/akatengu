package request

import "errors"

type UpdateLedgerAccountTypeConfig struct {
	AccountID string `json:"account_id"`
}

func (r *UpdateLedgerAccountTypeConfig) Validate() error {
	if r.AccountID == "" {
		return errors.New("account_id is required")
	}
	return nil
}

type UpdateAssetTypeAccountConfig struct {
	RealizedGainAccountID   string  `json:"realized_gain_account_id"`
	RealizedLossAccountID   string  `json:"realized_loss_account_id"`
	UnrealizedGainAccountID string  `json:"unrealized_gain_account_id"`
	UnrealizedLossAccountID string  `json:"unrealized_loss_account_id"`
	OCIAccountID            *string `json:"oci_account_id"`
	FeeAccountID            string  `json:"fee_account_id"`
	TaxAccountID            string  `json:"tax_account_id"`
	AccountID               *string `json:"account_id"`
}

func (r *UpdateAssetTypeAccountConfig) Validate() error {
	if r.RealizedGainAccountID == "" {
		return errors.New("realized_gain_account_id is required")
	}
	if r.RealizedLossAccountID == "" {
		return errors.New("realized_loss_account_id is required")
	}
	if r.UnrealizedGainAccountID == "" {
		return errors.New("unrealized_gain_account_id is required")
	}
	if r.UnrealizedLossAccountID == "" {
		return errors.New("unrealized_loss_account_id is required")
	}
	if r.FeeAccountID == "" {
		return errors.New("fee_account_id is required")
	}
	if r.TaxAccountID == "" {
		return errors.New("tax_account_id is required")
	}
	return nil
}
