package services

import (
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type AuditService interface {
	GetAggregateVersions(ctx context.Context) ([]projection.AggregateVersionAudit, error)
	GetEventStorePaged(ctx context.Context, req model.PaginationParams) ([]projection.EventStoreAudit, int64, error)
	GetCheckpoints(ctx context.Context) ([]projection.CheckpointAudit, error)
	GetSnapshots(ctx context.Context) ([]projection.SnapshotAudit, error)
	GetExchangeRates(ctx context.Context, currency string) ([]projection.ExchangeRate, error)
}

type auditService struct {
	r query.AuditRepo
}

func NewAuditService(db *sqlx.DB) AuditService {
	return &auditService{r: query.NewAuditRepo(db)}
}

func (s *auditService) GetAggregateVersions(ctx context.Context) ([]projection.AggregateVersionAudit, error) {
	return s.r.GetAggregateVersions(ctx)
}

func (s *auditService) GetEventStorePaged(ctx context.Context, req model.PaginationParams) ([]projection.EventStoreAudit, int64, error) {
	return s.r.GetEventStorePaged(ctx, req)
}

func (s *auditService) GetCheckpoints(ctx context.Context) ([]projection.CheckpointAudit, error) {
	return s.r.GetCheckpoints(ctx)
}

func (s *auditService) GetSnapshots(ctx context.Context) ([]projection.SnapshotAudit, error) {
	return s.r.GetSnapshots(ctx)
}

func (s *auditService) GetExchangeRates(ctx context.Context, currency string) ([]projection.ExchangeRate, error) {
	return s.r.GetExchangeRates(ctx, currency)
}
