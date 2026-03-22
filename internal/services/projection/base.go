package projection

import (
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"

	"akatengu/internal/model/db"
)

// ─────────────────────────────────────────
// Projection 結果
// ─────────────────────────────────────────

// Result 前面處理完的結果傳遞給下一個
type Result struct {
	values map[string]any
}

func NewResult() Result {
	return Result{values: map[string]any{}}
}

// Set 存入任意值
func (r Result) Set(key string, value any) Result {
	copied := map[string]any{}
	for k, v := range r.values {
		copied[k] = v
	}
	copied[key] = value
	return Result{values: copied}
}

// Get 取出，需要 type assert
func (r Result) Get(key string) (any, bool) {
	v, ok := r.values[key]
	return v, ok
}

// ─────────────────────────────────────────
// Projection 介面
// ─────────────────────────────────────────

// Projection 負責把事件轉成 read model（projection 表）
// 每個 projection 只處理自己關心的 event_type，其餘忽略
type Projection interface {
	Name() string
	Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, prevResult Result) (Result, error)
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
