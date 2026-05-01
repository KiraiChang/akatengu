package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
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
	return a.q.GetAggregateVersion(ctx, sqlcdb.GetAggregateVersionParams{
		AggregateType: t,
		AggregateID:   id,
	})
}

func newAggerateRepo(q *sqlcdb.Queries) AggerateRepo {
	return &aggerateRepo{q: q}
}

func NewAggerateRepo(db *sqlx.DB) AggerateRepo {
	return newAggerateRepo(sqlcdb.New(db))
}
