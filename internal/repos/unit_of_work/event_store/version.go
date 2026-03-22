package event_store

import (
	"akatengu/internal/model/enums"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type sqlxTxVersionRepository struct{ tx *sqlx.Tx }

func (r *sqlxTxVersionRepository) Get(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (int64, error) {
	var version int64
	err := r.tx.QueryRowxContext(ctx, `
        SELECT current_version FROM aggregate_versions
        WHERE aggregate_type = ? AND aggregate_id = ?`,
		aggregateType, aggregateID,
	).Scan(&version)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	return version, err
}

func (r *sqlxTxVersionRepository) Upsert(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) error {
	_, err := r.tx.ExecContext(ctx, `
        INSERT INTO aggregate_versions (aggregate_type, aggregate_id, current_version)
        VALUES (?, ?, ?)
        ON CONFLICT(aggregate_type, aggregate_id)
        DO UPDATE SET current_version = excluded.current_version`,
		aggregateType, aggregateID, version,
	)
	return err
}
