-- +goose Up
CREATE TABLE IF NOT EXISTS entry_cf_categories (
    entry_uuid   TEXT    NOT NULL,
    merchant_id  INTEGER NOT NULL,
    cf_category  TEXT    CHECK (cf_category IN ('OPERATING', 'INVESTING', 'FINANCING')),
    is_confirmed INTEGER NOT NULL DEFAULT 0,
    updated_at   TEXT,
    updated_by   TEXT,
    PRIMARY KEY (entry_uuid, merchant_id)
);

CREATE INDEX IF NOT EXISTS idx_entry_cf_categories_merchant ON entry_cf_categories(merchant_id);

-- migrate existing data from journal_entries
INSERT OR IGNORE INTO entry_cf_categories (entry_uuid, merchant_id, cf_category, is_confirmed, updated_at)
SELECT je.entry_uuid, je.merchant_id, je.cash_flow_category, 1, datetime('now')
FROM journal_entries je
WHERE je.cash_flow_category IS NOT NULL AND je.cash_flow_category != ''
  AND je.entry_uuid IS NOT NULL AND je.entry_uuid != '';

-- +goose Down
DROP INDEX IF EXISTS idx_entry_cf_categories_merchant;
DROP TABLE IF EXISTS entry_cf_categories;
