package projection

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/enums"
	"context"
)

func (r *sqlxTxProjectionRepository) InsertClose(ctx context.Context, c db.PeriodClosing) error {
	_, err := r.tx.ExecContext(ctx, `
        INSERT INTO period_closings
            (closed_id, period_type, period_start, period_end, status, note)
        VALUES (?, ?, ?, ?, ?, ?)`,
		c.ClosingId, c.PeriodType, c.PeriodStart, c.PeriodEnd, c.Status, c.Note,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *sqlxTxProjectionRepository) UpdateCloseStatus(ctx context.Context, closingId int64, status enums.ClosingStatus, closedAt *string) error {
	_, err := r.tx.ExecContext(ctx, `
        UPDATE period_closings
        SET status = ?, closed_at = ?
        WHERE closing_id = ?`,
		status, closedAt, closingId,
	)
	return err
}

func (r *sqlxTxProjectionRepository) UpdateCloseSnapshotByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string, snapshot string) error {
	_, err := r.tx.ExecContext(ctx, `
        UPDATE period_closings
        SET snapshot = ?
        WHERE period_type = ?
        	AND period_start = ?`,
		snapshot, periodType, periodStart,
	)
	return err
}
