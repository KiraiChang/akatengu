-- +goose Up
ALTER TABLE transactions ADD COLUMN txn_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE journal_entries ADD COLUMN entry_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE journal_entries ADD COLUMN txn_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE journal_entries ADD COLUMN ledger_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE ledger_accounts ADD COLUMN ledger_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE investment_movements ADD COLUMN movement_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_movements ADD COLUMN investment_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_movements ADD COLUMN event_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE investment_lots ADD COLUMN lot_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_lots ADD COLUMN investment_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_lots ADD COLUMN movement_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE investment_lot_disposals ADD COLUMN disposal_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_lot_disposals ADD COLUMN lot_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_lot_disposals ADD COLUMN movement_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE investment_positions ADD COLUMN position_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE investment_positions ADD COLUMN investment_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE installments ADD COLUMN installment_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE installment_payments ADD COLUMN payment_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE installment_payments ADD COLUMN installment_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE prepaids ADD COLUMN prepaid_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE prepaid_amortizations ADD COLUMN amortization_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE prepaid_amortizations ADD COLUMN prepaid_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE fixed_assets ADD COLUMN asset_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE fixed_asset_depreciations ADD COLUMN depreciation_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE fixed_asset_depreciations ADD COLUMN asset_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE period_closings ADD COLUMN closing_uuid TEXT NOT NULL DEFAULT '';

ALTER TABLE snapshots ADD COLUMN snapshot_uuid TEXT NOT NULL DEFAULT '';

-- +goose Down
SELECT 1;
