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

type PrepaidCategoryProjectionService struct{}

func (s *PrepaidCategoryProjectionService) Name() string {
	return enums.ProjectionTypePrepaidCategory.String()
}

func (s *PrepaidCategoryProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventPrepaidCategoryCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventPrepaidCategoryUpdated:
		return s.applyUpdated(ctx, tx, ct)
	case event_types.EventPrepaidCategoryDeleted:
		return s.applyDeleted(ctx, tx, ct)
	}
	return nil
}

func (s *PrepaidCategoryProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidCategoryCreatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.PrepaidCategoryRepo.InsertPrepaidCategory(ctx, sqlcdb.InsertPrepaidCategoryParams{
		MerchantID:       ct.MerchantID,
		CategoryUuid:     ct.Event.EventUuid,
		Name:             p.Name,
		AccountID:        p.AccountID,
		ExpenseAccountID: p.ExpenseAccountID,
		UpdatedBy:        updatedBy,
	})
}

func (s *PrepaidCategoryProjectionService) applyUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidCategoryUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.PrepaidCategoryRepo.UpdatePrepaidCategory(ctx, sqlcdb.UpdatePrepaidCategoryParams{
		MerchantID:       ct.MerchantID,
		CategoryUuid:     p.CategoryUUID,
		Name:             p.Name,
		AccountID:        p.AccountID,
		ExpenseAccountID: p.ExpenseAccountID,
		UpdatedBy:        updatedBy,
	})
}

func (s *PrepaidCategoryProjectionService) applyDeleted(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidCategoryDeletedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.PrepaidCategoryRepo.SoftDeletePrepaidCategory(ctx, sqlcdb.SoftDeletePrepaidCategoryParams{
		MerchantID:   ct.MerchantID,
		CategoryUuid: p.CategoryUUID,
		UpdatedBy:    updatedBy,
	})
}
