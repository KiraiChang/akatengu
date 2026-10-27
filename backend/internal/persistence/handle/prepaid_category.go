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

// NewPrepaidCategoryProjector builds the prepaid_categories read model, the
// same rows the legacy PrepaidCategoryProjectionService
// (services/projection/prepaid_category.go) writes.
func NewPrepaidCategoryProjector(db *sqlx.DB) Projector {
	c := &prepaidCategoryProjector{
		handles: make(map[event_types.EventType]handle),
	}
	c.handles[event_types.EventPrepaidCategoryCreated.Enum()] = c.applyCreated
	c.handles[event_types.EventPrepaidCategoryUpdated.Enum()] = c.applyUpdated
	c.handles[event_types.EventPrepaidCategoryDeleted.Enum()] = c.applyDeleted
	return c
}

type prepaidCategoryProjector struct {
	handles map[event_types.EventType]handle
}

func (c prepaidCategoryProjector) EventTypes() []event_types.EventType {
	return utils.Keys(c.handles)
}

func (c prepaidCategoryProjector) Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	h, ok := c.handles[evt.EventType]
	if !ok {
		return errors.NewRuntimeError(errors.ErrProjectorNotFound, evt)
	}
	return h(ctx, tx, evt)
}

func (c prepaidCategoryProjector) Name() string {
	return "prepaidCategoryProjector"
}

// Reset deletes the merchant's prepaid categories, deleted ones included.
func (c prepaidCategoryProjector) Reset(ctx context.Context, tx *repos.DbTransaction) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return tx.Projection.DeletePrepaidCategories(ctx, merchantID)
}

// applyCreated names the category after its creation event's UUID, which is
// also its aggregate ID (BFS R2 D6).
func (c *prepaidCategoryProjector) applyCreated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.PrepaidCategoryCreatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.InsertPrepaidCategory(ctx, sqlcdb.InsertPrepaidCategoryParams{
		MerchantID:       evt.MerchantID,
		CategoryUuid:     evt.Uuid,
		Name:             p.Name,
		AccountID:        p.AccountID,
		ExpenseAccountID: p.ExpenseAccountID,
		UpdatedBy:        toUpdatedBy(evt.UpdatedBy),
	})
}

func (c *prepaidCategoryProjector) applyUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.PrepaidCategoryUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpdatePrepaidCategory(ctx, sqlcdb.UpdatePrepaidCategoryParams{
		MerchantID:       evt.MerchantID,
		CategoryUuid:     p.CategoryUUID,
		Name:             p.Name,
		AccountID:        p.AccountID,
		ExpenseAccountID: p.ExpenseAccountID,
		UpdatedBy:        toUpdatedBy(evt.UpdatedBy),
	})
}

func (c *prepaidCategoryProjector) applyDeleted(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.PrepaidCategoryDeletedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.SoftDeletePrepaidCategory(ctx, sqlcdb.SoftDeletePrepaidCategoryParams{
		MerchantID:   evt.MerchantID,
		CategoryUuid: p.CategoryUUID,
		UpdatedBy:    toUpdatedBy(evt.UpdatedBy),
	})
}
