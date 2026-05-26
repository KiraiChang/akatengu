package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/jmoiron/sqlx"
)

type AuditRepo interface {
	GetAggregateVersions(ctx context.Context) ([]projection.AggregateVersionAudit, error)
	GetEventStorePaged(ctx context.Context, req model.PaginationParams) ([]projection.EventStoreAudit, int64, error)
	GetCheckpoints(ctx context.Context) ([]projection.CheckpointAudit, error)
	GetSnapshots(ctx context.Context) ([]projection.SnapshotAudit, error)
	GetExchangeRates(ctx context.Context, currency string) ([]projection.ExchangeRate, error)
}

type sqlcdbAuditRepo struct {
	q *sqlcdb.Queries
}

func newAuditRepo(q *sqlcdb.Queries) AuditRepo {
	return &sqlcdbAuditRepo{q: q}
}

func NewAuditRepo(db *sqlx.DB) AuditRepo {
	return &sqlcdbAuditRepo{q: sqlcdb.New(db)}
}

func (r *sqlcdbAuditRepo) GetAggregateVersions(ctx context.Context) ([]projection.AggregateVersionAudit, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAggregateVersions(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.AggregateVersionAudit, len(rows))
	for i, row := range rows {
		result[i] = projection.AggregateVersionAudit{
			AggregateType:  row.AggregateType,
			AggregateId:    row.AggregateID,
			MerchantID:     row.MerchantID,
			CurrentVersion: row.CurrentVersion,
		}
	}
	return result, nil
}

func (r *sqlcdbAuditRepo) GetEventStorePaged(ctx context.Context, req model.PaginationParams) ([]projection.EventStoreAudit, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetEventStorePaged(ctx, sqlcdb.GetEventStorePagedParams{
		MerchantID: merchantID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.EventStoreAudit, len(rows))
	for i, row := range rows {
		result[i] = projection.EventStoreAudit{
			EventId:          row.EventID,
			EventUuid:        row.EventUuid,
			OccurredAt:       row.OccurredAt,
			MerchantID:       row.MerchantID,
			AggregateType:    row.AggregateType,
			AggregateId:      row.AggregateID,
			AggregateVersion: row.AggregateVersion,
			EventType:        row.EventType,
			Payload:          row.Payload,
			Metadata:         row.Metadata,
			UpdatedBy:        row.UpdatedBy,
		}
	}
	return result, total, nil
}

func (r *sqlcdbAuditRepo) GetCheckpoints(ctx context.Context) ([]projection.CheckpointAudit, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetCheckpoints(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.CheckpointAudit, len(rows))
	for i, row := range rows {
		result[i] = projection.CheckpointAudit{
			ProjectionName: row.ProjectionName,
			MerchantID:     row.MerchantID,
			LastEventID:    row.LastEventID,
			UpdatedAt:      row.UpdatedAt,
		}
	}
	return result, nil
}

func (r *sqlcdbAuditRepo) GetSnapshots(ctx context.Context) ([]projection.SnapshotAudit, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetSnapshots(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.SnapshotAudit, len(rows))
	for i, row := range rows {
		result[i] = projection.SnapshotAudit{
			SnapshotId:    row.SnapshotID,
			MerchantID:    row.MerchantID,
			AggregateType: row.AggregateType,
			AggregateId:   row.AggregateID,
			AtVersion:     row.AtVersion,
			State:         row.State,
			CreatedAt:     row.CreatedAt,
		}
	}
	return result, nil
}

func (r *sqlcdbAuditRepo) GetExchangeRates(ctx context.Context, currency string) ([]projection.ExchangeRate, error) {
	var rows []sqlcdb.ExchangeRate
	var err error
	if currency != "" {
		rows, err = r.q.GetExchangeRatesByCurrency(ctx, currency)
	} else {
		rows, err = r.q.GetExchangeRates(ctx)
	}
	if err != nil {
		return nil, err
	}
	result := make([]projection.ExchangeRate, len(rows))
	for i, row := range rows {
		result[i] = projection.ExchangeRate{
			RateId:   row.RateID,
			Currency: row.Currency,
			RateDate: row.RateDate,
			RateTWD:  row.RateTwd,
			Source:   row.Source,
		}
	}
	return result, nil
}
