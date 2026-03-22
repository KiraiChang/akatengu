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

func (s *AccountProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, prevResult Result) (Result, error) {
	switch event.EventType.String() {
	case event_types.EventAccountCreated.String():
		return s.applyAccountCreate(ctx, tx, event, prevResult)
	case event_types.EventLedgerAccountCreated.String():
		return s.applyLedgerCreate(ctx, tx, event, prevResult)
	}
	return prevResult, nil
}

func (s *AccountProjectionService) applyAccountCreate(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.AccountCreatePayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.CreateAccount(ctx, projection.Account{
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
		return result, err
	}

	return result, nil
}

func (s *AccountProjectionService) applyLedgerCreate(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.LedgerAccountCreatePayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	if err := tx.Projection.CreateLedgerAccount(ctx, projection.LedgerAccount{
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
		return result, err
	}

	return result, nil
}
