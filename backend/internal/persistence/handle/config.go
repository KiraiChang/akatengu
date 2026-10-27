package handle

import (
	"akatengu/internal/database/sqlcdb"
	event_payload "akatengu/internal/domain/event"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/shared/utils"
	"context"

	"github.com/jmoiron/sqlx"
)

// NewConfigProjector builds the sys_accounts, ledger_account_type_config and
// asset_type_account_config read models, the same rows the legacy
// ConfigProjectionService (services/projection/config.go) writes.
func NewConfigProjector(db *sqlx.DB) Projector {
	c := &configProjector{
		handles: make(map[event_types.EventType]handle),
	}
	c.handles[event_types.EventSysAccountUpdated.Enum()] = c.applySysAccountUpdated
	c.handles[event_types.EventLedgerAccountTypeConfigUpdated.Enum()] = c.applyLedgerAccountTypeConfigUpdated
	c.handles[event_types.EventAssetTypeAccountConfigUpdated.Enum()] = c.applyAssetTypeAccountConfigUpdated
	return c
}

type configProjector struct {
	handles map[event_types.EventType]handle
}

func (c configProjector) EventTypes() []event_types.EventType {
	return utils.Keys(c.handles)
}

func (c configProjector) Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	h, ok := c.handles[evt.EventType]
	if !ok {
		return errors.NewRuntimeError(errors.ErrProjectorNotFound, evt)
	}
	return h(ctx, tx, evt)
}

func (c configProjector) Name() string {
	return "configProjector"
}

// Reset deletes the merchant's system accounts and account settings.
func (c configProjector) Reset(ctx context.Context, tx *repos.DbTransaction) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	if err := tx.Projection.DeleteSysAccounts(ctx, merchantID); err != nil {
		return err
	}
	if err := tx.Projection.DeleteLedgerAccountTypeConfigs(ctx, merchantID); err != nil {
		return err
	}
	return tx.Projection.DeleteAssetTypeAccountConfigs(ctx, merchantID)
}

func (c *configProjector) applySysAccountUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.SysAccountUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpsertSysAccount(ctx, sqlcdb.UpsertSysAccountParams{
		MerchantID:  evt.MerchantID,
		SysCode:     p.SysCode,
		Description: p.Description,
		AccountID:   p.AccountID,
	})
}

func (c *configProjector) applyLedgerAccountTypeConfigUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.LedgerAccountTypeConfigUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpsertLedgerAccountTypeConfig(ctx, sqlcdb.UpsertLedgerAccountTypeConfigParams{
		MerchantID: evt.MerchantID,
		Type:       p.Type,
		AccountID:  p.AccountID,
		UpdatedBy:  toUpdatedBy(evt.UpdatedBy),
	})
}

func (c *configProjector) applyAssetTypeAccountConfigUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.AssetTypeAccountConfigUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpsertAssetTypeAccountConfig(ctx, sqlcdb.UpsertAssetTypeAccountConfigParams{
		MerchantID:              evt.MerchantID,
		AssetType:               p.AssetType,
		RealizedGainAccountID:   p.RealizedGainAccountID,
		RealizedLossAccountID:   p.RealizedLossAccountID,
		UnrealizedGainAccountID: p.UnrealizedGainAccountID,
		UnrealizedLossAccountID: p.UnrealizedLossAccountID,
		OciAccountID:            p.OCIAccountID,
		FeeAccountID:            p.FeeAccountID,
		TaxAccountID:            p.TaxAccountID,
		AccountID:               p.AccountID,
		UpdatedBy:               toUpdatedBy(evt.UpdatedBy),
	})
}
