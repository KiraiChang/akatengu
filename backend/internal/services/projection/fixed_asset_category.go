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

type FixedAssetCategoryProjectionService struct{}

func (s *FixedAssetCategoryProjectionService) Name() string {
	return enums.ProjectionTypeFixedAssetCategory.String()
}

func (s *FixedAssetCategoryProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventFixedAssetCategoryCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventFixedAssetCategoryUpdated:
		return s.applyUpdated(ctx, tx, ct)
	case event_types.EventFixedAssetCategoryDeleted:
		return s.applyDeleted(ctx, tx, ct)
	}
	return nil
}

func (s *FixedAssetCategoryProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.FixedAssetCategoryCreatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.FixedAssetCategoryRepo.InsertFixedAssetCategory(ctx, sqlcdb.InsertFixedAssetCategoryParams{
		MerchantID:                   ct.MerchantID,
		CategoryUuid:                 ct.Event.EventUuid,
		Name:                         p.Name,
		AssetAccountID:               p.AssetAccountID,
		AccumDepreciationAccountID:   p.AccumDepreciationAccountID,
		DepreciationExpenseAccountID: p.DepreciationExpenseAccountID,
		UpdatedBy:                    updatedBy,
	})
}

func (s *FixedAssetCategoryProjectionService) applyUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.FixedAssetCategoryUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.FixedAssetCategoryRepo.UpdateFixedAssetCategory(ctx, sqlcdb.UpdateFixedAssetCategoryParams{
		MerchantID:                   ct.MerchantID,
		CategoryUuid:                 p.CategoryUUID,
		Name:                         p.Name,
		AssetAccountID:               p.AssetAccountID,
		AccumDepreciationAccountID:   p.AccumDepreciationAccountID,
		DepreciationExpenseAccountID: p.DepreciationExpenseAccountID,
		UpdatedBy:                    updatedBy,
	})
}

func (s *FixedAssetCategoryProjectionService) applyDeleted(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.FixedAssetCategoryDeletedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.FixedAssetCategoryRepo.SoftDeleteFixedAssetCategory(ctx, sqlcdb.SoftDeleteFixedAssetCategoryParams{
		MerchantID:   ct.MerchantID,
		CategoryUuid: p.CategoryUUID,
		UpdatedBy:    updatedBy,
	})
}
