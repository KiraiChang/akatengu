package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type InstallmentService interface {
	GetInstallmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Installment, int64, error)
	GetPaymentPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InstallmentPayment, int64, error)
}

func NewInstallmentService(repo query.InstallmentRepo) InstallmentService {
	return &installmentService{
		repo,
	}
}

type installmentService struct {
	r query.InstallmentRepo
}

func (i installmentService) GetInstallmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Installment, int64, error) {
	return i.r.GetInstallmentPaged(ctx, req)
}

func (i installmentService) GetPaymentPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InstallmentPayment, int64, error) {
	return i.r.GetPaymentPaged(ctx, req, id)
}
