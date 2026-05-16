-- +goose Up
CREATE TABLE IF NOT EXISTS merchants (
    merchant_id  INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT    NOT NULL,
    display_name TEXT    NOT NULL,
    currency     TEXT    NOT NULL DEFAULT 'TWD',
    status       TEXT    NOT NULL DEFAULT 'ACTIVE',
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT chk_merchant_status CHECK (status IN ('ACTIVE', 'INACTIVE'))
);

CREATE TABLE IF NOT EXISTS user_merchants (
    user_id     INTEGER NOT NULL REFERENCES users(user_id),
    merchant_id INTEGER NOT NULL REFERENCES merchants(merchant_id),
    role        TEXT    NOT NULL DEFAULT 'MEMBER',
    joined_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, merchant_id),
    CONSTRAINT chk_um_role CHECK (role IN ('OWNER', 'ADMIN', 'MEMBER'))
);

CREATE INDEX IF NOT EXISTS idx_user_merchants_user     ON user_merchants(user_id);
CREATE INDEX IF NOT EXISTS idx_user_merchants_merchant ON user_merchants(merchant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_merchants_merchant;
DROP INDEX IF EXISTS idx_user_merchants_user;
DROP TABLE IF EXISTS user_merchants;
DROP TABLE IF EXISTS merchants;
