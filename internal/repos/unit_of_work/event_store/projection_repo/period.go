package projection_repo

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxPeriodCloseRepo struct {
	tx *sqlx.Tx
}

func NewPeriodCloseRepo(tx *sqlx.Tx) PeriodCloseRepo {
	return &sqlxPeriodCloseRepo{tx: tx}
}

func (r *sqlxPeriodCloseRepo) InsertPeriodClose(ctx context.Context, c projection.PeriodClosing) (*int64, error) {
	result, err := r.tx.ExecContext(ctx, `
        INSERT INTO period_closings
            (period_type, period_start, period_end, status, note)
        VALUES (?, ?, ?, ?, ?)`,
		c.PeriodType, c.PeriodStart, c.PeriodEnd, c.Status, c.Note,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (r *sqlxPeriodCloseRepo) UpdatePeriodCloseSnapshotByPeriod(ctx context.Context,
	id int64, snapshot string, closedAt *string) error {
	_, err := r.tx.ExecContext(ctx, `
        UPDATE period_closings
        SET status = ?, closed_at = ?, snapshot = ?
        WHERE closing_id = ?`,
		enums.PeriodTypeStatusClosed, closedAt, snapshot, id,
	)
	return err
}

func (r *sqlxPeriodCloseRepo) ReopenPeriodClose(ctx context.Context, id int64, reason string, at string) error {
	_, err := r.tx.ExecContext(ctx, `
        UPDATE period_closings
        SET status = ?, reopen_reason = ?, reopen_at = ?
        WHERE closing_id = ?`,
		enums.PeriodTypeStatusReopened, reason, at, id,
	)
	return err
}

func (r *sqlxPeriodCloseRepo) UpdatePeriodCloseTxnId(ctx context.Context, id int64, closingTxnId *int64, openingTxnId *int64) error {
	_, err := r.tx.ExecContext(ctx, `
        UPDATE period_closings
        SET closing_txn_id = ?, opening_txn_id = ?
        WHERE closing_id = ?`,
		closingTxnId, openingTxnId, id,
	)
	return err
}
