package event_store

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
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
	Projection ProjectionRepository
}

// EventRepository 是 transaction 內的操作，不需要傳 tx，由 UnitOfWork 管理
type EventRepository interface {
	Insert(ctx context.Context, p InsertEventParams) (int64, error)
}

type VersionRepository interface {
	Get(ctx context.Context, aggregateType enums.AggregateType, aggregateID string) (int64, error)
	Upsert(ctx context.Context, aggregateType enums.AggregateType, aggregateID string, version int64) error
}

type SnapshotRepository interface {
	Upsert(ctx context.Context, snap db.Snapshot) error
}

type CheckpointRepository interface {
	Update(ctx context.Context, projectionName string, lastEventID int64) error
}

// ProjectionRepository 是 projection 表的寫入操作
type ProjectionRepository interface {
	// transaction
	UpsertTransaction(ctx context.Context, p projection.Transaction) error
	UpsertJournalEntries(ctx context.Context, entries []projection.Entry) error

	//// installment
	//UpsertInstallment(ctx context.Context, p ProjInstallment) error
	//UpsertInstallmentPayment(ctx context.Context, p ProjInstallmentPayment) error
	//
	//// reconciliation
	//UpsertReconciliation(ctx context.Context, p ProjReconciliation) error
	//UpsertReconciliationAdjustment(ctx context.Context, p ProjReconciliationAdjustment) error

	// account
	CreateAccount(ctx context.Context, p projection.Account) error
	UpdateAccount(ctx context.Context, p projection.Account) error
	CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error
	UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error

	//// 重建用
	//TruncateAll(ctx context.Context) error
	//TruncateByAggregate(ctx context.Context, aggregateType enums.AggregateType) error

	// investment
	CreateInvestment(ctx context.Context, p projection.Investment) error
	UpdateInvestment(ctx context.Context, p projection.Investment) error

	InsertClose(ctx context.Context, c db.PeriodClosing) error
	UpdateCloseStatus(ctx context.Context, closingID int64, status enums.ClosingStatus, closedAt *string) error
	UpdateCloseSnapshotByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string, snapshot string) error
}
