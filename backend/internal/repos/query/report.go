package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/report"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type ReportRepo interface {
	GetBalanceSheet(ctx context.Context, reportDate string) (*report.BalanceSheet, error)
	GetIncomeStatement(ctx context.Context, startDate, endDate string) (*report.IncomeStatement, error)
	GetCashFlowStatement(ctx context.Context, startDate, endDate string) (*report.CashFlowStatement, error)
	GetDirectCashFlowStatement(ctx context.Context, startDate, endDate string) (*report.DirectCashFlowStatement, error)
	GetEquityStatement(ctx context.Context, startDate, endDate string) (*report.EquityStatement, error)
}

type sqlcReportRepo struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func newReportRepo(q *sqlcdb.Queries, db *sqlx.DB) ReportRepo {
	return &sqlcReportRepo{q: q, db: db}
}

func NewReportRepo(db *sqlx.DB) ReportRepo {
	return &sqlcReportRepo{q: sqlcdb.New(db), db: db}
}

// bsRawRow is the local scan target for balance sheet queries; total_debit/total_credit use decimal to avoid float rounding.
type bsRawRow struct {
	AccountId     string              `db:"account_id"`
	Name          string              `db:"name"`
	Type          enums.AccountType   `db:"type"`
	NormalBalance enums.NormalBalance `db:"normal_balance"`
	TotalDebit    decimal.Decimal     `db:"total_debit"`
	TotalCredit   decimal.Decimal     `db:"total_credit"`
	ParentId      *string             `db:"parent_id"`
	HasChild      bool                `db:"has_child"`
	Depth         int                 `db:"depth"`
}

// isRawRow is the local scan target for income statement queries; period_debit/period_credit may be negative after two-snapshot subtraction.
type isRawRow struct {
	AccountId     string              `db:"account_id"`
	Name          string              `db:"name"`
	Type          enums.AccountType   `db:"type"`
	NormalBalance enums.NormalBalance `db:"normal_balance"`
	PeriodDebit   decimal.Decimal     `db:"period_debit"`
	PeriodCredit  decimal.Decimal     `db:"period_credit"`
	ParentId      *string             `db:"parent_id"`
	HasChild      bool                `db:"has_child"`
	Depth         int                 `db:"depth"`
}

// queryBalanceSheetNoSnap scans all transactions up to reportDate, filtered by merchant_id.
// args merchant_id, report_date
const queryBalanceSheetNoSnap = `
WITH has_child_set AS (
    SELECT DISTINCT parent_id AS account_id FROM accounts WHERE parent_id IS NOT NULL AND merchant_id = :merchant_id
),
account_depths AS (
    SELECT descendant_id AS account_id, MAX(depth) AS depth FROM account_closure WHERE merchant_id = :merchant_id GROUP BY descendant_id 
),
leaf_balances AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(je.debit), 0)  AS total_debit,
           COALESCE(SUM(je.credit), 0) AS total_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    LEFT JOIN journal_entries je ON a.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE' AND t.txn_date <= :report_date AND t.merchant_id = :merchant_id
    WHERE a.is_summary = 0 AND a.is_active = 1 AND a.type IN ('ASSET','LIABILITY','EQUITY') AND a.merchant_id = :merchant_id
    GROUP BY a.account_id
),
parent_balances AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(je.debit), 0)  AS total_debit,
           COALESCE(SUM(je.credit), 0) AS total_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.depth > 0
    LEFT JOIN journal_entries je ON je.account_id = ac.descendant_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE' AND t.txn_date <= :report_date AND t.merchant_id = :merchant_id
    WHERE a.is_summary = 1 AND a.is_active = 1 AND a.type IN ('ASSET','LIABILITY','EQUITY') AND a.merchant_id = :merchant_id
    GROUP BY a.account_id
)
SELECT account_id, name, type, normal_balance, total_debit, total_credit, parent_id, has_child, depth FROM (
    SELECT account_id, name, type, normal_balance, total_debit, total_credit, parent_id, has_child, depth FROM leaf_balances
    UNION ALL
    SELECT account_id, name, type, normal_balance, total_debit, total_credit, parent_id, has_child, depth FROM parent_balances
) ORDER BY CASE type WHEN 'ASSET' THEN 1 WHEN 'LIABILITY' THEN 2 WHEN 'EQUITY' THEN 3 END, account_id`

