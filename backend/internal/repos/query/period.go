package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
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
	GetPeriodPagedByType(ctx context.Context, periodType enums.PeriodType, req model.PaginationParams) ([]projection.PeriodClosing, int64, error)
}

type sqlcdbPeriodRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbPeriodRepo) GetPeriodPagedByType(ctx context.Context, periodType enums.PeriodType, req model.PaginationParams) ([]projection.PeriodClosing, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetPeriodPagedByType(ctx, sqlcdb.GetPeriodPagedByTypeParams{
		MerchantID: merchantID,
		PeriodType: periodType,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.PeriodClosing, len(rows))
	for i, row := range rows {
		result[i] = projection.PeriodClosingFromGetPeriodPagedByTypeRow(row)
	}
	return result, total, nil
}

func newPeriodRepo(q *sqlcdb.Queries) PeriodRepo {
	return &sqlcdbPeriodRepo{q: q}
}

func NewPeriodRepo(db *sqlx.DB) PeriodRepo {
	return &sqlcdbPeriodRepo{q: sqlcdb.New(db)}
}

func (r *sqlcdbPeriodRepo) GetByPeriod(ctx context.Context, periodType enums.PeriodType, periodStart string) (*projection.PeriodClosing, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetPeriodByPeriod(ctx, sqlcdb.GetPeriodByPeriodParams{
		MerchantID:  merchantID,
		PeriodType:  periodType,
		PeriodStart: periodStart,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PeriodClosingPtrFromGetPeriodByPeriodRow(row), nil
}

func (r *sqlcdbPeriodRepo) GetByID(ctx context.Context, id int64) (*projection.PeriodClosing, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetPeriodByID(ctx, sqlcdb.GetPeriodByIDParams{
		ClosingID:  id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PeriodClosingPtrFromGetPeriodByIDRow(row), nil
}

func (r *sqlcdbPeriodRepo) GetLatestClosed(ctx context.Context, periodType enums.PeriodType) (*projection.PeriodClosing, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetLatestClosedPeriod(ctx, sqlcdb.GetLatestClosedPeriodParams{
		PeriodType: periodType,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.PeriodClosingPtrFromGetLatestClosedPeriodRow(row), nil
}

func (r *sqlcdbPeriodRepo) IsDateInClosedPeriod(ctx context.Context, date string) (bool, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return false, err
	}
	count, err := r.q.IsDateInClosedPeriod(ctx, sqlcdb.IsDateInClosedPeriodParams{
		Date:       date,
		MerchantID: merchantID,
	})
	return count > 0, err
}

func (r *sqlcdbPeriodRepo) AssertNoUnresolvedAdjustments(ctx context.Context, start string, end string) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	count, err := r.q.CountUnresolvedAdjustments(ctx, sqlcdb.CountUnresolvedAdjustmentsParams{
		MerchantID: merchantID,
		StartDate:  start,
		EndDate:    end,
	})
	if err != nil {
		return fmt.Errorf("check unresolved: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("period has %d unresolved reconciliation adjustments, please resolve before closing", count)
	}
	return nil
}
