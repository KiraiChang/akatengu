-- name: InsertPrepaid :one
INSERT INTO prepaids (
    merchant_id, prepaid_uuid, account_id, expense_account_id,
    name, total_amount, periods, start_date, updated_by
) VALUES (
    @merchant_id, @prepaid_uuid, @account_id, @expense_account_id,
    @name, @total_amount, @periods, @start_date, @updated_by
) RETURNING *;

-- name: UpdatePrepaidTxn :exec
UPDATE prepaids SET txn_id = @txn_id WHERE id = @id AND merchant_id = @merchant_id;

-- name: GetPrepaidByID :one
SELECT * FROM prepaids WHERE id = @id AND merchant_id = @merchant_id;

-- name: GetPrepaidByUUID :one
SELECT * FROM prepaids WHERE prepaid_uuid = @prepaid_uuid AND merchant_id = @merchant_id;

-- name: GetActivePrepaidsByMerchant :many
SELECT * FROM prepaids WHERE merchant_id = @merchant_id AND status = 'ACTIVE' ORDER BY id ASC;

-- name: GetAllPrepaidsByMerchant :many
SELECT * FROM prepaids WHERE merchant_id = @merchant_id ORDER BY id ASC;

-- name: UpdatePrepaidAmortization :exec
UPDATE prepaids SET
    amortized_amount  = amortized_amount + @delta_amount,
    amortized_periods = amortized_periods + 1,
    status            = @status,
    updated_by        = @updated_by,
    updated_at        = datetime('now'),
    version           = version + 1
WHERE id = @id AND merchant_id = @merchant_id;

-- name: UpdatePrepaidDisposed :exec
UPDATE prepaids SET
    status     = 'DISPOSED',
    updated_by = @updated_by,
    updated_at = datetime('now'),
    version    = version + 1
WHERE id = @id AND merchant_id = @merchant_id;

-- name: InsertPrepaidAmortization :exec
INSERT INTO prepaid_amortizations (
    merchant_id, amortization_uuid, prepaid_uuid, prepaid_id, txn_id, period_date, amount, updated_by
) VALUES (
    @merchant_id, @amortization_uuid, @prepaid_uuid, @prepaid_id, @txn_id, @period_date, @amount, @updated_by
);

-- name: GetPrepaidAmortizationsByPrepaidID :many
SELECT * FROM prepaid_amortizations
WHERE prepaid_id = @prepaid_id AND merchant_id = @merchant_id
ORDER BY period_date ASC;
