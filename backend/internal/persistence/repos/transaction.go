package repos

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db"
	"akatengu/internal/persistence/repos/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
)

type Transaction struct {
	Store      Store
	Check      Checkpoint
	Snap       Snapshot
	Projection projection.Projection
}

type Checkpoint interface {
	Update(ctx context.Context, projectionName string, lastEventID int64) error
	Upsert(ctx context.Context, projectionName string, lastEventID int64) error
}

type sqlcdbTxCheckpoint struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxCheckpoint) Upsert(ctx context.Context, projectionName string, lastEventID int64) error {
	merchantId, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return r.q.UpsertCheckpoint(ctx, sqlcdb.UpsertCheckpointParams{
		projectionName,
		merchantId,
		lastEventID,
	})
}

func (r *sqlcdbTxCheckpoint) Update(ctx context.Context, name string, lastEventID int64) error {
	merchantId, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return r.q.UpdateCheckpoint(ctx, sqlcdb.UpdateCheckpointParams{
		ProjectionName: name,
		MerchantID:     merchantId,
		LastEventID:    lastEventID,
	})
}

type Snapshot interface {
	Upsert(ctx context.Context, snap db.Snapshot) error
}

type sqlcdbTxSnapshot struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxSnapshot) Upsert(ctx context.Context, snap db.Snapshot) error {
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
