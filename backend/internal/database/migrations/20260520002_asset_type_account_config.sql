-- +goose Up
CREATE TABLE asset_type_account_config (
    id                         INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id                INTEGER NOT NULL,
    asset_type                 TEXT    NOT NULL CHECK(asset_type IN ('STOCK','FUND','GOLD','FX')),
    realized_gain_account_id   TEXT    NOT NULL,
    realized_loss_account_id   TEXT    NOT NULL,
    unrealized_gain_account_id TEXT    NOT NULL,
    unrealized_loss_account_id TEXT    NOT NULL,
    oci_account_id             TEXT,
    fee_account_id             TEXT    NOT NULL,
    tax_account_id             TEXT    NOT NULL,
    updated_by                 TEXT,
    updated_at                 TEXT    DEFAULT (datetime('now')),
    version                    INTEGER NOT NULL DEFAULT 1,
    UNIQUE(merchant_id, asset_type)
);

-- +goose Down
DROP TABLE IF EXISTS asset_type_account_config;
