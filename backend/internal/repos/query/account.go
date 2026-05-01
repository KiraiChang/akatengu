package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type AccountRepo interface {
	GetAccount(ctx context.Context, id string) (*projection.Account, error)
	GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error)
	GetAllAccounts(ctx context.Context) ([]projection.Account, error)
	GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error)
	GetChildrenAccount(ctx context.Context, id string) ([]projection.Account, error)
	GetLedgerPaged(ctx context.Context, req model.PaginationParams) ([]projection.LedgerAccount, int64, error)
	GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error)
	GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error)
}

type sqlxAccountRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlxAccountRepo) GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error) {
	rows, err := r.q.GetAllLedgerBalances(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccountBalance, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountBalanceFromGetAllLedgerBalancesRow(row)
		if result[i].NormalBalance.Is(enums.BalanceDebit) {
			result[i].Balance = result[i].DebitTotal.Sub(result[i].CreditTotal)
		} else {
			result[i].Balance = result[i].CreditTotal.Sub(result[i].DebitTotal)
		}
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetLedgerPaged(ctx context.Context, params model.PaginationParams) ([]projection.LedgerAccount, int64, error) {
	rows, err := r.q.GetLedgerPaged(ctx, sqlcdb.GetLedgerPagedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.LedgerAccount, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountFromGetLedgerPagedRow(row)
	}
	return result, total, nil
}

func (r *sqlxAccountRepo) GetChildrenAccount(ctx context.Context, id string) ([]projection.Account, error) {
	rows, err := r.q.GetChildrenAccount(ctx, &id)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromGetChildrenAccountRow(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error) {
	rows, err := r.q.GetAccountPaged(ctx, sqlcdb.GetAccountPagedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromGetAccountPagedRow(row)
	}
	return result, total, nil
}

func newAccountRepo(q *sqlcdb.Queries) AccountRepo {
	return &sqlxAccountRepo{q: q}
}

func NewAccountRepo(db *sqlx.DB) AccountRepo {
	return &sqlxAccountRepo{q: sqlcdb.New(db)}
}

func (r *sqlxAccountRepo) GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error) {
	rows, err := r.q.GetAllLedgers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccount, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountFromLedgerAccount(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAllAccounts(ctx context.Context) ([]projection.Account, error) {
	rows, err := r.q.GetAllAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromGetAllAccountsRow(row)
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
