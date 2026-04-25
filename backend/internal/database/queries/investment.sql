-- name: GetInvestment :one
SELECT investment_id, account_id, asset_type, currency, symbol,
       name, cost_method, is_active, version
FROM investments
WHERE investment_id = ?;

-- name: GetInvestmentBySymbol :one
SELECT investment_id, account_id, asset_type, currency, symbol,
       name, cost_method, is_active, version
FROM investments
WHERE symbol = ? AND currency = ?;

-- name: GetInvestmentPosition :one
SELECT id, investment_id, total_quantity, total_cost, avg_cost
FROM investment_positions
WHERE investment_id = ?;

-- name: GetOpenLots :many
SELECT lot_id, investment_id, movement_id, acquired_date, txn_id,
       quantity, unit_cost, total_cost, remaining_qty, status
FROM investment_lots
WHERE investment_id = ? AND status != ?
ORDER BY acquired_date, lot_id;

-- name: GetInvestmentMovements :many
SELECT movement_id, investment_id, event_id, txn_id, movement_type,
       movement_date, quantity, unit_price, unit_price_twd, exchange_rate,
       fee, tax, realized_gain, cost_basis, gross_amount, net_amount,
       withholding_tax, split_ratio
FROM investment_movements
WHERE investment_id = ?
ORDER BY movement_date, movement_id;

-- name: CreateInvestment :exec
INSERT INTO investments
    (account_id, asset_type, currency, symbol, name, cost_method, is_active, version)
VALUES (?, ?, ?, ?, ?, ?, ?, 1);

-- name: UpdateInvestment :exec
UPDATE investments
SET account_id  = ?,
    asset_type  = ?,
    currency    = ?,
    symbol      = ?,
    name        = ?,
    cost_method = ?,
    is_active   = ?,
    version     = version + 1
WHERE investment_id = ? AND version = ?;

-- name: UpsertInvestmentPosition :exec
INSERT INTO investment_positions (investment_id, total_quantity, total_cost)
VALUES (?, ?, ?)
ON CONFLICT(investment_id)
DO UPDATE SET
    total_quantity = total_quantity + excluded.total_quantity,
    total_cost     = total_cost     + excluded.total_cost;

-- name: InsertInvestmentLot :execlastid
INSERT INTO investment_lots
    (investment_id, acquired_date, movement_id, quantity, unit_cost, total_cost, remaining_qty)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateInvestmentLotTxn :exec
UPDATE investment_lots
SET txn_id = ?
WHERE lot_id = ?;

-- name: InsertInvestmentLotDisposal :exec
INSERT INTO investment_lot_disposals
    (lot_id, movement_id, quantity, cost_basis, sale_proceeds, capital_gain,
     holding_period_days, disposal_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertInvestmentMovement :execlastid
INSERT INTO investment_movements
    (investment_id, movement_type, movement_date, event_id,
     quantity, unit_price, unit_price_twd, exchange_rate,
     fee, tax, realized_gain, cost_basis, split_ratio,
     gross_amount, net_amount, withholding_tax)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateInvestmentMovementTxn :exec
UPDATE investment_movements
SET txn_id = ?
WHERE movement_id = ?;

-- name: UpdateInvestmentPositionSplit :exec
UPDATE investment_positions
SET total_quantity = total_quantity * ?
WHERE investment_id = ?;

-- name: UpdateInvestmentLotSplit :exec
UPDATE investment_lots
SET quantity      = quantity      * ?,
    remaining_qty = remaining_qty * ?,
    unit_cost     = unit_cost     / ?
WHERE investment_id = ? AND status != 'CLOSED';

-- name: UpsertExchangeRate :exec
INSERT INTO exchange_rates (currency, rate_date, rate_twd, source)
VALUES (?, ?, ?, ?)
ON CONFLICT(currency, rate_date)
DO UPDATE SET rate_twd = excluded.rate_twd, source = excluded.source;