-- name: GetAccountJournalEntriesPaged :many
WITH total AS (
    SELECT COUNT(*) AS cnt
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id
    WHERE je.account_id = @account_id
      AND je.merchant_id = @merchant_id
      AND t.status = 'ACTIVE'
      AND t.txn_date >= @from_date
      AND t.txn_date <= @to_date
)
SELECT je.entry_id, je.txn_id, je.account_id, je.ledger_id,
       je.debit, je.credit, je.note,
       t.txn_date, t.description,
       total.cnt AS total
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id, total
WHERE je.account_id = @account_id
  AND je.merchant_id = @merchant_id
  AND t.status = 'ACTIVE'
  AND t.txn_date >= @from_date
  AND t.txn_date <= @to_date
ORDER BY t.txn_date DESC, je.entry_id DESC
LIMIT @limit OFFSET @offset;

-- name: GetClosedMonthlyPeriodsInRange :many
SELECT closing_id, period_start, period_end
FROM period_closings
WHERE merchant_id = @merchant_id
  AND period_type = 'MONTHLY'
  AND status = 'CLOSED'
  AND period_end >= @from_date
  AND period_end <= @to_date
ORDER BY period_end ASC;

-- name: GetAccountSnapshotByClosingID :one
SELECT debit_total, credit_total
FROM account_balance_snapshots
WHERE closing_id = @closing_id AND account_id = @account_id AND merchant_id = @merchant_id;

-- name: GetAccountJournalEntriesInDateRange :many
SELECT je.debit, je.credit, t.txn_date
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id
WHERE je.account_id = @account_id
  AND je.merchant_id = @merchant_id
  AND t.status = 'ACTIVE'
  AND t.txn_date >= @from_date
  AND t.txn_date <= @to_date
ORDER BY t.txn_date ASC;

-- name: GetLatestAccountSnapshotBefore :one
SELECT abs.debit_total, abs.credit_total, pc.period_end
FROM account_balance_snapshots abs
JOIN period_closings pc ON pc.closing_id = abs.closing_id
WHERE abs.account_id = @account_id AND abs.merchant_id = @merchant_id
  AND pc.merchant_id = @merchant_id
  AND pc.period_type = 'MONTHLY' AND pc.status = 'CLOSED'
  AND pc.period_end < @before_date
ORDER BY pc.period_end DESC
LIMIT 1;
