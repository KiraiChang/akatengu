package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type BankPdfTemplateService interface {
	GetAllTemplates(ctx context.Context) ([]projection.BankPdfTemplate, error)
	GetTemplateByUUID(ctx context.Context, uuid string) (*projection.BankPdfTemplate, error)
	GetTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankPdfTemplateLedger, error)
}

func NewBankPdfTemplateService(repo query.BankPdfTemplateRepo) BankPdfTemplateService {
	return &bankPdfTemplateService{r: repo}
}

type bankPdfTemplateService struct {
	r query.BankPdfTemplateRepo
}

func (s *bankPdfTemplateService) GetAllTemplates(ctx context.Context) ([]projection.BankPdfTemplate, error) {
	return s.r.GetBankPdfTemplates(ctx)
}

func (s *bankPdfTemplateService) GetTemplateByUUID(ctx context.Context, uuid string) (*projection.BankPdfTemplate, error) {
	return s.r.GetBankPdfTemplateByUUID(ctx, uuid)
}

func (s *bankPdfTemplateService) GetTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankPdfTemplateLedger, error) {
	return s.r.GetBankPdfTemplateLedgers(ctx, templateID)
}
