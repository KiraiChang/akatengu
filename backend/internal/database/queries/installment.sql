-- name: GetInstallment :one
SELECT installment_id, txn_id, ledger_id, description, total_amount,
       total_periods, paid_periods, amount_per_period, start_date, end_date,
       interest_rate, interest_type, status, note
FROM installments
WHERE installment_id = ?;

-- name: GetInstallmentPayment :one
SELECT payment_id, installment_id, txn_id, period_no, amount,
       interest, due_date, paid_date, status
FROM installment_payments
WHERE installment_id = ? AND period_no = ?;

-- name: InsertInstallment :execlastid
INSERT INTO installments
    (ledger_id, description, total_amount, total_periods, paid_periods,
     amount_per_period, start_date, interest_rate, interest_type, status, note)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInstallmentPayment :execlastid
INSERT INTO installment_payments
    (installment_id, period_no, amount, interest, due_date, status)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateInstallmentTxn :exec
UPDATE installments
SET txn_id = ?
WHERE installment_id = ?;

-- name: PaidInstallmentPayment :exec
UPDATE installment_payments
SET status    = ?,
    paid_date = ?
WHERE payment_id = ?;

-- name: UpdateInstallmentStatus :exec
UPDATE installments
SET status = ?
WHERE installment_id = ?;

-- name: UpdatePaymentTxn :exec
UPDATE installment_payments
SET txn_id = ?
WHERE payment_id = ?;