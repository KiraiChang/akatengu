-- +goose Up
CREATE TABLE IF NOT EXISTS account_balance_snapshots (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    closing_id   INTEGER NOT NULL REFERENCES period_closings(closing_id),
    account_id   TEXT    NOT NULL REFERENCES accounts(account_id),
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    UNIQUE(closing_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_abs_closing ON account_balance_snapshots(closing_id);

-- +goose Down
DROP INDEX IF EXISTS idx_abs_closing;
DROP TABLE IF EXISTS account_balance_snapshots;