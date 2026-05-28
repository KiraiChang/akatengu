-- +goose Up

-- SQLite does not support ALTER TABLE ... DROP/MODIFY CONSTRAINT.
-- Recreate both tables with the updated chk_aggregate_type CHECK constraint
-- to include SYS_CONFIG and ACCOUNT_CONFIG aggregate types.
-- Use PRAGMA legacy_alter_table = ON to skip view validation during renames.
PRAGMA legacy_alter_table = ON;

-- aggregate_versions
-- +goose StatementBegin
CREATE TABLE aggregate_versions_new (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG')),
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

-- event_store
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
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG')),
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

PRAGMA legacy_alter_table = OFF;

-- +goose Down
SELECT 1;
