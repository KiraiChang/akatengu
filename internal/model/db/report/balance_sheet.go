package report

import "github.com/shopspring/decimal"

type BalanceSheetRow struct {
	Type      string          `db:"type"`
	AccountID string          `db:"account_id"`
	Name      string          `db:"name"`
	Balance   decimal.Decimal `db:"balance"`
}

type BalanceSheet struct {
	ReportDate       string
	Assets           []BalanceSheetRow
	Liabilities      []BalanceSheetRow
	Equity           []BalanceSheetRow
	TotalAssets      decimal.Decimal
	TotalLiabilities decimal.Decimal
	TotalEquity      decimal.Decimal
	NetWorth         decimal.Decimal // TotalAssets - TotalLiabilities
}
