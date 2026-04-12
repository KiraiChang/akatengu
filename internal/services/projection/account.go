package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
)

// ─────────────────────────────────────────
// AccountProjectionService
// ─────────────────────────────────────────

type AccountProjectionService struct{}

func (s *AccountProjectionService) Name() string { return string(enums.AggregateAccount) }

func (s *AccountProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	// Account
	case event_types.EventAccountCreated:
		return s.applyAccountCreate(ctx, tx, ct)

	// Ledger
	case event_types.EventLedgerAccountCreated:
		return s.applyLedgerCreate(ctx, tx, ct)

	}
	return nil
}

func (s *AccountProjectionService) applyAccountCreate(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AccountCreatePayload](ct)
	if err != nil {
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

func (s *AccountProjectionService) applyLedgerCreate(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.LedgerAccountCreatePayload](ct)
	if err != nil {
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
