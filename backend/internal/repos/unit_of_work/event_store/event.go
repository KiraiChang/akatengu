package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"fmt"
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

func (r *sqlcdbTxEventRepository) InsertWithTimestamp(ctx context.Context, p InsertEventImportParams) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	if err := r.q.InsertEventImport(ctx, sqlcdb.InsertEventImportParams{
		MerchantID:       merchantID,
		AggregateType:    p.AggregateType,
		AggregateID:      p.AggregateID,
		AggregateVersion: p.AggregateVersion,
		EventType:        p.EventType,
		EventUuid:        p.EventUuid,
		OccurredAt:       p.OccurredAt.UTC().Format("2006-01-02 15:04:05"),
		Payload:          p.Payload,
		Metadata:         p.Metadata,
		UpdatedBy:        p.UpdatedBy,
	}); err != nil {
		return fmt.Errorf("insert event import: %w", err)
	}
	return nil
}
