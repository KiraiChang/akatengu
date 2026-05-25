package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type AccountAnalysisService interface {
	GetAccountDirectChildrenWithBalance(ctx context.Context, accountID string) ([]projection.AccountChildBalance, error)
	GetAccountJournalEntriesPaged(ctx context.Context, accountID, fromDate, toDate string, req model.PaginationParams) ([]projection.AccountJournalEntryRow, int64, error)
	GetAccountMonthlyBalances(ctx context.Context, accountID, fromMonth, toMonth string) ([]projection.AccountMonthlyBalance, error)
}

func NewAccountAnalysisService(db *sqlx.DB) AccountAnalysisService {
	return &accountAnalysisService{r: query.NewAccountAnalysisRepo(db)}
}

type accountAnalysisService struct {
	r query.AccountAnalysisRepo
}

func (s *accountAnalysisService) GetAccountDirectChildrenWithBalance(ctx context.Context, accountID string) ([]projection.AccountChildBalance, error) {
	return s.r.GetAccountDirectChildrenWithBalance(ctx, accountID)
}

func (s *accountAnalysisService) GetAccountJournalEntriesPaged(ctx context.Context, accountID, fromDate, toDate string, req model.PaginationParams) ([]projection.AccountJournalEntryRow, int64, error) {
	return s.r.GetAccountJournalEntriesPaged(ctx, accountID, fromDate, toDate, req)
}

func (s *accountAnalysisService) GetAccountMonthlyBalances(ctx context.Context, accountID, fromMonth, toMonth string) ([]projection.AccountMonthlyBalance, error) {
	return s.r.GetAccountMonthlyBalances(ctx, accountID, fromMonth, toMonth)
}
