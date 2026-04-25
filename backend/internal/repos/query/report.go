package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/report"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type ReportRepo interface {
	GetBalanceSheet(ctx context.Context, reportDate string) (*report.BalanceSheet, error)
	GetIncomeStatement(ctx context.Context, startDate, endDate string) (*report.IncomeStatement, error)
}

type sqlcReportRepo struct {
	q *sqlcdb.Queries
}

func newReportRepo(q *sqlcdb.Queries) ReportRepo {
	return &sqlcReportRepo{q: q}
}

func NewReportRepo(db *sqlx.DB) ReportRepo {
	return &sqlcReportRepo{q: sqlcdb.New(db)}
}

func (r *sqlcReportRepo) GetBalanceSheet(ctx context.Context, reportDate string) (*report.BalanceSheet, error) {
	rows, err := r.q.GetBalanceSheetRows(ctx, reportDate)
	if err != nil {
		return nil, err
	}

	bs := &report.BalanceSheet{ReportDate: reportDate}
	for _, row := range rows {
		balance := decimal.NewFromFloat(row.Balance)
		var accType enums.AccountType
		_ = accType.Scan(row.Type)
		bsRow := report.BalanceSheetRow{
			Type:      accType,
			AccountId: row.AccountID,
			Name:      row.Name,
			Balance:   balance,
		}
		switch accType.Val() {
		case enums.AccountAsset:
			bs.Assets = append(bs.Assets, bsRow)
			bs.TotalAssets = bs.TotalAssets.Add(balance)
		case enums.AccountLiability:
			bs.Liabilities = append(bs.Liabilities, bsRow)
			bs.TotalLiabilities = bs.TotalLiabilities.Add(balance)
		case enums.AccountEquity:
			bs.Equity = append(bs.Equity, bsRow)
			bs.TotalEquity = bs.TotalEquity.Add(balance)
		}
	}
	bs.NetWorth = bs.TotalAssets.Sub(bs.TotalLiabilities)

	return bs, nil
}

func (r *sqlcReportRepo) GetIncomeStatement(ctx context.Context, startDate, endDate string) (*report.IncomeStatement, error) {
	rows, err := r.q.GetIncomeStatementRows(ctx, sqlcdb.GetIncomeStatementRowsParams{
		TxnDate:   startDate,
		TxnDate_2: endDate,
	})
	if err != nil {
		return nil, err
	}

	is := &report.IncomeStatement{StartDate: startDate, EndDate: endDate}
	for _, row := range rows {
		amount := decimal.NewFromFloat(row.Amount)
		var accType enums.AccountType
		_ = accType.Scan(row.Type)
		isRow := report.IncomeStatementRow{
			Type:      accType,
			AccountID: row.AccountID,
			Name:      row.Name,
			Amount:    amount,
		}
		switch accType.Val() {
		case enums.AccountIncome:
			is.Income = append(is.Income, isRow)
			is.TotalIncome = is.TotalIncome.Add(amount)
		case enums.AccountExpense:
			is.Expenses = append(is.Expenses, isRow)
			is.TotalExpenses = is.TotalExpenses.Add(amount)
		}
	}
	is.NetIncome = is.TotalIncome.Sub(is.TotalExpenses)

	return is, nil
}