-- +goose Up

-- SQLite 3.26+ validates all views when renaming tables.
-- v_installments_active references i.paid_periods which does not exist (pre-existing bug).
-- Use legacy_alter_table to skip view validation during the renames below.
PRAGMA legacy_alter_table = ON;

-- aggregate_versions: PK (type, id) -> (type, id, merchant_id)
-- +goose StatementBegin
CREATE TABLE aggregate_versions_new (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION')),
    PRIMARY KEY (aggregate_type, aggregate_id, merchant_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO aggregate_versions_new (aggregate_type, aggregate_id, merchant_id, current_version)
SELECT aggregate_type, aggregate_id, 0, current_version FROM aggregate_versions;
-- +goose StatementEnd
DROP TABLE aggregate_versions;
-- +goose StatementBegin
ALTER TABLE aggregate_versions_new RENAME TO aggregate_versions;
-- +goose StatementEnd

-- snapshots: UNIQUE (type, id) -> (type, id, merchant_id)
-- +goose StatementBegin
CREATE TABLE snapshots_new (
    snapshot_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id    INTEGER NOT NULL DEFAULT 0,
    aggregate_type TEXT    NOT NULL,
    aggregate_id   TEXT    NOT NULL,
    at_version     INTEGER NOT NULL,
    state          TEXT    NOT NULL,
    created_at     TEXT    DEFAULT (datetime('now')),
    UNIQUE(aggregate_type, aggregate_id, merchant_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO snapshots_new (snapshot_id, merchant_id, aggregate_type, aggregate_id, at_version, state, created_at)
SELECT snapshot_id, 0, aggregate_type, aggregate_id, at_version, state, created_at FROM snapshots;
-- +goose StatementEnd
DROP TABLE snapshots;
-- +goose StatementBegin
ALTER TABLE snapshots_new RENAME TO snapshots;
-- +goose StatementEnd

-- event_store: UNIQUE (type, id, version) -> (type, id, version, merchant_id)
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
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION')),
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
    event_id, event_uuid, occurred_at, 0,
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
-- SQLite 不支援 DROP COLUMN，Down migration 僅提示需手動還原。
SELECT 1;
