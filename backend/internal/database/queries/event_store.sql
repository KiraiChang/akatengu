-- name: InsertEvent :execlastid
INSERT INTO event_store
    (aggregate_type, aggregate_id, aggregate_version, event_type, payload, metadata, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetEventsByAggregate :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE aggregate_type = ? AND aggregate_id = ?
ORDER BY aggregate_version;

-- name: GetEventsAfterVersion :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE aggregate_type    = ?
  AND aggregate_id      = ?
  AND aggregate_version > ?
ORDER BY aggregate_version;

-- name: GetEventsAfterEventID :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > ?
ORDER BY event_id;

-- name: GetEventsAfterEventIDByType :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > ? AND aggregate_type = ?
ORDER BY event_id;

-- name: GetEventsAfterEventIDLimited :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > ?
ORDER BY event_id
LIMIT ?;

-- name: GetEventsAfterEventIDByTypeLimited :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > ? AND aggregate_type = ?
ORDER BY event_id
LIMIT ?;

-- name: ReplayAllAggregates :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > ?
ORDER BY event_id;

-- name: ReplayByAggregate :many
SELECT event_id, event_uuid, occurred_at, aggregate_type, aggregate_id,
       aggregate_version, event_type, payload, metadata, updated_by
FROM event_store
WHERE event_id > ? AND aggregate_type = ?
ORDER BY event_id;

-- name: GetSnapshot :one
SELECT snapshot_id, aggregate_type, aggregate_id, at_version, state, created_at
FROM snapshots
WHERE aggregate_type = ? AND aggregate_id = ?;

-- name: UpsertSnapshot :exec
INSERT INTO snapshots (aggregate_type, aggregate_id, at_version, state)
VALUES (?, ?, ?, ?)
ON CONFLICT(aggregate_type, aggregate_id)
DO UPDATE SET at_version = excluded.at_version, state = excluded.state;

-- name: GetCheckpoint :one
SELECT last_event_id FROM projection_checkpoints
WHERE projection_name = ?;

-- name: UpdateCheckpoint :exec
UPDATE projection_checkpoints
SET last_event_id = ?, updated_at = datetime('now')
WHERE projection_name = ?;

-- name: UpsertCheckpoint :exec
INSERT INTO projection_checkpoints (projection_name, last_event_id, updated_at)
VALUES (@projection_name, @last_event_id, datetime('now'))
    ON CONFLICT(projection_name) DO UPDATE SET
    last_event_id = excluded.last_event_id,
    updated_at = excluded.updated_at;

-- name: GetAggregateVersion :one
SELECT current_version FROM aggregate_versions
WHERE aggregate_type = ? AND aggregate_id = ?;

-- name: InsertAggregateVersion :exec
INSERT INTO aggregate_versions (aggregate_type, aggregate_id, current_version)
VALUES (?, ?, ?);

-- name: UpdateVersionIfMatch :one
UPDATE aggregate_versions
SET current_version = current_version + 1
WHERE aggregate_type    = ?
  AND aggregate_id      = ?
  AND current_version   = ?
RETURNING aggregate_type, aggregate_id, current_version;
