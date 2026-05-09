-- +goose Up
CREATE TABLE IF NOT EXISTS event_store (
    event_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_uuid        TEXT    NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(16)))),
    occurred_at       TEXT    NOT NULL DEFAULT (datetime('now')),
    aggregate_type    TEXT    NOT NULL,
    aggregate_id      TEXT    NOT NULL,
    aggregate_version INTEGER NOT NULL,
    event_type        TEXT    NOT NULL,
    payload           TEXT    NOT NULL,
    metadata          TEXT,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION')),
    UNIQUE(aggregate_type, aggregate_id, aggregate_version)
);

CREATE INDEX IF NOT EXISTS idx_es_aggregate ON event_store(aggregate_type, aggregate_id);
CREATE INDEX IF NOT EXISTS idx_es_type      ON event_store(event_type);
CREATE INDEX IF NOT EXISTS idx_es_occurred  ON event_store(occurred_at);

CREATE TABLE IF NOT EXISTS aggregate_versions (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    current_version INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION')),
    PRIMARY KEY (aggregate_type, aggregate_id)
);

CREATE TABLE IF NOT EXISTS snapshots (
    snapshot_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    aggregate_type TEXT    NOT NULL,
    aggregate_id   TEXT    NOT NULL,
    at_version     INTEGER NOT NULL,
    state          TEXT    NOT NULL,
    created_at     TEXT    DEFAULT (datetime('now')),
    UNIQUE(aggregate_type, aggregate_id)
);

