package query

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type InstallmentRepo interface {
	GetInstallment(ctx context.Context, id int64) (*projection.Installment, error)
	GetPayment(ctx context.Context, id int64, period int) (*projection.InstallmentPayment, error)
}

type sqlInstallmentRepo struct {
	db *sqlx.DB
}

func (s sqlInstallmentRepo) GetInstallment(ctx context.Context, id int64) (*projection.Installment, error) {
	var result projection.Installment
	if err := s.db.QueryRowxContext(ctx, `
        SELECT * FROM installments
        WHERE installment_id = ?`,
		id,
	).StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s sqlInstallmentRepo) GetPayment(ctx context.Context, id int64, period int) (*projection.InstallmentPayment, error) {
	var result projection.InstallmentPayment
	if err := s.db.QueryRowxContext(ctx, `
        SELECT * FROM installment_payments
        WHERE installment_id = ?
        	AND period = ?`,
		id, period,
	).StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func NewInstallmentRepo(db *sqlx.DB) InstallmentRepo {
	return &sqlInstallmentRepo{db: db}
}
