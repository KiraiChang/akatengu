-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    user_id  INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT    NOT NULL UNIQUE,
    password TEXT    NOT NULL,
    status   TEXT    NOT NULL DEFAULT 'INACTIVE',
    CONSTRAINT chk_user_status CHECK (status IN ('INACTIVE', 'ACTIVE'))
);

-- +goose Down
DROP TABLE IF EXISTS users;
