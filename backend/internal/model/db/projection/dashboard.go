package projection

import "github.com/shopspring/decimal"

type DashboardSummary struct {
	AsOfDate         string          `json:"as_of_date"`
	Month            string          `json:"month"`
	MonthIncome      decimal.Decimal `json:"month_income"`
	MonthExpense     decimal.Decimal `json:"month_expense"`
	TotalAssets      decimal.Decimal `json:"total_assets"`
	TotalLiabilities decimal.Decimal `json:"total_liabilities"`
	TotalEquity      decimal.Decimal `json:"total_equity"`
	CashBalance      decimal.Decimal `json:"cash_balance"`
}

type MonthlyTrendItem struct {
	Month   string          `json:"month"`
	Income  decimal.Decimal `json:"income"`
	Expense decimal.Decimal `json:"expense"`
	Net     decimal.Decimal `json:"net"`
}

type LedgerBalance struct {
	LedgerID    int64           `json:"ledger_id"`
	Name        string          `json:"name"`
	Institution string          `json:"institution"`
	Type        string          `json:"type"`
	Balance     decimal.Decimal `json:"balance"`
}
