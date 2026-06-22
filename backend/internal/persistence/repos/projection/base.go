package projection

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type Projection interface {
	UpsertEntryCFCategory(ctx context.Context, p sqlcdb.UpsertEntryCFCategoryParams) error
}

type sqlxProjection struct {
	q *sqlcdb.Queries
}

func NewProjection(q *sqlcdb.Queries) Projection {
	return &sqlxProjection{q: q}
}
