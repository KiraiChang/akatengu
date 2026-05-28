package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
)

type ConfigProjectionService struct{}

func (s *ConfigProjectionService) Name() string { return enums.ProjectionTypeConfig.String() }

func (s *ConfigProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventSysAccountUpdated:
		return s.applySysAccountUpdated(ctx, tx, ct)
	case event_types.EventLedgerAccountTypeConfigUpdated:
		return s.applyLedgerAccountTypeConfigUpdated(ctx, tx, ct)
	case event_types.EventAssetTypeAccountConfigUpdated:
		return s.applyAssetTypeAccountConfigUpdated(ctx, tx, ct)
	}
	return nil
}

func (s *ConfigProjectionService) applySysAccountUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.SysAccountUpdatedPayload](ct)
	if err != nil {
		return err
	}
	return tx.Projection.ConfigProjectionRepo.UpsertSysAccount(ctx, sqlcdb.UpsertSysAccountParams{
		MerchantID:  ct.MerchantID,
		SysCode:     p.SysCode,
		Description: p.Description,
		AccountID:   p.AccountID,
	})
}

func (s *ConfigProjectionService) applyLedgerAccountTypeConfigUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.LedgerAccountTypeConfigUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.ConfigProjectionRepo.UpsertLedgerAccountTypeConfig(ctx, sqlcdb.UpsertLedgerAccountTypeConfigParams{
		MerchantID: ct.MerchantID,
		Type:       p.Type,
		AccountID:  p.AccountID,
		UpdatedBy:  updatedBy,
	})
}

func (s *ConfigProjectionService) applyAssetTypeAccountConfigUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AssetTypeAccountConfigUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.ConfigProjectionRepo.UpsertAssetTypeAccountConfig(ctx, sqlcdb.UpsertAssetTypeAccountConfigParams{
		MerchantID:              ct.MerchantID,
		AssetType:               p.AssetType,
		RealizedGainAccountID:   p.RealizedGainAccountID,
		RealizedLossAccountID:   p.RealizedLossAccountID,
		UnrealizedGainAccountID: p.UnrealizedGainAccountID,
		UnrealizedLossAccountID: p.UnrealizedLossAccountID,
		OciAccountID:            p.OCIAccountID,
		FeeAccountID:            p.FeeAccountID,
		TaxAccountID:            p.TaxAccountID,
		AccountID:               p.AccountID, // *string, nullable
		UpdatedBy:               updatedBy,
	})
}
