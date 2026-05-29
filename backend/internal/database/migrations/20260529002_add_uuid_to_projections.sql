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

DROP INDEX IF EXISTS idx_install_ledger;
DROP VIEW IF EXISTS v_installments_due;
DROP VIEW IF EXISTS v_installments_active;

ALTER TABLE installments ADD COLUMN installment_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE installments ADD COLUMN ledger_uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE installments DROP COLUMN ledger_id;

CREATE VIEW IF NOT EXISTS v_installments_active AS
SELECT
    i.installment_id,
    la.institution,
    la.name                                                            AS card_name,
    i.description,
    i.total_amount,
    i.total_periods,
    i.amount_per_period,
    i.start_date,
    i.end_date
FROM installments i
         JOIN ledger_accounts la ON i.ledger_uuid = la.ledger_uuid
WHERE i.status = 'ACTIVE'
ORDER BY i.end_date;

CREATE VIEW IF NOT EXISTS v_installments_due AS
SELECT
    ip.due_date,
    la.institution,
    la.name     AS card_name,
    i.description,
    ip.period_no,
    i.total_periods,
    ip.amount
FROM installment_payments ip
         JOIN installments i     ON ip.installment_id = i.installment_id
         JOIN ledger_accounts la ON i.ledger_uuid = la.ledger_uuid
WHERE ip.status = 'PENDING'
ORDER BY ip.due_date;

CREATE INDEX IF NOT EXISTS idx_install_ledger ON installments(ledger_uuid);

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
