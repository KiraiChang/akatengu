-- name: GetInvestment :one
SELECT investment_id, account_id, asset_type, currency, symbol,
       name, cost_method, ifrs_category, is_active, version,
       updated_by, updated_at, creation_event_uuid
FROM investments
WHERE investment_id = @investment_id AND merchant_id = @merchant_id;

-- name: GetInvestmentByCreationEventUuid :one
SELECT investment_id, account_id, asset_type, currency, symbol,
       name, cost_method, ifrs_category, is_active, version,
       updated_by, updated_at, creation_event_uuid
FROM investments
WHERE creation_event_uuid = @creation_event_uuid AND merchant_id = @merchant_id;

-- name: GetInvestmentPaged :many
WITH total AS (SELECT COUNT(*) AS cnt FROM investments WHERE merchant_id = @merchant_id)
SELECT i.investment_id, i.account_id, i.asset_type, i.currency, i.symbol,
       i.name, i.cost_method, i.ifrs_category, i.is_active, i.version,
       i.updated_by, i.updated_at, i.creation_event_uuid,
       total.cnt AS total
FROM investments AS i, total
WHERE i.merchant_id = @merchant_id
ORDER BY i.investment_id ASC
    LIMIT @limit
OFFSET @offset;

-- name: GetInvestmentBySymbol :one
SELECT investment_id, account_id, asset_type, currency, symbol,
       name, cost_method, ifrs_category, is_active, version,
       updated_by, updated_at, creation_event_uuid
FROM investments
WHERE symbol = @symbol AND currency = @currency AND merchant_id = @merchant_id;

-- name: GetPosition :one
SELECT id, investment_id, total_quantity, total_cost, avg_cost, market_price_twd,
       updated_by, updated_at
FROM investment_positions
WHERE investment_id = @investment_id AND merchant_id = @merchant_id;

-- name: GetNotStatusLots :many
SELECT lot_id, investment_id, movement_id, acquired_date, txn_id,
       quantity, unit_cost, total_cost, remaining_qty, status, unrealized_unit_twd,
       updated_by, updated_at
FROM investment_lots
WHERE investment_id = @investment_id AND status != @status AND merchant_id = @merchant_id
ORDER BY acquired_date, lot_id;

-- name: GetOpenLotsPaged :many
WITH total AS (SELECT COUNT(*) AS cnt FROM investment_lots AS p2 WHERE p2.investment_id = @investment_id AND p2.merchant_id = @merchant_id)
SELECT lot_id, p.investment_id, movement_id, acquired_date, txn_id,
       quantity, unit_cost, total_cost, remaining_qty, status, unrealized_unit_twd,
       p.updated_by, p.updated_at,
       total.cnt AS total
FROM investment_lots AS p, total
WHERE p.investment_id = @investment_id AND p.merchant_id = @merchant_id
ORDER BY acquired_date, lot_id ASC
    LIMIT @limit
OFFSET @offset;

-- name: GetOpenLotDisposalsPaged :many
WITH total AS (SELECT COUNT(*) AS cnt FROM investment_lot_disposals AS d2 WHERE d2.lot_id = @lot_id AND d2.merchant_id = @merchant_id)
SELECT d.lot_id, movement_id, quantity, cost_basis, sale_proceeds, capital_gain,
       holding_period_days, disposal_date, txn_id,
       total.cnt AS total
FROM investment_lot_disposals AS d, total
WHERE d.lot_id = @lot_id AND d.merchant_id = @merchant_id
ORDER BY disposal_date ASC
    LIMIT @limit
OFFSET @offset;

-- name: GetInvestmentMovements :many
SELECT movement_id, investment_id, event_id, txn_id, movement_type,
       movement_date, quantity, unit_price, unit_price_twd, exchange_rate,
       fee, tax, realized_gain, cost_basis, gross_amount, net_amount,
       withholding_tax, split_ratio,
       updated_by, updated_at
FROM investment_movements
WHERE investment_id = @investment_id AND merchant_id = @merchant_id
ORDER BY movement_date, movement_id;

-- name: GetInvestmentMovementsPaged :many
WITH total AS (SELECT COUNT(*) AS cnt FROM investment_movements AS m2 WHERE m2.investment_id = @investment_id AND m2.merchant_id = @merchant_id)
SELECT movement_id, m.investment_id, event_id, txn_id, movement_type,
       movement_date, quantity, unit_price, unit_price_twd, exchange_rate,
       fee, tax, realized_gain, cost_basis, gross_amount, net_amount,
       withholding_tax, split_ratio,
       m.updated_by, m.updated_at,
       total.cnt AS total
