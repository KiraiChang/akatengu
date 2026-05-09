package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbAccountBalanceSnapshotRepo struct {
	q *sqlcdb.Queries
}

func NewAccountBalanceSnapshotRepo(q *sqlcdb.Queries) AccountBalanceSnapshotRepo {
	return &sqlcdbAccountBalanceSnapshotRepo{q: q}
}

func (r *sqlcdbAccountBalanceSnapshotRepo) BulkInsert(ctx context.Context, closingId int64) error {
	err := r.q.BulkInsertBalanceSnapshot(ctx, closingId)
	return err
}

func (r *sqlcdbAccountBalanceSnapshotRepo) DeleteByClosingId(ctx context.Context, closingId int64) error {
	return r.q.DeleteAccountBalanceSnapshotsByClosingId(ctx, closingId)
}
