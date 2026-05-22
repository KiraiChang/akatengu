package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type PrepaidService interface {
	GetAllPrepaids(ctx context.Context) ([]projection.Prepaid, error)
	GetActivePrepaids(ctx context.Context) ([]projection.Prepaid, error)
	GetPrepaidByID(ctx context.Context, id int64) (*projection.Prepaid, error)
	GetPrepaidAmortizations(ctx context.Context, prepaidID int64) ([]projection.PrepaidAmortization, error)
}

func NewPrepaidService(repo query.PrepaidQueryRepo) PrepaidService {
	return &prepaidService{r: repo}
}

type prepaidService struct {
	r query.PrepaidQueryRepo
}

func (s *prepaidService) GetAllPrepaids(ctx context.Context) ([]projection.Prepaid, error) {
	return s.r.GetAllPrepaidsByMerchant(ctx)
}

func (s *prepaidService) GetActivePrepaids(ctx context.Context) ([]projection.Prepaid, error) {
	return s.r.GetActivePrepaidsByMerchant(ctx)
}

func (s *prepaidService) GetPrepaidByID(ctx context.Context, id int64) (*projection.Prepaid, error) {
	return s.r.GetPrepaidByID(ctx, id)
}

func (s *prepaidService) GetPrepaidAmortizations(ctx context.Context, prepaidID int64) ([]projection.PrepaidAmortization, error) {
	return s.r.GetPrepaidAmortizationsByPrepaidID(ctx, prepaidID)
}
