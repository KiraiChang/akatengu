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

func (s *AccountProjectionService) Name() string { return enums.ProjectionTypeAccount.String() }

func (s *AccountProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	// Account
	case event_types.EventAccountCreated:
		return s.applyAccountCreate(ctx, tx, ct)
	case event_types.EventAccountUpdated:
		return s.applyAccountUpdated(ctx, tx, ct)

	// Ledger
	case event_types.EventLedgerAccountCreated:
		return s.applyLedgerCreate(ctx, tx, ct)
	case event_types.EventLedgerAccountUpdated:
		return s.applyLedgerUpdated(ctx, tx, ct)
	}
	return nil
}

func (s *AccountProjectionService) applyAccountCreate(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AccountCreatePayload](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.AccountRepo.CreateAccount(ctx, projection.Account{
		MerchantID:       ct.MerchantID,
		AccountId:        p.AccountId,
		ParentId:         p.ParentId,
		Name:             p.Name,
		Type:             p.Type,
		NormalBalance:    p.NormalBalance,
		Currency:         coalesce(p.Currency, "TWD"),
		IsSummary:        p.IsSummary,
		IsActive:         p.IsActive,
		Note:             p.Note,
		CashFlowCategory: p.CashFlowCategory,
		UpdatedBy:        updatedBy,
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

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.AccountRepo.CreateLedgerAccount(ctx, projection.LedgerAccount{
		MerchantID:  ct.MerchantID,
		AccountId:   p.AccountId,
		Institution: p.Institution,
		Name:        p.Name,
		Type:        p.Type,
		AccountNo:   p.AccountNo,
		Currency:    coalesce(p.Currency, "TWD"),
		CreditLimit: p.CreditLimit,
		BillingDay:  p.BillingDay,
		DueDay:      p.DueDay,
		IsActive:    p.IsActive,
		Note:        p.Note,
		UpdatedBy:   updatedBy,
	}); err != nil {
		return err
	}

	return nil
}

func (s *AccountProjectionService) applyAccountUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AccountUpdatedPayload](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.AccountRepo.UpdateAccount(ctx, projection.Account{
		MerchantID:       ct.MerchantID,
		AccountId:        p.AccountId,
		ParentId:         p.ParentId,
		Name:             p.Name,
		Type:             p.Type,
		NormalBalance:    p.NormalBalance,
		Currency:         coalesce(p.Currency, "TWD"),
		IsSummary:        p.IsSummary,
		IsActive:         p.IsActive,
		Note:             p.Note,
		Version:          p.Version,
		CashFlowCategory: p.CashFlowCategory,
		UpdatedBy:        updatedBy,
	}); err != nil {
		return err
	}

	return nil
}

func (s *AccountProjectionService) applyLedgerUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.LedgerAccountUpdatedPayload](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.AccountRepo.UpdateLedgerAccount(ctx, projection.LedgerAccount{
		MerchantID:  ct.MerchantID,
		LedgerId:    p.LedgerId,
		AccountId:   p.AccountId,
		Institution: p.Institution,
		Name:        p.Name,
		Type:        p.Type,
		AccountNo:   p.AccountNo,
		Currency:    coalesce(p.Currency, "TWD"),
		CreditLimit: p.CreditLimit,
		BillingDay:  p.BillingDay,
		DueDay:      p.DueDay,
		IsActive:    p.IsActive,
		Note:        p.Note,
		Version:     p.Version,
		UpdatedBy:   updatedBy,
	}); err != nil {
		return err
	}

	return nil
}
