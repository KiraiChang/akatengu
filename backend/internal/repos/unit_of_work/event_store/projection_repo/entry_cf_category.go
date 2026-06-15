package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbEntryCFCategoryRepo struct {
	q *sqlcdb.Queries
}

func NewEntryCFCategoryRepo(q *sqlcdb.Queries) EntryCFCategoryRepo {
	return &sqlcdbEntryCFCategoryRepo{q: q}
}

func (r *sqlcdbEntryCFCategoryRepo) UpsertEntryCFCategory(ctx context.Context, p sqlcdb.UpsertEntryCFCategoryParams) error {
	return r.q.UpsertEntryCFCategory(ctx, p)
}
