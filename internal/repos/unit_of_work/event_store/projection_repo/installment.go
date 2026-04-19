package projection_repo

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxInstallmentRepo struct {
	tx *sqlx.Tx
}

func (s sqlxInstallmentRepo) UpdatePaymentTxn(ctx context.Context, paymentId int64, txnId int64) error {
	_, err := s.tx.ExecContext(ctx, `
        UPDATE installment_payments
        SET txn_id = ?
        WHERE payment_id = ?`,
		txnId, paymentId,
	)
	return err
}

func (s sqlxInstallmentRepo) PaidInstallmentPayment(ctx context.Context, id int64, date string) error {
	_, err := s.tx.ExecContext(ctx, `
        UPDATE installment_payments
        SET status = ?, paid_date = ?
        WHERE payment_id = ?`,
		enums.InstallmentPaymentStatusPaid.Enum(), date, id,
	)
	return err
}

func (s sqlxInstallmentRepo) UpdateInstallmentStatus(ctx context.Context, id int64, status enums.InstallmentStatus) error {
	_, err := s.tx.ExecContext(ctx, `
        UPDATE installment
        SET status = ?
        WHERE installment_id = ?`,
		status, id,
	)
	return err
}

func (s sqlxInstallmentRepo) InsertInstallment(ctx context.Context, p *projection.Installment) (int64, error) {
	result, err := s.tx.NamedExecContext(ctx, `
            INSERT INTO installments(
				ledger_id,
                description,
                total_amount,
                total_periods,
                paid_periods,
                amount_per_period,
                start_date,
                interest_rate,
                interest_type,
                status,
                note
            )
            VALUES (
					:ledger_id,
                    :description,
                    :total_amount,
                    :total_periods,
                    :paid_periods,
                    :amount_per_period,
                    :start_date,
                    :interest_rate,
                    :interest_type,
                    :status,
                    :note
            )`,
		p,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s sqlxInstallmentRepo) InsertInstallmentPayment(ctx context.Context, p *projection.InstallmentPayment) (int64, error) {
	result, err := s.tx.NamedExecContext(ctx, `
            INSERT INTO installments(
				installment_id,
                period_no,
                amount,
                interest,
                due_date,
                status
            )
            VALUES (
				:installment_id,
                :period_no,
                :amount,
                :interest,
                :due_date,
                :status
            )`,
		p,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s sqlxInstallmentRepo) UpdateInstallmentTxn(ctx context.Context, inst_id int64, txn_id int64) error {
	query := `
		UPDATE installments
		SET txn_id = :txn_id
		WHERE investment_id = :investment_id
	`

	args := map[string]interface{}{
		"investment_id": inst_id,
		"txn_id":        txn_id,
	}

	_, err := s.tx.NamedExecContext(ctx, query, args)
	return err
}

func NewInstallment(tx *sqlx.Tx) InstallmentRepo {
	return &sqlxInstallmentRepo{tx}
}
