package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbLedgerAccountBalanceSnapshotRepo struct {
	q *sqlcdb.Queries
}

func NewLedgerAccountBalanceSnapshotRepo(q *sqlcdb.Queries) LedgerAccountBalanceSnapshotRepo {
	return &sqlcdbLedgerAccountBalanceSnapshotRepo{q: q}
}

func (r *sqlcdbLedgerAccountBalanceSnapshotRepo) BulkInsert(ctx context.Context, merchantID int64, closingId int64) error {
	return r.q.BulkInsertLedgerBalanceSnapshot(ctx, sqlcdb.BulkInsertLedgerBalanceSnapshotParams{
		MerchantID: merchantID,
		ClosingID:  closingId,
	})
}

func (r *sqlcdbLedgerAccountBalanceSnapshotRepo) DeleteByClosingId(ctx context.Context, merchantID int64, closingId int64) error {
	return r.q.DeleteLedgerAccountBalanceSnapshotsByClosingId(ctx, sqlcdb.DeleteLedgerAccountBalanceSnapshotsByClosingIdParams{
		ClosingID:  closingId,
		MerchantID: merchantID,
	})
}
