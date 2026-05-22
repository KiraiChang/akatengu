CREATE TABLE event_store (
    event_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_uuid        TEXT    NOT NULL UNIQUE,
    occurred_at       TEXT    NOT NULL,
    aggregate_type    TEXT    NOT NULL,
    aggregate_id      TEXT    NOT NULL,
    aggregate_version INTEGER NOT NULL,
    event_type        TEXT    NOT NULL,
    payload           TEXT    NOT NULL,
    metadata          TEXT,
    updated_by        TEXT
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
    projection_name TEXT    NOT NULL,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    last_event_id   INTEGER NOT NULL DEFAULT 0,
    updated_at      TEXT,
    PRIMARY KEY (projection_name, merchant_id)
);

CREATE TABLE accounts (
    account_id         TEXT    NOT NULL,
    merchant_id        INTEGER NOT NULL DEFAULT 0,
    parent_id          TEXT,
    name               TEXT    NOT NULL,
    type               TEXT    NOT NULL,
    normal_balance     TEXT    NOT NULL,
    currency           TEXT    NOT NULL,
    is_summary         INTEGER NOT NULL DEFAULT 0,
    is_active          INTEGER NOT NULL DEFAULT 1,
    note               TEXT,
    version            INTEGER NOT NULL,
    cash_flow_category TEXT    CHECK (cash_flow_category IN ('CASH', 'OPERATING', 'INVESTING', 'FINANCING')),
    updated_by         TEXT,
    updated_at         TEXT,
    PRIMARY KEY (account_id, merchant_id)
);

CREATE TABLE ledger_accounts (
    ledger_id    INTEGER PRIMARY KEY,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
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
    version      INTEGER NOT NULL,
    updated_by   TEXT,
    updated_at   TEXT
);

CREATE TABLE sys_accounts (
    sys_code    TEXT    NOT NULL,
    merchant_id INTEGER NOT NULL DEFAULT 0,
    description TEXT    NOT NULL,
    account_id  TEXT    NOT NULL,
    PRIMARY KEY (sys_code, merchant_id)
);

CREATE TABLE transactions (
    txn_id         INTEGER PRIMARY KEY,
    merchant_id    INTEGER NOT NULL DEFAULT 0,
    txn_date       TEXT    NOT NULL,
    description    TEXT    NOT NULL,
    total_amount   REAL    NOT NULL,
    currency       TEXT    NOT NULL,
    status         TEXT    NOT NULL DEFAULT 'ACTIVE',
    installment_id INTEGER,
    receipt_no     TEXT,
    note           TEXT,
    version        INTEGER NOT NULL,
    ref_txn_id     INTEGER,
    updated_by     TEXT,
    updated_at     TEXT
);

CREATE TABLE journal_entries (
    entry_id            INTEGER PRIMARY KEY,
    merchant_id         INTEGER NOT NULL DEFAULT 0,
    txn_id              INTEGER NOT NULL,
    ledger_id           INTEGER,
    account_id          TEXT    NOT NULL,
    debit               REAL    NOT NULL DEFAULT 0,
    credit              REAL    NOT NULL DEFAULT 0,
    note                TEXT,
    cash_flow_category  TEXT    CHECK (cash_flow_category IN ('OPERATING', 'INVESTING', 'FINANCING')),
    updated_by          TEXT,
    updated_at          TEXT
);

CREATE TABLE installments (
    installment_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id       INTEGER NOT NULL DEFAULT 0,
    ledger_id         INTEGER NOT NULL,
    txn_id            INTEGER,
    description       TEXT    NOT NULL,
    total_amount      REAL    NOT NULL,
    total_periods     INTEGER NOT NULL,
    amount_per_period REAL    NOT NULL,
    start_date        TEXT    NOT NULL,
    end_date          TEXT,
    interest_rate     REAL    NOT NULL DEFAULT 0,
    interest_type     TEXT    NOT NULL,
    status            TEXT    NOT NULL DEFAULT 'ACTIVE',
    note              TEXT,
    updated_by        TEXT,
    updated_at        TEXT
);