// queryBalanceSheetWithSnap applies snapshot baseline + post-snapshot delta up to reportDate, filtered by merchant_id.
// args merchant_id, closing_id, start_date, end_date
const queryBalanceSheetWithSnap = `
WITH has_child_set AS (
    SELECT DISTINCT parent_id AS account_id FROM accounts WHERE parent_id IS NOT NULL AND merchant_id = :merchant_id
),
account_depths AS (
    SELECT descendant_id AS account_id, MAX(depth) AS depth FROM account_closure WHERE merchant_id = :merchant_id GROUP BY descendant_id 
),
snap AS (
    SELECT account_id, debit_total, credit_total
    FROM account_balance_snapshots WHERE closing_id = :closing_id AND merchant_id = :merchant_id
),
delta AS (
    SELECT je.account_id,
           COALESCE(SUM(je.debit), 0)  AS debit_total,
           COALESCE(SUM(je.credit), 0) AS credit_total
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE t.txn_date > :start_date AND t.txn_date <= :end_date AND t.merchant_id = :merchant_id
    GROUP BY je.account_id
),
parent_delta AS (
    -- Aggregate only leaf-account deltas through closure; snapshot for parents is fetched directly below.
    SELECT ac.ancestor_id AS account_id,
           COALESCE(SUM(delta.debit_total), 0)  AS debit_total,
           COALESCE(SUM(delta.credit_total), 0) AS credit_total
    FROM account_closure ac
    JOIN delta ON delta.account_id = ac.descendant_id
    JOIN accounts leaf ON leaf.account_id = ac.descendant_id AND leaf.is_summary = 0
    WHERE ac.depth > 0 GROUP BY ac.ancestor_id
),
leaf_balances AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(snap.debit_total, 0)  + COALESCE(delta.debit_total, 0)  AS total_debit,
           COALESCE(snap.credit_total, 0) + COALESCE(delta.credit_total, 0) AS total_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    LEFT JOIN snap  ON a.account_id = snap.account_id
    LEFT JOIN delta ON a.account_id = delta.account_id
    WHERE a.is_summary = 0 AND a.is_active = 1 AND a.type IN ('ASSET','LIABILITY','EQUITY') AND a.merchant_id = :merchant_id
),
parent_balances AS (
    -- Use snapshot directly (BulkInsert already wrote aggregated parent values); only delta needs closure JOIN.
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(snap.debit_total, 0)  + COALESCE(pd.debit_total, 0) AS total_debit,
           COALESCE(snap.credit_total, 0) + COALESCE(pd.credit_total, 0) AS total_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    LEFT JOIN snap         ON a.account_id = snap.account_id
    LEFT JOIN parent_delta pd ON a.account_id = pd.account_id
    WHERE a.is_summary = 1 AND a.is_active = 1 AND a.type IN ('ASSET','LIABILITY','EQUITY') AND a.merchant_id = :merchant_id
)
SELECT account_id, name, type, normal_balance, total_debit, total_credit, parent_id, has_child, depth FROM (
    SELECT account_id, name, type, normal_balance, total_debit, total_credit, parent_id, has_child, depth FROM leaf_balances
    UNION ALL
    SELECT account_id, name, type, normal_balance, total_debit, total_credit, parent_id, has_child, depth FROM parent_balances
) ORDER BY CASE type WHEN 'ASSET' THEN 1 WHEN 'LIABILITY' THEN 2 WHEN 'EQUITY' THEN 3 END, account_id`

// queryIncomeStatementNoSnap scans transactions within [startDate, endDate], filtered by merchant_id.
// args: merchant_id, start_date, end_date
const queryIncomeStatementNoSnap = `
WITH has_child_set AS (
    SELECT DISTINCT parent_id AS account_id FROM accounts WHERE parent_id IS NOT NULL AND merchant_id = :merchant_id
),
account_depths AS (
    SELECT descendant_id AS account_id, MAX(depth) AS depth FROM account_closure WHERE merchant_id = :merchant_id GROUP BY descendant_id
),
period_entries AS (
    SELECT je.account_id,
           COALESCE(SUM(je.debit), 0)  AS period_debit,
           COALESCE(SUM(je.credit), 0) AS period_credit
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
        AND t.txn_date >= :start_date AND t.txn_date <= :end_date
        AND t.merchant_id = :merchant_id
    WHERE je.merchant_id = :merchant_id
    GROUP BY je.account_id
),
leaf_period AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(pe.period_debit, 0)  AS period_debit,
           COALESCE(pe.period_credit, 0) AS period_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    LEFT JOIN period_entries  pe ON a.account_id   = pe.account_id
    WHERE a.is_summary = 0 AND a.is_active = 1 AND a.type IN ('INCOME','EXPENSE') AND a.merchant_id = :merchant_id
),
parent_period AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           COALESCE(SUM(pe.period_debit), 0)  AS period_debit,
           COALESCE(SUM(pe.period_credit), 0) AS period_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.depth > 0 AND ac.merchant_id = a.merchant_id
    LEFT JOIN period_entries  pe ON pe.account_id  = ac.descendant_id
    WHERE a.is_summary = 1 AND a.is_active = 1 AND a.type IN ('INCOME','EXPENSE') AND a.merchant_id = :merchant_id
    GROUP BY a.account_id
)
SELECT account_id, name, type, normal_balance, period_debit, period_credit, parent_id, has_child, depth FROM (
    SELECT account_id, name, type, normal_balance, period_debit, period_credit, parent_id, has_child, depth FROM leaf_period
    UNION ALL
    SELECT account_id, name, type, normal_balance, period_debit, period_credit, parent_id, has_child, depth FROM parent_period
) ORDER BY CASE type WHEN 'INCOME' THEN 1 WHEN 'EXPENSE' THEN 2 END, account_id`

