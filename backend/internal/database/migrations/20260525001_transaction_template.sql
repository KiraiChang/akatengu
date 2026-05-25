-- +goose Up
CREATE TABLE transaction_templates (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    name        TEXT    NOT NULL,
    description TEXT,
    tag         TEXT,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now')),
    version     INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE transaction_template_entries (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id        INTEGER NOT NULL,
    template_id        INTEGER NOT NULL,
    sort_order         INTEGER NOT NULL DEFAULT 0,
    account_id         TEXT    NOT NULL,
    ledger_id          INTEGER,
    debit              REAL    NOT NULL DEFAULT 0,
    credit             REAL    NOT NULL DEFAULT 0,
    note               TEXT,
    cash_flow_category TEXT CHECK(cash_flow_category IN ('CASH','OPERATING','INVESTING','FINANCING')
                            OR cash_flow_category IS NULL),
    FOREIGN KEY(template_id) REFERENCES transaction_templates(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS transaction_template_entries;
DROP TABLE IF EXISTS transaction_templates;
