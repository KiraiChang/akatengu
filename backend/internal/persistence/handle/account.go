package handle

import (
	event_payload "akatengu/internal/domain/event"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	dbprojection "akatengu/internal/model/db/projection"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/shared/utils"
	"context"

	"github.com/jmoiron/sqlx"
)

// NewAccountProjector builds the accounts and ledger_accounts read models. The
// rows it writes are the same as the legacy AccountProjectionService
// (services/projection/account.go) writes for the same event.
func NewAccountProjector(db *sqlx.DB) Projector {
	c := &accountProjector{
		handles: make(map[event_types.EventType]handle),
	}
	c.handles[event_types.EventAccountBulkImported.Enum()] = c.applyAccountBulkImported
	c.handles[event_types.EventAccountCreated.Enum()] = c.applyAccountCreated
	c.handles[event_types.EventAccountUpdated.Enum()] = c.applyAccountUpdated
	c.handles[event_types.EventLedgerAccountCreated.Enum()] = c.applyLedgerCreated
	c.handles[event_types.EventLedgerAccountUpdated.Enum()] = c.applyLedgerUpdated
	return c
}

type accountProjector struct {
	handles map[event_types.EventType]handle
}

func (c accountProjector) EventTypes() []event_types.EventType {
	return utils.Keys(c.handles)
}

func (c accountProjector) Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	h, ok := c.handles[evt.EventType]
	if !ok {
		return errors.NewRuntimeError(errors.ErrProjectorNotFound, evt)
	}
	return h(ctx, tx, evt)
}

func (c accountProjector) Name() string {
	return "accountProjector"
}

// applyAccountBulkImported writes the accounts in payload order; a parent must
// come before its children, or the foreign key fails.
func (c *accountProjector) applyAccountBulkImported(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.AccountBulkImportedPayload](&evt)
	if err != nil {
		return err
	}
	updBy := toUpdatedBy(evt.UpdatedBy)
	for _, a := range p.Accounts {
		if err := tx.Projection.UpsertAccount(ctx, dbprojection.Account{
			MerchantID:       evt.MerchantID,
			AccountId:        a.AccountId,
			ParentId:         a.ParentId,
			Name:             a.Name,
			Type:             a.Type,
			NormalBalance:    a.NormalBalance,
			Currency:         coalesce(a.Currency, "TWD"),
			IsSummary:        a.IsSummary,
			IsActive:         a.IsActive,
			Note:             a.Note,
			CashFlowCategory: a.CashFlowCategory,
			UpdatedBy:        updBy,
			Version:          1, // same initial version as CreateAccount
		}); err != nil {
			return err
		}
	}
	return nil
}

// applyAccountCreated inserts at version 1: a duplicate account_id fails on
// the unique key instead of overwriting the account.
func (c *accountProjector) applyAccountCreated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.AccountCreatePayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.CreateAccount(ctx, dbprojection.Account{
		MerchantID:       evt.MerchantID,
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
		UpdatedBy:        toUpdatedBy(evt.UpdatedBy),
	})
}

// applyAccountUpdated keeps the versioned UPDATE of F6 as a second guard
// behind the mediator handler's version check (BFS R2 D9).
func (c *accountProjector) applyAccountUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.AccountUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpdateAccount(ctx, dbprojection.Account{
		MerchantID:       evt.MerchantID,
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
		UpdatedBy:        toUpdatedBy(evt.UpdatedBy),
	})
}

// applyLedgerCreated names the ledger account after its creation event's UUID,
// the same as the legacy projection (BFS R2 D12).
func (c *accountProjector) applyLedgerCreated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.LedgerAccountCreatePayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.CreateLedgerAccount(ctx, dbprojection.LedgerAccount{
		MerchantID:  evt.MerchantID,
		LedgerUuid:  evt.Uuid,
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
		UpdatedBy:   toUpdatedBy(evt.UpdatedBy),
	})
}

func (c *accountProjector) applyLedgerUpdated(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.LedgerAccountUpdatedPayload](&evt)
	if err != nil {
		return err
	}
	return tx.Projection.UpdateLedgerAccount(ctx, dbprojection.LedgerAccount{
		MerchantID:  evt.MerchantID,
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
		UpdatedBy:   toUpdatedBy(evt.UpdatedBy),
	})
}
