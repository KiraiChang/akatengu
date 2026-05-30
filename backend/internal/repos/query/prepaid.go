package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

type PrepaidQueryRepo interface {
	GetPrepaidByID(ctx context.Context, id int64) (*projection.Prepaid, error)
	GetPrepaidByUUID(ctx context.Context, uuid string) (*projection.Prepaid, error)
	GetActivePrepaidsByMerchant(ctx context.Context) ([]projection.Prepaid, error)
	GetAllPrepaidsByMerchant(ctx context.Context) ([]projection.Prepaid, error)
	GetPrepaidAmortizationsByPrepaidID(ctx context.Context, prepaidID int64) ([]projection.PrepaidAmortization, error)
}

type sqlcdbPrepaidRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbPrepaidRepo) GetPrepaidByID(ctx context.Context, id int64) (*projection.Prepaid, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetPrepaidByID(ctx, sqlcdb.GetPrepaidByIDParams{
		ID:         id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PrepaidPtrFromPrepaid(row), nil
}

func (r *sqlcdbPrepaidRepo) GetPrepaidByUUID(ctx context.Context, uuid string) (*projection.Prepaid, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetPrepaidByUUID(ctx, sqlcdb.GetPrepaidByUUIDParams{
		PrepaidUuid: uuid,
		MerchantID:  merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PrepaidPtrFromPrepaid(row), nil
}

func (r *sqlcdbPrepaidRepo) GetActivePrepaidsByMerchant(ctx context.Context) ([]projection.Prepaid, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetActivePrepaidsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Prepaid, len(rows))
	for i, row := range rows {
		result[i] = projection.PrepaidFromPrepaid(row)
	}
	return result, nil
}

func (r *sqlcdbPrepaidRepo) GetAllPrepaidsByMerchant(ctx context.Context) ([]projection.Prepaid, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllPrepaidsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Prepaid, len(rows))
	for i, row := range rows {
		result[i] = projection.PrepaidFromPrepaid(row)
	}
	return result, nil
}

func (r *sqlcdbPrepaidRepo) GetPrepaidAmortizationsByPrepaidID(ctx context.Context, prepaidID int64) ([]projection.PrepaidAmortization, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetPrepaidAmortizationsByPrepaidID(ctx, sqlcdb.GetPrepaidAmortizationsByPrepaidIDParams{
		PrepaidID:  prepaidID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.PrepaidAmortization, len(rows))
	for i, row := range rows {
		result[i] = projection.PrepaidAmortizationFromPrepaidAmortization(row)
	}
	return result, nil
}

func newPrepaidQueryRepo(q *sqlcdb.Queries) PrepaidQueryRepo {
	return &sqlcdbPrepaidRepo{q: q}
}
