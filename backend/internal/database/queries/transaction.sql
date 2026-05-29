-- name: GetTransaction :one
SELECT txn_id, txn_uuid, txn_date, description, total_amount, currency,
       status, installment_id, receipt_no, note, version, ref_txn_id,
       updated_by, updated_at
FROM transactions
WHERE txn_id = @txn_id AND merchant_id = @merchant_id;

-- name: GetTransactionPaged :many
WITH total AS (SELECT COUNT(*) AS cnt FROM transactions WHERE merchant_id = @merchant_id)
SELECT t.txn_id, t.txn_uuid, t.txn_date, t.description, t.total_amount, t.currency,
       t.status, t.installment_id, t.receipt_no, t.note, t.version, t.ref_txn_id,
       t.updated_by, t.updated_at,
       total.cnt AS total
FROM transactions AS t, total
WHERE t.merchant_id = @merchant_id
ORDER BY t.txn_id DESC
    LIMIT @limit
OFFSET @offset;

-- name: InsertTransaction :execlastid
INSERT INTO transactions
    (merchant_id, txn_uuid, txn_date, description, total_amount, currency, status, receipt_no, note, version, ref_txn_id, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?);

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET status     = ?,
    updated_by = ?,
    updated_at = datetime('now'),
    version    = version + 1
WHERE txn_id = ? AND version = ?;

-- name: SysUpdateTransactionStatus :exec
UPDATE transactions
SET status     = ?,
    ref_txn_id = ?,
    updated_by = ?,
    updated_at = datetime('now'),
    version    = version + 1
WHERE txn_id = ?;

-- name: GetJournalEntries :many
SELECT entry_id, entry_uuid, txn_uuid, ledger_uuid, txn_id, ledger_id, account_id, debit, credit, note,
       updated_by, updated_at
FROM journal_entries
WHERE txn_id = @txn_id AND merchant_id = @merchant_id;

-- name: InsertJournalEntry :exec
INSERT INTO journal_entries
    (merchant_id, entry_uuid, txn_uuid, ledger_uuid, txn_id, ledger_id, account_id, debit, credit, note, cash_flow_category, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertJournalEntryWithID :exec
INSERT INTO journal_entries
    (entry_id, merchant_id, entry_uuid, txn_uuid, ledger_uuid, txn_id, ledger_id, account_id, debit, credit, note, cash_flow_category, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(entry_id) DO NOTHING;
