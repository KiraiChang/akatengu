-- +goose Up
ALTER TABLE journal_entries DROP COLUMN cash_flow_category;

-- +goose Down
ALTER TABLE journal_entries ADD COLUMN cash_flow_category TEXT CHECK (cash_flow_category IN ('OPERATING', 'INVESTING', 'FINANCING'));
