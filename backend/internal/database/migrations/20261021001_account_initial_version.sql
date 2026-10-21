-- +goose Up
-- Accounts written by account.created / account.bulk_imported used to start at
-- version 0, which account.updated rejects. New accounts now start at 1, the
-- same as ledger accounts and investments (backend ISSUE-027).
UPDATE accounts SET version = 1 WHERE version = 0;

-- +goose Down
-- No-op: rows that were originally 0 cannot be told apart after the update.
