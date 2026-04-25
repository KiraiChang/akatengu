package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type AccountRepo interface {
	GetAccount(ctx context.Context, id string) (*projection.Account, error)
	GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error)
	GetAllAccounts(ctx context.Context) ([]projection.Account, error)
	GetAccountPaged(ctx context.Context, params model.PaginationParams) (*model.PaginatedResponse[projection.Account], error)
	GetChildrenAccount(ctx context.Context, id string) ([]projection.Account, error)
}

type sqlxAccountRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlxAccountRepo) GetChildrenAccount(ctx context.Context, id string) ([]projection.Account, error) {
	rows, err := r.q.GetChildrenAccount(ctx, &id)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromAccount(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAccountPaged(ctx context.Context, params model.PaginationParams) (*model.PaginatedResponse[projection.Account], error) {
	rows, err := r.q.GetAccountsPaged(ctx, sqlcdb.GetAccountsPagedParams{
		PageSize: params.PageSize,
		Offset:   params.Page,
	})
	result := &model.PaginatedResponse[projection.Account]{
		Data: make([]projection.Account, len(rows)),
		Meta: model.PaginatedMeta{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		result.Data[i] = projection.AccountFromGetAccountsPagedRow(row)
	}
	result.SetTotalCount(total)
	return result, nil
}

func newAccountRepo(q *sqlcdb.Queries) AccountRepo {
	return &sqlxAccountRepo{q: q}
}

func NewAccountRepo(db *sqlx.DB) AccountRepo {
	return &sqlxAccountRepo{q: sqlcdb.New(db)}
}

func (r *sqlxAccountRepo) GetAllAccounts(ctx context.Context) ([]projection.Account, error) {
	rows, err := r.q.GetAllAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromAccount(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAccount(ctx context.Context, id string) (*projection.Account, error) {
	row, err := r.q.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	a := projection.AccountFromAccount(row)
	return &a, nil
}

func (r *sqlxAccountRepo) GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error) {
	row, err := r.q.GetLedger(ctx, id)
	if err != nil {
		return nil, err
	}
	la := projection.LedgerAccountFromLedgerAccount(row)
	return &la, nil
}
