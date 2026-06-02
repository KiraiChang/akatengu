package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbFixedAssetCategoryRepo struct {
	q *sqlcdb.Queries
}

func NewFixedAssetCategoryRepo(q *sqlcdb.Queries) FixedAssetCategoryRepo {
	return &sqlcdbFixedAssetCategoryRepo{q: q}
}

func (r *sqlcdbFixedAssetCategoryRepo) InsertFixedAssetCategory(ctx context.Context, p sqlcdb.InsertFixedAssetCategoryParams) error {
	return r.q.InsertFixedAssetCategory(ctx, p)
}

func (r *sqlcdbFixedAssetCategoryRepo) UpdateFixedAssetCategory(ctx context.Context, p sqlcdb.UpdateFixedAssetCategoryParams) error {
	return r.q.UpdateFixedAssetCategory(ctx, p)
}

func (r *sqlcdbFixedAssetCategoryRepo) SoftDeleteFixedAssetCategory(ctx context.Context, p sqlcdb.SoftDeleteFixedAssetCategoryParams) error {
	return r.q.SoftDeleteFixedAssetCategory(ctx, p)
}
