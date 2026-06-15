-- name: UpsertEntryCFCategory :exec
INSERT INTO entry_cf_categories (entry_uuid, merchant_id, cf_category, is_confirmed, updated_at, updated_by)
VALUES (@entry_uuid, @merchant_id, @cf_category, @is_confirmed, datetime('now'), @updated_by)
ON CONFLICT(entry_uuid, merchant_id) DO UPDATE SET
    cf_category  = excluded.cf_category,
    is_confirmed = excluded.is_confirmed,
    updated_at   = excluded.updated_at,
    updated_by   = excluded.updated_by;

-- name: GetEntryCFCategoryByUUID :one
SELECT * FROM entry_cf_categories
WHERE entry_uuid = @entry_uuid AND merchant_id = @merchant_id;

-- name: ListTransactionCFReview :many
SELECT
    t.txn_id,
    t.txn_uuid,
    t.txn_date,
    t.description,
    t.total_amount,
    t.currency,
    je.entry_id,
    je.entry_uuid,
    je.account_id,
    je.ledger_id,
    je.debit,
    je.credit,
    ecc.cf_category,
    COALESCE(ecc.is_confirmed, 0) AS is_confirmed
FROM transactions t
JOIN journal_entries je ON je.txn_id = t.txn_id AND je.merchant_id = t.merchant_id
LEFT JOIN entry_cf_categories ecc ON ecc.entry_uuid = je.entry_uuid AND ecc.merchant_id = je.merchant_id
WHERE t.merchant_id = @merchant_id
  AND t.status = 'ACTIVE'
  AND (@date_from = '' OR t.txn_date >= @date_from)
  AND (@date_to   = '' OR t.txn_date <= @date_to)
ORDER BY t.txn_id DESC, je.entry_id ASC;