CREATE TABLE installment_payments (
    payment_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id    INTEGER NOT NULL DEFAULT 0,
    installment_id INTEGER NOT NULL,
    txn_id         INTEGER,
    period_no      INTEGER NOT NULL,
    amount         REAL    NOT NULL,
    interest       REAL    NOT NULL,
    due_date       TEXT    NOT NULL,
    paid_date      TEXT,
    status         TEXT    NOT NULL DEFAULT 'PENDING',
    updated_by     TEXT,
    updated_at     TEXT
);

CREATE TABLE reconciliations (
    recon_id          INTEGER PRIMARY KEY,
    merchant_id       INTEGER NOT NULL DEFAULT 0,
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
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    recon_id        INTEGER NOT NULL,
    txn_id          INTEGER NOT NULL,
    adjustment_type TEXT    NOT NULL,
    amount          REAL    NOT NULL,
    note            TEXT
);

CREATE TABLE period_closings (
    closing_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id    INTEGER NOT NULL DEFAULT 0,
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
    updated_by     TEXT,
    updated_at     TEXT
);

CREATE TABLE account_balance_snapshots (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
    closing_id   INTEGER NOT NULL,
    account_id   TEXT    NOT NULL,
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    UNIQUE(closing_id, account_id, merchant_id)
);

CREATE TABLE investments (
    investment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id   INTEGER NOT NULL DEFAULT 0,
    account_id    TEXT    NOT NULL,
    asset_type    TEXT    NOT NULL,
    currency      TEXT    NOT NULL,
    symbol        TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    cost_method   TEXT    NOT NULL DEFAULT 'FIFO',
    ifrs_category TEXT    NOT NULL DEFAULT 'FVTPL',
    is_active     INTEGER NOT NULL DEFAULT 1,
    version       INTEGER NOT NULL,
    updated_by    TEXT,
    updated_at    TEXT
);

CREATE TABLE investment_lots (
    lot_id        INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id   INTEGER NOT NULL DEFAULT 0,
    investment_id INTEGER NOT NULL,
    movement_id   INTEGER NOT NULL,
    acquired_date TEXT    NOT NULL,
    txn_id        INTEGER,
    quantity      REAL    NOT NULL,
    unit_cost     REAL    NOT NULL,
    total_cost    REAL    NOT NULL,
    remaining_qty       REAL NOT NULL,
    status              TEXT NOT NULL DEFAULT 'OPEN',
    unrealized_unit_twd REAL NOT NULL DEFAULT 0,
    updated_by          TEXT,
    updated_at          TEXT
);

CREATE TABLE investment_movements (
    movement_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
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
    withholding_tax REAL,
    updated_by      TEXT,
    updated_at      TEXT
);

CREATE TABLE investment_positions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id     INTEGER NOT NULL DEFAULT 0,
    investment_id   INTEGER NOT NULL UNIQUE,
    total_quantity  REAL    NOT NULL,
    total_cost      REAL    NOT NULL,
    avg_cost        REAL    NOT NULL DEFAULT 0,
    market_price_twd REAL   NOT NULL DEFAULT 0,
    updated_by      TEXT,
    updated_at      TEXT
);

CREATE TABLE investment_lot_disposals (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id         INTEGER NOT NULL DEFAULT 0,
    lot_id              INTEGER,
    txn_id              INTEGER,
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

CREATE TABLE merchants (
    merchant_id  INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT    NOT NULL,
    display_name TEXT    NOT NULL,
    currency     TEXT    NOT NULL DEFAULT 'TWD',
    status       TEXT    NOT NULL DEFAULT 'ACTIVE',
    created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE user_merchants (
    user_id     INTEGER NOT NULL REFERENCES users(user_id),
    merchant_id INTEGER NOT NULL REFERENCES merchants(merchant_id),
    role        TEXT    NOT NULL DEFAULT 'MEMBER',
    joined_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, merchant_id)
);

CREATE TABLE account_closure (
    ancestor_id   TEXT    NOT NULL,
    descendant_id TEXT    NOT NULL,
    merchant_id   INTEGER NOT NULL DEFAULT 0,
    depth         INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id, merchant_id)
);

