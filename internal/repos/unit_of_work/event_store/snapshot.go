package event_store

import (
	"akatengu/internal/model/db"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxTxSnapshotRepository struct{ tx *sqlx.Tx }

func (r *sqlxTxSnapshotRepository) Upsert(ctx context.Context, snap db.Snapshot) error {
	_, err := r.tx.ExecContext(ctx, `
        INSERT INTO snapshots (aggregate_type, aggregate_id, at_version, state)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(aggregate_type, aggregate_id)
        DO UPDATE SET at_version = excluded.at_version, state = excluded.state`,
		snap.AggregateType, snap.AggregateId, snap.AtVersion, snap.State,
	)
	return err
}
