package services

import (
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type PeriodService interface {
	GetPeriodPagedByType(ctx context.Context, periodType enums.PeriodType, req model.PaginationParams) ([]projection.PeriodClosing, int64, error)
}

type periodService struct {
	q query.PeriodRepo
}

func (p periodService) GetPeriodPagedByType(ctx context.Context, periodType enums.PeriodType, req model.PaginationParams) ([]projection.PeriodClosing, int64, error) {
	return p.q.GetPeriodPagedByType(ctx, periodType, req)
}

func NewPeriodService(db *sqlx.DB) PeriodService {
	return &periodService{q: query.NewPeriodRepo(db)}
}
