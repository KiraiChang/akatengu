package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"context"
)

type sqlxAccountRepo struct {
	q *sqlcdb.Queries
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
	err := r.q.UpdateAccount(ctx, p.ToUpdateAccountParams())
	return err
}

func (r *sqlxAccountRepo) CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	err := r.q.CreateLedgerAccount(ctx, p.ToCreateLedgerAccountParams())
	return err
}

func (r *sqlxAccountRepo) UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	err := r.q.UpdateLedgerAccount(ctx, p.ToUpdateLedgerAccountParams())
	return err
}
