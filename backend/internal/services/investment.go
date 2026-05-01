package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type InvestmentService interface {
	GetInvestmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Investment, int64, error)
	GetOpenLotsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLot, int64, error)
	GetLotDisposalsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLotDisposals, int64, error)
	GetPosition(ctx context.Context, id int64) (*projection.InvestmentPosition, error)
}

type investmentService struct {
	r query.InvestmentRepo
}

func (i investmentService) GetOpenLotsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLot, int64, error) {
	return i.r.GetOpenLotsPaged(ctx, req, id)
}

func (i investmentService) GetLotDisposalsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLotDisposals, int64, error) {
	return i.r.GetLotDisposalsPaged(ctx, req, id)
}

func (i investmentService) GetPosition(ctx context.Context, id int64) (*projection.InvestmentPosition, error) {
	return i.r.GetPosition(ctx, id)
}

func (i investmentService) GetInvestmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Investment, int64, error) {
	return i.r.GetInvestmentPaged(ctx, req)
}

func NewInvestmentService(repo query.InvestmentRepo) InvestmentService {
	return &investmentService{
		r: repo,
	}
}
