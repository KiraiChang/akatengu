package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type InstallmentRepo interface {
	GetInstallment(ctx context.Context, id int64) (*projection.Installment, error)
	GetPayment(ctx context.Context, id int64, period int) (*projection.InstallmentPayment, error)
}

type sqlcdbInstallmentRepo struct {
	q *sqlcdb.Queries
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
