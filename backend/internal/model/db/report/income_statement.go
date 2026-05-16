package report

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type IncomeStatementRow struct {
	Type      enums.AccountType `db:"type"       json:"type"`
	AccountID string            `db:"account_id" json:"account_id"`
	Name      string            `db:"name"       json:"name"`
	Amount    decimal.Decimal   `db:"amount"     json:"amount"`
	ParentId  *string           `db:"parent_id"  json:"parent_id"`
	HasChild  bool              `db:"has_child"  json:"has_child"`
	Depth     int               `db:"depth"      json:"depth"`
}

type IncomeStatement struct {
	StartDate     string               `json:"start_date"`
	EndDate       string               `json:"end_date"`
	Income        []IncomeStatementRow `json:"income"`
	Expenses      []IncomeStatementRow `json:"expenses"`
	TotalIncome   decimal.Decimal      `json:"total_income"`
	TotalExpenses decimal.Decimal      `json:"total_expenses"`
	NetIncome     decimal.Decimal      `json:"net_income"` // TotalIncome - TotalExpenses
}
