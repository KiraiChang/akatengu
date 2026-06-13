-- +goose Up

-- bank_csv_template_ledgers: CSV/XLSX 模板的帳本清單（1 個模板可對應多個帳本）
CREATE TABLE bank_csv_template_ledgers (
    tpl_ledger_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    tpl_ledger_uuid  TEXT    NOT NULL UNIQUE,
    template_id      INTEGER NOT NULL REFERENCES bank_csv_templates(template_id),
    ledger_uuid      TEXT    NOT NULL,
    ledger_id        INTEGER REFERENCES ledger_accounts(ledger_id),
    account_type     TEXT    NOT NULL,
    sort_order       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_bctl_template ON bank_csv_template_ledgers(template_id);

-- bank_pdf_templates: PDF 對帳單主檔
CREATE TABLE bank_pdf_templates (
    template_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    template_uuid  TEXT    NOT NULL UNIQUE,
    merchant_id    INTEGER NOT NULL,
    template_name  TEXT    NOT NULL,
    bank_type      TEXT    NOT NULL,
    is_active      INTEGER NOT NULL DEFAULT 1,
    version        INTEGER NOT NULL DEFAULT 0,
    updated_by     TEXT,
    updated_at     TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_bpt_merchant ON bank_pdf_templates(merchant_id);

-- bank_pdf_template_ledgers: PDF 模板的帳本清單（1 個模板可對應多個帳本）
CREATE TABLE bank_pdf_template_ledgers (
    tpl_ledger_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    tpl_ledger_uuid  TEXT    NOT NULL UNIQUE,
    template_id      INTEGER NOT NULL REFERENCES bank_pdf_templates(template_id),
    ledger_uuid      TEXT    NOT NULL,
    ledger_id        INTEGER REFERENCES ledger_accounts(ledger_id),
    account_type     TEXT    NOT NULL,
    sort_order       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_bptl_template ON bank_pdf_template_ledgers(template_id);

-- bank_statement_import_ledgers: 每筆 import 實際關聯的帳本清單（CSV/XLSX/PDF 共用）
CREATE TABLE bank_statement_import_ledgers (
    import_ledger_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    import_ledger_uuid  TEXT    NOT NULL UNIQUE,
    import_id           INTEGER NOT NULL REFERENCES bank_statement_imports(import_id),
    ledger_uuid         TEXT    NOT NULL,
    ledger_id           INTEGER REFERENCES ledger_accounts(ledger_id),
    account_type        TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bsil_import ON bank_statement_import_ledgers(import_id);

-- 重建 bank_statement_imports：
--   1. import_source CHECK 加入 'XLSX' 和 'PDF'
--   2. 新增 template_uuid、pdf_template_id、bank_type 欄位
PRAGMA legacy_alter_table = ON;

-- +goose StatementBegin
CREATE TABLE bank_statement_imports_new (
    import_id         INTEGER PRIMARY KEY AUTOINCREMENT,
    import_uuid       TEXT    NOT NULL UNIQUE,
    merchant_id       INTEGER NOT NULL,
    ledger_id         INTEGER NOT NULL REFERENCES ledger_accounts(ledger_id),
    template_id       INTEGER REFERENCES bank_csv_templates(template_id),
    template_uuid     TEXT,
    pdf_template_id   INTEGER REFERENCES bank_pdf_templates(template_id),
    pdf_template_uuid TEXT,
    bank_type         TEXT,
    statement_date    TEXT    NOT NULL,
    import_source     TEXT    NOT NULL DEFAULT 'CSV' CHECK (import_source IN ('CSV', 'XLSX', 'PDF')),
    filename          TEXT,
    status            TEXT    NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'REVIEWING', 'COMPLETED')),
    note              TEXT,
    version           INTEGER NOT NULL DEFAULT 0,
    updated_by        TEXT,
    updated_at        TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO bank_statement_imports_new (
    import_id, import_uuid, merchant_id, ledger_id, template_id, template_uuid,
    pdf_template_id, pdf_template_uuid, bank_type,
    statement_date, import_source, filename, status, note, version, updated_by, updated_at
)
SELECT
    import_id, import_uuid, merchant_id, ledger_id, template_id, NULL,
    NULL, NULL, NULL,
    statement_date, import_source, filename, status, note, version, updated_by, updated_at
FROM bank_statement_imports;
-- +goose StatementEnd
DROP TABLE bank_statement_imports;
-- +goose StatementBegin
ALTER TABLE bank_statement_imports_new RENAME TO bank_statement_imports;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_bsi_merchant ON bank_statement_imports(merchant_id);
CREATE INDEX IF NOT EXISTS idx_bsi_ledger   ON bank_statement_imports(merchant_id, ledger_id);
CREATE INDEX IF NOT EXISTS idx_bsi_status   ON bank_statement_imports(merchant_id, status);

-- 新增 bank_statement_txns 的 ledger 欄位（多帳本交易歸屬）
ALTER TABLE bank_statement_txns ADD COLUMN ledger_uuid TEXT;
ALTER TABLE bank_statement_txns ADD COLUMN ledger_id   INTEGER REFERENCES ledger_accounts(ledger_id);

-- 更新 aggregate_versions CHECK 約束，加入 BANK_PDF_TEMPLATE
-- +goose StatementBegin
CREATE TABLE aggregate_versions_new (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN (
        'ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG',
        'FIXED_ASSET_CATEGORY', 'PREPAID_CATEGORY',
        'BANK_CSV_TEMPLATE', 'BANK_STATEMENT', 'BANK_PDF_TEMPLATE'
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

-- 更新 event_store CHECK 約束，加入 BANK_PDF_TEMPLATE
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
        'BANK_CSV_TEMPLATE', 'BANK_STATEMENT', 'BANK_PDF_TEMPLATE'
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
DROP TABLE IF EXISTS bank_statement_import_ledgers;
DROP TABLE IF EXISTS bank_pdf_template_ledgers;
DROP TABLE IF EXISTS bank_pdf_templates;
DROP TABLE IF EXISTS bank_csv_template_ledgers;
