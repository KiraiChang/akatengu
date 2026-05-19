-- +goose Up

-- event_store：只加 updated_by（occurred_at 已是時間戳）
ALTER TABLE event_store ADD COLUMN updated_by TEXT;

-- accounts
ALTER TABLE accounts ADD COLUMN updated_by TEXT;
ALTER TABLE accounts ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- ledger_accounts
ALTER TABLE ledger_accounts ADD COLUMN updated_by TEXT;
ALTER TABLE ledger_accounts ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- transactions
ALTER TABLE transactions ADD COLUMN updated_by TEXT;
ALTER TABLE transactions ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- journal_entries
ALTER TABLE journal_entries ADD COLUMN updated_by TEXT;
ALTER TABLE journal_entries ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- investments
ALTER TABLE investments ADD COLUMN updated_by TEXT;
ALTER TABLE investments ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- investment_movements
ALTER TABLE investment_movements ADD COLUMN updated_by TEXT;
ALTER TABLE investment_movements ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- investment_lots
ALTER TABLE investment_lots ADD COLUMN updated_by TEXT;
ALTER TABLE investment_lots ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- investment_positions
ALTER TABLE investment_positions ADD COLUMN updated_by TEXT;
ALTER TABLE investment_positions ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- installments
ALTER TABLE installments ADD COLUMN updated_by TEXT;
ALTER TABLE installments ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- installment_payments
ALTER TABLE installment_payments ADD COLUMN updated_by TEXT;
ALTER TABLE installment_payments ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- period_closings
ALTER TABLE period_closings ADD COLUMN updated_by TEXT;
ALTER TABLE period_closings ADD COLUMN updated_at TEXT DEFAULT (datetime('now'));

-- +goose Down

ALTER TABLE event_store DROP COLUMN updated_by;

ALTER TABLE accounts DROP COLUMN updated_by;
ALTER TABLE accounts DROP COLUMN updated_at;

ALTER TABLE ledger_accounts DROP COLUMN updated_by;
ALTER TABLE ledger_accounts DROP COLUMN updated_at;

ALTER TABLE transactions DROP COLUMN updated_by;
ALTER TABLE transactions DROP COLUMN updated_at;

ALTER TABLE journal_entries DROP COLUMN updated_by;
ALTER TABLE journal_entries DROP COLUMN updated_at;

ALTER TABLE investments DROP COLUMN updated_by;
ALTER TABLE investments DROP COLUMN updated_at;

ALTER TABLE investment_movements DROP COLUMN updated_by;
ALTER TABLE investment_movements DROP COLUMN updated_at;

ALTER TABLE investment_lots DROP COLUMN updated_by;
ALTER TABLE investment_lots DROP COLUMN updated_at;

ALTER TABLE investment_positions DROP COLUMN updated_by;
ALTER TABLE investment_positions DROP COLUMN updated_at;

ALTER TABLE installments DROP COLUMN updated_by;
ALTER TABLE installments DROP COLUMN updated_at;

ALTER TABLE installment_payments DROP COLUMN updated_by;
ALTER TABLE installment_payments DROP COLUMN updated_at;

ALTER TABLE period_closings DROP COLUMN updated_by;
ALTER TABLE period_closings DROP COLUMN updated_at;
