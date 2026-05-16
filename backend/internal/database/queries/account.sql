-- name: GetAllAccounts :many
SELECT a.account_id, a.parent_id, a.name, a.type, a.normal_balance,
       a.currency, a.is_summary, a.is_active, a.note, a.version, a.cash_flow_category,
       EXISTS (SELECT 1 FROM accounts c WHERE c.parent_id = a.account_id AND c.merchant_id = a.merchant_id) AS has_child
FROM accounts a
WHERE a.merchant_id = @merchant_id;

-- name: GetAccount :one
SELECT account_id, parent_id, name, type, normal_balance,
       currency, is_summary, is_active, note, version, cash_flow_category
FROM accounts
WHERE account_id = @account_id AND merchant_id = @merchant_id;

-- name: GetAccountPaged :many
SELECT
    a.account_id, a.parent_id, a.name, a.type, a.normal_balance,
    a.currency, a.is_summary, a.is_active, a.note, a.version, a.cash_flow_category,
    COUNT(*) OVER() AS total,
    EXISTS (SELECT 1 FROM accounts c WHERE c.parent_id = a.account_id AND c.merchant_id = a.merchant_id) AS has_child
FROM accounts a
WHERE a.parent_id IS NULL AND a.merchant_id = @merchant_id
ORDER BY a.account_id ASC
    LIMIT @limit
OFFSET @offset;

-- name: GetChildrenAccount :many
SELECT
    account_id, parent_id, name, type, normal_balance,
    currency, is_summary, is_active, note, version, cash_flow_category,
    EXISTS (SELECT 1 FROM accounts c WHERE c.parent_id = a.account_id AND c.merchant_id = a.merchant_id) AS has_child
FROM accounts a
WHERE a.parent_id = @parent_id AND a.merchant_id = @merchant_id
ORDER BY a.account_id ASC;

-- name: GetLedger :one
SELECT ledger_id, account_id, institution, name, type,
       account_no, currency, credit_limit, billing_day, due_day,
       is_active, note, version
FROM ledger_accounts
WHERE ledger_id = @ledger_id AND merchant_id = @merchant_id;

-- name: GetAllLedgers :many
SELECT ledger_id, account_id, institution, name, type,
       account_no, currency, credit_limit, billing_day, due_day,
       is_active, note, version
FROM ledger_accounts
WHERE merchant_id = @merchant_id
ORDER BY ledger_id ASC;

-- name: GetAllLedgerBalances :many
SELECT ledger_id, debit_total, credit_total, normal_balance
FROM v_ledger_account_balances
WHERE merchant_id = @merchant_id
ORDER BY ledger_id ASC;

-- name: GetAllAccountBalances :many
SELECT account_id, debit_total, credit_total, normal_balance
FROM v_account_balances
WHERE merchant_id = @merchant_id
ORDER BY account_id ASC;

-- name: GetLedgerPaged :many
SELECT ledger_id, account_id, institution, name, type,
       account_no, currency, credit_limit, billing_day, due_day,
       is_active, note, version,
       COUNT(*) OVER() AS total
FROM ledger_accounts
WHERE merchant_id = @merchant_id
ORDER BY ledger_id ASC
    LIMIT @limit
OFFSET @offset;

-- name: CreateAccount :exec
INSERT INTO accounts
    (account_id, merchant_id, parent_id, name, type, normal_balance, currency, is_summary, is_active, note, version, cash_flow_category)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?);

-- name: UpdateAccount :exec
UPDATE accounts
SET parent_id          = ?,
    name               = ?,
    type               = ?,
    normal_balance     = ?,
    currency           = ?,
    is_summary         = ?,
    is_active          = ?,
    note               = ?,
    cash_flow_category = ?,
    version            = version + 1
WHERE account_id = ? AND merchant_id = ? AND version = ?;

-- name: CreateLedgerAccount :exec
INSERT INTO ledger_accounts
    (merchant_id, account_id, institution, name, type, account_no, currency, credit_limit, billing_day, due_day, is_active, note, version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1);

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
WHERE ledger_id = ? AND merchant_id = ? AND version = ?;
