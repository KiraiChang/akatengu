package projection

import (
	dbprojection "akatengu/internal/model/db/projection"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"context"
	"fmt"
)

// The account methods run the same sqlc queries as the legacy
// projection_repo/account.go, so both paths build the same read model.

func (r *sqlxProjection) CreateAccount(ctx context.Context, p dbprojection.Account) error {
	return r.q.CreateAccount(ctx, p.ToCreateAccountParams())
}

// UpdateAccount only matches the version the client read and increments it;
// no matching row is a conflict, with the message of the legacy path (F6).
func (r *sqlxProjection) UpdateAccount(ctx context.Context, p dbprojection.Account) error {
	n, err := r.q.UpdateAccount(ctx, p.ToUpdateAccountParams())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("account %s version %d: %w", p.AccountId, p.Version, projection_repo.ErrVersionConflict)
	}
	return nil
}

func (r *sqlxProjection) UpsertAccount(ctx context.Context, p dbprojection.Account) error {
	return r.q.UpsertAccount(ctx, p.ToUpsertAccountParams())
}

func (r *sqlxProjection) CreateLedgerAccount(ctx context.Context, p dbprojection.LedgerAccount) error {
	return r.q.CreateLedgerAccount(ctx, p.ToCreateLedgerAccountParams())
}

// UpdateLedgerAccount is versioned as UpdateAccount.
func (r *sqlxProjection) UpdateLedgerAccount(ctx context.Context, p dbprojection.LedgerAccount) error {
	n, err := r.q.UpdateLedgerAccount(ctx, p.ToUpdateLedgerAccountParams())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("ledger account %d version %d: %w", p.LedgerId, p.Version, projection_repo.ErrVersionConflict)
	}
	return nil
}
