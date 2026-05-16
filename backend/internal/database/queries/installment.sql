-- name: GetInstallment :one
SELECT installment_id, txn_id, ledger_id, description, total_amount,
       total_periods, amount_per_period, start_date, end_date,
       interest_rate, interest_type, status, note
FROM installments
WHERE installment_id = @installment_id AND merchant_id = @merchant_id;

-- name: GetInstallmentPaged :many
WITH total AS (
    SELECT COUNT(*) AS cnt FROM installments WHERE merchant_id = @merchant_id
)
SELECT
             i.installment_id, i.txn_id, i.ledger_id, i.description, i.total_amount,
             i.total_periods, i.amount_per_period, i.start_date, i.end_date,
             i.interest_rate, i.interest_type, i.status, i.note,
             total.cnt AS total,
             COUNT(p.payment_id) AS paid_periods
         FROM installments AS i, total
                  LEFT JOIN installment_payments AS p
                            ON p.installment_id = i.installment_id AND p.status = 'PAID'
         WHERE i.merchant_id = @merchant_id
         GROUP BY
             i.installment_id, i.txn_id, i.ledger_id, i.description, i.total_amount,
             i.total_periods, i.amount_per_period, i.start_date, i.end_date,
             i.interest_rate, i.interest_type, i.status, i.note
         ORDER BY i.start_date DESC
    LIMIT @limit OFFSET @offset;

-- name: GetInstallmentPayment :one
SELECT payment_id, installment_id, txn_id, period_no, amount,
       interest, due_date, paid_date, status
FROM installment_payments
WHERE installment_id = @installment_id AND period_no = @period_no AND merchant_id = @merchant_id;

-- name: GetPaymentPaged :many
WITH total AS (
    SELECT COUNT(*) AS cnt
    FROM installment_payments AS p2
    WHERE p2.installment_id = @installment_id AND p2.merchant_id = @merchant_id
)
         SELECT p.payment_id, p.installment_id, p.txn_id, p.period_no, p.amount,
                p.interest, p.due_date, p.paid_date, p.status,
                total.cnt AS total
         FROM installment_payments AS p, total
         WHERE p.installment_id = @installment_id AND p.merchant_id = @merchant_id
         ORDER BY p.due_date DESC
    LIMIT @limit OFFSET @offset;

-- name: InsertInstallment :execlastid
INSERT INTO installments
    (merchant_id, ledger_id, description, total_amount, total_periods,
     amount_per_period, start_date, interest_rate, interest_type, status, note)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInstallmentPayment :execlastid
INSERT INTO installment_payments
    (merchant_id, installment_id, period_no, amount, interest, due_date, status)
VALUES (?, ?, ?, ?, ?, ?, ?);

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
