package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

type FixedAssetCategoryQueryRepo interface {
	GetFixedAssetCategoryByUUID(ctx context.Context, uuid string) (*projection.FixedAssetCategory, error)
	GetAllFixedAssetCategoriesByMerchant(ctx context.Context) ([]projection.FixedAssetCategory, error)
}

type sqlcdbFixedAssetCategoryQueryRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbFixedAssetCategoryQueryRepo) GetFixedAssetCategoryByUUID(ctx context.Context, uuid string) (*projection.FixedAssetCategory, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetFixedAssetCategoryByUUID(ctx, sqlcdb.GetFixedAssetCategoryByUUIDParams{
		CategoryUuid: uuid,
		MerchantID:   merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.FixedAssetCategoryPtrFromFixedAssetCategory(row), nil
}

func (r *sqlcdbFixedAssetCategoryQueryRepo) GetAllFixedAssetCategoriesByMerchant(ctx context.Context) ([]projection.FixedAssetCategory, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllFixedAssetCategoriesByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.FixedAssetCategory, len(rows))
	for i, row := range rows {
		result[i] = projection.FixedAssetCategoryFromFixedAssetCategory(row)
	}
	return result, nil
}

func newFixedAssetCategoryQueryRepo(q *sqlcdb.Queries) FixedAssetCategoryQueryRepo {
	return &sqlcdbFixedAssetCategoryQueryRepo{q: q}
}
