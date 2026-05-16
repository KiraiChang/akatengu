-- name: UpsertAccountRunningBalance :exec
INSERT INTO account_running_balances (account_id, merchant_id, debit_total, credit_total)
VALUES (?, ?, ?, ?)
ON CONFLICT(account_id, merchant_id) DO UPDATE SET
    debit_total  = debit_total  + excluded.debit_total,
    credit_total = credit_total + excluded.credit_total;

-- name: UpsertLedgerRunningBalance :exec
INSERT INTO ledger_running_balances (ledger_id, merchant_id, debit_total, credit_total)
VALUES (?, ?, ?, ?)
ON CONFLICT(ledger_id, merchant_id) DO UPDATE SET
    debit_total  = debit_total  + excluded.debit_total,
    credit_total = credit_total + excluded.credit_total;

-- name: GetAncestorAccountIds :many
SELECT ancestor_id FROM account_closure
WHERE descendant_id = ? AND depth > 0 AND merchant_id = ?;

-- name: GetAllAccountRunningBalance :many
SELECT b.account_id, b.merchant_id, b.debit_total, b.credit_total, a.name, a.normal_balance
FROM account_running_balances AS b
INNER JOIN accounts AS a ON a.account_id = b.account_id
WHERE b.merchant_id = @merchant_id;

-- name: GetAllLedgerRunningBalance :many
SELECT lb.ledger_id, lb.merchant_id, lb.debit_total, lb.credit_total, a.normal_balance
FROM ledger_running_balances AS lb
    INNER JOIN ledger_accounts AS la ON la.ledger_id = lb.ledger_id
    INNER JOIN accounts AS a ON a.account_id = la.account_id
WHERE lb.merchant_id = @merchant_id;
