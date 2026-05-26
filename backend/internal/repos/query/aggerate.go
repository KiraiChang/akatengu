package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/jmoiron/sqlx"
)

type AggerateRepo interface {
	GetVersion(ctx context.Context, t enums.AggregateType, id string) (int64, error)
}

type aggerateRepo struct {
	q *sqlcdb.Queries
}

func (a *aggerateRepo) GetVersion(ctx context.Context, t enums.AggregateType, id string) (int64, error) {
	merchantId, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return 0, err
	}
	return a.q.GetAggregateVersion(ctx, sqlcdb.GetAggregateVersionParams{
		AggregateType: t,
		AggregateID:   id,
		MerchantID:    merchantId,
	})
}

func newAggerateRepo(q *sqlcdb.Queries) AggerateRepo {
	return &aggerateRepo{q: q}
}

func NewAggerateRepo(db *sqlx.DB) AggerateRepo {
	return newAggerateRepo(sqlcdb.New(db))
}
