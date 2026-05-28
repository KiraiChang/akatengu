package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/pkg/ctxkey"
	"context"
)

type sqlcdbTxEventRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxEventRepository) Insert(ctx context.Context, p InsertEventParams) (int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return 0, err
	}
	return r.q.InsertEvent(ctx, sqlcdb.InsertEventParams{
		MerchantID:       merchantID,
		AggregateType:    p.AggregateType,
		AggregateID:      p.AggregateID,
		AggregateVersion: p.AggregateVersion,
		EventType:        p.EventType,
		EventUuid:        p.EventUuid,
		Payload:          p.Payload,
		Metadata:         p.Metadata,
		UpdatedBy:        p.UpdatedBy,
	})
}
