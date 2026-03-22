package event_store

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxTxEventRepository struct{ tx *sqlx.Tx }

func (r *sqlxTxEventRepository) Insert(ctx context.Context, p InsertEventParams) (int64, error) {
	result, err := r.tx.ExecContext(ctx, `
        INSERT INTO event_store
            (aggregate_type, aggregate_id, aggregate_version, event_type, payload, metadata)
        VALUES (?, ?, ?, ?, ?, ?)`,
		p.AggregateType, p.AggregateID, p.AggregateVersion,
		p.EventType, p.Payload, p.Metadata,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
