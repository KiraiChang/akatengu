-- name: GetAggregateVersions :many
SELECT aggregate_type, aggregate_id, merchant_id, current_version
FROM aggregate_versions
WHERE merchant_id = @merchant_id
ORDER BY aggregate_type, aggregate_id;

-- name: GetEventStorePaged :many
WITH total AS (SELECT COUNT(*) AS cnt FROM event_store WHERE merchant_id = @merchant_id)
SELECT e.event_id, e.event_uuid, e.occurred_at, e.merchant_id,
       e.aggregate_type, e.aggregate_id, e.aggregate_version,
       e.event_type, e.payload, e.metadata, e.updated_by,
       total.cnt AS total
FROM event_store AS e, total
WHERE e.merchant_id = @merchant_id
ORDER BY e.event_id DESC
LIMIT @limit OFFSET @offset;

-- name: GetCheckpoints :many
SELECT projection_name, merchant_id, last_event_id, updated_at
FROM projection_checkpoints
WHERE merchant_id = @merchant_id
ORDER BY projection_name;

-- name: GetSnapshots :many
SELECT snapshot_id, merchant_id, aggregate_type, aggregate_id, at_version, state, created_at
FROM snapshots
WHERE merchant_id = @merchant_id
ORDER BY aggregate_type, aggregate_id;

-- name: GetExchangeRates :many
SELECT rate_id, currency, rate_date, rate_twd, source
FROM exchange_rates
ORDER BY rate_date DESC, currency;

-- name: GetExchangeRatesByCurrency :many
SELECT rate_id, currency, rate_date, rate_twd, source
FROM exchange_rates
WHERE currency = @currency
ORDER BY rate_date DESC;
