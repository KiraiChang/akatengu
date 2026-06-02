package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbPrepaidCategoryRepo struct {
	q *sqlcdb.Queries
}

func NewPrepaidCategoryRepo(q *sqlcdb.Queries) PrepaidCategoryRepo {
	return &sqlcdbPrepaidCategoryRepo{q: q}
}

func (r *sqlcdbPrepaidCategoryRepo) InsertPrepaidCategory(ctx context.Context, p sqlcdb.InsertPrepaidCategoryParams) error {
	return r.q.InsertPrepaidCategory(ctx, p)
}

func (r *sqlcdbPrepaidCategoryRepo) UpdatePrepaidCategory(ctx context.Context, p sqlcdb.UpdatePrepaidCategoryParams) error {
	return r.q.UpdatePrepaidCategory(ctx, p)
}

func (r *sqlcdbPrepaidCategoryRepo) SoftDeletePrepaidCategory(ctx context.Context, p sqlcdb.SoftDeletePrepaidCategoryParams) error {
	return r.q.SoftDeletePrepaidCategory(ctx, p)
}