CREATE INDEX idx_account_closure_descendant ON account_closure(descendant_id, merchant_id);

CREATE TRIGGER trg_account_insert_closure
AFTER INSERT ON accounts
BEGIN
    INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, merchant_id, depth)
    VALUES (NEW.account_id, NEW.account_id, NEW.merchant_id, 0);
    INSERT OR IGNORE INTO account_closure (ancestor_id, descendant_id, merchant_id, depth)
    SELECT ancestor_id, NEW.account_id, NEW.merchant_id, depth + 1
    FROM account_closure
    WHERE descendant_id = NEW.parent_id AND merchant_id = NEW.merchant_id;
END;

CREATE VIEW v_parent_balance_agg AS
SELECT snap.closing_id, snap.merchant_id, a.account_id,
       COALESCE(SUM(snap.debit_total),  0) AS debit_total,
       COALESCE(SUM(snap.credit_total), 0) AS credit_total
FROM accounts a
JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.merchant_id = a.merchant_id AND ac.depth > 0
JOIN accounts leaf      ON leaf.account_id = ac.descendant_id AND leaf.merchant_id = a.merchant_id
                       AND leaf.is_summary = 0 AND leaf.is_active = 1
JOIN account_balance_snapshots snap ON snap.account_id = ac.descendant_id AND snap.merchant_id = a.merchant_id
WHERE a.is_summary = 1
GROUP BY snap.closing_id, snap.merchant_id, a.account_id;

CREATE TABLE ledger_account_balance_snapshots (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
    closing_id   INTEGER NOT NULL,
    ledger_id    INTEGER NOT NULL,
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    UNIQUE(closing_id, ledger_id, merchant_id)
);

CREATE INDEX IF NOT EXISTS idx_labs_closing ON ledger_account_balance_snapshots(closing_id);

CREATE VIEW v_ledger_account_balances AS
SELECT
    la.ledger_id,
    la.merchant_id,
    a.normal_balance,
    COALESCE(SUM(je.debit), 0)  AS debit_total,
    COALESCE(SUM(je.credit), 0) AS credit_total
FROM ledger_accounts la
JOIN accounts a             ON la.account_id = a.account_id AND la.merchant_id = a.merchant_id
LEFT JOIN journal_entries je ON la.ledger_id = je.ledger_id AND la.account_id = je.account_id
                             AND la.merchant_id = je.merchant_id
LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
WHERE la.is_active = 1
GROUP BY la.ledger_id, la.merchant_id;

CREATE VIEW v_account_balances AS
WITH
leaf_balances AS (
    SELECT a.account_id, a.merchant_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(je.debit), 0)  AS debit_total,
           COALESCE(SUM(je.credit), 0) AS credit_total
    FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id = je.account_id AND a.merchant_id = je.merchant_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE a.is_active = 1 AND a.is_summary = 0
    GROUP BY a.account_id, a.merchant_id
),
parent_balances AS (
    SELECT a.account_id, a.merchant_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(je.debit), 0)  AS debit_total,
           COALESCE(SUM(je.credit), 0) AS credit_total
    FROM accounts a
    JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.merchant_id = a.merchant_id AND ac.depth > 0
    LEFT JOIN journal_entries je ON je.account_id = ac.descendant_id AND je.merchant_id = a.merchant_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE a.is_active = 1 AND a.is_summary = 1
    GROUP BY a.account_id, a.merchant_id
)
SELECT account_id, merchant_id, name, type, normal_balance, debit_total, credit_total FROM leaf_balances
UNION ALL
SELECT account_id, merchant_id, name, type, normal_balance, debit_total, credit_total FROM parent_balances;

CREATE TABLE account_running_balances (
    account_id   TEXT    NOT NULL,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    PRIMARY KEY (account_id, merchant_id),
    FOREIGN KEY (account_id) REFERENCES accounts(account_id)
);

