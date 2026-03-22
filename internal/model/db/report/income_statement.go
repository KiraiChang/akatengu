package report

import "github.com/shopspring/decimal"

type IncomeStatementRow struct {
	Type      string          `db:"type"`
	AccountID string          `db:"account_id"`
	Name      string          `db:"name"`
	Amount    decimal.Decimal `db:"amount"`
}

type IncomeStatement struct {
	StartDate     string
	EndDate       string
	Income        []IncomeStatementRow
	Expenses      []IncomeStatementRow
	TotalIncome   decimal.Decimal
	TotalExpenses decimal.Decimal
	NetIncome     decimal.Decimal // TotalIncome - TotalExpenses
}
