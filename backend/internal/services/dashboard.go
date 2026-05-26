package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type DashboardService interface {
	GetSummary(ctx context.Context, today string) (projection.DashboardSummary, error)
	GetMonthlyTrend(ctx context.Context, fromMonth, toMonth string) ([]projection.MonthlyTrendItem, error)
	GetLedgerBalances(ctx context.Context) ([]projection.LedgerBalance, error)
}

type dashboardService struct {
	r query.DashboardRepo
}

func NewDashboardService(db *sqlx.DB) DashboardService {
	return &dashboardService{r: query.NewDashboardRepo(db)}
}

func (s *dashboardService) GetSummary(ctx context.Context, today string) (projection.DashboardSummary, error) {
	return s.r.GetSummary(ctx, today)
}

func (s *dashboardService) GetMonthlyTrend(ctx context.Context, fromMonth, toMonth string) ([]projection.MonthlyTrendItem, error) {
	return s.r.GetMonthlyTrend(ctx, fromMonth, toMonth)
}

func (s *dashboardService) GetLedgerBalances(ctx context.Context) ([]projection.LedgerBalance, error) {
	return s.r.GetLedgerBalances(ctx)
}
