package projection

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

func (r *sqlxProjection) UpsertEntryCFCategory(ctx context.Context, p sqlcdb.UpsertEntryCFCategoryParams) error {
	return r.q.UpsertEntryCFCategory(ctx, p)
}
