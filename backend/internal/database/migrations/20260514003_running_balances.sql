-- +goose Up
CREATE TABLE account_running_balances (
    account_id   TEXT    NOT NULL,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    PRIMARY KEY (account_id, merchant_id),
    FOREIGN KEY (account_id) REFERENCES accounts(account_id)
);

CREATE TABLE ledger_running_balances (
    ledger_id    INTEGER NOT NULL,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    PRIMARY KEY (ledger_id, merchant_id),
    FOREIGN KEY (ledger_id) REFERENCES ledger_accounts(ledger_id)
);

-- 從現有 ACTIVE 交易回填葉科目餘額（排除沖銷參照交易）
INSERT OR IGNORE INTO account_running_balances (account_id, merchant_id, debit_total, credit_total)
SELECT je.account_id, je.merchant_id,
    COALESCE(SUM(je.debit), 0),
    COALESCE(SUM(je.credit), 0)
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND t.merchant_id = je.merchant_id
WHERE t.status = 'ACTIVE' AND t.ref_txn_id IS NULL
GROUP BY je.account_id, je.merchant_id;

-- 回填摘要科目（透過 account_closure 向上聚合）
INSERT INTO account_running_balances (account_id, merchant_id, debit_total, credit_total)
SELECT ac.ancestor_id, je.merchant_id,
    COALESCE(SUM(je.debit), 0),
    COALESCE(SUM(je.credit), 0)
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND t.merchant_id = je.merchant_id
JOIN account_closure ac ON ac.descendant_id = je.account_id
    AND ac.depth > 0 AND ac.merchant_id = je.merchant_id
WHERE t.status = 'ACTIVE' AND t.ref_txn_id IS NULL
GROUP BY ac.ancestor_id, je.merchant_id
ON CONFLICT(account_id, merchant_id) DO UPDATE SET
    debit_total = debit_total + excluded.debit_total,
    credit_total = credit_total + excluded.credit_total;

-- 回填 ledger 餘額
INSERT OR IGNORE INTO ledger_running_balances (ledger_id, merchant_id, debit_total, credit_total)
SELECT je.ledger_id, je.merchant_id,
    COALESCE(SUM(je.debit), 0),
    COALESCE(SUM(je.credit), 0)
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND t.merchant_id = je.merchant_id
WHERE t.status = 'ACTIVE' AND t.ref_txn_id IS NULL AND je.ledger_id IS NOT NULL
GROUP BY je.ledger_id, je.merchant_id;

-- +goose Down
DROP TABLE IF EXISTS ledger_running_balances;
DROP TABLE IF EXISTS account_running_balances;
