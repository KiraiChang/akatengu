-- +goose Up
ALTER TABLE investments ADD COLUMN creation_event_uuid TEXT NOT NULL DEFAULT '';

-- +goose Down
SELECT 1;
