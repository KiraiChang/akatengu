-- +goose Up
-- period_closings was unique on (period_type, period_start), so only one
-- merchant could open a period (backend ISSUE-030). SQLite cannot drop a table
-- UNIQUE constraint, so the table is rebuilt with the merchant in the key.
--
-- The table keeps its name throughout: the snapshot tables reference
-- period_closings(closing_id), and rows are written back with their original
-- closing_id. With foreign keys on, the violations caused by DROP TABLE are
-- deferred to commit and cleared by the write-back; with them off the pragma
-- has no effect. It resets when the migration transaction ends.
PRAGMA defer_foreign_keys = ON;

CREATE TABLE period_closings_backup AS SELECT * FROM period_closings;
-- DROP TABLE removes the AUTOINCREMENT counter; keep it so deleted IDs are not reused.
CREATE TABLE period_closings_seq_backup AS
SELECT seq FROM sqlite_sequence WHERE name = 'period_closings';

DROP TABLE period_closings;

CREATE TABLE period_closings (
    closing_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    period_type    TEXT    NOT NULL,
    period_start   TEXT    NOT NULL,
    period_end     TEXT    NOT NULL,
    status         TEXT    NOT NULL DEFAULT 'OPEN',
    opening_txn_id INTEGER,
    closing_txn_id INTEGER,
    snapshot       TEXT,
    closed_at      TEXT,
    note           TEXT,
    reopen_at      TEXT,
    reopen_reason  TEXT,
    merchant_id    INTEGER NOT NULL DEFAULT 0,
    updated_by     TEXT,
    updated_at     TEXT    DEFAULT (datetime('now')),
    closing_uuid   TEXT    NOT NULL DEFAULT '',
    CONSTRAINT chk_period_type CHECK (period_type IN ('MONTHLY', 'ANNUAL')),
    CONSTRAINT chk_status CHECK (status IN ('OPEN', 'CLOSED', 'REOPENED')),
    CONSTRAINT chk_period_dates CHECK (period_start <= period_end),
    CONSTRAINT chk_closed_at CHECK (status != 'CLOSED' OR closed_at IS NOT NULL),
    UNIQUE(merchant_id, period_type, period_start)
);

INSERT INTO period_closings
    (closing_id, period_type, period_start, period_end, status, opening_txn_id, closing_txn_id,
     snapshot, closed_at, note, reopen_at, reopen_reason, merchant_id, updated_by, updated_at, closing_uuid)
SELECT closing_id, period_type, period_start, period_end, status, opening_txn_id, closing_txn_id,
       snapshot, closed_at, note, reopen_at, reopen_reason, merchant_id, updated_by, updated_at, closing_uuid
FROM period_closings_backup;

-- The write-back sets the counter to the largest closing_id; restore the larger saved value.
UPDATE sqlite_sequence
SET seq = (SELECT seq FROM period_closings_seq_backup)
WHERE name = 'period_closings'
  AND seq < (SELECT seq FROM period_closings_seq_backup);
INSERT INTO sqlite_sequence (name, seq)
SELECT 'period_closings', seq FROM period_closings_seq_backup
WHERE NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = 'period_closings');

DROP TABLE period_closings_backup;
DROP TABLE period_closings_seq_backup;

CREATE INDEX IF NOT EXISTS idx_period_closings_status   ON period_closings(status);
CREATE INDEX IF NOT EXISTS idx_period_closings_end      ON period_closings(period_end);
CREATE INDEX IF NOT EXISTS idx_period_closings_merchant ON period_closings(merchant_id, status);

-- +goose Down
-- No-op: restoring UNIQUE(period_type, period_start) would bring ISSUE-030 back,
-- and fails once two merchants have opened the same period.
