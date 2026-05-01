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
}

type accountServiceImpl struct {
	repo query.AccountRepo
}

func (a accountServiceImpl) GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error) {
	return a.repo.GetAllLedgerBalances(ctx)
}

func (a accountServiceImpl) GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error) {
	return a.repo.GetAllLedgers(ctx)
}

func (a accountServiceImpl) GetAllAccount(ctx context.Context) ([]projection.Account, error) {
	return a.repo.GetAllAccounts(ctx)
}

func (a accountServiceImpl) GetLedgerPaged(ctx context.Context, req model.PaginationParams) ([]projection.LedgerAccount, int64, error) {
	return a.repo.GetLedgerPaged(ctx, req)
}

func (a accountServiceImpl) GetChildrenAccount(ctx context.Context, parentId string) ([]projection.Account, error) {
	return a.repo.GetChildrenAccount(ctx, parentId)
}

func (a accountServiceImpl) GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error) {
	return a.repo.GetAccountPaged(ctx, params)
}

func NewAccountService(repo query.AccountRepo) AccountService {
	return &accountServiceImpl{repo}
}
