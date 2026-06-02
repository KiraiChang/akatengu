package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type FixedAssetCategoryService interface {
	GetAllFixedAssetCategories(ctx context.Context) ([]projection.FixedAssetCategory, error)
}

func NewFixedAssetCategoryService(repo query.FixedAssetCategoryQueryRepo) FixedAssetCategoryService {
	return &fixedAssetCategoryService{r: repo}
}

type fixedAssetCategoryService struct {
	r query.FixedAssetCategoryQueryRepo
}

func (s *fixedAssetCategoryService) GetAllFixedAssetCategories(ctx context.Context) ([]projection.FixedAssetCategory, error) {
	return s.r.GetAllFixedAssetCategoriesByMerchant(ctx)
}
