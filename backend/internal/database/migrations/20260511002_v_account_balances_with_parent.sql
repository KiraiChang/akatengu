-- +goose Up
DROP VIEW IF EXISTS v_account_balances;
CREATE VIEW IF NOT EXISTS v_account_balances AS
WITH RECURSIVE
leaf_balances AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(je.debit), 0)  AS debit_total,
           COALESCE(SUM(je.credit), 0) AS credit_total
    FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE a.is_active = 1 AND a.is_summary = 0
    GROUP BY a.account_id
),
subtree(ancestor_id, leaf_id) AS (
    SELECT a.parent_id, a.account_id
    FROM accounts a
    WHERE a.is_summary = 0 AND a.is_active = 1 AND a.parent_id IS NOT NULL
    UNION ALL
    SELECT a.parent_id, s.leaf_id
    FROM accounts a
    JOIN subtree s ON a.account_id = s.ancestor_id
    WHERE a.parent_id IS NOT NULL
),
parent_balances AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(lb.debit_total), 0) AS debit_total,
           COALESCE(SUM(lb.credit_total), 0) AS credit_total
    FROM accounts a
    JOIN subtree st ON a.account_id = st.ancestor_id
    JOIN leaf_balances lb ON lb.account_id = st.leaf_id
    WHERE a.is_summary = 1 AND a.is_active = 1
    GROUP BY a.account_id
)
SELECT account_id, name, type, normal_balance, debit_total, credit_total FROM leaf_balances
UNION ALL
SELECT account_id, name, type, normal_balance, debit_total, credit_total FROM parent_balances;

-- +goose Down
DROP VIEW IF EXISTS v_account_balances;
CREATE VIEW IF NOT EXISTS v_account_balances AS
SELECT a.account_id, a.name, a.type, a.normal_balance,
       COALESCE(SUM(je.debit), 0)  AS debit_total,
       COALESCE(SUM(je.credit), 0) AS credit_total
FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
WHERE a.is_active = 1
GROUP BY a.account_id;
