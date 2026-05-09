-- name: DeleteAccountBalanceSnapshotsByClosingId :exec
DELETE FROM account_balance_snapshots WHERE closing_id = ?;


-- name: BulkInsertBalanceSnapshot :exec
INSERT OR REPLACE INTO account_balance_snapshots (closing_id, account_id, debit_total, credit_total)
SELECT
    :closing_id,
    a.account_id,
    COALESCE(ps.debit_total, 0) + COALESCE(d.debit_sum, 0),
    COALESCE(ps.credit_total, 0) + COALESCE(d.credit_sum, 0)
FROM accounts a
LEFT JOIN (
    SELECT s.account_id, s.debit_total, s.credit_total
    FROM account_balance_snapshots s
    WHERE s.closing_id = (
        SELECT MAX(pc.closing_id)
        FROM period_closings pc
        WHERE pc.status = 'CLOSED'
          AND pc.period_end < (SELECT period_start FROM period_closings WHERE closing_id = :closing_id)
    )
) ps ON ps.account_id = a.account_id
LEFT JOIN (
    SELECT je.account_id,
           SUM(je.debit)  AS debit_sum,
           SUM(je.credit) AS credit_sum
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE t.txn_date BETWEEN
        (SELECT period_start FROM period_closings WHERE closing_id = :closing_id) AND
        (SELECT period_end   FROM period_closings WHERE closing_id = :closing_id)
    GROUP BY je.account_id
) d ON d.account_id = a.account_id
WHERE a.is_active = 1
GROUP BY a.account_id