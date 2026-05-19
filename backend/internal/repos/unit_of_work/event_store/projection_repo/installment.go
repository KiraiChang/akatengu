package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
)

type sqlxInstallmentRepo struct {
	q *sqlcdb.Queries
}

func (s sqlxInstallmentRepo) UpdatePaymentTxn(ctx context.Context, paymentId int64, txnId int64, updatedBy *string) error {
	return s.q.UpdatePaymentTxn(ctx, sqlcdb.UpdatePaymentTxnParams{
		PaymentID: paymentId,
		TxnID:     &txnId,
		UpdatedBy: updatedBy,
	})
}

func (s sqlxInstallmentRepo) PaidInstallmentPayment(ctx context.Context, id int64, date string, updatedBy *string) error {
	return s.q.PaidInstallmentPayment(ctx, sqlcdb.PaidInstallmentPaymentParams{
		PaymentID: id,
		Status:    enums.InstallmentPaymentStatusPaid.Enum(),
		PaidDate:  &date,
		UpdatedBy: updatedBy,
	})
}

func (s sqlxInstallmentRepo) UpdateInstallmentStatus(ctx context.Context, id int64, status enums.InstallmentStatus, updatedBy *string) error {
	return s.q.UpdateInstallmentStatus(ctx, sqlcdb.UpdateInstallmentStatusParams{
		InstallmentID: id,
		Status:        status,
		UpdatedBy:     updatedBy,
	})
}

func (s sqlxInstallmentRepo) InsertInstallment(ctx context.Context, p *projection.Installment) (int64, error) {
	return s.q.InsertInstallment(ctx, p.ToInsertInstallmentParams())
}

func (s sqlxInstallmentRepo) InsertInstallmentPayment(ctx context.Context, p *projection.InstallmentPayment) (int64, error) {
	return s.q.InsertInstallmentPayment(ctx, p.ToInsertInstallmentPaymentParams())
}

func (s sqlxInstallmentRepo) UpdateInstallmentTxn(ctx context.Context, inst_id int64, txn_id int64, updatedBy *string) error {
	return s.q.UpdateInstallmentTxn(ctx, sqlcdb.UpdateInstallmentTxnParams{
		InstallmentID: inst_id,
		TxnID:         &txn_id,
		UpdatedBy:     updatedBy,
	})
}

func NewInstallment(q *sqlcdb.Queries) InstallmentRepo {
	return &sqlxInstallmentRepo{q}
}
