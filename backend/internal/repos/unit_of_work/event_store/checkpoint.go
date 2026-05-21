package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/pkg/ctxkey"
	"context"
)

type sqlcdbTxCheckpointRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxCheckpointRepository) Upsert(ctx context.Context, projectionName string, lastEventID int64) error {
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

func (r *sqlcdbTxCheckpointRepository) Update(ctx context.Context, name string, lastEventID int64) error {
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