// queryIncomeStatementWithSnaps computes period income as (snap_end + delta_end) - (snap_pre + delta_pre), filtered by merchant_id.
// args order: merchantID, snapEndID, merchantID, snapPreID, merchantID, snapEndPeriodEnd, endDate, merchantID, snapPrePeriodEnd, startDate, merchantID, merchantID, merchantID
// args: end_id, period_id, start_date, period_end_date, end_end_date, end_date, merchant_id
const queryIncomeStatementWithSnaps = `
WITH has_child_set AS (
    SELECT DISTINCT parent_id AS account_id FROM accounts WHERE parent_id IS NOT NULL AND merchant_id = :merchant_id
),
account_depths AS (
    SELECT descendant_id AS account_id, MAX(depth) AS depth FROM account_closure WHERE merchant_id = :merchant_id GROUP BY descendant_id 
),
snap_end AS (
    SELECT account_id, debit_total, credit_total FROM account_balance_snapshots WHERE closing_id = :end_id AND merchant_id = :merchant_id
),
snap_pre AS (
    SELECT account_id, debit_total, credit_total FROM account_balance_snapshots WHERE closing_id = :period_id AND merchant_id = :merchant_id
),
delta_end AS (
    SELECT je.account_id,
           COALESCE(SUM(je.debit), 0)  AS debit,
           COALESCE(SUM(je.credit), 0) AS credit
    FROM journal_entries je JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE t.txn_date > :end_end_date AND t.txn_date <= :end_date AND t.merchant_id = :merchant_id
    GROUP BY je.account_id
),
delta_pre AS (
    SELECT je.account_id,
           COALESCE(SUM(je.debit), 0)  AS debit,
           COALESCE(SUM(je.credit), 0) AS credit
    FROM journal_entries je JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE t.txn_date > :period_end_date AND t.txn_date < :start_date AND t.merchant_id = :merchant_id
    GROUP BY je.account_id
),
parent_delta_end AS (
    -- Aggregate only leaf-account deltas; snapshot for parents is fetched directly below.
    SELECT ac.ancestor_id AS account_id,
           COALESCE(SUM(de.debit), 0)  AS debit,
           COALESCE(SUM(de.credit), 0) AS credit
    FROM account_closure ac
    JOIN delta_end de ON de.account_id = ac.descendant_id
    JOIN accounts leaf ON leaf.account_id = ac.descendant_id AND leaf.is_summary = 0
    WHERE ac.depth > 0 GROUP BY ac.ancestor_id
),
parent_delta_pre AS (
    -- Aggregate only leaf-account deltas; snapshot for parents is fetched directly below.
    SELECT ac.ancestor_id AS account_id,
           COALESCE(SUM(dp.debit), 0)  AS debit,
           COALESCE(SUM(dp.credit), 0) AS credit
    FROM account_closure ac
    JOIN delta_pre dp ON dp.account_id = ac.descendant_id
    JOIN accounts leaf ON leaf.account_id = ac.descendant_id AND leaf.is_summary = 0
    WHERE ac.depth > 0 GROUP BY ac.ancestor_id
),
leaf_period AS (
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           (COALESCE(se.debit_total, 0) + COALESCE(de.debit, 0)) - (COALESCE(sp.debit_total, 0) + COALESCE(dp.debit, 0)) AS period_debit,
           (COALESCE(se.credit_total, 0) + COALESCE(de.credit, 0)) - (COALESCE(sp.credit_total, 0) + COALESCE(dp.credit, 0)) AS period_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    LEFT JOIN snap_end se ON a.account_id = se.account_id
    LEFT JOIN snap_pre sp ON a.account_id = sp.account_id
    LEFT JOIN delta_end de ON a.account_id = de.account_id
    LEFT JOIN delta_pre dp ON a.account_id = dp.account_id
    WHERE a.is_summary = 0 AND a.is_active = 1 AND a.type IN ('INCOME','EXPENSE') AND a.merchant_id = :merchant_id
),
parent_period AS (
    -- Snapshot for parents is taken directly (BulkInsert already wrote aggregated parent values).
    SELECT a.account_id, a.name, a.type, a.normal_balance,
           (COALESCE(se.debit_total, 0) + COALESCE(pde.debit, 0)) - (COALESCE(sp.debit_total, 0) + COALESCE(pdp.debit, 0)) AS period_debit,
           (COALESCE(se.credit_total, 0) + COALESCE(pde.credit, 0)) - (COALESCE(sp.credit_total, 0) + COALESCE(pdp.credit, 0)) AS period_credit,
           a.parent_id,
           hcs.account_id IS NOT NULL  AS has_child,
           COALESCE(ad.depth, 0)       AS depth
    FROM accounts a
    LEFT JOIN has_child_set hcs ON hcs.account_id = a.account_id
    LEFT JOIN account_depths  ad ON ad.account_id  = a.account_id
    LEFT JOIN snap_end          se  ON a.account_id = se.account_id
    LEFT JOIN snap_pre          sp  ON a.account_id = sp.account_id
    LEFT JOIN parent_delta_end  pde ON a.account_id = pde.account_id
    LEFT JOIN parent_delta_pre  pdp ON a.account_id = pdp.account_id
    WHERE a.is_summary = 1 AND a.is_active = 1 AND a.type IN ('INCOME','EXPENSE') AND a.merchant_id = :merchant_id
)
SELECT account_id, name, type, normal_balance, period_debit, period_credit, parent_id, has_child, depth FROM (
    SELECT account_id, name, type, normal_balance, period_debit, period_credit, parent_id, has_child, depth FROM leaf_period
    UNION ALL
    SELECT account_id, name, type, normal_balance, period_debit, period_credit, parent_id, has_child, depth FROM parent_period
) ORDER BY CASE type WHEN 'INCOME' THEN 1 WHEN 'EXPENSE' THEN 2 END, account_id`

