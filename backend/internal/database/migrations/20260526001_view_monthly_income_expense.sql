-- +goose Up
-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS v_monthly_income_expense AS
SELECT
    je.merchant_id,
    substr(t.txn_date, 1, 7) AS month,
    COALESCE(SUM(CASE WHEN a.type = 'INCOME'  THEN je.credit - je.debit ELSE 0 END), 0) AS income,
    COALESCE(SUM(CASE WHEN a.type = 'EXPENSE' THEN je.debit - je.credit ELSE 0 END), 0) AS expense
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND je.merchant_id = t.merchant_id
JOIN accounts a     ON je.account_id = a.account_id AND je.merchant_id = a.merchant_id
WHERE t.status = 'ACTIVE'
  AND a.is_summary = 0
  AND a.type IN ('INCOME', 'EXPENSE')
GROUP BY je.merchant_id, substr(t.txn_date, 1, 7);
-- +goose StatementEnd

-- +goose Down
DROP VIEW IF EXISTS v_monthly_income_expense;
