package payload

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type InvestmentCreatedPayload struct {
	AccountId  string           `json:"account_id"`
	AssetType  enums.AssetType  `json:"asset_type"`
	Currency   string           `json:"currency"`
	Symbol     string           `json:"symbol"`
	Name       string           `json:"name"`
	CostMethod enums.CostMethod `json:"cost_method"`
	IsActive   bool             `json:"is_active"`
}

type InvestmentUpdatedPayload struct {
	InvestmentCreatedPayload
	InvestmentId int64 `json:"investment_id"`
	Version      int64 `json:"version"`
}

type InvestmentBoughtPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitPrice    decimal.Decimal `json:"unit_price"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	Fee          decimal.Decimal `json:"fee"`
	Tax          decimal.Decimal `json:"tax"`
	LedgerId     int64           `json:"ledger_id"` // 由哪個帳戶扣款
}

type InvestmentSoldPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitPrice    decimal.Decimal `json:"unit_price"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	Fee          decimal.Decimal `json:"fee"`
	Tax          decimal.Decimal `json:"tax"`
	LedgerId     int64           `json:"ledger_id"` // 入到哪個帳戶
}

type DividendReceivedPayload struct {
	InvestmentId   int64           `json:"investment_id"`
	Date           string          `json:"date"`
	Amount         decimal.Decimal `json:"amount"`          // 現金股利
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`   // 匯率
	WithholdingTax decimal.Decimal `json:"withholding_tax"` // 稅金
	Ratio          decimal.Decimal `json:"ratio"`           // 股票股利
	LedgerId       int64           `json:"ledger_id"`       // 入到哪個帳戶
}

type StockSplitPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Ratio        decimal.Decimal `json:"ratio"`
}

type UnrealizedMarkedPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	MarketPrice  decimal.Decimal `json:"market_price"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
}

type RateUpdatedPayload struct {
	Currency string          `json:"currency"`
	Date     string          `json:"date"`
	RateTWD  decimal.Decimal `json:"rate_twd"`
}
