-- name: GetBalanceSheetRows :many
WITH filtered_entries AS (
    SELECT je.account_id,
           je.debit,
           je.credit
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id
    WHERE t.status   = 'ACTIVE'
      AND t.txn_date <= ?
),
balances AS (
    SELECT
        a.account_id,
        a.parent_id,
        a.name,
        a.type,
        a.normal_balance,
        COALESCE(SUM(fe.debit) - SUM(fe.credit), 0) AS raw_balance
    FROM accounts a
    LEFT JOIN filtered_entries fe ON a.account_id = fe.account_id
    WHERE a.is_summary = 0 AND a.is_active = 1
    GROUP BY a.account_id
),
normalized AS (
    SELECT
        account_id,
        parent_id,
        name,
        type,
        CAST(CASE normal_balance
            WHEN 'DEBIT'  THEN  raw_balance
            WHEN 'CREDIT' THEN -raw_balance
        END AS REAL) AS balance
    FROM balances
)
SELECT
    type,
    account_id,
    name,
    balance
FROM normalized
WHERE type IN ('ASSET', 'LIABILITY', 'EQUITY')
ORDER BY
    CASE type
        WHEN 'ASSET'     THEN 1
        WHEN 'LIABILITY' THEN 2
        WHEN 'EQUITY'    THEN 3
    END,
    account_id;

-- name: GetIncomeStatementRows :many
WITH filtered_entries AS (
    SELECT je.account_id,
           je.debit,
           je.credit
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id
    WHERE t.status   = 'ACTIVE'
      AND t.txn_date >= ?
      AND t.txn_date <= ?
),
period_balances AS (
    SELECT
        a.account_id,
        a.parent_id,
        a.name,
        a.type,
        a.normal_balance,
        COALESCE(SUM(fe.debit) - SUM(fe.credit), 0) AS raw_balance
    FROM accounts a
    LEFT JOIN filtered_entries fe ON a.account_id = fe.account_id
    WHERE a.is_summary = 0
      AND a.is_active  = 1
      AND a.type IN ('INCOME', 'EXPENSE')
    GROUP BY a.account_id
),
normalized AS (
    SELECT
        account_id,
        parent_id,
        name,
        type,
        CAST(CASE normal_balance
            WHEN 'DEBIT'  THEN  raw_balance
            WHEN 'CREDIT' THEN -raw_balance
        END AS REAL) AS amount
    FROM period_balances
)
SELECT
    type,
    account_id,
    name,
    amount
FROM normalized
ORDER BY
    CASE type
        WHEN 'INCOME'  THEN 1
        WHEN 'EXPENSE' THEN 2
    END,
    account_id;
