package services

import (
	"akatengu/internal/model/db/report"
	"akatengu/internal/repos/query"
	"context"
	"fmt"
)

type ReportService struct {
	repo *query.ReportRepo
}

func NewReportService(repo *query.ReportRepo) *ReportService {
	return &ReportService{repo: repo}
}

func (r *ReportService) GetBalanceSheet(ctx context.Context, reportDate string) (*report.BalanceSheet, error) {
	return r.repo.GetBalanceSheet(ctx, reportDate)
}

func (r *ReportService) GetIncomeStatement(ctx context.Context, startDate, endDate string) (*report.IncomeStatement, error) {
	return r.repo.GetIncomeStatement(ctx, startDate, endDate)
}

func (r *ReportService) Verify(ctx context.Context, date string) error {
	bs, err := r.GetBalanceSheet(ctx, date)
	if err != nil {
		return err
	}

	is, err := r.GetIncomeStatement(ctx, date[:4]+"-01-01", date)
	if err != nil {
		return err
	}

	// 資產 = 負債 + 淨資產 + 本期損益
	right := bs.TotalLiabilities.Add(bs.TotalEquity).Add(is.NetIncome)
	if !bs.TotalAssets.Sub(right).IsZero() {
		return fmt.Errorf(
			"balance sheet equation broken: assets=%.2f, liabilities+equity+net=%.2f",
			bs.TotalAssets, right,
		)
	}
	return nil
}
