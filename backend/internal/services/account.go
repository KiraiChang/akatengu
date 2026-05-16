package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type AccountService interface {
	GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error)
	GetChildrenAccount(ctx context.Context, parentId string) ([]projection.Account, error)
	GetLedgerPaged(ctx context.Context, req model.PaginationParams) ([]projection.LedgerAccount, int64, error)
	GetAllAccount(ctx context.Context) ([]projection.Account, error)
	GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error)
	GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error)
	GetAllAccountBalances(ctx context.Context) ([]projection.AccountBalance, error)
}

type accountServiceImpl struct {
	r query.AccountRepo
	b query.RunningBalanceRepo
}

func (a accountServiceImpl) GetAllAccountBalances(ctx context.Context) ([]projection.AccountBalance, error) {
	return a.b.GetAllAccount(ctx)
}

func (a accountServiceImpl) GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error) {
	return a.b.GetAllLedger(ctx)
}

func (a accountServiceImpl) GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error) {
	return a.r.GetAllLedgers(ctx)
}

func (a accountServiceImpl) GetAllAccount(ctx context.Context) ([]projection.Account, error) {
	return a.r.GetAllAccounts(ctx)
}

func (a accountServiceImpl) GetLedgerPaged(ctx context.Context, req model.PaginationParams) ([]projection.LedgerAccount, int64, error) {
	return a.r.GetLedgerPaged(ctx, req)
}

func (a accountServiceImpl) GetChildrenAccount(ctx context.Context, parentId string) ([]projection.Account, error) {
	return a.r.GetChildrenAccount(ctx, parentId)
}

func (a accountServiceImpl) GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error) {
	return a.r.GetAccountPaged(ctx, params)
}

func NewAccountService(r query.AccountRepo, b query.RunningBalanceRepo) AccountService {
	return &accountServiceImpl{r, b}
}
