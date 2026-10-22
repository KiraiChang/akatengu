package projection

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

// The category methods run the same sqlc queries as the legacy
// projection_repo/prepaid_category.go and fixed_asset_category.go; since F8 the
// update and soft delete increment the read-model version.

func (r *sqlxProjection) InsertPrepaidCategory(ctx context.Context, p sqlcdb.InsertPrepaidCategoryParams) error {
	return r.q.InsertPrepaidCategory(ctx, p)
}

func (r *sqlxProjection) UpdatePrepaidCategory(ctx context.Context, p sqlcdb.UpdatePrepaidCategoryParams) error {
	return r.q.UpdatePrepaidCategory(ctx, p)
}

func (r *sqlxProjection) SoftDeletePrepaidCategory(ctx context.Context, p sqlcdb.SoftDeletePrepaidCategoryParams) error {
	return r.q.SoftDeletePrepaidCategory(ctx, p)
}

func (r *sqlxProjection) InsertFixedAssetCategory(ctx context.Context, p sqlcdb.InsertFixedAssetCategoryParams) error {
	return r.q.InsertFixedAssetCategory(ctx, p)
}

func (r *sqlxProjection) UpdateFixedAssetCategory(ctx context.Context, p sqlcdb.UpdateFixedAssetCategoryParams) error {
	return r.q.UpdateFixedAssetCategory(ctx, p)
}

func (r *sqlxProjection) SoftDeleteFixedAssetCategory(ctx context.Context, p sqlcdb.SoftDeleteFixedAssetCategoryParams) error {
	return r.q.SoftDeleteFixedAssetCategory(ctx, p)
}
