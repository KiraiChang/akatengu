package projection

import (
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"

	"akatengu/internal/model/db"
)

// ─────────────────────────────────────────
// Projection 介面
// ─────────────────────────────────────────

// Projection 負責把事件轉成 read model（projection 表）
// 每個 projection 只處理自己關心的 event_type，其餘忽略
type Projection interface {
	Name() string
	Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error
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
