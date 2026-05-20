-- +goose Up
CREATE TABLE ledger_account_type_config (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    type        TEXT    NOT NULL CHECK(type IN ('BANK_ACCOUNT','LOAN','CREDIT_CARD')),
    account_id  TEXT    NOT NULL,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now')),
    version     INTEGER NOT NULL DEFAULT 1,
    UNIQUE(merchant_id, type),
    FOREIGN KEY(account_id, merchant_id) REFERENCES accounts(account_id, merchant_id)
);

-- +goose Down
DROP TABLE IF EXISTS ledger_account_type_config;
