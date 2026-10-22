-- +goose Up
-- BFS R2 segment 5 writes period openings to the merchant's PERIOD aggregate
-- (BFS R2 D8, ADR-024 A7). SQLite cannot alter a CHECK constraint, so
-- event_store and aggregate_versions are rebuilt with PERIOD added to
-- chk_aggregate_type; everything else is declared as before.
--
-- No foreign key references either table. Both keep their name throughout and
-- rows are written back with their original keys; the event_store
-- AUTOINCREMENT counter is restored, so a deleted event_id is never reused.

CREATE TABLE aggregate_versions_backup AS SELECT * FROM aggregate_versions;
DROP TABLE aggregate_versions;

CREATE TABLE aggregate_versions (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN (
        'ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG',
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY',
        'BANK_CSV_TEMPLATE', 'BANK_STATEMENT', 'BANK_PDF_TEMPLATE',
        'PERIOD'
    )),
    PRIMARY KEY (aggregate_type, aggregate_id, merchant_id)
);

INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
SELECT aggregate_type, aggregate_id, merchant_id, current_version FROM aggregate_versions_backup;

DROP TABLE aggregate_versions_backup;

CREATE TABLE event_store_backup AS SELECT * FROM event_store;
-- DROP TABLE removes the AUTOINCREMENT counter; keep it so deleted IDs are not reused.
CREATE TABLE event_store_seq_backup AS
SELECT seq FROM sqlite_sequence WHERE name = 'event_store';

DROP TABLE event_store;

CREATE TABLE event_store (
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
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY',
        'BANK_CSV_TEMPLATE', 'BANK_STATEMENT', 'BANK_PDF_TEMPLATE',
        'PERIOD'
    )),
    UNIQUE(aggregate_type, aggregate_id, aggregate_version, merchant_id)
);

INSERT INTO event_store
    (event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id, aggregate_version,
     event_type, payload, metadata, updated_by)
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id, aggregate_version,
       event_type, payload, metadata, updated_by
FROM event_store_backup;

-- The write-back sets the counter to the largest event_id; restore the larger saved value.
UPDATE sqlite_sequence
SET seq = (SELECT seq FROM event_store_seq_backup)
WHERE name = 'event_store'
  AND seq < (SELECT seq FROM event_store_seq_backup);
INSERT INTO sqlite_sequence (name, seq)
SELECT 'event_store', seq FROM event_store_seq_backup
WHERE NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = 'event_store');

DROP TABLE event_store_backup;
DROP TABLE event_store_seq_backup;

CREATE INDEX idx_es_aggregate ON event_store(aggregate_type, aggregate_id, merchant_id);
CREATE INDEX idx_es_type      ON event_store(event_type);
CREATE INDEX idx_es_occurred  ON event_store(occurred_at);

-- +goose Down
-- No-op: removing PERIOD from the constraint fails once a period has been
-- opened through BFS.