// cfRawRow is the local scan target for cash flow account period changes.
type cfRawRow struct {
	AccountId        string          `db:"account_id"`
	Name             string          `db:"name"`
	CashFlowCategory string          `db:"cash_flow_category"`
	IsSummary        bool            `db:"is_summary"`
	PeriodDebit      decimal.Decimal `db:"period_debit"`
	PeriodCredit     decimal.Decimal `db:"period_credit"`
}

// cashSumRow scans debit/credit totals for cash account balance queries.
type cashSumRow struct {
	DebitTotal  decimal.Decimal `db:"debit_total"`
	CreditTotal decimal.Decimal `db:"credit_total"`
}

// queryCashFlowChanges fetches period debit/credit grouped by journal_entries.cash_flow_category.
// Leaf accounts join directly; summary accounts aggregate descendants via account_closure.
// A parent account can appear in multiple sections if its descendants have entries with different categories.
// args: merchant_id, start_date, end_date
const queryCashFlowChanges = `
WITH period_entries AS (
    SELECT je.account_id, je.cash_flow_category,
           COALESCE(SUM(je.debit), 0)  AS period_debit,
           COALESCE(SUM(je.credit), 0) AS period_credit
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
        AND t.txn_date >= :start_date AND t.txn_date <= :end_date
        AND t.merchant_id = :merchant_id
    WHERE je.merchant_id = :merchant_id
      AND je.cash_flow_category IN ('OPERATING', 'INVESTING', 'FINANCING')
    GROUP BY je.account_id, je.cash_flow_category
),
leaf_cf AS (
    SELECT a.account_id, a.name, pe.cash_flow_category, 0 AS is_summary,
           pe.period_debit, pe.period_credit
    FROM period_entries pe
    JOIN accounts a ON a.account_id = pe.account_id
        AND a.is_active = 1 AND a.is_summary = 0 AND a.merchant_id = :merchant_id
),
summary_cf AS (
    SELECT a.account_id, a.name, pe.cash_flow_category, 1 AS is_summary,
           COALESCE(SUM(pe.period_debit), 0)  AS period_debit,
           COALESCE(SUM(pe.period_credit), 0) AS period_credit
    FROM accounts a
    JOIN account_closure ac ON ac.ancestor_id = a.account_id AND ac.depth > 0 AND ac.merchant_id = a.merchant_id
    JOIN accounts leaf ON leaf.account_id = ac.descendant_id AND leaf.is_summary = 0
                       AND leaf.is_active = 1 AND leaf.merchant_id = a.merchant_id
    JOIN period_entries pe ON pe.account_id = ac.descendant_id
    WHERE a.is_active = 1 AND a.is_summary = 1 AND a.merchant_id = :merchant_id
    GROUP BY a.account_id, a.name, pe.cash_flow_category
)
SELECT account_id, name, cash_flow_category, is_summary, period_debit, period_credit FROM (
    SELECT account_id, name, cash_flow_category, is_summary, period_debit, period_credit FROM leaf_cf
    UNION ALL
    SELECT account_id, name, cash_flow_category, is_summary, period_debit, period_credit FROM summary_cf
) ORDER BY cash_flow_category, account_id`

// queryCashBefore fetches cumulative debit/credit for CASH accounts before a date (exclusive).
// args: merchant_id, date
const queryCashBefore = `
SELECT COALESCE(SUM(je.debit), 0)  AS debit_total,
       COALESCE(SUM(je.credit), 0) AS credit_total
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    AND t.txn_date < :date AND t.merchant_id = :merchant_id
JOIN accounts a ON je.account_id = a.account_id
    AND a.cash_flow_category = 'CASH' AND a.merchant_id = :merchant_id
WHERE je.merchant_id = :merchant_id`

// queryCashUpTo fetches cumulative debit/credit for CASH accounts up to a date (inclusive).
// args: merchant_id, date
const queryCashUpTo = `
SELECT COALESCE(SUM(je.debit), 0)  AS debit_total,
       COALESCE(SUM(je.credit), 0) AS credit_total
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    AND t.txn_date <= :date AND t.merchant_id = :merchant_id
JOIN accounts a ON je.account_id = a.account_id
    AND a.cash_flow_category = 'CASH' AND a.merchant_id = :merchant_id
WHERE je.merchant_id = :merchant_id`

