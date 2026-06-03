-- +goose Up

-- bank_csv_templates
CREATE TABLE bank_csv_templates (
    template_id        INTEGER PRIMARY KEY AUTOINCREMENT,
    template_uuid      TEXT    NOT NULL UNIQUE,
    merchant_id        INTEGER NOT NULL,
    template_name      TEXT    NOT NULL,
    encoding           TEXT    NOT NULL DEFAULT 'UTF-8',
    skip_rows          INTEGER NOT NULL DEFAULT 1,
    date_column        INTEGER NOT NULL,
    date_format        TEXT    NOT NULL,
    description_column INTEGER NOT NULL,
    debit_column       INTEGER,
    credit_column      INTEGER,
    amount_column      INTEGER,
    balance_column     INTEGER,
    reference_column   INTEGER,
    is_active          INTEGER NOT NULL DEFAULT 1,
    note               TEXT,
    version            INTEGER NOT NULL DEFAULT 0,
    updated_by         TEXT,
    updated_at         TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_bct_merchant ON bank_csv_templates(merchant_id);

-- bank_statement_imports
CREATE TABLE bank_statement_imports (
    import_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    import_uuid    TEXT    NOT NULL UNIQUE,
    merchant_id    INTEGER NOT NULL,
    ledger_id      INTEGER NOT NULL REFERENCES ledger_accounts(ledger_id),
    template_id    INTEGER REFERENCES bank_csv_templates(template_id),
    statement_date TEXT    NOT NULL,
    import_source  TEXT    NOT NULL DEFAULT 'CSV' CHECK (import_source IN ('CSV')),
    filename       TEXT,
    status         TEXT    NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'REVIEWING', 'COMPLETED')),
    note           TEXT,
    version        INTEGER NOT NULL DEFAULT 0,
    updated_by     TEXT,
    updated_at     TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_bsi_merchant ON bank_statement_imports(merchant_id);
CREATE INDEX IF NOT EXISTS idx_bsi_ledger   ON bank_statement_imports(merchant_id, ledger_id);
CREATE INDEX IF NOT EXISTS idx_bsi_status   ON bank_statement_imports(merchant_id, status);

-- bank_statement_txns
CREATE TABLE bank_statement_txns (
    bank_txn_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    bank_txn_uuid    TEXT    NOT NULL UNIQUE,
    import_id        INTEGER NOT NULL REFERENCES bank_statement_imports(import_id),
    merchant_id      INTEGER NOT NULL,
    txn_date         TEXT    NOT NULL,
    description      TEXT    NOT NULL,
    debit            REAL    NOT NULL DEFAULT 0,
    credit           REAL    NOT NULL DEFAULT 0,
    balance          REAL,
    reference_no     TEXT,
    match_status     TEXT    NOT NULL DEFAULT 'UNMATCHED'
        CHECK (match_status IN ('UNMATCHED', 'MATCHED', 'IGNORED', 'APPROVED')),
    matched_entry_id INTEGER REFERENCES journal_entries(entry_id),
    match_confidence TEXT    CHECK (match_confidence IN ('EXACT', 'FUZZY', 'CONFIRMED', 'MANUAL')),
    created_txn_id   INTEGER REFERENCES transactions(txn_id)
);

CREATE INDEX IF NOT EXISTS idx_bst_import  ON bank_statement_txns(import_id);
CREATE INDEX IF NOT EXISTS idx_bst_status  ON bank_statement_txns(import_id, match_status);
CREATE INDEX IF NOT EXISTS idx_bst_entry   ON bank_statement_txns(matched_entry_id);

-- Update aggregate_versions CHECK constraint to include new aggregate types
PRAGMA legacy_alter_table = ON;

-- +goose StatementBegin
CREATE TABLE aggregate_versions_new (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN (
        'ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG',
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY',
        'BANK_CSV_TEMPLATE', 'BANK_STATEMENT'
    )),
    PRIMARY KEY (aggregate_type, aggregate_id, merchant_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO aggregate_versions_new SELECT * FROM aggregate_versions;
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
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY',
        'BANK_CSV_TEMPLATE', 'BANK_STATEMENT'
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

PRAGMA legacy_alter_table = OFF;

-- +goose Down
DROP TABLE IF EXISTS bank_statement_txns;
DROP TABLE IF EXISTS bank_statement_imports;
DROP TABLE IF EXISTS bank_csv_templates;
