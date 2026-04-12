package projection

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ─────────────────────────────────────────
// Projection 介面
// ─────────────────────────────────────────

// Projection 負責把事件轉成 read model（projection 表）
// 每個 projection 只處理自己關心的 event_type，其餘忽略
type Projection interface {
	Name() string
	Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error
}

// ─────────────────────────────────────────
// helper
// ─────────────────────────────────────────

func coalesce(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func checkAndGetPayload[T any](ct *pipelines.Result) (*T, error) {
	p, ok := ct.Payload.(T)
	if !ok {
		return nil, fmt.Errorf("payload type error")
	}
	return &p, nil
}

func checkAndGetState[T any](ct *pipelines.Result) (*T, error) {
	p, ok := ct.State.(*T)
	if !ok {
		return nil, fmt.Errorf("state type error")
	}
	return p, nil
}

func NewProjection() []Projection {
	return []Projection{
		&InvestmentProjectionService{},
		&AccountProjectionService{},
		&PeriodProjectionService{},
		&TransactionProjectionService{},
	}
}
