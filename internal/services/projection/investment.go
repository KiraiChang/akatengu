package projection

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
	"encoding/json"
)

// ─────────────────────────────────────────
// InvestmentProjectionService
// ─────────────────────────────────────────

type InvestmentProjectionService struct{}

func (s *InvestmentProjectionService) Name() string { return enums.AggreagateInvestment.String() }

func (s *InvestmentProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, prevResult Result) (Result, error) {
	switch event.EventType.String() {
	case event_types.EventInvestmentCreate.String():
		return s.applyCreated(ctx, tx, event, prevResult)
	case event_types.EventInvestmentUpdate.String():
		return s.applyUpdate(ctx, tx, event, prevResult)
	}
	return prevResult, nil
}

func (s *InvestmentProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.InvestmentCreatedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.CreateInvestment(ctx, projection.Investment{
		AccountId:  p.AccountId,
		AssetType:  p.AssetType,
		Currency:   coalesce(p.Currency, "TWD"),
		Symbol:     p.Symbol,
		Name:       p.Name,
		CostMethod: p.CostMethod,
		IsActive:   p.IsActive,
	}); err != nil {
		return result, err
	}

	return result, nil
}

func (s *InvestmentProjectionService) applyUpdate(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.InvestmentUpdatedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.UpdateInvestment(ctx, projection.Investment{
		InvestmentId: p.InvestmentId,
		AccountId:    p.AccountId,
		AssetType:    p.AssetType,
		Currency:     coalesce(p.Currency, "TWD"),
		Symbol:       p.Symbol,
		Name:         p.Name,
		CostMethod:   p.CostMethod,
		IsActive:     p.IsActive,
	}); err != nil {
		return result, err
	}

	return result, nil
}
