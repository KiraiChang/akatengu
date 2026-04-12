package event_store

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"context"
	"database/sql"
	"fmt"

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

func (r *sqlxTxVersionRepository) UpdateIfVersionMatch(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) (int64, error) {
	var result db.AggregateVersion
	err := r.tx.GetContext(ctx, &result, `
        UPDATE aggregate_versions 
        	SET current_version = current_version +1 
        WHERE aggregate_type = ?
        	AND aggregate_id = ?
        	AND current_version = ?
        RETURNING *`,
		aggregateType, aggregateID, version,
	)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("version %d not found in aggregate %s", version, aggregateType)
	}
	return result.AggregateVersion, nil
}
