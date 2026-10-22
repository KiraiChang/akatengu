package projection

import (
	"akatengu/internal/database/sqlcdb"
	dbprojection "akatengu/internal/model/db/projection"
	"context"
)

type Projection interface {
	UpsertEntryCFCategory(ctx context.Context, p sqlcdb.UpsertEntryCFCategoryParams) error

	CreateAccount(ctx context.Context, p dbprojection.Account) error
	UpdateAccount(ctx context.Context, p dbprojection.Account) error
	UpsertAccount(ctx context.Context, p dbprojection.Account) error
	CreateLedgerAccount(ctx context.Context, p dbprojection.LedgerAccount) error
	UpdateLedgerAccount(ctx context.Context, p dbprojection.LedgerAccount) error
}

type sqlxProjection struct {
	q *sqlcdb.Queries
}

func NewProjection(q *sqlcdb.Queries) Projection {
	return &sqlxProjection{q: q}
}
