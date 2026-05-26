package event_store

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

type sqlcdbTxVersionRepository struct{ q *sqlcdb.Queries }

func (r *sqlcdbTxVersionRepository) Get(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return 0, err
	}
	version, err := r.q.GetAggregateVersion(ctx, sqlcdb.GetAggregateVersionParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		MerchantID:    merchantID,
	})
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return version, err
}

func (r *sqlcdbTxVersionRepository) Insert(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return r.q.InsertAggregateVersion(ctx, sqlcdb.InsertAggregateVersionParams{
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		MerchantID:     merchantID,
		CurrentVersion: version,
	})
}

func (r *sqlcdbTxVersionRepository) UpdateIfVersionMatch(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) (int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return 0, err
	}
	result, err := r.q.UpdateVersionIfMatch(ctx, sqlcdb.UpdateVersionIfMatchParams{
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		MerchantID:     merchantID,
		CurrentVersion: version,
	})
	if err != nil {
		return 0, err
	}
	return result.CurrentVersion, nil
}
