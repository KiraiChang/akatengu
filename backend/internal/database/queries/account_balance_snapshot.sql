-- name: DeleteAccountBalanceSnapshotsByClosingId :exec
DELETE FROM account_balance_snapshots WHERE closing_id = ?;


-- name: BulkInsertBalanceSnapshot :exec
INSERT OR REPLACE INTO account_balance_snapshots (closing_id, account_id, debit_total, credit_total)
SELECT ?, a.account_id,
       COALESCE(SUM(je.debit), 0),
       COALESCE(SUM(je.credit), 0)
FROM accounts a
         LEFT JOIN journal_entries je ON a.account_id = je.account_id
         LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
WHERE a.is_active = 1
GROUP BY a.account_id