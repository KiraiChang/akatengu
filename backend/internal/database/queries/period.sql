-- name: GetPeriodByPeriod :one
SELECT closing_id, period_type, period_start, period_end, status,
       opening_txn_id, closing_txn_id, snapshot, closed_at, note,
       reopen_at, reopen_reason
FROM period_closings
WHERE period_type = @period_type AND period_start = @period_start AND merchant_id = @merchant_id;

-- name: GetPeriodPagedByType :many
WITH total AS (SELECT COUNT(*) AS cnt FROM period_closings AS p2
               WHERE p2.period_type = @period_type AND p2.merchant_id = @merchant_id)
SELECT closing_id, period_type, period_start, period_end, status,
       opening_txn_id, closing_txn_id, snapshot, closed_at, note,
       reopen_at, reopen_reason,
       total.cnt AS total
FROM period_closings AS p, total
WHERE p.period_type = @period_type AND p.merchant_id = @merchant_id
ORDER BY period_end DESC
    LIMIT @limit
OFFSET @offset;

-- name: GetPeriodByID :one
SELECT closing_id, period_type, period_start, period_end, status,
       opening_txn_id, closing_txn_id, snapshot, closed_at, note,
       reopen_at, reopen_reason
FROM period_closings
WHERE closing_id = @closing_id AND merchant_id = @merchant_id;

-- name: GetLatestClosedPeriod :one
SELECT closing_id, period_type, period_start, period_end, status,
       opening_txn_id, closing_txn_id, snapshot, closed_at, note,
       reopen_at, reopen_reason
FROM period_closings
WHERE period_type = @period_type AND status = 'CLOSED' AND merchant_id = @merchant_id
ORDER BY period_end DESC
LIMIT 1;

-- name: GetLatestClosedPeriodBeforeOrOn :one
SELECT closing_id, period_end FROM period_closings
WHERE period_type = @period_type AND status = 'CLOSED' AND period_end <= @period_end AND merchant_id = @merchant_id
ORDER BY period_end DESC LIMIT 1;

-- name: GetLatestClosedPeriodBefore :one
SELECT closing_id, period_end FROM period_closings
WHERE period_type = @period_type AND status = 'CLOSED' AND period_end < @period_end AND merchant_id = @merchant_id
ORDER BY period_end DESC LIMIT 1;

-- name: GetLatestClosedPeriodBetween :one
SELECT closing_id, period_end FROM period_closings
WHERE period_type = @period_type AND status = 'CLOSED' AND period_end >= @start AND period_end <= @end AND merchant_id = @merchant_id
ORDER BY period_end DESC LIMIT 1;

-- name: IsDateInClosedPeriod :one
SELECT COUNT(*) FROM period_closings
WHERE status       = 'CLOSED'
  AND period_start <= @date
  AND period_end   >= @date
  AND merchant_id  = @merchant_id;

-- name: CountUnresolvedAdjustments :one
SELECT COUNT(*)
FROM reconciliation_adjustments ra
         JOIN reconciliations r ON ra.recon_id = r.recon_id
WHERE ra.adjustment_type = 'UNRESOLVED'
  AND r.recon_date >= sqlc.arg(start_date)
  AND r.recon_date <  sqlc.arg(end_date)
  AND r.merchant_id = sqlc.arg(merchant_id);

-- name: InsertPeriodClose :execlastid
INSERT INTO period_closings
    (merchant_id, period_type, period_start, period_end, status, note)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdatePeriodCloseSnapshot :exec
UPDATE period_closings
SET status    = ?,
    closed_at = ?,
    snapshot  = ?
WHERE closing_id = ?;

-- name: ReopenPeriodClose :exec
UPDATE period_closings
SET status        = ?,
    reopen_reason = ?,
    reopen_at     = ?,
    snapshot      = NULL
WHERE closing_id = ?;

-- name: UpdatePeriodCloseTxnID :exec
UPDATE period_closings
SET closing_txn_id = ?,
    opening_txn_id = ?
WHERE closing_id = ?;
