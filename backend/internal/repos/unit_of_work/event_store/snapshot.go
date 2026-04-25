package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db"
	"context"
)

type sqlcdbTxSnapshotRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxSnapshotRepository) Upsert(ctx context.Context, snap db.Snapshot) error {
	return r.q.UpsertSnapshot(ctx, sqlcdb.UpsertSnapshotParams{
		AggregateType: snap.AggregateType,
		AggregateID:   snap.AggregateId,
		AtVersion:     snap.AtVersion,
		State:         snap.State,
	})
}