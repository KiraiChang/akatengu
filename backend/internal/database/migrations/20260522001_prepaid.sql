-- +goose Up
CREATE TABLE prepaids (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id         INTEGER NOT NULL,
    txn_id              INTEGER,
    account_id          TEXT    NOT NULL,
    expense_account_id  TEXT    NOT NULL,
    name                TEXT    NOT NULL,
    total_amount        REAL    NOT NULL,
    amortized_amount    REAL    NOT NULL DEFAULT 0,
    periods             INTEGER NOT NULL,
    amortized_periods   INTEGER NOT NULL DEFAULT 0,
    start_date          TEXT    NOT NULL,
    status              TEXT    NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','COMPLETED','DISPOSED')),
    updated_by          TEXT,
    updated_at          TEXT    DEFAULT (datetime('now')),
    version             INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE prepaid_amortizations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    prepaid_id  INTEGER NOT NULL REFERENCES prepaids(id),
    txn_id      INTEGER NOT NULL,
    period_date TEXT    NOT NULL,
    amount      REAL    NOT NULL,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS prepaid_amortizations;
DROP TABLE IF EXISTS prepaids;
