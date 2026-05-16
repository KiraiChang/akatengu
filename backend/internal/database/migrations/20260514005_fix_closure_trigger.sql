-- +goose Up

-- Fix trigger: include merchant_id so closures match the account's merchant.
DROP TRIGGER IF EXISTS trg_account_insert_closure;

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_account_insert_closure
AFTER INSERT ON accounts
BEGIN
    INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, merchant_id, depth)
    VALUES (NEW.account_id, NEW.account_id, NEW.merchant_id, 0);
    INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, merchant_id, depth)
    SELECT ancestor_id, NEW.account_id, NEW.merchant_id, depth + 1
    FROM account_closure
    WHERE descendant_id = NEW.parent_id AND merchant_id = NEW.merchant_id;
END;
-- +goose StatementEnd

-- Rebuild account_closure from current accounts to fix rows inserted by old trigger (merchant_id=0).
DELETE FROM account_closure;

INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, merchant_id, depth)
WITH RECURSIVE anc(account_id, ancestor_id, merchant_id, depth) AS (
    SELECT account_id, account_id, merchant_id, 0 FROM accounts
    UNION ALL
    SELECT anc.account_id, a.parent_id, anc.merchant_id, anc.depth + 1
    FROM anc
    JOIN accounts a ON a.account_id = anc.ancestor_id AND a.merchant_id = anc.merchant_id
    WHERE a.parent_id IS NOT NULL
)
SELECT ancestor_id, account_id, merchant_id, depth FROM anc;

-- Replace index to cover merchant_id for GetAncestorAccountIds queries.
DROP INDEX IF EXISTS idx_account_closure_descendant;
CREATE INDEX IF NOT EXISTS idx_account_closure_descendant ON account_closure(descendant_id, merchant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_account_closure_descendant;
CREATE INDEX IF NOT EXISTS idx_account_closure_descendant ON account_closure(descendant_id);

DROP TRIGGER IF EXISTS trg_account_insert_closure;

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