CREATE TABLE ledger_running_balances (
    ledger_id    INTEGER NOT NULL,
    merchant_id  INTEGER NOT NULL DEFAULT 0,
    debit_total  REAL    NOT NULL DEFAULT 0,
    credit_total REAL    NOT NULL DEFAULT 0,
    PRIMARY KEY (ledger_id, merchant_id),
    FOREIGN KEY (ledger_id) REFERENCES ledger_accounts(ledger_id)
);

CREATE TABLE ledger_account_type_config (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    type        TEXT    NOT NULL CHECK(type IN ('BANK_ACCOUNT','LOAN','CREDIT_CARD')),
    account_id  TEXT    NOT NULL,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now')),
    version     INTEGER NOT NULL DEFAULT 1,
    UNIQUE(merchant_id, type),
    FOREIGN KEY(account_id, merchant_id) REFERENCES accounts(account_id, merchant_id)
);

CREATE TABLE asset_type_account_config (
    id                         INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id                INTEGER NOT NULL,
    asset_type                 TEXT    NOT NULL CHECK(asset_type IN ('STOCK','FUND','GOLD','FX')),
    realized_gain_account_id   TEXT    NOT NULL,
    realized_loss_account_id   TEXT    NOT NULL,
    unrealized_gain_account_id TEXT    NOT NULL,
    unrealized_loss_account_id TEXT    NOT NULL,
    oci_account_id             TEXT,
    fee_account_id             TEXT    NOT NULL,
    tax_account_id             TEXT    NOT NULL,
    updated_by                 TEXT,
    updated_at                 TEXT    DEFAULT (datetime('now')),
    version                    INTEGER NOT NULL DEFAULT 1,
    UNIQUE(merchant_id, asset_type)
);

CREATE TABLE prepaids (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id         INTEGER NOT NULL,
    txn_id              INTEGER,
    account_id          TEXT    NOT NULL,
    expense_account_id  TEXT    NOT NULL,
    name                TEXT    NOT NULL,
    total_amount        REAL    NOT NULL,
    amortized_amount    REAL    NOT NULL DEFAULT 0,
    periods             INTEGER NOT NULL,
    amortized_periods   INTEGER NOT NULL DEFAULT 0,
    start_date          TEXT    NOT NULL,
    status              TEXT    NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','COMPLETED','DISPOSED')),
    updated_by          TEXT,
    updated_at          TEXT    DEFAULT (datetime('now')),
    version             INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE prepaid_amortizations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    prepaid_id  INTEGER NOT NULL REFERENCES prepaids(id),
    txn_id      INTEGER NOT NULL,
    period_date TEXT    NOT NULL,
    amount      REAL    NOT NULL,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now'))
);

CREATE TABLE fixed_assets (
    id                              INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id                     INTEGER NOT NULL,
    txn_id                          INTEGER,
    name                            TEXT    NOT NULL,
    asset_account_id                TEXT    NOT NULL,
    accum_depreciation_account_id   TEXT    NOT NULL,
    depreciation_expense_account_id TEXT    NOT NULL,
    cost                            REAL    NOT NULL,
    residual_value                  REAL    NOT NULL DEFAULT 0,
    useful_life_months              INTEGER NOT NULL,
    depreciation_method             TEXT    NOT NULL DEFAULT 'STRAIGHT_LINE' CHECK(depreciation_method IN ('STRAIGHT_LINE')),
    payment_type                    TEXT    NOT NULL CHECK(payment_type IN ('CASH','LEASE')),
    total_depreciated               REAL    NOT NULL DEFAULT 0,
    depreciated_periods             INTEGER NOT NULL DEFAULT 0,
    purchase_date                   TEXT    NOT NULL,
    disposal_date                   TEXT,
    status                          TEXT    NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','DISPOSED')),
    updated_by                      TEXT,
    updated_at                      TEXT    DEFAULT (datetime('now')),
    version                         INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE fixed_asset_depreciations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL,
    asset_id    INTEGER NOT NULL REFERENCES fixed_assets(id),
    txn_id      INTEGER NOT NULL,
    period_date TEXT    NOT NULL,
    amount      REAL    NOT NULL,
    updated_by  TEXT,
    updated_at  TEXT    DEFAULT (datetime('now'))
);
