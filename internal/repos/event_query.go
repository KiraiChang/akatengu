package repos

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// EventQueryRepository 是 transaction 外的查詢（read-only，不需要 tx）
type EventQueryRepository interface {
	GetByAggregate(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) ([]db.EventStore, error)
	GetAfterVersion(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, afterVersion int64) ([]db.EventStore, error)
	GetAfterEventID(ctx context.Context, aggregateType enums.AggregateType, afterEventID int64, limit int) ([]db.EventStore, error)
	GetSnapshot(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (*db.Snapshot, error)
	GetCheckpoint(ctx context.Context, projectionName string) (int64, error)
	Replay(ctx context.Context, fromEventID int64, aggregateType *enums.AggregateType) ([]db.EventStore, error)
}

// ─────────────────────────────────────────
// Read-only repository（不需要 tx）
// ─────────────────────────────────────────

type sqlxEventQueryRepository struct{ db *sqlx.DB }

func NewEventQueryRepository(db *sqlx.DB) EventQueryRepository {
	return &sqlxEventQueryRepository{db: db}
}

func (r *sqlxEventQueryRepository) GetByAggregate(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) ([]db.EventStore, error) {
	var events []db.EventStore
	err := r.db.SelectContext(ctx, &events, `
        SELECT * FROM event_store
        WHERE aggregate_type = ? AND aggregate_id = ?
        ORDER BY aggregate_version`,
		aggregateType, aggregateID,
	)
	return events, err
}

func (r *sqlxEventQueryRepository) GetAfterVersion(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, afterVersion int64) ([]db.EventStore, error) {
	var events []db.EventStore
	err := r.db.SelectContext(ctx, &events, `
        SELECT * FROM event_store
        WHERE aggregate_type    = ?
          AND aggregate_id      = ?
          AND aggregate_version > ?
        ORDER BY aggregate_version`,
		aggregateType, aggregateID, afterVersion,
	)
	return events, err
}

func (r *sqlxEventQueryRepository) GetAfterEventID(ctx context.Context, aggregateType enums.AggregateType, afterEventID int64, limit int) ([]db.EventStore, error) {
	var events []db.EventStore

	query := `
        SELECT * FROM event_store
        WHERE event_id > ?`
	args := []interface{}{afterEventID}

	if aggregateType.String() != "" {
		query += " AND aggregate_type = ?"
		args = append(args, aggregateType)
	}
	query += " ORDER BY event_id"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	err := r.db.SelectContext(ctx, &events, query, args...)
	return events, err
}

func (r *sqlxEventQueryRepository) GetSnapshot(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (*db.Snapshot, error) {
	var snap db.Snapshot
	err := r.db.QueryRowxContext(ctx, `
        SELECT * FROM snapshots
        WHERE aggregate_type = ? AND aggregate_id = ?`,
		aggregateType, aggregateID,
	).StructScan(&snap)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &snap, err
}

func (r *sqlxEventQueryRepository) GetCheckpoint(ctx context.Context, projectionName string) (int64, error) {
	var id int64
	err := r.db.QueryRowxContext(ctx, `
        SELECT last_event_id FROM projection_checkpoints
        WHERE projection_name = ?`,
		projectionName,
	).Scan(&id)
	return id, err
}

// Replay 從指定 event_id 之後重播事件（用於重建 projection）
func (es *sqlxEventQueryRepository) Replay(ctx context.Context, fromEventID int64, aggregateType *enums.AggregateType) ([]db.EventStore, error) {
	query := `
		SELECT event_id, event_uuid, occurred_at,
		       aggregate_type, aggregate_id, aggregate_version,
		       event_type, payload, metadata
		FROM event_store
		WHERE event_id > ?`

	args := []interface{}{fromEventID}
	if aggregateType != nil {
		query += " AND aggregate_type = ?"
		args = append(args, aggregateType)
	}
	query += " ORDER BY event_id"
	var events []db.EventStore
	err := es.db.SelectContext(ctx, &events, query, args...)
	if err != nil {
		return nil, err
	}
	return events, nil
}