CREATE TABLE IF NOT EXISTS projection_checkpoints (
    projection_name TEXT    PRIMARY KEY,
    last_event_id   INTEGER NOT NULL DEFAULT 0,
    updated_at      TEXT    DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS accounts (
    account_id     TEXT    PRIMARY KEY,
    parent_id      TEXT    REFERENCES accounts(account_id),
    name           TEXT    NOT NULL,
    type           TEXT    NOT NULL,
    normal_balance TEXT    NOT NULL,
    currency       TEXT    DEFAULT 'TWD',
    is_summary     INTEGER DEFAULT 0,
    is_active      INTEGER DEFAULT 1,
    note           TEXT,
    version        INTEGER NOT NULL,
    CONSTRAINT chk_normal_balance CHECK (normal_balance IN ('DEBIT', 'CREDIT')),
    CONSTRAINT chk_type CHECK (type IN ('ASSET', 'LIABILITY', 'INCOME', 'EXPENSE', 'EQUITY'))
);

CREATE TABLE IF NOT EXISTS ledger_accounts (
    ledger_id    INTEGER PRIMARY KEY,
    account_id   TEXT    NOT NULL REFERENCES accounts(account_id),
    institution  TEXT    NOT NULL,
    name         TEXT    NOT NULL,
    type         TEXT    NOT NULL,
    account_no   TEXT,
    currency     TEXT    DEFAULT 'TWD',
    credit_limit REAL,
    billing_day  TEXT,
    due_day      TEXT,
    is_active    INTEGER DEFAULT 1,
    note         TEXT,
    version      INTEGER NOT NULL,
    CONSTRAINT chk_type CHECK (type IN ('BANK_ACCOUNT', 'LOAN', 'CREDIT_CARD'))
);

CREATE TABLE IF NOT EXISTS sys_accounts (
    sys_code    TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    account_id  TEXT NOT NULL REFERENCES accounts(account_id)
);

CREATE TABLE IF NOT EXISTS transactions (
    txn_id         INTEGER PRIMARY KEY,
    txn_date       TEXT    NOT NULL,
    description    TEXT    NOT NULL,
    total_amount   REAL    NOT NULL,
    currency       TEXT    DEFAULT 'TWD',
    status         TEXT    DEFAULT 'ACTIVE',
    installment_id INTEGER,
    receipt_no     TEXT,
    note           TEXT,
    version        INTEGER NOT NULL,
    ref_txn_id     INTEGER,
    CONSTRAINT chk_status CHECK (status IN ('ACTIVE', 'CORRECTED', 'VOIDED', 'VOID_REF'))
);

CREATE INDEX IF NOT EXISTS idx_txn_date   ON transactions(txn_date);
CREATE INDEX IF NOT EXISTS idx_txn_status ON transactions(status);

CREATE TABLE IF NOT EXISTS journal_entries (
    entry_id   INTEGER PRIMARY KEY,
    txn_id     INTEGER NOT NULL REFERENCES transactions(txn_id),
    ledger_id  INTEGER REFERENCES ledger_accounts(ledger_id),
    account_id TEXT    NOT NULL REFERENCES accounts(account_id),
    debit      REAL    DEFAULT 0,
    credit     REAL    DEFAULT 0,
    note       TEXT
);

CREATE INDEX IF NOT EXISTS idx_je_txn     ON journal_entries(txn_id);
CREATE INDEX IF NOT EXISTS idx_je_account ON journal_entries(account_id);
CREATE INDEX IF NOT EXISTS idx_je_ledger  ON journal_entries(ledger_id);

CREATE TABLE IF NOT EXISTS installments (
    installment_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    ledger_id         INTEGER NOT NULL REFERENCES ledger_accounts(ledger_id),
    txn_id            INTEGER,
    description       TEXT    NOT NULL,
    total_amount      REAL    NOT NULL,
    total_periods     INTEGER NOT NULL,
    amount_per_period REAL    NOT NULL,
    start_date        TEXT    NOT NULL,
    end_date          TEXT,
    interest_rate     REAL    DEFAULT 0,
    interest_type     TEXT    NOT NULL,
    status            TEXT    DEFAULT 'ACTIVE',
    note              TEXT,
    CONSTRAINT chk_status CHECK (status IN ('ACTIVE', 'COMPLETED', 'CANCELED')),
    CONSTRAINT chk_interest_type CHECK (interest_type IN ('FREE', 'FIXED_RATE'))
);

CREATE INDEX IF NOT EXISTS idx_install_status ON installments(status);
CREATE INDEX IF NOT EXISTS idx_install_ledger ON installments(ledger_id);

CREATE TABLE IF NOT EXISTS installment_payments (
    payment_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    installment_id INTEGER NOT NULL REFERENCES installments(installment_id),
    txn_id         INTEGER,
    period_no      INTEGER NOT NULL,
    amount         REAL    NOT NULL,
    interest       REAL    NOT NULL,
    due_date       TEXT    NOT NULL,
    paid_date      TEXT,
    status         TEXT    DEFAULT 'PENDING',
    CONSTRAINT chk_status CHECK (status IN ('PENDING', 'PAID'))
);

CREATE INDEX IF NOT EXISTS idx_ip_install     ON installment_payments(installment_id);
CREATE INDEX IF NOT EXISTS idx_ip_in_period_no ON installment_payments(installment_id, period_no);
CREATE INDEX IF NOT EXISTS idx_ip_status      ON installment_payments(status);

CREATE TABLE IF NOT EXISTS reconciliations (
    recon_id          INTEGER PRIMARY KEY,
    ledger_id         INTEGER NOT NULL REFERENCES ledger_accounts(ledger_id),
    recon_date        TEXT    NOT NULL,
    statement_balance REAL    NOT NULL,
    book_balance      REAL    NOT NULL,
    difference        REAL,
    is_balanced       INTEGER DEFAULT 0,
    note              TEXT,
    version           INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_recon_ledger   ON reconciliations(ledger_id);
CREATE INDEX IF NOT EXISTS idx_recon_balanced ON reconciliations(is_balanced);

CREATE TABLE IF NOT EXISTS reconciliation_adjustments (
    adjustment_id   INTEGER PRIMARY KEY,
    recon_id        INTEGER NOT NULL REFERENCES reconciliations(recon_id),
    txn_id          INTEGER NOT NULL REFERENCES transactions(txn_id),
    adjustment_type TEXT    NOT NULL,
    amount          REAL    NOT NULL,
    note            TEXT,
    CONSTRAINT chk_adjustment_type CHECK (adjustment_type IN ('MISSING', 'DUPLICATE', 'CORRECTION', 'UNRESOLVED'))
);

CREATE INDEX IF NOT EXISTS idx_ra_recon ON reconciliation_adjustments(recon_id);
CREATE INDEX IF NOT EXISTS idx_ra_txn   ON reconciliation_adjustments(txn_id);
CREATE INDEX IF NOT EXISTS idx_ra_type  ON reconciliation_adjustments(adjustment_type);

CREATE TABLE IF NOT EXISTS period_closings (
    closing_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    period_type    TEXT    NOT NULL,
    period_start   TEXT    NOT NULL,
    period_end     TEXT    NOT NULL,
    status         TEXT    NOT NULL DEFAULT 'OPEN',
    opening_txn_id INTEGER,
    closing_txn_id INTEGER,
    snapshot       TEXT,
    closed_at      TEXT,
    note           TEXT,
    reopen_at      TEXT,
    reopen_reason  TEXT,
    CONSTRAINT chk_period_type CHECK (period_type IN ('MONTHLY', 'ANNUAL')),
    CONSTRAINT chk_status CHECK (status IN ('OPEN', 'CLOSED', 'REOPENED')),
    CONSTRAINT chk_period_dates CHECK (period_start <= period_end),
    CONSTRAINT chk_closed_at CHECK (status != 'CLOSED' OR closed_at IS NOT NULL),
    UNIQUE(period_type, period_start)
);

CREATE INDEX IF NOT EXISTS idx_period_closings_status ON period_closings(status);
CREATE INDEX IF NOT EXISTS idx_period_closings_end    ON period_closings(period_end);

CREATE TABLE IF NOT EXISTS investments (
    investment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id    TEXT    NOT NULL REFERENCES accounts(account_id),
    asset_type    TEXT    NOT NULL,
    currency      TEXT    NOT NULL,
    symbol        TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    cost_method   TEXT    NOT NULL DEFAULT 'FIFO',
    is_active     INTEGER NOT NULL DEFAULT 1,
    version       INTEGER NOT NULL,
    ifrs_category TEXT NOT NULL DEFAULT 'FVTPL',
    CONSTRAINT chk_ifrs_category CHECK(ifrs_category IN ('FVTPL','FVOCI','AC')),
    CONSTRAINT chk_asset_type  CHECK (asset_type IN ('STOCK', 'FUND', 'GOLD', 'FX')),
    CONSTRAINT chk_cost_method CHECK (cost_method IN ('AVG', 'FIFO')),
    CONSTRAINT chk_currency    CHECK (length(currency) = 3)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_investments_symbol ON investments(symbol, currency);

CREATE TABLE IF NOT EXISTS investment_lots (
    lot_id        INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id INTEGER NOT NULL REFERENCES investments(investment_id),
    movement_id   INTEGER NOT NULL,
    acquired_date TEXT    NOT NULL,
    txn_id        INTEGER,
    quantity      REAL    NOT NULL,
    unit_cost     REAL    NOT NULL,
    total_cost    REAL    NOT NULL,
    remaining_qty REAL    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'OPEN',
    unrealized_unit_twd REAL NOT NULL DEFAULT 0,
    CONSTRAINT chk_lot_status     CHECK (status IN ('OPEN', 'PARTIAL', 'CLOSED')),
    CONSTRAINT chk_lot_qty        CHECK (quantity > 0),
    CONSTRAINT chk_remaining_qty  CHECK (remaining_qty >= 0 AND remaining_qty <= quantity)
);

CREATE INDEX IF NOT EXISTS idx_lots_investment_status ON investment_lots(investment_id, status, acquired_date);

CREATE TABLE IF NOT EXISTS investment_movements (
    movement_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id   INTEGER NOT NULL REFERENCES investments(investment_id),
    event_id        INTEGER NOT NULL,
    txn_id          INTEGER,
    movement_type   TEXT    NOT NULL,
    movement_date   TEXT    NOT NULL,
    quantity        REAL    NOT NULL DEFAULT 0,
    unit_price      REAL    NOT NULL DEFAULT 0,
    unit_price_twd  REAL    NOT NULL DEFAULT 0,
    exchange_rate   REAL    NOT NULL DEFAULT 1,
    fee             REAL    NOT NULL DEFAULT 0,
    tax             REAL    NOT NULL DEFAULT 0,
    realized_gain   REAL,
    cost_basis      REAL,
    split_ratio     REAL,
    gross_amount    REAL,
    net_amount      REAL,
    withholding_tax REAL,
    UNIQUE(investment_id, event_id),
    CONSTRAINT chk_movement_type CHECK (movement_type IN ('BUY', 'SELL', 'DIVIDEND', 'SPLIT', 'CONVERT', 'MARK'))
);

CREATE INDEX IF NOT EXISTS idx_movements_investment ON investment_movements(investment_id, movement_date);
CREATE INDEX IF NOT EXISTS idx_movements_txn        ON investment_movements(txn_id);

CREATE TABLE IF NOT EXISTS investment_positions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id INTEGER NOT NULL REFERENCES investments(investment_id),
    total_quantity REAL   NOT NULL,
    total_cost    REAL    NOT NULL,
    market_price_twd REAL NOT NULL DEFAULT 0,
    avg_cost      REAL    GENERATED ALWAYS AS (
        CASE WHEN total_quantity = 0 THEN 0
             ELSE total_cost / total_quantity
        END
    ),
    UNIQUE(investment_id)
);

CREATE TABLE IF NOT EXISTS investment_lot_disposals (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    lot_id              INTEGER REFERENCES investment_lots(lot_id),
    txn_id              INTEGER,
    movement_id         INTEGER REFERENCES investment_movements(movement_id),
    quantity            REAL    NOT NULL,
    cost_basis          REAL    NOT NULL,
    sale_proceeds       REAL    NOT NULL,
    capital_gain        REAL    NOT NULL,
    holding_period_days INTEGER,
    disposal_date       TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_disposals_lot ON investment_lot_disposals(lot_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS check_oversold_insert
BEFORE INSERT ON investment_lot_disposals
FOR EACH ROW
WHEN (
    COALESCE((
        SELECT SUM(quantity)
        FROM investment_lot_disposals
        WHERE lot_id = NEW.lot_id
    ), 0) + NEW.quantity
) > (
    COALESCE((
        SELECT quantity
        FROM investment_lots
        WHERE lot_id = NEW.lot_id
    ), 0)
)
BEGIN
SELECT RAISE(ABORT, '賣出數量超過持有數量');
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS apply_disposal
AFTER INSERT ON investment_lot_disposals
FOR EACH ROW
BEGIN
    UPDATE investment_lots
    SET remaining_qty = remaining_qty - NEW.quantity
    WHERE lot_id = NEW.lot_id;
END;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS exchange_rates (
    rate_id   INTEGER PRIMARY KEY AUTOINCREMENT,
    currency  TEXT    NOT NULL,
    rate_date TEXT    NOT NULL,
    rate_twd  REAL    NOT NULL,
    source    TEXT    NOT NULL DEFAULT 'MANUAL',
    CONSTRAINT chk_rate_positive CHECK (rate_twd > 0),
    CONSTRAINT chk_rate_source   CHECK (source IN ('MANUAL', 'IMPORT')),
    UNIQUE(currency, rate_date)
);

CREATE INDEX IF NOT EXISTS idx_rates_currency_date ON exchange_rates(currency, rate_date DESC);

CREATE VIEW IF NOT EXISTS v_account_balances AS
SELECT
    la.ledger_id,
    a.normal_balance,
    COALESCE(SUM(je.debit), 0)     AS debit_total,
    COALESCE(SUM(je.credit), 0)     AS credit_total
FROM ledger_accounts la
    JOIN accounts a          ON la.account_id = a.account_id
    LEFT JOIN journal_entries je ON la.ledger_id = je.ledger_id
                                        AND la.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
WHERE la.is_active = 1
GROUP BY la.ledger_id;

CREATE VIEW IF NOT EXISTS v_account_summary AS
SELECT
    a.account_id,
    a.name,
    a.type,
    COALESCE(SUM(je.debit) - SUM(je.credit), 0)     AS balance
FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
WHERE a.is_active = 1
GROUP BY a.account_id;

CREATE VIEW IF NOT EXISTS v_installments_active AS
SELECT
    i.installment_id,
    la.institution,
    la.name                                                            AS card_name,
    i.description,
    i.total_amount,
    i.total_periods,
    i.paid_periods,
    (i.total_periods - i.paid_periods)                                 AS remaining_periods,
    i.amount_per_period,
    ROUND(i.amount_per_period * (i.total_periods - i.paid_periods), 0) AS remaining_amount,
    i.start_date,
    i.end_date
FROM installments i
    JOIN ledger_accounts la ON i.ledger_id = la.ledger_id
WHERE i.status = 'ACTIVE'
ORDER BY i.end_date;

CREATE VIEW IF NOT EXISTS v_installments_due AS
SELECT
    ip.due_date,
    la.institution,
    la.name     AS card_name,
    i.description,
    ip.period_no,
    i.total_periods,
    ip.amount
FROM installment_payments ip
    JOIN installments i     ON ip.installment_id = i.installment_id
    JOIN ledger_accounts la ON i.ledger_id = la.ledger_id
WHERE ip.status = 'PENDING'
ORDER BY ip.due_date;

CREATE VIEW IF NOT EXISTS v_monthly_summary AS
SELECT
    strftime('%Y-%m', t.txn_date) AS month,
    a.type                         AS account_type,
    a.name                         AS account_name,
    SUM(je.debit)                  AS total_debit,
    SUM(je.credit)                 AS total_credit
FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id
    JOIN accounts a     ON je.account_id = a.account_id
WHERE t.status = 'ACTIVE'
GROUP BY month, a.account_id
ORDER BY month DESC, a.type;

CREATE VIEW IF NOT EXISTS v_reconciliation_detail AS
SELECT
    r.recon_id,
    r.recon_date,
    la.institution,
    la.name             AS account_name,
    r.statement_balance,
    r.book_balance,
    r.difference,
    r.is_balanced,
    ra.adjustment_id,
    ra.adjustment_type,
    ra.amount           AS adjustment_amount,
    t.description       AS txn_description,
    t.txn_date,
    ra.note             AS adjustment_note
FROM reconciliations r
    JOIN ledger_accounts la                 ON r.ledger_id  = la.ledger_id
    LEFT JOIN reconciliation_adjustments ra ON r.recon_id   = ra.recon_id
    LEFT JOIN transactions t                ON ra.txn_id    = t.txn_id
ORDER BY r.recon_date DESC, ra.adjustment_id;

CREATE VIEW IF NOT EXISTS v_unresolved_adjustments AS
SELECT
    r.recon_date,
    la.institution,
    la.name    AS account_name,
    ra.amount,
    ra.note,
    r.recon_id,
    ra.adjustment_id
FROM reconciliation_adjustments ra
    JOIN reconciliations r  ON ra.recon_id = r.recon_id
    JOIN ledger_accounts la ON r.ledger_id = la.ledger_id
WHERE ra.adjustment_type = 'UNRESOLVED'
ORDER BY r.recon_date;

CREATE VIEW IF NOT EXISTS v_investment_summary AS
SELECT
    i.investment_id,
    i.symbol,
    i.name,
    i.asset_type,
    i.currency,
    i.cost_method,
    COALESCE(SUM(l.remaining_qty), 0)                         AS total_qty,
    CASE
        WHEN COALESCE(SUM(l.remaining_qty), 0) = 0 THEN 0
        ELSE COALESCE(SUM(l.remaining_qty * l.unit_cost), 0) / SUM(l.remaining_qty)
    END                                                        AS avg_cost_twd,
    COALESCE(SUM(l.remaining_qty * l.unit_cost), 0)           AS total_cost_twd
FROM investments i
    LEFT JOIN investment_lots l ON i.investment_id = l.investment_id AND l.status != 'CLOSED'
WHERE i.is_active = 1
GROUP BY i.investment_id;

CREATE VIEW IF NOT EXISTS v_realized_gains AS
SELECT
    i.investment_id,
    i.symbol,
    i.name,
    i.asset_type,
    strftime('%Y', m.movement_date)    AS year,
    strftime('%Y-%m', m.movement_date) AS month,
    SUM(m.realized_gain)               AS total_realized_gain,
    SUM(m.fee + m.tax)                 AS total_cost,
    COUNT(*)                           AS sell_count
FROM investment_movements m
    JOIN investments i ON m.investment_id = i.investment_id
WHERE m.movement_type IN ('SELL')
GROUP BY i.investment_id, month;

CREATE VIEW IF NOT EXISTS v_open_lots AS
SELECT
    l.lot_id,
    i.symbol,
    i.name,
    i.currency,
    l.acquired_date,
    l.quantity,
    l.remaining_qty,
    l.unit_cost,
    l.remaining_qty * l.unit_cost AS remaining_cost_twd
FROM investment_lots l
    JOIN investments i ON l.investment_id = i.investment_id
WHERE l.status != 'CLOSED'
ORDER BY l.investment_id, l.acquired_date;

-- +goose Down
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

DROP TRIGGER IF EXISTS check_oversold;
DROP TRIGGER IF EXISTS apply_disposal;

DROP INDEX IF EXISTS idx_rates_currency_date;
DROP TABLE IF EXISTS exchange_rates;

DROP INDEX IF EXISTS idx_disposals_lot;
DROP TABLE IF EXISTS investment_lot_disposals;

DROP TABLE IF EXISTS investment_positions;

DROP INDEX IF EXISTS idx_movements_investment;
DROP INDEX IF EXISTS idx_movements_txn;
DROP TABLE IF EXISTS investment_movements;

DROP INDEX IF EXISTS idx_lots_investment_status;
DROP TABLE IF EXISTS investment_lots;

DROP INDEX IF EXISTS idx_investments_symbol;
DROP TABLE IF EXISTS investments;

DROP INDEX IF EXISTS idx_period_closings_status;
DROP INDEX IF EXISTS idx_period_closings_end;
DROP TABLE IF EXISTS period_closings;

DROP INDEX IF EXISTS idx_ra_recon;
DROP INDEX IF EXISTS idx_ra_txn;
DROP INDEX IF EXISTS idx_ra_type;
DROP TABLE IF EXISTS reconciliation_adjustments;

DROP INDEX IF EXISTS idx_recon_ledger;
DROP INDEX IF EXISTS idx_recon_balanced;
DROP TABLE IF EXISTS reconciliations;

DROP INDEX IF EXISTS idx_ip_install;
DROP INDEX IF EXISTS idx_ip_in_period_no;
DROP INDEX IF EXISTS idx_ip_status;
DROP TABLE IF EXISTS installment_payments;

DROP INDEX IF EXISTS idx_install_status;
DROP INDEX IF EXISTS idx_install_ledger;
DROP TABLE IF EXISTS installments;

DROP INDEX IF EXISTS idx_je_txn;
DROP INDEX IF EXISTS idx_je_account;
DROP INDEX IF EXISTS idx_je_ledger;
DROP TABLE IF EXISTS journal_entries;

DROP INDEX IF EXISTS idx_txn_date;
DROP INDEX IF EXISTS idx_txn_status;
DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS sys_accounts;
DROP TABLE IF EXISTS ledger_accounts;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS projection_checkpoints;
DROP TABLE IF EXISTS snapshots;
DROP TABLE IF EXISTS aggregate_versions;

DROP INDEX IF EXISTS idx_es_aggregate;
DROP INDEX IF EXISTS idx_es_type;
DROP INDEX IF EXISTS idx_es_occurred;
DROP TABLE IF EXISTS event_store;