package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"
)

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
