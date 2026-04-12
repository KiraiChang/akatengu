package repos

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/report"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

const (
	balanceSheetSQL = `
WITH balances AS (
    SELECT
        a.account_id,
        a.parent_id,
        a.name,
        a.type,
        a.normal_balance,
        COALESCE(SUM(je.debit) - SUM(je.credit), 0) AS raw_balance
    FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id  = je.account_id
    LEFT JOIN transactions t     ON je.txn_id     = t.txn_id
        AND t.status   = 'ACTIVE'
        AND t.txn_date <= :report_date   -- 只算到指定日期
    WHERE a.is_summary = 0               -- 只看明細科目，匯總科目用加總
      AND a.is_active  = 1
    GROUP BY a.account_id
),
normalized AS (
    SELECT
        account_id,
        parent_id,
        name,
        type,
        -- 轉成正數：借方科目用借貸差，貸方科目反過來
        CASE normal_balance
            WHEN 'DEBIT'  THEN  raw_balance
            WHEN 'CREDIT' THEN -raw_balance
        END AS balance
    FROM balances
)
SELECT
    type,
    account_id,
    name,
    balance
FROM normalized
WHERE type IN ('ASSET', 'LIABILITY', 'EQUITY')
ORDER BY
    CASE type
        WHEN 'ASSET'     THEN 1
        WHEN 'LIABILITY' THEN 2
        WHEN 'EQUITY'    THEN 3
    END,
    account_id;
`
	incomeStatementSQL = `
WITH period_balances AS (
    SELECT
        a.account_id,
        a.parent_id,
        a.name,
        a.type,
        a.normal_balance,
        COALESCE(SUM(je.debit) - SUM(je.credit), 0) AS raw_balance
    FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id    = t.txn_id
        AND t.status   = 'ACTIVE'
        AND t.txn_date BETWEEN :start_date AND :end_date  -- 只算期間內
    WHERE a.is_summary = 0
      AND a.is_active  = 1
      AND a.type IN ('INCOME', 'EXPENSE')                 -- 只看損益科目
    GROUP BY a.account_id
),
normalized AS (
    SELECT
        account_id,
        parent_id,
        name,
        type,
        CASE normal_balance
            WHEN 'DEBIT'  THEN  raw_balance
            WHEN 'CREDIT' THEN -raw_balance
        END AS amount
    FROM period_balances
)
SELECT
    type,
    account_id,
    name,
    amount
FROM normalized
ORDER BY
    CASE type
        WHEN 'INCOME'  THEN 1
        WHEN 'EXPENSE' THEN 2
    END,
    account_id;
`
)

type ReportRepo struct {
	db *sqlx.DB
}

func NewReportRepo(db *sqlx.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) GetBalanceSheet(ctx context.Context, reportDate string) (*report.BalanceSheet, error) {
	var rows []report.BalanceSheetRow
	err := r.db.SelectContext(ctx, &rows, balanceSheetSQL, sql.Named("report_date", reportDate))
	if err != nil {
		return nil, err
	}

	bs := &report.BalanceSheet{ReportDate: reportDate}
	for _, row := range rows {
		switch row.Type.Val() {
		case enums.AccountAsset:
			bs.Assets = append(bs.Assets, row)
			bs.TotalAssets = bs.TotalAssets.Add(row.Balance)
		case enums.AccountLiability:
			bs.Liabilities = append(bs.Liabilities, row)
			bs.TotalLiabilities = bs.TotalLiabilities.Add(row.Balance)
		case enums.AccountEquity:
			bs.Equity = append(bs.Equity, row)
			bs.TotalEquity = bs.TotalEquity.Add(row.Balance)
		}
	}
	bs.NetWorth = bs.TotalAssets.Sub(bs.TotalLiabilities)

	return bs, nil
}

func (r *ReportRepo) GetIncomeStatement(ctx context.Context, startDate, endDate string) (*report.IncomeStatement, error) {
	var rows []report.IncomeStatementRow
	err := r.db.SelectContext(ctx, &rows, incomeStatementSQL,
		sql.Named("start_date", startDate),
		sql.Named("end_date", endDate),
	)
	if err != nil {
		return nil, err
	}

	is := &report.IncomeStatement{StartDate: startDate, EndDate: endDate}
	for _, row := range rows {
		switch row.Type.Val() {
		case enums.AccountIncome:
			is.Income = append(is.Income, row)
			is.TotalIncome = is.TotalIncome.Add(row.Amount)
		case enums.AccountExpense:
			is.Expenses = append(is.Expenses, row)
			is.TotalExpenses = is.TotalExpenses.Add(row.Amount)
		}
	}
	is.NetIncome = is.TotalIncome.Sub(is.TotalExpenses)

	return is, nil
}
