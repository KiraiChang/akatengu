-- +goose Up
DROP VIEW IF EXISTS v_installments_active;

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
         JOIN ledger_accounts la ON i.ledger_id = la.ledger_id
WHERE i.status = 'ACTIVE'
ORDER BY i.end_date;

ALTER TABLE investments RENAME COLUMN creation_event_uuid TO uuid;

-- +goose Down
DROP VIEW IF EXISTS v_installments_active;

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
         JOIN ledger_accounts la ON i.ledger_id = la.ledger_id
WHERE i.status = 'ACTIVE'
ORDER BY i.end_date;

ALTER TABLE investments RENAME COLUMN uuid TO creation_event_uuid;