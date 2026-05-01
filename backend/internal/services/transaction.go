package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type TransactionService interface {
	GetTransactionPaged(ctx context.Context, req model.PaginationParams) ([]projection.Transaction, int64, error)
	GetEntries(ctx context.Context, id int64) ([]projection.Entry, error)
}

type transactionService struct {
	repo query.TransactionRepo
}

func (t transactionService) GetEntries(ctx context.Context, id int64) ([]projection.Entry, error) {
	return t.repo.GetEntries(ctx, id)
}

func (t transactionService) GetTransactionPaged(ctx context.Context, req model.PaginationParams) ([]projection.Transaction, int64, error) {
	return t.repo.GetTransactionPaged(ctx, req)
}

func NewTransactionService(db *sqlx.DB) TransactionService {
	return &transactionService{
		repo: query.NewTransactionRepo(db),
	}
}
