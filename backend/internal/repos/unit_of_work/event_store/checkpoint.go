package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbTxCheckpointRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxCheckpointRepository) Update(ctx context.Context, name string, lastEventID int64) error {
	return r.q.UpdateCheckpoint(ctx, sqlcdb.UpdateCheckpointParams{
		ProjectionName: name,
		LastEventID:    lastEventID,
	})
}