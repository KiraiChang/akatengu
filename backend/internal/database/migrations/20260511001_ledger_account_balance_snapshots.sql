-- +goose Up
CREATE TABLE IF NOT EXISTS ledger_account_balance_snapshots (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    closing_id   INTEGER NOT NULL REFERENCES period_closings(closing_id),
    ledger_id    INTEGER NOT NULL REFERENCES ledger_accounts(ledger_id),
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    UNIQUE(closing_id, ledger_id)
);

CREATE INDEX IF NOT EXISTS idx_labs_closing ON ledger_account_balance_snapshots(closing_id);

-- +goose Down
DROP INDEX IF EXISTS idx_labs_closing;
DROP TABLE IF EXISTS ledger_account_balance_snapshots;
