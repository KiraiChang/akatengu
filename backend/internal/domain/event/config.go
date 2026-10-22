package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
)

// The config payloads are copies of model/payload/sysconfig.go and
// accountconfig.go (BFS R2 D1); the JSON tags must stay identical until R10.

// SysAccountUpdatedPayload maps a system code to an account. The event keeps
// Description, so a replay rebuilds the initial description too.
type SysAccountUpdatedPayload struct {
	SysCode     string `json:"sys_code"`
	Description string `json:"description"`
	AccountID   string `json:"account_id"`
}

func (p SysAccountUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventSysAccountUpdated.Enum()
}

func (p SysAccountUpdatedPayload) Validate() error {
	var errs []string
	if p.SysCode == "" {
		errs = append(errs, "sys_code is required")
	}
	if p.AccountID == "" {
		errs = append(errs, "account_id is required")
	}
	return joinErrors(errs)
}

// LedgerAccountTypeConfigUpdatedPayload sets the anchor account of a ledger
// account type.
type LedgerAccountTypeConfigUpdatedPayload struct {
	Type      enums.LedgerAccountType `json:"type"`
	AccountID string                  `json:"account_id"`
}

func (p LedgerAccountTypeConfigUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventLedgerAccountTypeConfigUpdated.Enum()
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

// AssetTypeAccountConfigUpdatedPayload sets the accounts of an asset type.
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

func (p AssetTypeAccountConfigUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventAssetTypeAccountConfigUpdated.Enum()
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
