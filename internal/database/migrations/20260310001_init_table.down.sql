DROP VIEW IF EXISTS v_open_lots;
DROP VIEW IF EXISTS v_realized_gains;
DROP VIEW IF EXISTS v_investment_summary;
DROP VIEW IF EXISTS v_unresolved_adjustments;
DROP VIEW IF EXISTS v_reconciliation_detail;
DROP VIEW IF EXISTS v_monthly_summary;
DROP VIEW IF EXISTS v_installments_due;
DROP VIEW IF EXISTS v_installments_active;
DROP VIEW IF EXISTS v_account_summary;
DROP VIEW IF EXISTS v_account_balances;


DROP TRIGGER IF EXISTS check_oversold
DROP TRIGGER IF EXISTS apply_disposal

DROP INDEX IF EXISTS idx_rates_currency_date;

DROP TABLE IF EXISTS exchange_rates;

DROP INDEX IF EXISTS idx_disposals_lot;
DROP TABLE IF EXISTS investment_lot_disposals

DROP TABLE IF EXISTS investment_positions

DROP INDEX IF EXISTS idx_movements_investment;
DROP INDEX IF EXISTS idx_movements_txn;

DROP TABLE IF EXISTS investment_movements;


DROP INDEX IF EXISTS idx_lots_investment_status;
DROP INDEX IF EXISTS idx_lots_lookup;

DROP TABLE IF EXISTS investment_lots;

DROP INDEX IF EXISTS idx_investments_symbol;

DROP TABLE IF EXISTS investments;


DROP INDEX IF EXISTS idx_period_closings_status;
DROP INDEX IF EXISTS idx_period_closings_end;

DROP TABLE IF EXISTS period_closings;

DROP INDEX IF EXISTS idx_ra_recon ON reconciliation_adjustments(recon_id);
DROP INDEX IF EXISTS idx_ra_txn   ON reconciliation_adjustments(txn_id);
DROP INDEX IF EXISTS idx_ra_type  ON reconciliation_adjustments(adjustment_type);

DROP TABLE IF EXISTS reconciliation_adjustments;

DROP INDEX IF EXISTS idx_recon_ledger   ON reconciliations(ledger_id);
DROP INDEX IF EXISTS idx_recon_balanced ON reconciliations(is_balanced);

DROP TABLE IF EXISTS reconciliations;

DROP INDEX IF EXISTS idx_install_status ON installments(status);
DROP INDEX IF EXISTS idx_install_ledger ON installments(ledger_id);

DROP TABLE IF EXISTS installments;

DROP INDEX IF EXISTS idx_je_txn     ON journal_entries(txn_id);
DROP INDEX IF EXISTS idx_je_account ON journal_entries(account_id);
DROP INDEX IF EXISTS idx_je_ledger  ON journal_entries(ledger_id);

DROP TABLE IF EXISTS journal_entries;

DROP INDEX IF EXISTS idx_txn_date   ON transactions(txn_date);
DROP INDEX IF EXISTS idx_txn_status ON transactions(status);

DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS ledger_accounts;

DROP TABLE IF EXISTS accounts;

DROP TABLE IF EXISTS projection_checkpoints;

DROP TABLE IF EXISTS snapshots;

DROP TABLE IF EXISTS aggregate_versions;

DROP INDEX IF EXISTS idx_es_aggregate ON event_store(aggregate_type, aggregate_id);
DROP INDEX IF EXISTS idx_es_type      ON event_store(event_type);
DROP INDEX IF EXISTS idx_es_occurred  ON event_store(occurred_at);

DROP TABLE IF EXISTS event_store;