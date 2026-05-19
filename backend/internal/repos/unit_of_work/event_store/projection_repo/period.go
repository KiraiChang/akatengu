package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
)

type sqlxPeriodCloseRepo struct {
	q *sqlcdb.Queries
}

func NewPeriodCloseRepo(q *sqlcdb.Queries) PeriodCloseRepo {
	return &sqlxPeriodCloseRepo{q}
}

func (r *sqlxPeriodCloseRepo) InsertPeriodClose(ctx context.Context, c projection.PeriodClosing) (*int64, error) {
	id, err := r.q.InsertPeriodClose(ctx, c.ToInsertPeriodCloseParams())
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (r *sqlxPeriodCloseRepo) UpdatePeriodCloseSnapshotByPeriod(ctx context.Context,
	id int64, snapshot string, closedAt *string, updatedBy *string) error {
	return r.q.UpdatePeriodCloseSnapshot(ctx, sqlcdb.UpdatePeriodCloseSnapshotParams{
		Status:    enums.PeriodTypeStatusClosed.Enum(),
		ClosedAt:  closedAt,
		Snapshot:  &snapshot,
		UpdatedBy: updatedBy,
		ClosingID: id,
	})
}

func (r *sqlxPeriodCloseRepo) ReopenPeriodClose(ctx context.Context, id int64, reason string, at string, updatedBy *string) error {
	return r.q.ReopenPeriodClose(ctx, sqlcdb.ReopenPeriodCloseParams{
		Status:       enums.PeriodTypeStatusReopened.Enum(),
		ReopenReason: &reason,
		ReopenAt:     &at,
		UpdatedBy:    updatedBy,
		ClosingID:    id,
	})
}

func (r *sqlxPeriodCloseRepo) UpdatePeriodCloseTxnId(ctx context.Context, id int64, closingTxnId *int64, openingTxnId *int64, updatedBy *string) error {
	return r.q.UpdatePeriodCloseTxnID(ctx, sqlcdb.UpdatePeriodCloseTxnIDParams{
		ClosingID:    id,
		ClosingTxnID: closingTxnId,
		OpeningTxnID: openingTxnId,
		UpdatedBy:    updatedBy,
	})
}
