package query

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────
// Interface
// ─────────────────────────────────────────

//// ClosingTxRepository transaction of closing unit of work
//type ClosingTxRepository interface {
//	Insert(ctx context.Context, c db.PeriodClosing) (int64, error)
//	UpdateStatus(ctx context.Context, closingID int64, status enums.ClosingStatus, closedAt *string) error
//	UpdateSnapshot(ctx context.Context, closingID int64, snapshot string) error
//	UpdateClosingTxn(ctx context.Context, closingID int64, openingTxnID int64, txnID int64) error
//}

type ClosingRepo interface {
	// ─────────────────────────────────────────
	// 查詢
	// ─────────────────────────────────────────

	// GetByPeriod 依照期間來找關帳資料
	GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*db.PeriodClosing, error)
	GetByID(ctx context.Context, id int64) (*db.PeriodClosing, error)
	GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*db.PeriodClosing, error)
	IsDateInClosedPeriod(ctx context.Context, date string) (bool, error)
	GetClosingEntries(ctx context.Context, txnID int64) ([]projection.Entry, error)
	AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error
}

// ─────────────────────────────────────────
// sqlx 實作
// ─────────────────────────────────────────

type sqlxClosingRepo struct {
	db *sqlx.DB
}

func (r *sqlxClosingRepo) AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error {
	// 查 v_unresolved_adjustments 是否有在這個期間的未解決差異
	var count int
	err := r.db.QueryRowxContext(ctx, `
		SELECT COUNT(*)
		FROM v_unresolved_adjustments va
		JOIN reconciliations r ON va.recon_id = r.recon_id
		WHERE r.recon_date BETWEEN ? AND ?`,
		start, end,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check unresolved: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("period has %d unresolved reconciliation adjustments, please resolve before closing", count)
	}
	return nil
}

func NewClosingRepo(db *sqlx.DB) ClosingRepo {
	return &sqlxClosingRepo{db: db}
}

func (r *sqlxClosingRepo) GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*db.PeriodClosing, error) {
	var c db.PeriodClosing
	err := r.db.QueryRowxContext(ctx, `
		SELECT * FROM period_closings
		WHERE period_type = ? AND period_start = ?`,
		periodType, periodStart,
	).StructScan(&c)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *sqlxClosingRepo) GetByID(ctx context.Context, closeId int64) (*db.PeriodClosing, error) {
	var c db.PeriodClosing
	err := r.db.QueryRowxContext(ctx, `
		SELECT * FROM period_closings
		WHERE ClosingId = ?`,
		closeId,
	).StructScan(&c)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *sqlxClosingRepo) GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*db.PeriodClosing, error) {
	var c db.PeriodClosing
	err := r.db.QueryRowxContext(ctx, `
		SELECT * FROM period_closings
		WHERE period_type = ? AND status = 'closed'
		ORDER BY period_end DESC
		LIMIT 1`,
		periodType,
	).StructScan(&c)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *sqlxClosingRepo) IsDateInClosedPeriod(ctx context.Context, date string) (bool, error) {
	var count int
	err := r.db.QueryRowxContext(ctx, `
		SELECT COUNT(*) FROM period_closings
		WHERE status      = 'closed'
		  AND period_start <= ?
		  AND period_end   >= ?`,
		date, date,
	).Scan(&count)
	return count > 0, err
}

func (r *sqlxClosingRepo) CountUnclosedMonths(ctx context.Context, year int) (int, error) {
	var count int
	err := r.db.QueryRowxContext(ctx, `
        SELECT COUNT(*) FROM period_closings
        WHERE period_type  = 'monthly'
          AND period_start >= ?
          AND period_end   <= ?
          AND status       = 'closed'`,
		fmt.Sprintf("%d-01-01", year),
		fmt.Sprintf("%d-12-31", year),
	).Scan(&count)
	return count, err
}

func (r *sqlxClosingRepo) GetClosingEntries(ctx context.Context, txnID int64) ([]projection.Entry, error) {
	var entries []projection.Entry
	err := r.db.SelectContext(ctx, &entries, `
        SELECT account_id, ledger_id, debit, credit
        FROM journal_entries
        WHERE txn_id = ?`,
		txnID,
	)
	return entries, err
}

//// ─────────────────────────────────────────
//// tx 實作
//// ─────────────────────────────────────────
//
//type sqlxClosingTxRepository struct{ tx *sqlx.Tx }
//
//func (r *sqlxClosingTxRepository) Insert(ctx context.Context, c db.PeriodClosing) (int64, error) {
//	result, err := r.tx.ExecContext(ctx, `
//        INSERT INTO period_closings
//            (period_type, period_start, period_end, status, note)
//        VALUES (?, ?, ?, ?, ?)`,
//		c.PeriodType, c.PeriodStart, c.PeriodEnd, c.Status, c.Note,
//	)
//	if err != nil {
//		return 0, err
//	}
//	return result.LastInsertId()
//}
//
//func (r *sqlxClosingTxRepository) UpdateStatus(ctx context.Context, closingID int64, status enums.ClosingStatus, closedAt *string) error {
//	_, err := r.tx.ExecContext(ctx, `
//        UPDATE period_closings
//        SET status = ?, closed_at = ?
//        WHERE closing_id = ?`,
//		status, closedAt, closingID,
//	)
//	return err
//}
//
//func (r *sqlxClosingTxRepository) UpdateSnapshot(ctx context.Context, closingID int64, snapshot string) error {
//	_, err := r.tx.ExecContext(ctx, `
//        UPDATE period_closings
//        SET snapshot = ?
//        WHERE closing_id = ?`,
//		snapshot, closingID,
//	)
//	return err
//}
//
//func (r *sqlxClosingTxRepository) UpdateClosingTxn(ctx context.Context, closingID int64, openingTxnID int64, txnID int64) error {
//	_, err := r.tx.ExecContext(ctx, `
//        UPDATE period_closings
//        SET closing_txn_id = ?,
//            opening_txn_id = ?
//        WHERE closing_id = ?`,
//		txnID, openingTxnID, closingID,
//	)
//	return err
//}
