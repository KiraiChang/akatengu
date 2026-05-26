-- +goose Up
ALTER TABLE asset_type_account_config ADD COLUMN account_id TEXT;

-- +goose Down
SELECT 1;
