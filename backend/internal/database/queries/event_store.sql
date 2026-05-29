-- name: InsertEvent :execlastid
INSERT INTO event_store
    (merchant_id, aggregate_type, aggregate_id, aggregate_version, event_type, event_uuid, payload, metadata, updated_by)
VALUES (@merchant_id, @aggregate_type, @aggregate_id, @aggregate_version, @event_type, @event_uuid, @payload, @metadata, @updated_by);

-- name: GetEventsByAggregate :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE aggregate_type = @aggregate_type AND aggregate_id = @aggregate_id AND merchant_id = @merchant_id
ORDER BY aggregate_version;

-- name: GetEventsAfterVersion :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE aggregate_type     = @aggregate_type
  AND aggregate_id       = @aggregate_id
  AND merchant_id        = @merchant_id
  AND aggregate_version  > @aggregate_version
ORDER BY aggregate_version;

-- name: GetEventsAfterEventID :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > @event_id AND merchant_id = @merchant_id
ORDER BY event_id;

-- name: GetEventsAfterEventIDByType :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > @event_id AND aggregate_type = @aggregate_type AND merchant_id = @merchant_id
ORDER BY event_id;

-- name: GetEventsAfterEventIDLimited :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > @event_id AND merchant_id = @merchant_id
ORDER BY event_id
LIMIT @limit;

-- name: GetEventsAfterEventIDByTypeLimited :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > @event_id AND aggregate_type = @aggregate_type AND merchant_id = @merchant_id
ORDER BY event_id
LIMIT @limit;

-- name: ReplayAllAggregates :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > @event_id AND merchant_id = @merchant_id
ORDER BY event_id;

-- name: ReplayByAggregate :many
SELECT event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > @event_id AND aggregate_type = @aggregate_type AND merchant_id = @merchant_id
ORDER BY event_id;

-- name: GetSnapshot :one
SELECT snapshot_id, snapshot_uuid, merchant_id, aggregate_type, aggregate_id, at_version, state, created_at
FROM snapshots
WHERE aggregate_type = @aggregate_type AND aggregate_id = @aggregate_id AND merchant_id = @merchant_id;

-- name: UpsertSnapshot :exec
INSERT INTO snapshots (merchant_id, snapshot_uuid, aggregate_type, aggregate_id, at_version, state)
VALUES (@merchant_id, @snapshot_uuid, @aggregate_type, @aggregate_id, @at_version, @state)
ON CONFLICT(aggregate_type, aggregate_id, merchant_id)
DO UPDATE SET at_version = excluded.at_version, state = excluded.state;

-- name: GetCheckpoint :one
SELECT last_event_id FROM projection_checkpoints
WHERE projection_name = ? AND merchant_id = ?;

-- name: UpdateCheckpoint :exec
UPDATE projection_checkpoints
SET last_event_id = ?, updated_at = datetime('now')
WHERE projection_name = ? AND merchant_id = ?;

-- name: UpsertCheckpoint :exec
INSERT INTO projection_checkpoints (projection_name, merchant_id, last_event_id, updated_at)
VALUES (@projection_name, @merchant_id, @last_event_id, datetime('now'))
    ON CONFLICT(projection_name, merchant_id) DO UPDATE SET
    last_event_id = excluded.last_event_id,
    updated_at = excluded.updated_at;

-- name: GetAggregateVersion :one
SELECT current_version FROM aggregate_versions
WHERE aggregate_type = @aggregate_type AND aggregate_id = @aggregate_id AND merchant_id = @merchant_id;

-- name: InsertAggregateVersion :exec
INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
VALUES (@aggregate_type, @aggregate_id, @merchant_id, @current_version);

-- name: UpdateVersionIfMatch :one
UPDATE aggregate_versions
SET current_version = current_version + 1
WHERE aggregate_type  = @aggregate_type
  AND aggregate_id    = @aggregate_id
  AND merchant_id     = @merchant_id
  AND current_version = @current_version
RETURNING aggregate_type, aggregate_id, merchant_id, current_version;

-- name: UpsertAggregateVersion :exec
INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
VALUES (@aggregate_type, @aggregate_id, @merchant_id, @current_version)
ON CONFLICT(aggregate_type, aggregate_id, merchant_id) DO UPDATE SET
    current_version = MAX(current_version, excluded.current_version);
