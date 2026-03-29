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
// AccountProjectionService
// ─────────────────────────────────────────

type AccountProjectionService struct{}

func (s *AccountProjectionService) Name() string { return enums.AggregateAccount.String() }

func (s *AccountProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	switch event.EventType.String() {
	// Account
	case event_types.EventAccountCreated.String():
		return s.applyAccountCreate(ctx, tx, event)

	// Ledger
	case event_types.EventLedgerAccountCreated.String():
		return s.applyLedgerCreate(ctx, tx, event)

	// Investment
	case event_types.EventInvestmentCreate.String():
		return s.applyCreated(ctx, tx, event)
	case event_types.EventInvestmentUpdate.String():
		return s.applyUpdate(ctx, tx, event)
	}
	return nil
}

func (s *AccountProjectionService) applyAccountCreate(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.AccountCreatePayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.AccountRepo.CreateAccount(ctx, projection.Account{
		AccountId:     p.AccountId,
		ParentId:      p.ParentId,
		Name:          p.Name,
		Type:          p.Type,
		NormalBalance: p.NormalBalance,
		Currency:      coalesce(p.Currency, "TWD"),
		IsSummary:     p.IsSummary,
		IsActive:      p.IsActive,
		Note:          p.Note,
	}); err != nil {
		return err
	}

	return nil
}

func (s *AccountProjectionService) applyLedgerCreate(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.LedgerAccountCreatePayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if err := tx.Projection.AccountRepo.CreateLedgerAccount(ctx, projection.LedgerAccount{
		AccountId:   p.AccountId,
		Institution: p.Institution,
		Name:        p.Name,
		AccountNo:   p.AccountNo,
		Currency:    coalesce(p.Currency, "TWD"),
		CreditLimit: p.CreditLimit,
		BillingDay:  p.BillingDay,
		DueDay:      p.DueDay,
		IsActive:    p.IsActive,
		Note:        p.Note,
	}); err != nil {
		return err
	}

	return nil
}

func (s *AccountProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.InvestmentCreatedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.InvestmentRepo.CreateInvestment(ctx, projection.Investment{
		AccountId:  p.AccountId,
		AssetType:  p.AssetType,
		Currency:   coalesce(p.Currency, "TWD"),
		Symbol:     p.Symbol,
		Name:       p.Name,
		CostMethod: p.CostMethod,
		IsActive:   p.IsActive,
	}); err != nil {
		return err
	}

	return nil
}

func (s *AccountProjectionService) applyUpdate(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.InvestmentUpdatedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.InvestmentRepo.UpdateInvestment(ctx, projection.Investment{
		InvestmentId: p.InvestmentId,
		AccountId:    p.AccountId,
		AssetType:    p.AssetType,
		Currency:     coalesce(p.Currency, "TWD"),
		Symbol:       p.Symbol,
		Name:         p.Name,
		CostMethod:   p.CostMethod,
		IsActive:     p.IsActive,
	}); err != nil {
		return err
	}

	return nil
}
