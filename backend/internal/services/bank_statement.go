package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type BankStatementService interface {
	GetImportsPaged(ctx context.Context, ledgerID int64, page, pageSize int64) ([]projection.BankStatementImport, int64, error)
	GetImportByID(ctx context.Context, id int64) (*projection.BankStatementImport, error)
	GetTxnsByImport(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error)
	GetTxnsForReview(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error)
	GetUnmatchedTxns(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error)
}

func NewBankStatementService(repo query.BankStatementImportRepo) BankStatementService {
	return &bankStatementService{r: repo}
}

type bankStatementService struct {
	r query.BankStatementImportRepo
}

func (s *bankStatementService) GetImportsPaged(ctx context.Context, ledgerID int64, page, pageSize int64) ([]projection.BankStatementImport, int64, error) {
	offset := (page - 1) * pageSize
	return s.r.GetBankStatementImportsPaged(ctx, ledgerID, pageSize, offset)
}

func (s *bankStatementService) GetImportByID(ctx context.Context, id int64) (*projection.BankStatementImport, error) {
	return s.r.GetBankStatementImportByID(ctx, id)
}

func (s *bankStatementService) GetTxnsByImport(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error) {
	return s.r.GetBankStatementTxnsByImport(ctx, importID)
}

func (s *bankStatementService) GetTxnsForReview(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error) {
	return s.r.GetBankTxnsForReview(ctx, importID)
}

func (s *bankStatementService) GetUnmatchedTxns(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error) {
	return s.r.GetUnmatchedBankTxns(ctx, importID)
}
