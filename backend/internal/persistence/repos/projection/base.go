package projection

import (
	"akatengu/internal/database/sqlcdb"
)

type Projection interface {
}

type sqlxProjection struct {
	q *sqlcdb.Queries
}

func NewProjection(q *sqlcdb.Queries) Projection {
	return &sqlxProjection{q: q}
}
