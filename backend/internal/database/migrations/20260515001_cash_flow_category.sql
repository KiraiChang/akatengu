-- +goose Up
ALTER TABLE accounts ADD COLUMN cash_flow_category TEXT
    CHECK (cash_flow_category IN ('CASH', 'OPERATING', 'INVESTING', 'FINANCING'));

-- +goose Down
ALTER TABLE accounts DROP COLUMN cash_flow_category;
