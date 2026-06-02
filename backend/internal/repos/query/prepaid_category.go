package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

type PrepaidCategoryQueryRepo interface {
	GetPrepaidCategoryByUUID(ctx context.Context, uuid string) (*projection.PrepaidCategory, error)
	GetAllPrepaidCategoriesByMerchant(ctx context.Context) ([]projection.PrepaidCategory, error)
}

type sqlcdbPrepaidCategoryQueryRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbPrepaidCategoryQueryRepo) GetPrepaidCategoryByUUID(ctx context.Context, uuid string) (*projection.PrepaidCategory, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetPrepaidCategoryByUUID(ctx, sqlcdb.GetPrepaidCategoryByUUIDParams{
		CategoryUuid: uuid,
		MerchantID:   merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PrepaidCategoryPtrFromPrepaidCategory(row), nil
}

func (r *sqlcdbPrepaidCategoryQueryRepo) GetAllPrepaidCategoriesByMerchant(ctx context.Context) ([]projection.PrepaidCategory, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllPrepaidCategoriesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.PrepaidCategory, len(rows))
	for i, row := range rows {
		result[i] = projection.PrepaidCategoryFromPrepaidCategory(row)
	}
	return result, nil
}

func newPrepaidCategoryQueryRepo(q *sqlcdb.Queries) PrepaidCategoryQueryRepo {
	return &sqlcdbPrepaidCategoryQueryRepo{q: q}
}
