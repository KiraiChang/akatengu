package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"context"
	"database/sql"
	"fmt"
)

type sqlcdbTxVersionRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxVersionRepository) Get(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (int64, error) {
	version, err := r.q.GetAggregateVersion(ctx, sqlcdb.GetAggregateVersionParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
	})
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return version, err
}

func (r *sqlcdbTxVersionRepository) Upsert(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) error {
	return r.q.UpsertAggregateVersion(ctx, sqlcdb.UpsertAggregateVersionParams{
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		CurrentVersion: version,
	})
}

func (r *sqlcdbTxVersionRepository) UpdateIfVersionMatch(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) (int64, error) {
	result, err := r.q.UpdateVersionIfMatch(ctx, sqlcdb.UpdateVersionIfMatchParams{
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		CurrentVersion: version,
	})
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("version %d not found in aggregate %s", version, aggregateType)
	}
	if err != nil {
		return 0, err
	}
	return result.CurrentVersion, nil
}