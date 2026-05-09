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
	if err := r.q.BulkInsertBalanceSnapshot(ctx, closingId); err != nil {
		return err
	}
	parents, err := r.q.GetParentBalanceAggregations(ctx, closingId)
	if err != nil {
		return err
	}
	for _, p := range parents {
		if err := r.q.UpsertAccountBalanceSnapshot(ctx, sqlcdb.UpsertAccountBalanceSnapshotParams{
			ClosingID:   closingId,
			AccountID:   p.AccountID,
			DebitTotal:  p.DebitTotal,
			CreditTotal: p.CreditTotal,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *sqlcdbAccountBalanceSnapshotRepo) DeleteByClosingId(ctx context.Context, closingId int64) error {
	return r.q.DeleteAccountBalanceSnapshotsByClosingId(ctx, closingId)
}
