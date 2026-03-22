package event_store

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxTxCheckpointRepository struct{ tx *sqlx.Tx }

func (r *sqlxTxCheckpointRepository) Update(ctx context.Context, name string, lastEventID int64) error {
	_, err := r.tx.ExecContext(ctx, `
        UPDATE projection_checkpoints
        SET last_event_id = ?, updated_at = datetime('now')
        WHERE projection_name = ?`,
		lastEventID, name,
	)
	return err
}
