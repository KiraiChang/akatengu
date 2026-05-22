-- +goose Up
CREATE TABLE fixed_assets (
    id                              INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id                     INTEGER NOT NULL,
    txn_id                          INTEGER,
    name                            TEXT    NOT NULL,
    asset_account_id                TEXT    NOT NULL,
    accum_depreciation_account_id   TEXT    NOT NULL,
    depreciation_expense_account_id TEXT    NOT NULL,
    cost                            REAL    NOT NULL,
    residual_value                  REAL    NOT NULL DEFAULT 0,
    useful_life_months              INTEGER NOT NULL,
    depreciation_method             TEXT    NOT NULL DEFAULT 'STRAIGHT_LINE' CHECK(depreciation_method IN ('STRAIGHT_LINE')),
    payment_type                    TEXT    NOT NULL CHECK(payment_type IN ('CASH','LEASE')),
    total_depreciated               REAL    NOT NULL DEFAULT 0,
    depreciated_periods             INTEGER NOT NULL DEFAULT 0,
    purchase_date                   TEXT    NOT NULL,
    disposal_date                   TEXT,
    status                          TEXT    NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','DISPOSED')),
    updated_by                      TEXT,
    updated_at                      TEXT    DEFAULT (datetime('now')),
    version                         INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE fixed_asset_depreciations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    asset_id    INTEGER NOT NULL REFERENCES fixed_assets(id),
    txn_id      INTEGER NOT NULL,
    period_date TEXT    NOT NULL,
    amount      REAL    NOT NULL,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS fixed_asset_depreciations;
DROP TABLE IF EXISTS fixed_assets;
