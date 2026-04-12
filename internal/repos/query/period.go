package query

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ─────────────────────────────────────────
// Interface
// ─────────────────────────────────────────

type PeriodRepo interface {
	// ─────────────────────────────────────────
	// 查詢
	// ─────────────────────────────────────────

	// GetByPeriod 依照期間來找關帳資料
	GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*projection.PeriodClosing, error)
	GetByID(ctx context.Context, id int64) (*projection.PeriodClosing, error)
	GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*projection.PeriodClosing, error)
	IsDateInClosedPeriod(ctx context.Context, date string) (bool, error)
	AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error
}

// ─────────────────────────────────────────
// sqlx 實作
// ─────────────────────────────────────────

type sqlxPeriodRepo struct {
	db *sqlx.DB
}

func (r *sqlxPeriodRepo) AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error {
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

func NewPeriodRepo(db *sqlx.DB) PeriodRepo {
	return &sqlxPeriodRepo{db: db}
}

func (r *sqlxPeriodRepo) GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*projection.PeriodClosing, error) {
	var c projection.PeriodClosing
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

func (r *sqlxPeriodRepo) GetByID(ctx context.Context, closeId int64) (*projection.PeriodClosing, error) {
	var c projection.PeriodClosing
	err := r.db.QueryRowxContext(ctx, `
		SELECT * FROM period_closings
		WHERE closing_id = ?`,
		closeId,
	).StructScan(&c)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *sqlxPeriodRepo) GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*projection.PeriodClosing, error) {
	var c projection.PeriodClosing
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

func (r *sqlxPeriodRepo) IsDateInClosedPeriod(ctx context.Context, date string) (bool, error) {
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

func (r *sqlxPeriodRepo) CountUnclosedMonths(ctx context.Context, year int) (int, error) {
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
