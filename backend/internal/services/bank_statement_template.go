package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type BankStatementTemplateService interface {
	GetAllTemplates(ctx context.Context) ([]projection.BankCsvTemplate, error)
	GetTemplateByID(ctx context.Context, id int64) (*projection.BankCsvTemplate, error)
	GetTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankCsvTemplateLedger, error)
}

func NewBankStatementTemplateService(repo query.BankCsvTemplateRepo) BankStatementTemplateService {
	return &bankStatementTemplateService{r: repo}
}

type bankStatementTemplateService struct {
	r query.BankCsvTemplateRepo
}

func (s *bankStatementTemplateService) GetAllTemplates(ctx context.Context) ([]projection.BankCsvTemplate, error) {
	return s.r.GetBankCsvTemplates(ctx)
}

func (s *bankStatementTemplateService) GetTemplateByID(ctx context.Context, id int64) (*projection.BankCsvTemplate, error) {
	return s.r.GetBankCsvTemplateByID(ctx, id)
}

func (s *bankStatementTemplateService) GetTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankCsvTemplateLedger, error) {
	return s.r.GetBankCsvTemplateLedgers(ctx, templateID)
}
