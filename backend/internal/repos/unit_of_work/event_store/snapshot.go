package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"context"
)

type sqlcdbTxSnapshotRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxSnapshotRepository) Upsert(ctx context.Context, snap db.Snapshot) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return r.q.UpsertSnapshot(ctx, sqlcdb.UpsertSnapshotParams{
		MerchantID:    merchantID,
		SnapshotUuid:  snap.SnapshotUuid,
		AggregateType: snap.AggregateType,
		AggregateID:   snap.AggregateId,
		AtVersion:     snap.AtVersion,
		State:         snap.State,
	})
}
