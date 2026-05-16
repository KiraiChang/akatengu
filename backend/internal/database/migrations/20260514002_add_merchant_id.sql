-- +goose Up

-- ── 其他業務表：新增 merchant_id 欄位（ALTER TABLE ADD COLUMN）────────────────
ALTER TABLE ledger_accounts                 ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE transactions                    ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE journal_entries                 ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE installments                    ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE installment_payments            ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE investments                     ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE investment_lots                 ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE investment_movements            ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE investment_positions            ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE investment_lot_disposals        ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE reconciliations                 ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE reconciliation_adjustments      ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE period_closings                 ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE account_balance_snapshots       ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ledger_account_balance_snapshots ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0;

-- ── 常用查詢 Index ─────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_accounts_merchant          ON accounts(merchant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_ledger_accounts_merchant   ON ledger_accounts(merchant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_transactions_merchant      ON transactions(merchant_id, status, txn_date);
CREATE INDEX IF NOT EXISTS idx_investments_merchant       ON investments(merchant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_installments_merchant      ON installments(merchant_id, status);
CREATE INDEX IF NOT EXISTS idx_period_closings_merchant   ON period_closings(merchant_id, status);
CREATE INDEX IF NOT EXISTS idx_sys_accounts_merchant      ON sys_accounts(merchant_id);
CREATE INDEX IF NOT EXISTS idx_projection_checkpoint_mid  ON projection_checkpoints(merchant_id, projection_name);

-- +goose Down
SELECT 1;
