-- name: GetTransaction :one
SELECT txn_id, txn_date, description, total_amount, currency,
       status, installment_id, receipt_no, note, version, ref_txn_id
FROM transactions
WHERE txn_id = ?;

-- name: GetTransactionPaged :many
WITH total AS(SELECT COUNT(*) AS cnt FROM transactions)
SELECT txn_id, txn_date, description, total_amount, currency,
       status, installment_id, receipt_no, note, version, ref_txn_id,
       total.cnt AS total
FROM transactions, total
ORDER BY txn_id DESC
    LIMIT @limit
OFFSET @offset;


-- name: InsertTransaction :execlastid
INSERT INTO transactions
    (txn_date, description, total_amount, currency, status, receipt_no, note, version, ref_txn_id)
VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?);

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET status  = ?,
    version = version + 1
WHERE txn_id = ? AND version = ?;

-- name: SysUpdateTransactionStatus :exec
UPDATE transactions
SET status     = ?,
    ref_txn_id = ?,
    version    = version + 1
WHERE txn_id = ?;

-- name: GetJournalEntries :many
SELECT entry_id, txn_id, ledger_id, account_id, debit, credit, note
FROM journal_entries
WHERE txn_id = ?;

-- name: InsertJournalEntry :exec
INSERT INTO journal_entries
    (txn_id, ledger_id, account_id, debit, credit, note)
VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertJournalEntryWithID :exec
INSERT INTO journal_entries
    (entry_id, txn_id, ledger_id, account_id, debit, credit, note)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(entry_id) DO NOTHING;