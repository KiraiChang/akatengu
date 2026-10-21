package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrVersionConflict is returned when a versioned UPDATE affects no row: the
// data does not exist, or it was changed after the client read its version.
// Projections run in the event's transaction, so the event is rolled back too.
var ErrVersionConflict = errors.New("data not found or modified by someone else, please reload and try again")

type sqlxAccountRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlxAccountRepo) UpsertAccount(ctx context.Context, p projection.Account) error {
	err := r.q.UpsertAccount(ctx, p.ToUpsertAccountParams())
	return err
}

func NewAccountRepo(q *sqlcdb.Queries) AccountRepo {
	return &sqlxAccountRepo{
		q: q,
	}
}

func (r *sqlxAccountRepo) CreateAccount(ctx context.Context, p projection.Account) error {
	err := r.q.CreateAccount(ctx, p.ToCreateAccountParams())
	return err
}

func (r *sqlxAccountRepo) UpdateAccount(ctx context.Context, p projection.Account) error {
	n, err := r.q.UpdateAccount(ctx, p.ToUpdateAccountParams())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("account %s version %d: %w", p.AccountId, p.Version, ErrVersionConflict)
	}
	return nil
}

func (r *sqlxAccountRepo) CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	err := r.q.CreateLedgerAccount(ctx, p.ToCreateLedgerAccountParams())
	return err
}

func (r *sqlxAccountRepo) UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	n, err := r.q.UpdateLedgerAccount(ctx, p.ToUpdateLedgerAccountParams())
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("ledger account %d version %d: %w", p.LedgerId, p.Version, ErrVersionConflict)
	}
	return nil
}

func (r *sqlxAccountRepo) GetAccountCFCategory(ctx context.Context, accountId string, merchantID int64) (enums.CashFlowCategory, error) {
	row, err := r.q.GetAccount(ctx, sqlcdb.GetAccountParams{
		AccountID:  accountId,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return enums.CashFlowCategory{}, nil
	}
	if err != nil {
		return enums.CashFlowCategory{}, err
	}
	return row.CashFlowCategory, nil
}
