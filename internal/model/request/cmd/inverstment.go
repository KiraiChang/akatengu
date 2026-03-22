package cmd

import "github.com/shopspring/decimal"

type BuyCmd struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitPrice    decimal.Decimal `json:"unit_price"`    // 原幣單價
	ExchangeRate decimal.Decimal `json:"exchange_rate"` // 原幣:TWD，台幣計價資產填 1
	Fee          decimal.Decimal `json:"fee"`           // 手續費（台幣）
	Tax          decimal.Decimal `json:"tax"`           // 交易稅（台幣）
	LedgerId     int64           `json:"ledger_id"`     // 從哪個帳戶扣款
}

type SellCmd struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitPrice    decimal.Decimal `json:"unit_price"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	Fee          decimal.Decimal `json:"fee"`
	Tax          decimal.Decimal `json:"tax"`
	LedgerId     int64           `json:"ledger_id"` // 款項入哪個帳戶
}

type DividendCmd struct {
	InvestmentId   int64           `json:"investment_id"`
	Date           string          `json:"date"`
	Amount         decimal.Decimal `json:"amount"` // 原幣金額
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`
	WithholdingTax decimal.Decimal `json:"withholding_tax"` // 扣繳稅額（台幣）
	LedgerId       int64           `json:"ledger_id"`
}

type SplitCmd struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Ratio        decimal.Decimal `json:"ratio"` // 2:1 填 2.0
}

type FxBuyCmd struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Amount       decimal.Decimal `json:"amount"`        // 買入外幣金額
	ExchangeRate decimal.Decimal `json:"exchange_rate"` // 買入匯率
	Fee          decimal.Decimal `json:"fee"`
	LedgerId     int64           `json:"ledger_id"` // 台幣帳戶
}

type FxSellCmd struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Amount       decimal.Decimal `json:"amount"`        // 賣出外幣金額
	ExchangeRate decimal.Decimal `json:"exchange_rate"` // 賣出匯率
	Fee          decimal.Decimal `json:"fee"`
	LedgerId     int64           `json:"ledger_id"`
}

type MarkUnrealizedCmd struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	MarketPrice  decimal.Decimal `json:"market_price"` // 市價（原幣）
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
}
