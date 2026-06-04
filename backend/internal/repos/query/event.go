package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type EventRepo interface {
	GetByAggregate(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) ([]db.EventStore, error)
	GetAfterVersion(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, afterVersion int64) ([]db.EventStore, error)
	GetAfterEventID(ctx context.Context, aggregateType enums.AggregateType, afterEventID int64, limit int) ([]db.EventStore, error)
	GetSnapshot(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (*db.Snapshot, error)
	GetCheckpoint(ctx context.Context, projectionName string) (int64, error)
	Replay(ctx context.Context, fromEventID int64, aggregateType *enums.AggregateType) ([]db.EventStore, error)
	GetAllByMerchant(ctx context.Context) ([]db.EventStore, error) // 匯出用
}

type sqlcdbEventRepository struct {
	q *sqlcdb.Queries
}

func newEventRepo(q *sqlcdb.Queries) EventRepo {
	return &sqlcdbEventRepository{q: q}
}

func NewEventRepo(db *sqlx.DB) EventRepo {
	return &sqlcdbEventRepository{q: sqlcdb.New(db)}
}

func (r *sqlcdbEventRepository) GetByAggregate(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) ([]db.EventStore, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetEventsByAggregate(ctx, sqlcdb.GetEventsByAggregateParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		MerchantID:    merchantID,
	})
	if err != nil {
		return nil, err
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetAfterVersion(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, afterVersion int64) ([]db.EventStore, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetEventsAfterVersion(ctx, sqlcdb.GetEventsAfterVersionParams{
		AggregateType:    aggregateType,
		AggregateID:      aggregateID,
		MerchantID:       merchantID,
		AggregateVersion: afterVersion,
	})
	if err != nil {
		return nil, err
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetAfterEventID(ctx context.Context, aggregateType enums.AggregateType, afterEventID int64, limit int) ([]db.EventStore, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	hasType := aggregateType.String() != ""
	hasLimit := limit > 0

	var rows []sqlcdb.EventStore
	var queryErr error

	switch {
	case hasType && hasLimit:
		rows, queryErr = r.q.GetEventsAfterEventIDByTypeLimited(ctx, sqlcdb.GetEventsAfterEventIDByTypeLimitedParams{
			EventID:       afterEventID,
			AggregateType: aggregateType,
			MerchantID:    merchantID,
			Limit:         int64(limit),
		})
	case hasType:
		rows, queryErr = r.q.GetEventsAfterEventIDByType(ctx, sqlcdb.GetEventsAfterEventIDByTypeParams{
			EventID:       afterEventID,
			AggregateType: aggregateType,
			MerchantID:    merchantID,
		})
	case hasLimit:
		rows, queryErr = r.q.GetEventsAfterEventIDLimited(ctx, sqlcdb.GetEventsAfterEventIDLimitedParams{
			EventID:    afterEventID,
			MerchantID: merchantID,
			Limit:      int64(limit),
		})
	default:
		rows, queryErr = r.q.GetEventsAfterEventID(ctx, sqlcdb.GetEventsAfterEventIDParams{
			EventID:    afterEventID,
			MerchantID: merchantID,
		})
	}

	if queryErr != nil {
		return nil, queryErr
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetSnapshot(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (*db.Snapshot, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetSnapshot(ctx, sqlcdb.GetSnapshotParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		MerchantID:    merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return db.SnapshotPtrFromSnapshot(row), nil
}

func (r *sqlcdbEventRepository) GetCheckpoint(ctx context.Context, projectionName string) (int64, error) {
	merchantId, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return 0, err
	}

	return r.q.GetCheckpoint(ctx, sqlcdb.GetCheckpointParams{
		MerchantID:     merchantId,
		ProjectionName: projectionName,
	})
}

func (r *sqlcdbEventRepository) Replay(ctx context.Context, fromEventID int64, aggregateType *enums.AggregateType) ([]db.EventStore, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	var rows []sqlcdb.EventStore
	var queryErr error

	if aggregateType != nil {
		rows, queryErr = r.q.ReplayByAggregate(ctx, sqlcdb.ReplayByAggregateParams{
			EventID:       fromEventID,
			AggregateType: *aggregateType,
			MerchantID:    merchantID,
		})
	} else {
		rows, queryErr = r.q.ReplayAllAggregates(ctx, sqlcdb.ReplayAllAggregatesParams{
			EventID:    fromEventID,
			MerchantID: merchantID,
		})
	}

	if queryErr != nil {
		return nil, queryErr
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetAllByMerchant(ctx context.Context) ([]db.EventStore, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllEventsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	return toDBEvents(rows), nil
}

func toDBEvents(rows []sqlcdb.EventStore) []db.EventStore {
	result := make([]db.EventStore, len(rows))
	for i, row := range rows {
		result[i] = db.EventStoreFromEventStore(row)
	}
	return result
}
