-- +goose Up
CREATE TABLE IF NOT EXISTS account_closure (
    ancestor_id   TEXT    NOT NULL,
    descendant_id TEXT    NOT NULL,
    merchant_id   INTEGER NOT NULL DEFAULT 0,
    depth         INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id, merchant_id)
);

CREATE INDEX IF NOT EXISTS idx_account_closure_descendant ON account_closure(descendant_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_account_insert_closure
AFTER INSERT ON accounts
BEGIN
    INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, depth)
    VALUES (NEW.account_id, NEW.account_id, 0);
    INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, depth)
    SELECT ancestor_id, NEW.account_id, depth + 1
    FROM account_closure
    WHERE descendant_id = NEW.parent_id;
END;
-- +goose StatementEnd

INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, depth)
WITH RECURSIVE anc(account_id, ancestor_id, depth) AS (
    SELECT account_id, account_id, 0 FROM accounts
    UNION ALL
    SELECT anc.account_id, a.parent_id, anc.depth + 1
    FROM anc
    JOIN accounts a ON a.account_id = anc.ancestor_id
    WHERE a.parent_id IS NOT NULL
)
SELECT ancestor_id, account_id, depth FROM anc;

-- +goose Down
DROP TRIGGER IF EXISTS trg_account_insert_closure;
DROP INDEX IF EXISTS idx_account_closure_descendant;
DROP TABLE IF EXISTS account_closure;