// queryDirectOperatingCash fetches period debit/credit of CASH accounts for operating transactions.
// Operating transactions = those containing INCOME/EXPENSE entries OR OPERATING-tagged entries.
// args: merchant_id, start_date, end_date
// sqlx.Named is required because the IN clause for account types cannot be expressed statically.
const queryDirectOperatingCash = `
WITH period_txns AS (
    SELECT DISTINCT je.txn_id
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
        AND t.txn_date >= :start_date AND t.txn_date <= :end_date
        AND t.merchant_id = :merchant_id
    WHERE je.merchant_id = :merchant_id
),
operating_txns AS (
    SELECT DISTINCT je.txn_id
    FROM journal_entries je
    JOIN accounts a ON je.account_id = a.account_id AND a.merchant_id = je.merchant_id
    WHERE je.txn_id IN (SELECT txn_id FROM period_txns)
      AND je.merchant_id = :merchant_id
      AND (a.type IN ('INCOME', 'EXPENSE') OR je.cash_flow_category = 'OPERATING')
)
SELECT
    COALESCE(SUM(je.debit),  0) AS debit_total,
    COALESCE(SUM(je.credit), 0) AS credit_total
FROM journal_entries je
JOIN accounts a ON je.account_id = a.account_id AND a.merchant_id = je.merchant_id
    AND a.cash_flow_category = 'CASH'
WHERE je.txn_id IN (SELECT txn_id FROM operating_txns)
  AND je.merchant_id = :merchant_id`

func bsRowFromRaw(row bsRawRow) (report.BalanceSheetRow, decimal.Decimal) {
	balance := row.TotalDebit.Sub(row.TotalCredit)
	if row.NormalBalance.Is(enums.BalanceCredit) {
		balance = row.TotalCredit.Sub(row.TotalDebit)
	}
	return report.BalanceSheetRow{
		Type:      row.Type,
		AccountId: row.AccountId,
		Name:      row.Name,
		Balance:   balance,
		ParentId:  row.ParentId,
		HasChild:  row.HasChild,
		Depth:     row.Depth,
	}, balance
}

func isRowFromRaw(row isRawRow) (report.IncomeStatementRow, decimal.Decimal) {
	amount := row.PeriodDebit.Sub(row.PeriodCredit)
	if row.NormalBalance.Is(enums.BalanceCredit) {
		amount = row.PeriodCredit.Sub(row.PeriodDebit)
	}
	return report.IncomeStatementRow{
		Type:      row.Type,
		AccountID: row.AccountId,
		Name:      row.Name,
		Amount:    amount,
		ParentId:  row.ParentId,
		HasChild:  row.HasChild,
		Depth:     row.Depth,
	}, amount
}

func (r *sqlcReportRepo) GetBalanceSheet(ctx context.Context, reportDate string) (*report.BalanceSheet, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	snap, err := r.q.GetLatestClosedPeriodBeforeOrOn(ctx, sqlcdb.GetLatestClosedPeriodBeforeOrOnParams{
		MerchantID: merchantID,
		PeriodType: enums.PeriodMonthly.Enum(),
		PeriodEnd:  reportDate,
	})

	var rawRows []bsRawRow
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		params := map[string]any{
			"merchant_id": merchantID,
			"report_date": reportDate,
		}
		query, args, err := sqlx.Named(queryBalanceSheetNoSnap, params)
		if err != nil {
			return nil, err
		}

		query = r.db.Rebind(query)
		if err := r.db.SelectContext(ctx, &rawRows, query, args...); err != nil {
			return nil, err
		}
	} else {
		params := map[string]any{
			"merchant_id": merchantID,
			"closing_id":  snap.ClosingID,
			"start_date":  snap.PeriodEnd,
			"end_date":    reportDate,
		}
		query, args, err := sqlx.Named(queryBalanceSheetWithSnap, params)
		if err != nil {
			return nil, err
		}

		query = r.db.Rebind(query)
		if err := r.db.SelectContext(ctx, &rawRows, query, args...); err != nil {
			return nil, err
		}
	}

	bs := &report.BalanceSheet{ReportDate: reportDate}
	for _, row := range rawRows {
		bsRow, balance := bsRowFromRaw(row)
		switch row.Type.Val() {
		case enums.AccountAsset:
			bs.Assets = append(bs.Assets, bsRow)
			if !bsRow.HasChild {
				bs.TotalAssets = bs.TotalAssets.Add(balance)
			}
		case enums.AccountLiability:
			bs.Liabilities = append(bs.Liabilities, bsRow)
			if !bsRow.HasChild {
				bs.TotalLiabilities = bs.TotalLiabilities.Add(balance)
			}
		case enums.AccountEquity:
			bs.Equity = append(bs.Equity, bsRow)
			if !bsRow.HasChild {
				bs.TotalEquity = bs.TotalEquity.Add(balance)
			}
		}
	}
	bs.NetWorth = bs.TotalAssets.Sub(bs.TotalLiabilities)
	return bs, nil
}

