package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"
)

type PrepaidCategoryService interface {
	GetAllPrepaidCategories(ctx context.Context) ([]projection.PrepaidCategory, error)
}

func NewPrepaidCategoryService(repo query.PrepaidCategoryQueryRepo) PrepaidCategoryService {
	return &prepaidCategoryService{r: repo}
}

type prepaidCategoryService struct {
	r query.PrepaidCategoryQueryRepo
}

func (s *prepaidCategoryService) GetAllPrepaidCategories(ctx context.Context) ([]projection.PrepaidCategory, error) {
	return s.r.GetAllPrepaidCategoriesByMerchant(ctx)
}
