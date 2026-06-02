-- +goose Up

-- fixed_asset_categories
CREATE TABLE fixed_asset_categories (
    id                              INTEGER PRIMARY KEY AUTOINCREMENT,
    category_uuid                   TEXT    NOT NULL UNIQUE,
    merchant_id                     INTEGER NOT NULL,
    name                            TEXT    NOT NULL,
    asset_account_id                TEXT    NOT NULL,
    accum_depreciation_account_id   TEXT    NOT NULL,
    depreciation_expense_account_id TEXT    NOT NULL,
    is_active                       INTEGER NOT NULL DEFAULT 1,
    updated_by                      TEXT,
    updated_at                      TEXT    NOT NULL DEFAULT (datetime('now')),
    version                         INTEGER NOT NULL DEFAULT 1
);

-- prepaid_categories
CREATE TABLE prepaid_categories (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    category_uuid     TEXT    NOT NULL UNIQUE,
    merchant_id       INTEGER NOT NULL,
    name              TEXT    NOT NULL,
    account_id        TEXT    NOT NULL,
    expense_account_id TEXT   NOT NULL,
    is_active         INTEGER NOT NULL DEFAULT 1,
    updated_by        TEXT,
    updated_at        TEXT    NOT NULL DEFAULT (datetime('now')),
    version           INTEGER NOT NULL DEFAULT 1
);

-- Update aggregate_versions to include new aggregate types
-- SQLite requires table recreation to modify CHECK constraints
PRAGMA legacy_alter_table = ON;

-- +goose StatementBegin
CREATE TABLE aggregate_versions_new (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN (
        'ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG',
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY'
    )),
    PRIMARY KEY (aggregate_type, aggregate_id, merchant_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO aggregate_versions_new (aggregate_type, aggregate_id, merchant_id, current_version)
SELECT aggregate_type, aggregate_id, merchant_id, current_version FROM aggregate_versions;
-- +goose StatementEnd
DROP TABLE aggregate_versions;
-- +goose StatementBegin
ALTER TABLE aggregate_versions_new RENAME TO aggregate_versions;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE event_store_new (
    event_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_uuid        TEXT    NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(16)))),
    occurred_at       TEXT    NOT NULL DEFAULT (datetime('now')),
    merchant_id       INTEGER NOT NULL DEFAULT 0,
    aggregate_type    TEXT    NOT NULL,
    aggregate_id      TEXT    NOT NULL,
    aggregate_version INTEGER NOT NULL,
    event_type        TEXT    NOT NULL,
    payload           TEXT    NOT NULL,
    metadata          TEXT,
    updated_by        TEXT,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN (
        'ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG',
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY'
    )),
    UNIQUE(aggregate_type, aggregate_id, aggregate_version, merchant_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO event_store_new (
    event_id, event_uuid, occurred_at, merchant_id,
    aggregate_type, aggregate_id, aggregate_version,
    event_type, payload, metadata, updated_by
)
SELECT
    event_id, event_uuid, occurred_at, merchant_id,
    aggregate_type, aggregate_id, aggregate_version,
    event_type, payload, metadata, updated_by
FROM event_store;
-- +goose StatementEnd
DROP TABLE event_store;
-- +goose StatementBegin
ALTER TABLE event_store_new RENAME TO event_store;
-- +goose StatementEnd
CREATE INDEX idx_es_aggregate ON event_store(aggregate_type, aggregate_id, merchant_id);
CREATE INDEX idx_es_type      ON event_store(event_type);
CREATE INDEX idx_es_occurred  ON event_store(occurred_at);

-- Fix fixed_assets payment_type constraint to include INSTALLMENT
-- +goose StatementBegin
CREATE TABLE fixed_assets_new (
    id                              INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id                     INTEGER NOT NULL,
    asset_uuid                      TEXT,
    txn_id                          INTEGER,
    name                            TEXT    NOT NULL,
    asset_account_id                TEXT    NOT NULL,
    accum_depreciation_account_id   TEXT    NOT NULL,
    depreciation_expense_account_id TEXT    NOT NULL,
    cost                            REAL    NOT NULL,
    residual_value                  REAL    NOT NULL DEFAULT 0,
    useful_life_months              INTEGER NOT NULL,
    depreciation_method             TEXT    NOT NULL DEFAULT 'STRAIGHT_LINE' CHECK(depreciation_method IN ('STRAIGHT_LINE')),
    payment_type                    TEXT    NOT NULL CHECK(payment_type IN ('CASH','LEASE','INSTALLMENT')),
    total_depreciated               REAL    NOT NULL DEFAULT 0,
    depreciated_periods             INTEGER NOT NULL DEFAULT 0,
    purchase_date                   TEXT    NOT NULL,
    disposal_date                   TEXT,
    status                          TEXT    NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','DISPOSED')),
    updated_by                      TEXT,
    updated_at                      TEXT    DEFAULT (datetime('now')),
    version                         INTEGER NOT NULL DEFAULT 1
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO fixed_assets_new (
    id, merchant_id, asset_uuid, txn_id, name,
    asset_account_id, accum_depreciation_account_id, depreciation_expense_account_id,
    cost, residual_value, useful_life_months, depreciation_method, payment_type,
    total_depreciated, depreciated_periods, purchase_date, disposal_date, status,
    updated_by, updated_at, version
)
SELECT
    id, merchant_id, asset_uuid, txn_id, name,
    asset_account_id, accum_depreciation_account_id, depreciation_expense_account_id,
    cost, residual_value, useful_life_months, depreciation_method, payment_type,
    total_depreciated, depreciated_periods, purchase_date, disposal_date, status,
    updated_by, updated_at, version
FROM fixed_assets;
-- +goose StatementEnd
DROP TABLE fixed_assets;
-- +goose StatementBegin
ALTER TABLE fixed_assets_new RENAME TO fixed_assets;
-- +goose StatementEnd

PRAGMA legacy_alter_table = OFF;

-- +goose Down
DROP TABLE IF EXISTS prepaid_categories;
DROP TABLE IF EXISTS fixed_asset_categories;