func (r *sqlcReportRepo) GetIncomeStatement(ctx context.Context, startDate, endDate string) (*report.IncomeStatement, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	snapEnd, errEnd := r.q.GetLatestClosedPeriodBetween(ctx, sqlcdb.GetLatestClosedPeriodBetweenParams{
		PeriodType: enums.PeriodMonthly.Enum(),
		Start:      startDate,
		End:        endDate,
		MerchantID: merchantID,
	})
	snapPre, errPre := r.q.GetLatestClosedPeriodBefore(ctx, sqlcdb.GetLatestClosedPeriodBeforeParams{
		PeriodType: enums.PeriodMonthly.Enum(),
		PeriodEnd:  startDate,
		MerchantID: merchantID,
	})

	if errEnd != nil && !errors.Is(errEnd, sql.ErrNoRows) {
		return nil, errEnd
	}
	if errPre != nil && !errors.Is(errPre, sql.ErrNoRows) {
		return nil, errPre
	}

	var rawRows []isRawRow
	if errEnd == nil && errPre == nil {
		params := map[string]any{
			"end_id":          snapEnd.ClosingID,
			"period_id":       snapPre.ClosingID,
			"start_date":      startDate,
			"period_end_date": snapPre.PeriodEnd,
			"end_end_date":    snapEnd.PeriodEnd,
			"end_date":        endDate,
			"merchant_id":     merchantID,
		}
		query, args, err := sqlx.Named(queryIncomeStatementWithSnaps, params)
		if err != nil {
			return nil, err
		}
		query = r.db.Rebind(query)
		if err := r.db.SelectContext(ctx, &rawRows, query, args...); err != nil {
			return nil, err
		}
	} else {
		params := map[string]any{
			"merchant_id": merchantID,
			"start_date":  startDate,
			"end_date":    endDate,
		}
		query, args, err := sqlx.Named(queryIncomeStatementNoSnap, params)
		if err != nil {
			return nil, err
		}

		query = r.db.Rebind(query)

		if err = r.db.SelectContext(ctx, &rawRows, query, args...); err != nil {
			return nil, err
		}
	}

	is := &report.IncomeStatement{StartDate: startDate, EndDate: endDate}
	for _, row := range rawRows {
		isRow, amount := isRowFromRaw(row)
		switch row.Type.Val() {
		case enums.AccountIncome:
			is.Income = append(is.Income, isRow)
			if !isRow.HasChild {
				is.TotalIncome = is.TotalIncome.Add(amount)
			}
		case enums.AccountExpense:
			is.Expenses = append(is.Expenses, isRow)
			if !isRow.HasChild {
				is.TotalExpenses = is.TotalExpenses.Add(amount)
			}
		}
	}
	is.NetIncome = is.TotalIncome.Sub(is.TotalExpenses)
	return is, nil
}

func (r *sqlcReportRepo) GetCashFlowStatement(ctx context.Context, startDate, endDate string) (*report.CashFlowStatement, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	is, err := r.GetIncomeStatement(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	q, args, err := sqlx.Named(queryCashFlowChanges, map[string]any{
		"merchant_id": merchantID,
		"start_date":  startDate,
		"end_date":    endDate,
	})
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)
	var cfRows []cfRawRow
	if err := r.db.SelectContext(ctx, &cfRows, q, args...); err != nil {
		return nil, err
	}

	qBefore, argsBefore, err := sqlx.Named(queryCashBefore, map[string]any{"merchant_id": merchantID, "date": startDate})
	if err != nil {
		return nil, err
	}
	qBefore = r.db.Rebind(qBefore)
	var beginRow cashSumRow
	if err := r.db.GetContext(ctx, &beginRow, qBefore, argsBefore...); err != nil {
		return nil, err
	}

	qUpTo, argsUpTo, err := sqlx.Named(queryCashUpTo, map[string]any{"merchant_id": merchantID, "date": endDate})
	if err != nil {
		return nil, err
	}
	qUpTo = r.db.Rebind(qUpTo)
	var endRow cashSumRow
	if err := r.db.GetContext(ctx, &endRow, qUpTo, argsUpTo...); err != nil {
		return nil, err
	}

	cf := &report.CashFlowStatement{
		StartDate:     startDate,
		EndDate:       endDate,
		BeginningCash: beginRow.DebitTotal.Sub(beginRow.CreditTotal),
		EndingCash:    endRow.DebitTotal.Sub(endRow.CreditTotal),
	}
	cf.OperatingActivities.NetIncome = is.NetIncome

	for _, row := range cfRows {
		amount := row.PeriodCredit.Sub(row.PeriodDebit)
		item := report.CashFlowItem{AccountId: row.AccountId, Name: row.Name, IsSummary: row.IsSummary, Amount: amount}
		switch row.CashFlowCategory {
		case "OPERATING":
			cf.OperatingActivities.Adjustments = append(cf.OperatingActivities.Adjustments, item)
			if !row.IsSummary {
				cf.OperatingActivities.Total = cf.OperatingActivities.Total.Add(amount)
			}
		case "INVESTING":
			cf.InvestingActivities.Items = append(cf.InvestingActivities.Items, item)
			if !row.IsSummary {
				cf.InvestingActivities.Total = cf.InvestingActivities.Total.Add(amount)
			}
		case "FINANCING":
			cf.FinancingActivities.Items = append(cf.FinancingActivities.Items, item)
			if !row.IsSummary {
				cf.FinancingActivities.Total = cf.FinancingActivities.Total.Add(amount)
			}
		}
	}
	cf.OperatingActivities.Total = cf.OperatingActivities.Total.Add(is.NetIncome)
	cf.NetChange = cf.OperatingActivities.Total.Add(cf.InvestingActivities.Total).Add(cf.FinancingActivities.Total)

	return cf, nil
}

