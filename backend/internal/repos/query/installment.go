package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type InstallmentRepo interface {
	GetInstallment(ctx context.Context, id int64) (*projection.Installment, error)
	GetPayment(ctx context.Context, id int64, period int) (*projection.InstallmentPayment, error)
	GetInstallmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Installment, int64, error)
	GetPaymentPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InstallmentPayment, int64, error)
}

type sqlcdbInstallmentRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbInstallmentRepo) GetInstallmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Installment, int64, error) {
	rows, err := r.q.GetInstallmentPaged(ctx, sqlcdb.GetInstallmentPagedParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.Installment, len(rows))
	for i, row := range rows {
		result[i] = projection.InstallmentFromGetInstallmentPagedRow(row)
	}
	return result, total, nil
}

func (r *sqlcdbInstallmentRepo) GetPaymentPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InstallmentPayment, int64, error) {
	rows, err := r.q.GetPaymentPaged(ctx, sqlcdb.GetPaymentPagedParams{
		Limit:         req.Limit,
		Offset:        req.Offset,
		InstallmentID: id,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.InstallmentPayment, len(rows))
	for i, row := range rows {
		result[i] = projection.InstallmentPaymentFromGetPaymentPagedRow(row)
	}
	return result, total, nil
}

func newInstallmentRepo(q *sqlcdb.Queries) InstallmentRepo {
	return &sqlcdbInstallmentRepo{q: q}
}

func NewInstallmentRepo(db *sqlx.DB) InstallmentRepo {
	return &sqlcdbInstallmentRepo{q: sqlcdb.New(db)}
}

func (r *sqlcdbInstallmentRepo) GetInstallment(ctx context.Context, id int64) (*projection.Installment, error) {
	row, err := r.q.GetInstallment(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	note := ""
	if row.Note != nil {
		note = *row.Note
	}
	result := projection.InstallmentPtrFromGetInstallmentRow(row)
	result.Note = note
	return result, nil
}

func (r *sqlcdbInstallmentRepo) GetPayment(ctx context.Context, id int64, period int) (*projection.InstallmentPayment, error) {
	row, err := r.q.GetInstallmentPayment(ctx, sqlcdb.GetInstallmentPaymentParams{
		InstallmentID: id,
		PeriodNo:      int64(period),
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.InstallmentPaymentPtrFromInstallmentPayment(row), nil
}
