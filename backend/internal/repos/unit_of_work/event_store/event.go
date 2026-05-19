package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbTxEventRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxEventRepository) Insert(ctx context.Context, p InsertEventParams) (int64, error) {
	return r.q.InsertEvent(ctx, sqlcdb.InsertEventParams{
		AggregateType:    p.AggregateType,
		AggregateID:      p.AggregateID,
		AggregateVersion: p.AggregateVersion,
		EventType:        p.EventType,
		Payload:          p.Payload,
		Metadata:         p.Metadata,
		UpdatedBy:        p.UpdatedBy, // *string, nil when no user (replay)
	})
}