func (r *sqlcReportRepo) GetDirectCashFlowStatement(ctx context.Context, startDate, endDate string) (*report.DirectCashFlowStatement, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	qOp, argsOp, err := sqlx.Named(queryDirectOperatingCash, map[string]any{
		"merchant_id": merchantID,
		"start_date":  startDate,
		"end_date":    endDate,
	})
	if err != nil {
		return nil, err
	}
	qOp = r.db.Rebind(qOp)
	var opRow cashSumRow
	if err := r.db.GetContext(ctx, &opRow, qOp, argsOp...); err != nil {
		return nil, err
	}

	qCF, argsCF, err := sqlx.Named(queryCashFlowChanges, map[string]any{
		"merchant_id": merchantID,
		"start_date":  startDate,
		"end_date":    endDate,
	})
	if err != nil {
		return nil, err
	}
	qCF = r.db.Rebind(qCF)
	var cfRows []cfRawRow
	if err := r.db.SelectContext(ctx, &cfRows, qCF, argsCF...); err != nil {
		return nil, err
	}

	qBefore, argsBefore, err := sqlx.Named(queryCashBefore, map[string]any{"merchant_id": merchantID, "date": startDate})
	if err != nil {
		return nil, err
	}
	qBefore = r.db.Rebind(qBefore)
	var beginRow cashSumRow
	if err := r.db.GetContext(ctx, &beginRow, qBefore, argsBefore...); err != nil {
		return nil, err
	}

	qUpTo, argsUpTo, err := sqlx.Named(queryCashUpTo, map[string]any{"merchant_id": merchantID, "date": endDate})
	if err != nil {
		return nil, err
	}
	qUpTo = r.db.Rebind(qUpTo)
	var endRow cashSumRow
	if err := r.db.GetContext(ctx, &endRow, qUpTo, argsUpTo...); err != nil {
		return nil, err
	}

	cf := &report.DirectCashFlowStatement{
		StartDate:     startDate,
		EndDate:       endDate,
		BeginningCash: beginRow.DebitTotal.Sub(beginRow.CreditTotal),
		EndingCash:    endRow.DebitTotal.Sub(endRow.CreditTotal),
	}
	cf.OperatingActivities.CashReceived = opRow.DebitTotal
	cf.OperatingActivities.CashPaid = opRow.CreditTotal
	cf.OperatingActivities.Total = opRow.DebitTotal.Sub(opRow.CreditTotal)

	for _, row := range cfRows {
		amount := row.PeriodCredit.Sub(row.PeriodDebit)
		item := report.CashFlowItem{AccountId: row.AccountId, Name: row.Name, IsSummary: row.IsSummary, Amount: amount}
		switch row.CashFlowCategory {
		case "INVESTING":
			cf.InvestingActivities.Items = append(cf.InvestingActivities.Items, item)
			if !row.IsSummary {
				cf.InvestingActivities.Total = cf.InvestingActivities.Total.Add(amount)
			}
		case "FINANCING":
			cf.FinancingActivities.Items = append(cf.FinancingActivities.Items, item)
			if !row.IsSummary {
				cf.FinancingActivities.Total = cf.FinancingActivities.Total.Add(amount)
			}
		}
	}
	cf.NetChange = cf.OperatingActivities.Total.Add(cf.InvestingActivities.Total).Add(cf.FinancingActivities.Total)

	return cf, nil
}

// eqRawRow scans equity account begin/period debit-credit totals.
type eqRawRow struct {
	AccountId    string          `db:"account_id"`
	Name         string          `db:"name"`
	IsSummary    bool            `db:"is_summary"`
	ParentId     *string         `db:"parent_id"`
	BeginDebit   decimal.Decimal `db:"begin_debit"`
	BeginCredit  decimal.Decimal `db:"begin_credit"`
	PeriodDebit  decimal.Decimal `db:"period_debit"`
	PeriodCredit decimal.Decimal `db:"period_credit"`
}