FROM investment_movements AS m, total
WHERE m.investment_id = @investment_id AND m.merchant_id = @merchant_id
ORDER BY movement_date, movement_id
    LIMIT @limit
OFFSET @offset;

-- name: CreateInvestment :exec
INSERT INTO investments
    (merchant_id, account_id, asset_type, currency, symbol, name, cost_method, ifrs_category, is_active, version, updated_by, creation_event_uuid)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?);

-- name: UpdateInvestment :exec
UPDATE investments
SET account_id  = ?,
    asset_type  = ?,
    currency    = ?,
    symbol      = ?,
    name        = ?,
    cost_method = ?,
    is_active   = ?,
    updated_by  = ?,
    updated_at  = datetime('now'),
    version     = version + 1
WHERE investment_id = ? AND merchant_id = ? AND version = ?;

-- name: UpsertInvestmentPosition :exec
INSERT INTO investment_positions (merchant_id, investment_id, total_quantity, total_cost, updated_by)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(investment_id)
DO UPDATE SET
    total_quantity = total_quantity + excluded.total_quantity,
    total_cost     = total_cost     + excluded.total_cost,
    updated_by     = excluded.updated_by,
    updated_at     = datetime('now');

-- name: InsertInvestmentLot :execlastid
INSERT INTO investment_lots
    (merchant_id, investment_id, acquired_date, movement_id, quantity, unit_cost, total_cost, remaining_qty, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateInvestmentLotTxn :exec
UPDATE investment_lots
SET txn_id     = ?,
    updated_by = ?,
    updated_at = datetime('now')
WHERE lot_id = ?;

-- name: UpdateLotUnrealizedUnit :exec
UPDATE investment_lots
SET unrealized_unit_twd = @unrealized_unit_twd,
    updated_by          = @updated_by,
    updated_at          = datetime('now')
WHERE lot_id = @lot_id;

-- name: InsertInvestmentLotDisposal :execlastid
INSERT INTO investment_lot_disposals
    (merchant_id, lot_id, movement_id, quantity, cost_basis, sale_proceeds, capital_gain,
     holding_period_days, disposal_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateInvestmentLotDisposalTxn :exec
UPDATE investment_lot_disposals
SET txn_id = @txn_id
WHERE id = @id;

-- name: InsertInvestmentMovement :execlastid
INSERT INTO investment_movements
    (merchant_id, investment_id, movement_type, movement_date, event_id,
     quantity, unit_price, unit_price_twd, exchange_rate,
     fee, tax, realized_gain, cost_basis, split_ratio,
     gross_amount, net_amount, withholding_tax, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateInvestmentMovementTxn :exec
UPDATE investment_movements
SET txn_id     = ?,
    updated_by = ?,
    updated_at = datetime('now')
WHERE movement_id = ?;

-- name: UpdateInvestmentPositionSplit :exec
UPDATE investment_positions
SET total_quantity = total_quantity * ?,
    updated_by     = ?,
    updated_at     = datetime('now')
WHERE investment_id = ?;

-- name: UpdateInvestmentPositionSold :exec
UPDATE investment_positions
SET total_quantity = total_quantity - @total_quantity,
    total_cost     = total_cost - @total_cost,
    updated_by     = @updated_by,
    updated_at     = datetime('now')
WHERE investment_id = @investment_id
    AND total_quantity - @total_quantity >= 0;

-- name: UpdateInvestmentLotSplit :exec
UPDATE investment_lots
SET quantity      = quantity      * ?,
    remaining_qty = remaining_qty * ?,
    unit_cost     = unit_cost     / ?,
    updated_by    = ?,
    updated_at    = datetime('now')
WHERE investment_id = ? AND status != 'CLOSED';

-- name: UpdateInvestmentPositionFairValue :exec
UPDATE investment_positions
SET market_price_twd = ?,
    updated_by       = ?,
    updated_at       = datetime('now')
WHERE investment_id = ?;

-- name: UpsertExchangeRate :exec
INSERT INTO exchange_rates (currency, rate_date, rate_twd, source)
VALUES (?, ?, ?, ?)
ON CONFLICT(currency, rate_date)
DO UPDATE SET rate_twd = excluded.rate_twd, source = excluded.source;
