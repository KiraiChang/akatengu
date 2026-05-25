package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type TemplateService interface {
	GetTemplates(ctx context.Context, search string) ([]projection.TransactionTemplate, error)
	GetTemplate(ctx context.Context, id int64) (*projection.TransactionTemplateDetail, error)
	CreateTemplate(ctx context.Context, cmd query.CreateTemplateCmd) (*projection.TransactionTemplate, error)
	UpdateTemplate(ctx context.Context, id int64, cmd query.UpdateTemplateCmd) error
	DeleteTemplate(ctx context.Context, id int64) error
}

func NewTemplateService(db *sqlx.DB) TemplateService {
	return &templateService{r: query.NewTemplateRepo(db)}
}

type templateService struct {
	r query.TemplateRepo
}

func (s *templateService) GetTemplates(ctx context.Context, search string) ([]projection.TransactionTemplate, error) {
	return s.r.GetTemplates(ctx, search)
}

func (s *templateService) GetTemplate(ctx context.Context, id int64) (*projection.TransactionTemplateDetail, error) {
	return s.r.GetTemplate(ctx, id)
}

func (s *templateService) CreateTemplate(ctx context.Context, cmd query.CreateTemplateCmd) (*projection.TransactionTemplate, error) {
	return s.r.CreateTemplate(ctx, cmd)
}

func (s *templateService) UpdateTemplate(ctx context.Context, id int64, cmd query.UpdateTemplateCmd) error {
	return s.r.UpdateTemplate(ctx, id, cmd)
}

func (s *templateService) DeleteTemplate(ctx context.Context, id int64) error {
	return s.r.DeleteTemplate(ctx, id)
}
