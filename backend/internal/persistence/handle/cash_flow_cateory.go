package handle

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/model/payload"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/shared/utils"
	"context"

	"github.com/jmoiron/sqlx"
)

func NewCashFlowCategoryProjector(db *sqlx.DB) Projector {
	c := &cashFlowCategoryProjector{
		handles: make(map[event_types.EventType]handle),
	}
	c.handles[event_types.EventTransactionCFCategoryUpdated.Enum()] = c.applyUserUpdate
	return c
}

type cashFlowCategoryProjector struct {
	handles map[event_types.EventType]handle
}

func (c cashFlowCategoryProjector) EventTypes() []event_types.EventType {
	return utils.Keys(c.handles)
}

func (c cashFlowCategoryProjector) Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event, state any) error {
	h, ok := c.handles[evt.EventType]
	if !ok {
		return errors.NewRuntimeError(errors.ErrProjectorNotFound, evt)
	}
	return h(ctx, tx, evt, state)
}

func (c cashFlowCategoryProjector) Name() string {
	return "cashFlowCategoryProjector"
}

func (s *cashFlowCategoryProjector) applyUserUpdate(ctx context.Context, tx *repos.DbTransaction, evt event.Event, state any) error {
	p, err := checkAndGetPayload[payload.TransactionCFCategoryUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	updBy := toUpdatedBy(evt.UpdatedBy)
	for _, item := range p.Entries {
		if err := tx.Projection.UpsertEntryCFCategory(ctx, sqlcdb.UpsertEntryCFCategoryParams{
			EntryUuid:   item.EntryUUID,
			MerchantID:  evt.MerchantID,
			CfCategory:  item.CFCategory,
			IsConfirmed: true,
			UpdatedBy:   updBy,
		}); err != nil {
			return err
		}
	}
	return nil
}
