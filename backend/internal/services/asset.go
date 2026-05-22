package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type FixedAssetService interface {
	GetAllFixedAssets(ctx context.Context) ([]projection.FixedAsset, error)
	GetActiveFixedAssets(ctx context.Context) ([]projection.FixedAsset, error)
	GetFixedAssetByID(ctx context.Context, id int64) (*projection.FixedAsset, error)
	GetFixedAssetDepreciations(ctx context.Context, assetID int64) ([]projection.FixedAssetDepreciation, error)
}

func NewFixedAssetService(repo query.FixedAssetQueryRepo) FixedAssetService {
	return &fixedAssetService{r: repo}
}

type fixedAssetService struct {
	r query.FixedAssetQueryRepo
}

func (s *fixedAssetService) GetAllFixedAssets(ctx context.Context) ([]projection.FixedAsset, error) {
	return s.r.GetAllFixedAssetsByMerchant(ctx)
}

func (s *fixedAssetService) GetActiveFixedAssets(ctx context.Context) ([]projection.FixedAsset, error) {
	return s.r.GetActiveFixedAssetsByMerchant(ctx)
}

func (s *fixedAssetService) GetFixedAssetByID(ctx context.Context, id int64) (*projection.FixedAsset, error) {
	return s.r.GetFixedAssetByID(ctx, id)
}

func (s *fixedAssetService) GetFixedAssetDepreciations(ctx context.Context, assetID int64) ([]projection.FixedAssetDepreciation, error) {
	return s.r.GetFixedAssetDepreciationsByAssetID(ctx, assetID)
}
