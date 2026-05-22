package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

type FixedAssetQueryRepo interface {
	GetFixedAssetByID(ctx context.Context, id int64) (*projection.FixedAsset, error)
	GetActiveFixedAssetsByMerchant(ctx context.Context) ([]projection.FixedAsset, error)
	GetAllFixedAssetsByMerchant(ctx context.Context) ([]projection.FixedAsset, error)
	GetFixedAssetDepreciationsByAssetID(ctx context.Context, assetID int64) ([]projection.FixedAssetDepreciation, error)
}

type sqlcdbFixedAssetRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbFixedAssetRepo) GetFixedAssetByID(ctx context.Context, id int64) (*projection.FixedAsset, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetFixedAssetByID(ctx, sqlcdb.GetFixedAssetByIDParams{
		ID:         id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.FixedAssetPtrFromFixedAsset(row), nil
}

func (r *sqlcdbFixedAssetRepo) GetActiveFixedAssetsByMerchant(ctx context.Context) ([]projection.FixedAsset, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetActiveFixedAssetsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.FixedAsset, len(rows))
	for i, row := range rows {
		result[i] = projection.FixedAssetFromFixedAsset(row)
	}
	return result, nil
}

func (r *sqlcdbFixedAssetRepo) GetAllFixedAssetsByMerchant(ctx context.Context) ([]projection.FixedAsset, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllFixedAssetsByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.FixedAsset, len(rows))
	for i, row := range rows {
		result[i] = projection.FixedAssetFromFixedAsset(row)
	}
	return result, nil
}

func (r *sqlcdbFixedAssetRepo) GetFixedAssetDepreciationsByAssetID(ctx context.Context, assetID int64) ([]projection.FixedAssetDepreciation, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetFixedAssetDepreciationsByAssetID(ctx, sqlcdb.GetFixedAssetDepreciationsByAssetIDParams{
		AssetID:    assetID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.FixedAssetDepreciation, len(rows))
	for i, row := range rows {
		result[i] = projection.FixedAssetDepreciationFromFixedAssetDepreciation(row)
	}
	return result, nil
}

func newFixedAssetQueryRepo(q *sqlcdb.Queries) FixedAssetQueryRepo {
	return &sqlcdbFixedAssetRepo{q: q}
}
