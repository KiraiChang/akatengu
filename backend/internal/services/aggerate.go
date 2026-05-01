package services

import (
	"akatengu/internal/enums"
	"akatengu/internal/repos/query"
	"context"
	"database/sql"
	"errors"
)

type AggerateService interface {
	GetVersion(ctx context.Context, t enums.AggregateType, id string) (int64, error)
}

type aggerateService struct {
	r query.AggerateRepo
}

func (a aggerateService) GetVersion(ctx context.Context, t enums.AggregateType, id string) (int64, error) {
	result, err := a.r.GetVersion(ctx, t, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return result, nil
}

func NewAggerateService(r query.AggerateRepo) AggerateService {
	return &aggerateService{r: r}
}
