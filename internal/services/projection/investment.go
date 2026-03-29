package projection

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/enums"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
)

// ─────────────────────────────────────────
// InvestmentProjectionService
// ─────────────────────────────────────────

type InvestmentProjectionService struct{}

func (s *InvestmentProjectionService) Name() string { return enums.AggregateTransaction.String() }

func (s *InvestmentProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	switch event.EventType.String() {
	}
	return nil
}
