package report

import "github.com/shopspring/decimal"

type EquityItem struct {
	AccountId    string          `json:"account_id"`
	Name         string          `json:"name"`
	IsSummary    bool            `json:"is_summary"`
	IsVirtual    bool            `json:"is_virtual"`
	BeginBalance decimal.Decimal `json:"begin_balance"`
	PeriodChange decimal.Decimal `json:"period_change"`
	EndBalance   decimal.Decimal `json:"end_balance"`
}

type EquityStatement struct {
	StartDate         string          `json:"start_date"`
	EndDate           string          `json:"end_date"`
	Items             []EquityItem    `json:"items"`
	NetIncome         decimal.Decimal `json:"net_income"`
	TotalBeginBalance decimal.Decimal `json:"total_begin_balance"`
	TotalPeriodChange decimal.Decimal `json:"total_period_change"`
	TotalEndBalance   decimal.Decimal `json:"total_end_balance"`
}
