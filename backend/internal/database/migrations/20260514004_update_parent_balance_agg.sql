-- +goose Up
DROP VIEW IF EXISTS v_parent_balance_agg;
CREATE VIEW IF NOT EXISTS v_parent_balance_agg AS
SELECT snap.closing_id, snap.merchant_id, a.account_id,
       COALESCE(SUM(snap.debit_total),  0) AS debit_total,
       COALESCE(SUM(snap.credit_total), 0) AS credit_total
FROM accounts a
JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.merchant_id = a.merchant_id AND ac.depth > 0
JOIN accounts leaf      ON leaf.account_id = ac.descendant_id AND leaf.merchant_id = a.merchant_id
                       AND leaf.is_summary = 0 AND leaf.is_active = 1
JOIN account_balance_snapshots snap ON snap.account_id = ac.descendant_id AND snap.merchant_id = a.merchant_id
WHERE a.is_summary = 1
GROUP BY snap.closing_id, snap.merchant_id, a.account_id;

-- +goose Down
DROP VIEW IF EXISTS v_parent_balance_agg;
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS v_parent_balance_agg AS
SELECT snap.closing_id, a.account_id,
       COALESCE(SUM(snap.debit_total),  0) AS debit_total,
       COALESCE(SUM(snap.credit_total), 0) AS credit_total
FROM accounts a
JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.depth > 0
JOIN accounts leaf      ON leaf.account_id = ac.descendant_id
                       AND leaf.is_summary = 0 AND leaf.is_active = 1
JOIN account_balance_snapshots snap ON snap.account_id = ac.descendant_id
WHERE a.is_summary = 1
GROUP BY snap.closing_id, a.account_id;
-- +goose StatementEnd
