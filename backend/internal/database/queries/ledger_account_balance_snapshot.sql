-- name: DeleteLedgerAccountBalanceSnapshotsByClosingId :exec
DELETE FROM ledger_account_balance_snapshots WHERE closing_id = @closing_id AND merchant_id = @merchant_id;

-- name: BulkInsertLedgerBalanceSnapshot :exec
INSERT OR REPLACE INTO ledger_account_balance_snapshots (merchant_id, closing_id, ledger_id, debit_total, credit_total)
SELECT
    :merchant_id,
    :closing_id,
    la.ledger_id,
    COALESCE(ps.debit_total, 0) + COALESCE(d.debit_sum, 0),
    COALESCE(ps.credit_total, 0) + COALESCE(d.credit_sum, 0)
FROM ledger_accounts la
LEFT JOIN (
    SELECT s.ledger_id, s.debit_total, s.credit_total
    FROM ledger_account_balance_snapshots s
    WHERE s.merchant_id = :merchant_id
      AND s.closing_id = (
        SELECT MAX(pc.closing_id)
        FROM period_closings pc
        WHERE pc.merchant_id = :merchant_id
          AND pc.status = 'CLOSED'
          AND pc.period_end < (SELECT period_start FROM period_closings WHERE closing_id = :closing_id AND merchant_id = :merchant_id)
    )
) ps ON ps.ledger_id = la.ledger_id
LEFT JOIN (
    SELECT je.ledger_id,
           SUM(je.debit)  AS debit_sum,
           SUM(je.credit) AS credit_sum
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE je.merchant_id = :merchant_id
      AND t.txn_date BETWEEN
        (SELECT period_start FROM period_closings WHERE closing_id = :closing_id AND merchant_id = :merchant_id) AND
        (SELECT period_end   FROM period_closings WHERE closing_id = :closing_id AND merchant_id = :merchant_id)
      AND je.ledger_id IS NOT NULL
    GROUP BY je.ledger_id
) d ON d.ledger_id = la.ledger_id
WHERE la.is_active = 1 AND la.merchant_id = :merchant_id;
