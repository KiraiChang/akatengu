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
	rows, err := r.q.GetEventsByAggregate(ctx, sqlcdb.GetEventsByAggregateParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
	})
	if err != nil {
		return nil, err
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetAfterVersion(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, afterVersion int64) ([]db.EventStore, error) {
	rows, err := r.q.GetEventsAfterVersion(ctx, sqlcdb.GetEventsAfterVersionParams{
		AggregateType:    aggregateType,
		AggregateID:      aggregateID,
		AggregateVersion: afterVersion,
	})
	if err != nil {
		return nil, err
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetAfterEventID(ctx context.Context, aggregateType enums.AggregateType, afterEventID int64, limit int) ([]db.EventStore, error) {
	hasType := aggregateType.String() != ""
	hasLimit := limit > 0

	var rows []sqlcdb.EventStore
	var err error

	switch {
	case hasType && hasLimit:
		rows, err = r.q.GetEventsAfterEventIDByTypeLimited(ctx, sqlcdb.GetEventsAfterEventIDByTypeLimitedParams{
			EventID:       afterEventID,
			AggregateType: aggregateType,
			Limit:         int64(limit),
		})
	case hasType:
		rows, err = r.q.GetEventsAfterEventIDByType(ctx, sqlcdb.GetEventsAfterEventIDByTypeParams{
			EventID:       afterEventID,
			AggregateType: aggregateType,
		})
	case hasLimit:
		rows, err = r.q.GetEventsAfterEventIDLimited(ctx, sqlcdb.GetEventsAfterEventIDLimitedParams{
			EventID: afterEventID,
			Limit:   int64(limit),
		})
	default:
		rows, err = r.q.GetEventsAfterEventID(ctx, afterEventID)
	}

	if err != nil {
		return nil, err
	}
	return toDBEvents(rows), nil
}

func (r *sqlcdbEventRepository) GetSnapshot(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (*db.Snapshot, error) {
	row, err := r.q.GetSnapshot(ctx, sqlcdb.GetSnapshotParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
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
	var rows []sqlcdb.EventStore
	var err error

	if aggregateType != nil {
		rows, err = r.q.ReplayByAggregate(ctx, sqlcdb.ReplayByAggregateParams{
			EventID:       fromEventID,
			AggregateType: *aggregateType,
		})
	} else {
		rows, err = r.q.ReplayAllAggregates(ctx, fromEventID)
	}

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
