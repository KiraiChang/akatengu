CREATE TABLE event_store (
    event_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_uuid        TEXT    NOT NULL UNIQUE,
    occurred_at       TEXT    NOT NULL,
    aggregate_type    TEXT    NOT NULL,
    aggregate_id      TEXT    NOT NULL,
    aggregate_version INTEGER NOT NULL,
    event_type        TEXT    NOT NULL,
    payload           TEXT    NOT NULL,
    metadata          TEXT
);

CREATE TABLE aggregate_versions (
    aggregate_type  TEXT    NOT NULL,
    aggregate_id    TEXT    NOT NULL,
    current_version INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (aggregate_type, aggregate_id)
);

CREATE TABLE snapshots (
    snapshot_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    aggregate_type TEXT    NOT NULL,
    aggregate_id   TEXT    NOT NULL,
    at_version     INTEGER NOT NULL,
    state          TEXT    NOT NULL,
    created_at     TEXT    NOT NULL
);

CREATE TABLE projection_checkpoints (
    projection_name TEXT    PRIMARY KEY,
    last_event_id   INTEGER NOT NULL DEFAULT 0,
    updated_at      TEXT
);

CREATE TABLE accounts (
    account_id     TEXT    PRIMARY KEY,
    parent_id      TEXT,
    name           TEXT    NOT NULL,
    type           TEXT    NOT NULL,
    normal_balance TEXT    NOT NULL,
    currency       TEXT    NOT NULL,
    is_summary     INTEGER NOT NULL DEFAULT 0,
    is_active      INTEGER NOT NULL DEFAULT 1,
    note           TEXT,
    version        INTEGER NOT NULL
);

CREATE TABLE ledger_accounts (
    ledger_id    INTEGER PRIMARY KEY,
    account_id   TEXT    NOT NULL,
    institution  TEXT    NOT NULL,
    name         TEXT    NOT NULL,
    type         TEXT    NOT NULL,
    account_no   TEXT,
    currency     TEXT    NOT NULL,
    credit_limit REAL,
    billing_day  TEXT,
    due_day      TEXT,
    is_active    INTEGER NOT NULL DEFAULT 1,
    note         TEXT,
    version      INTEGER NOT NULL
);

CREATE TABLE sys_accounts (
    sys_code   TEXT PRIMARY KEY,
    account_id TEXT NOT NULL
);

CREATE TABLE transactions (
    txn_id         INTEGER PRIMARY KEY,
    txn_date       TEXT    NOT NULL,
    description    TEXT    NOT NULL,
    total_amount   REAL    NOT NULL,
    currency       TEXT    NOT NULL,
    status         TEXT    NOT NULL DEFAULT 'ACTIVE',
    installment_id INTEGER,
    receipt_no     TEXT,
    note           TEXT,
    version        INTEGER NOT NULL,
    ref_txn_id     INTEGER
);

CREATE TABLE journal_entries (
    entry_id   INTEGER PRIMARY KEY,
    txn_id     INTEGER NOT NULL,
    ledger_id  INTEGER,
    account_id TEXT    NOT NULL,
    debit      REAL    NOT NULL DEFAULT 0,
    credit     REAL    NOT NULL DEFAULT 0,
    note       TEXT
);

CREATE TABLE installments (
    installment_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    ledger_id         INTEGER NOT NULL,
    txn_id            INTEGER,
    description       TEXT    NOT NULL,
    total_amount      REAL    NOT NULL,
    total_periods     INTEGER NOT NULL,
    paid_periods      INTEGER NOT NULL DEFAULT 0,
    amount_per_period REAL    NOT NULL,
    start_date        TEXT    NOT NULL,
    end_date          TEXT,
    interest_rate     REAL    NOT NULL DEFAULT 0,
    interest_type     TEXT    NOT NULL,
    status            TEXT    NOT NULL DEFAULT 'ACTIVE',
    note              TEXT
);

CREATE TABLE installment_payments (
    payment_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    installment_id INTEGER NOT NULL,
    txn_id         INTEGER,
    period_no      INTEGER NOT NULL,
    amount         REAL    NOT NULL,
    interest       REAL    NOT NULL,
    due_date       TEXT    NOT NULL,
    paid_date      TEXT,
    status         TEXT    NOT NULL DEFAULT 'PENDING'
);

CREATE TABLE reconciliations (
    recon_id          INTEGER PRIMARY KEY,
    ledger_id         INTEGER NOT NULL,
    recon_date        TEXT    NOT NULL,
    statement_balance REAL    NOT NULL,
    book_balance      REAL    NOT NULL,
    difference        REAL,
    is_balanced       INTEGER NOT NULL DEFAULT 0,
    note              TEXT,
    version           INTEGER NOT NULL
);

CREATE TABLE reconciliation_adjustments (
    adjustment_id   INTEGER PRIMARY KEY,
    recon_id        INTEGER NOT NULL,
    txn_id          INTEGER NOT NULL,
    adjustment_type TEXT    NOT NULL,
    amount          REAL    NOT NULL,
    note            TEXT
);

CREATE TABLE period_closings (
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
    reopen_reason  TEXT
);

CREATE TABLE investments (
    investment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id    TEXT    NOT NULL,
    asset_type    TEXT    NOT NULL,
    currency      TEXT    NOT NULL,
    symbol        TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    cost_method   TEXT    NOT NULL DEFAULT 'FIFO',
    is_active     INTEGER NOT NULL DEFAULT 1,
    version       INTEGER NOT NULL
);

CREATE TABLE investment_lots (
    lot_id        INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id INTEGER NOT NULL,
    movement_id   INTEGER NOT NULL,
    acquired_date TEXT    NOT NULL,
    txn_id        INTEGER,
    quantity      REAL    NOT NULL,
    unit_cost     REAL    NOT NULL,
    total_cost    REAL    NOT NULL,
    remaining_qty REAL    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'OPEN'
);

CREATE TABLE investment_movements (
    movement_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id   INTEGER NOT NULL,
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
    withholding_tax REAL
);

CREATE TABLE investment_positions (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id  INTEGER NOT NULL UNIQUE,
    total_quantity REAL    NOT NULL,
    total_cost     REAL    NOT NULL,
    avg_cost       REAL    NOT NULL DEFAULT 0
);

CREATE TABLE investment_lot_disposals (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    lot_id              INTEGER,
    movement_id         INTEGER,
    quantity            REAL    NOT NULL,
    cost_basis          REAL    NOT NULL,
    sale_proceeds       REAL    NOT NULL,
    capital_gain        REAL    NOT NULL,
    holding_period_days INTEGER,
    disposal_date       TEXT    NOT NULL
);

CREATE TABLE exchange_rates (
    rate_id   INTEGER PRIMARY KEY AUTOINCREMENT,
    currency  TEXT    NOT NULL,
    rate_date TEXT    NOT NULL,
    rate_twd  REAL    NOT NULL,
    source    TEXT    NOT NULL DEFAULT 'MANUAL'
);

CREATE TABLE users (
    user_id  INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT    NOT NULL UNIQUE,
    password TEXT    NOT NULL,
    status   TEXT    NOT NULL DEFAULT 'INACTIVE'
);

CREATE VIEW v_account_balances AS
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
