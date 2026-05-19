package event_store

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/db"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"context"
	"encoding/json"
)

type InsertEventParams struct {
	AggregateType    enums.AggregateType
	AggregateID      string
	AggregateVersion int64
	EventType        event_types.EventType
	Payload          json.RawMessage
	Metadata         json.RawMessage
	UpdatedBy        *string
}

type UnitOfWork interface {
	// Do 開啟 transaction，執行 fn，自動 commit 或 rollback
	Do(ctx context.Context, fn func(tx EventStoreRepositories) error) error
}

// EventStoreRepositories 是 transaction 內可用的所有 repository
// Service 只看這個介面，不知道底層是 sqlx
type EventStoreRepositories struct {
	Event      EventRepository
	Version    VersionRepository
	Snap       SnapshotRepository
	Check      CheckpointRepository
	Projection *projection_repo.TxProjectionRepository
	Truncate   TruncateRepository
}

// TruncateRepository 提供重建 projection 前的清除能力
type TruncateRepository interface {
	TruncateProjections(ctx context.Context) error
}

// EventRepository 是 transaction 內的操作，不需要傳 tx，由 UnitOfWork 管理
type EventRepository interface {
	Insert(ctx context.Context, p InsertEventParams) (int64, error)
}

type VersionRepository interface {
	Get(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (int64, error)
	Insert(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) error
	UpdateIfVersionMatch(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) (int64, error)
}

type SnapshotRepository interface {
	Upsert(ctx context.Context, snap db.Snapshot) error
}

type CheckpointRepository interface {
	Update(ctx context.Context, projectionName string, lastEventID int64) error
	Upsert(ctx context.Context, projectionName string, lastEventID int64) error
}
