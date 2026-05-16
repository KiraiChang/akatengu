package report

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type BalanceSheetRow struct {
	Type      enums.AccountType `db:"type"       json:"type"`
	AccountId string            `db:"account_id" json:"account_id"`
	Name      string            `db:"name"       json:"name"`
	Balance   decimal.Decimal   `db:"balance"    json:"balance"`
	ParentId  *string           `db:"parent_id"  json:"parent_id"`
	HasChild  bool              `db:"has_child"  json:"has_child"`
	Depth     int               `db:"depth"      json:"depth"`
}

type BalanceSheet struct {
	ReportDate       string            `json:"report_date"`
	Assets           []BalanceSheetRow `json:"assets"`
	Liabilities      []BalanceSheetRow `json:"liabilities"`
	Equity           []BalanceSheetRow `json:"equity"`
	TotalAssets      decimal.Decimal   `json:"total_assets"`
	TotalLiabilities decimal.Decimal   `json:"total_liabilities"`
	TotalEquity      decimal.Decimal   `json:"total_equity"`
	NetWorth         decimal.Decimal   `json:"net_worth"` // TotalAssets - TotalLiabilities
}
