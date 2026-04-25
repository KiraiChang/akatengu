package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PeriodRepo interface {
	GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*projection.PeriodClosing, error)
	GetByID(ctx context.Context, id int64) (*projection.PeriodClosing, error)
	GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*projection.PeriodClosing, error)
	IsDateInClosedPeriod(ctx context.Context, date string) (bool, error)
	AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error
}

type sqlcdbPeriodRepo struct {
	q *sqlcdb.Queries
}

func newPeriodRepo(q *sqlcdb.Queries) PeriodRepo {
	return &sqlcdbPeriodRepo{q: q}
}

func NewPeriodRepo(db *sqlx.DB) PeriodRepo {
	return &sqlcdbPeriodRepo{q: sqlcdb.New(db)}
}

func (r *sqlcdbPeriodRepo) GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*projection.PeriodClosing, error) {
	row, err := r.q.GetPeriodByPeriod(ctx, sqlcdb.GetPeriodByPeriodParams{
		PeriodType:  periodType,
		PeriodStart: periodStart,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PeriodClosingPtrFromPeriodClosing(row), nil
}

func (r *sqlcdbPeriodRepo) GetByID(ctx context.Context, id int64) (*projection.PeriodClosing, error) {
	row, err := r.q.GetPeriodByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PeriodClosingPtrFromPeriodClosing(row), nil
}

func (r *sqlcdbPeriodRepo) GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*projection.PeriodClosing, error) {
	row, err := r.q.GetLatestClosedPeriod(ctx, periodType)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PeriodClosingPtrFromPeriodClosing(row), nil
}

func (r *sqlcdbPeriodRepo) IsDateInClosedPeriod(ctx context.Context, date string) (bool, error) {
	count, err := r.q.IsDateInClosedPeriod(ctx, sqlcdb.IsDateInClosedPeriodParams{
		PeriodStart: date,
		PeriodEnd:   date,
	})
	return count > 0, err
}

func (r *sqlcdbPeriodRepo) AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error {
	count, err := r.q.CountUnresolvedAdjustments(ctx, sqlcdb.CountUnresolvedAdjustmentsParams{
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return fmt.Errorf("check unresolved: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("period has %d unresolved reconciliation adjustments, please resolve before closing", count)
	}
	return nil
}
