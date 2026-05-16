package report

import "github.com/shopspring/decimal"

type CashFlowItem struct {
	AccountId string          `json:"account_id"`
	Name      string          `json:"name"`
	IsSummary bool            `json:"is_summary"`
	Amount    decimal.Decimal `json:"amount"`
}

type CashFlowSection struct {
	Items []CashFlowItem  `json:"items"`
	Total decimal.Decimal `json:"total"`
}

type OperatingActivities struct {
	NetIncome   decimal.Decimal `json:"net_income"`
	Adjustments []CashFlowItem  `json:"adjustments"`
	Total       decimal.Decimal `json:"total"`
}

type CashFlowStatement struct {
	StartDate           string              `json:"start_date"`
	EndDate             string              `json:"end_date"`
	OperatingActivities OperatingActivities `json:"operating_activities"`
	InvestingActivities CashFlowSection     `json:"investing_activities"`
	FinancingActivities CashFlowSection     `json:"financing_activities"`
	NetChange           decimal.Decimal     `json:"net_change"`
	BeginningCash       decimal.Decimal     `json:"beginning_cash"`
	EndingCash          decimal.Decimal     `json:"ending_cash"`
}
