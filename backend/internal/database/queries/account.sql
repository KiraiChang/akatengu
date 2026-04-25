-- name: GetAllAccounts :many
SELECT account_id, parent_id, name, type, normal_balance,
       currency, is_summary, is_active, note, version
FROM accounts;

-- name: GetAccount :one
SELECT account_id, parent_id, name, type, normal_balance,
       currency, is_summary, is_active, note, version
FROM accounts
WHERE account_id = ?;

-- name: GetAccountsPaged :many
SELECT
    account_id, parent_id, name, type, normal_balance,
    currency, is_summary, is_active, note, version,
    COUNT(*) OVER() AS total
FROM accounts
ORDER BY account_id ASC
    LIMIT  @page_size
OFFSET @offset;

-- name: GetLedger :one
SELECT ledger_id, account_id, institution, name, type,
       account_no, currency, credit_limit, billing_day, due_day,
       is_active, note, version
FROM ledger_accounts
WHERE ledger_id = ?;

-- name: CreateAccount :exec
INSERT INTO accounts
    (account_id, parent_id, name, type, normal_balance, currency, is_summary, is_active, note, version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1);

-- name: UpdateAccount :exec
UPDATE accounts
SET parent_id      = ?,
    name           = ?,
    type           = ?,
    normal_balance = ?,
    currency       = ?,
    is_summary     = ?,
    is_active      = ?,
    note           = ?,
    version        = version + 1
WHERE account_id = ? AND version = ?;

-- name: CreateLedgerAccount :exec
INSERT INTO ledger_accounts
    (account_id, institution, name, type, account_no, currency, credit_limit, billing_day, due_day, is_active, note, version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1);

-- name: UpdateLedgerAccount :exec
UPDATE ledger_accounts
SET account_id   = ?,
    institution  = ?,
    name         = ?,
    type         = ?,
    account_no   = ?,
    currency     = ?,
    credit_limit = ?,
    billing_day  = ?,
    due_day      = ?,
    is_active    = ?,
    note         = ?,
    version      = version + 1
WHERE ledger_id = ? AND version = ?;