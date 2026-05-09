package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbTxCheckpointRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxCheckpointRepository) Upsert(ctx context.Context, projectionName string, lastEventID int64) error {
	return r.q.UpsertCheckpoint(ctx, sqlcdb.UpsertCheckpointParams{
		projectionName,
		lastEventID,
	})
}

func (r *sqlcdbTxCheckpointRepository) Update(ctx context.Context, name string, lastEventID int64) error {
	return r.q.UpdateCheckpoint(ctx, sqlcdb.UpdateCheckpointParams{
		ProjectionName: name,
		LastEventID:    lastEventID,
	})
}
