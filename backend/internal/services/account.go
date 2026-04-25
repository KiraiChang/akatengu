package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type AccountService interface {
	GetAccountPaged(ctx context.Context, params model.PaginationParams) (*model.PaginatedResponse[projection.Account], error)
	GetChildrenAccount(ctx context.Context, parentId string) ([]projection.Account, error)
}

type accountServiceImpl struct {
	repo query.AccountRepo
}

func (a accountServiceImpl) GetChildrenAccount(ctx context.Context, parentId string) ([]projection.Account, error) {
	return a.repo.GetChildrenAccount(ctx, parentId)
}

func (a accountServiceImpl) GetAccountPaged(ctx context.Context, params model.PaginationParams) (*model.PaginatedResponse[projection.Account], error) {
	return a.repo.GetAccountPaged(ctx, params)
}

func NewAccountService(repo query.AccountRepo) AccountService {
	return &accountServiceImpl{repo}
}
