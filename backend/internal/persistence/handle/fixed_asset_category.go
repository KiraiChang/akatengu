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

// NewFixedAssetCategoryProjector builds the fixed_asset_categories read model,
// the same rows the legacy FixedAssetCategoryProjectionService
// (services/projection/fixed_asset_category.go) writes.
func NewFixedAssetCategoryProjector(db *sqlx.DB) Projector {
	c := &fixedAssetCategoryProjector{
		handles: make(map[event_types.EventType]handle),
	}
	c.handles[event_types.EventFixedAssetCategoryCreated.Enum()] = c.applyCreated
	c.handles[event_types.EventFixedAssetCategoryUpdated.Enum()] = c.applyUpdated
	c.handles[event_types.EventFixedAssetCategoryDeleted.Enum()] = c.applyDeleted
	return c
}

type fixedAssetCategoryProjector struct {
	handles map[event_types.EventType]handle
}

func (c fixedAssetCategoryProjector) EventTypes() []event_types.EventType {
	return utils.Keys(c.handles)
}

func (c fixedAssetCategoryProjector) Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	h, ok := c.handles[evt.EventType]
	if !ok {
		return errors.NewRuntimeError(errors.ErrProjectorNotFound, evt)
	}
	return h(ctx, tx, evt)
}

func (c fixedAssetCategoryProjector) Name() string {
	return "fixedAssetCategoryProjector"
}

// Reset deletes the merchant's fixed asset categories, deleted ones included.
func (c fixedAssetCategoryProjector) Reset(ctx context.Context, tx *repos.DbTransaction) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return tx.Projection.DeleteFixedAssetCategories(ctx, merchantID)
}

// applyCreated names the category after its creation event's UUID, which is
// also its aggregate ID (BFS R2 D6).
func (c *fixedAssetCategoryProjector) applyCreated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.FixedAssetCategoryCreatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.InsertFixedAssetCategory(ctx, sqlcdb.InsertFixedAssetCategoryParams{
		MerchantID:                   evt.MerchantID,
		CategoryUuid:                 evt.Uuid,
		Name:                         p.Name,
		AssetAccountID:               p.AssetAccountID,
		AccumDepreciationAccountID:   p.AccumDepreciationAccountID,
		DepreciationExpenseAccountID: p.DepreciationExpenseAccountID,
		UpdatedBy:                    toUpdatedBy(evt.UpdatedBy),
	})
}

func (c *fixedAssetCategoryProjector) applyUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.FixedAssetCategoryUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpdateFixedAssetCategory(ctx, sqlcdb.UpdateFixedAssetCategoryParams{
		MerchantID:                   evt.MerchantID,
		CategoryUuid:                 p.CategoryUUID,
		Name:                         p.Name,
		AssetAccountID:               p.AssetAccountID,
		AccumDepreciationAccountID:   p.AccumDepreciationAccountID,
		DepreciationExpenseAccountID: p.DepreciationExpenseAccountID,
		UpdatedBy:                    toUpdatedBy(evt.UpdatedBy),
	})
}

func (c *fixedAssetCategoryProjector) applyDeleted(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.FixedAssetCategoryDeletedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.SoftDeleteFixedAssetCategory(ctx, sqlcdb.SoftDeleteFixedAssetCategoryParams{
		MerchantID:   evt.MerchantID,
		CategoryUuid: p.CategoryUUID,
		UpdatedBy:    toUpdatedBy(evt.UpdatedBy),
	})
}