// queryEquityAccounts fetches begin and period debit/credit for all active EQUITY accounts.
// Leaf accounts join journal_entries directly; summary accounts aggregate through account_closure.
// args: merchant_id, start_date, end_date
const queryEquityAccounts = `
WITH period_entries AS (
    SELECT je.account_id,
           COALESCE(SUM(CASE WHEN t.txn_date <  :start_date                               THEN je.debit  ELSE 0 END), 0) AS begin_debit,
           COALESCE(SUM(CASE WHEN t.txn_date <  :start_date                               THEN je.credit ELSE 0 END), 0) AS begin_credit,
           COALESCE(SUM(CASE WHEN t.txn_date >= :start_date AND t.txn_date <= :end_date   THEN je.debit  ELSE 0 END), 0) AS period_debit,
           COALESCE(SUM(CASE WHEN t.txn_date >= :start_date AND t.txn_date <= :end_date   THEN je.credit ELSE 0 END), 0) AS period_credit
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE' AND t.merchant_id = :merchant_id
    WHERE je.merchant_id = :merchant_id
    GROUP BY je.account_id
),
leaf_eq AS (
    SELECT a.account_id, a.name, 0 AS is_summary, a.parent_id,
           COALESCE(pe.begin_debit,   0) AS begin_debit,
           COALESCE(pe.begin_credit,  0) AS begin_credit,
           COALESCE(pe.period_debit,  0) AS period_debit,
           COALESCE(pe.period_credit, 0) AS period_credit
    FROM accounts a
    LEFT JOIN period_entries pe ON pe.account_id = a.account_id
    WHERE a.merchant_id = :merchant_id AND a.type = 'EQUITY' AND a.is_summary = 0 AND a.is_active = 1
),
summary_eq AS (
    SELECT a.account_id, a.name, 1 AS is_summary, a.parent_id,
           COALESCE(SUM(pe.begin_debit),   0) AS begin_debit,
           COALESCE(SUM(pe.begin_credit),  0) AS begin_credit,
           COALESCE(SUM(pe.period_debit),  0) AS period_debit,
           COALESCE(SUM(pe.period_credit), 0) AS period_credit
    FROM accounts a
    JOIN account_closure ac   ON ac.ancestor_id = a.account_id AND ac.merchant_id = a.merchant_id AND ac.depth > 0
    JOIN accounts leaf         ON leaf.account_id = ac.descendant_id AND leaf.merchant_id = a.merchant_id
                               AND leaf.is_summary = 0 AND leaf.is_active = 1
    LEFT JOIN period_entries pe ON pe.account_id = ac.descendant_id
    WHERE a.merchant_id = :merchant_id AND a.type = 'EQUITY' AND a.is_summary = 1 AND a.is_active = 1
    GROUP BY a.account_id, a.name, a.parent_id
)
SELECT account_id, name, is_summary, parent_id, begin_debit, begin_credit, period_debit, period_credit FROM (
    SELECT account_id, name, is_summary, parent_id, begin_debit, begin_credit, period_debit, period_credit FROM leaf_eq
    UNION ALL
    SELECT account_id, name, is_summary, parent_id, begin_debit, begin_credit, period_debit, period_credit FROM summary_eq
) ORDER BY account_id`

func (r *sqlcReportRepo) GetEquityStatement(ctx context.Context, startDate, endDate string) (*report.EquityStatement, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	is, err := r.GetIncomeStatement(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	q, args, err := sqlx.Named(queryEquityAccounts, map[string]any{
		"merchant_id": merchantID,
		"start_date":  startDate,
		"end_date":    endDate,
	})
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)
	var rows []eqRawRow
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, err
	}

	stmt := &report.EquityStatement{
		StartDate: startDate,
		EndDate:   endDate,
		Items:     make([]report.EquityItem, 0, len(rows)+1),
	}

	for _, row := range rows {
		beginBalance := row.BeginCredit.Sub(row.BeginDebit)
		periodChange := row.PeriodCredit.Sub(row.PeriodDebit)
		stmt.Items = append(stmt.Items, report.EquityItem{
			AccountId:    row.AccountId,
			Name:         row.Name,
			IsSummary:    row.IsSummary,
			IsVirtual:    false,
			BeginBalance: beginBalance,
			PeriodChange: periodChange,
			EndBalance:   beginBalance.Add(periodChange),
		})
		if row.ParentId == nil {
			stmt.TotalBeginBalance = stmt.TotalBeginBalance.Add(beginBalance)
			stmt.TotalPeriodChange = stmt.TotalPeriodChange.Add(periodChange)
		}
	}

	stmt.NetIncome = is.NetIncome
	stmt.Items = append(stmt.Items, report.EquityItem{
		AccountId:    "",
		Name:         "本期淨利",
		IsSummary:    false,
		IsVirtual:    true,
		BeginBalance: decimal.Zero,
		PeriodChange: is.NetIncome,
		EndBalance:   is.NetIncome,
	})
	stmt.TotalPeriodChange = stmt.TotalPeriodChange.Add(is.NetIncome)
	stmt.TotalEndBalance = stmt.TotalBeginBalance.Add(stmt.TotalPeriodChange)

	return stmt, nil
}
