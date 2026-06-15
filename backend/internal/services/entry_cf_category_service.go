package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type EntryCFCategoryService interface {
	ListTransactionCFReview(ctx context.Context, from string, to string) ([]projection.TransactionCFReview, error)
}

type entryCFCategoryService struct {
	repo *query.Repo
}

func NewEntryCFCategoryService(db *sqlx.DB) EntryCFCategoryService {
	return &entryCFCategoryService{
		repo: query.NewQueryRepository(db),
	}
}

func (s *entryCFCategoryService) ListTransactionCFReview(ctx context.Context, from string, to string) ([]projection.TransactionCFReview, error) {
	return s.repo.EntryCFCategory.ListTransactionCFReview(ctx, from, to)
}
