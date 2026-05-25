-- +goose Up
CREATE INDEX IF NOT EXISTS idx_je_merchant_account
    ON journal_entries(merchant_id, account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_je_merchant_account;
