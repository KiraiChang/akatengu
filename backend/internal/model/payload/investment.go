package payload

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type InvestmentCreatedPayload struct {
	AccountId    string             `json:"account_id"`
	AssetType    enums.AssetType    `json:"asset_type"`
	Currency     string             `json:"currency"`
	Symbol       string             `json:"symbol"`
	Name         string             `json:"name"`
	CostMethod   enums.CostMethod   `json:"cost_method"`
	IFRSCategory enums.IFRSCategory `json:"ifrs_category"`
	IsActive     bool               `json:"is_active"`
}

func (p InvestmentCreatedPayload) Validate() error {
	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}
	if p.Symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	return joinErrors(errs)
}

type InvestmentUpdatedPayload struct {
	InvestmentCreatedPayload
	InvestmentId int64 `json:"investment_id"`
	Version      int64 `json:"version"`
}

func (p InvestmentUpdatedPayload) Validate() error {
	var errs []string

	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}
	if p.Symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	return joinErrors(errs)
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

func (p InvestmentBoughtPayload) Validate() error {
	var errs []string
	if p.Date == "" {
		errs = append(errs, "date is required")
	}
	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}

	if p.LedgerId == 0 {
		errs = append(errs, "ledger_id is required")
	}

	if p.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "exchange rate is required")
	}

	if p.Quantity.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "quantity is required")
	}

	if p.UnitPrice.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "unit_price is required")
	}

	return joinErrors(errs)
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

func (p InvestmentSoldPayload) Validate() error {
	var errs []string
	if p.Date == "" {
		errs = append(errs, "date is required")
	}
	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}

	if p.LedgerId == 0 {
		errs = append(errs, "ledger_id is required")
	}

	if p.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "exchange rate is required")
	}

	if p.Quantity.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "quantity is required")
	}

	if p.UnitPrice.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "unit_price is required")
	}

	return joinErrors(errs)
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

func (p DividendReceivedPayload) Validate() error {
	var errs []string
	if p.Date == "" {
		errs = append(errs, "date is required")
	}
	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}
	if p.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "exchange_rate is required")
	}

	return joinErrors(errs)
}

type StockSplitPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Ratio        decimal.Decimal `json:"ratio"`
}

func (p StockSplitPayload) Validate() error {
	var errs []string
	if p.Date == "" {
		errs = append(errs, "date is required")
	}
	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}

	return joinErrors(errs)
}

type UnrealizedMarkedPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	MarketPrice  decimal.Decimal `json:"market_price"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
}

func (p UnrealizedMarkedPayload) Validate() error {
	var errs []string
	if p.Date == "" {
		errs = append(errs, "date is required")
	}
	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}
	if p.MarketPrice.IsNegative() {
		errs = append(errs, "market_price must be non-negative")
	}
	if p.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "exchange_rate is required")
	}
	return joinErrors(errs)
}

type RateUpdatedPayload struct {
	Currency string          `json:"currency"`
	Date     string          `json:"date"`
	RateTWD  decimal.Decimal `json:"rate_twd"`
}

func (p RateUpdatedPayload) Validate() error {
	var errs []string
	if p.Date == "" {
		errs = append(errs, "date is required")
	}
	if p.Currency == "" {
		errs = append(errs, "currency is required")
	}

	if p.RateTWD.IsZero() {
		errs = append(errs, "rate_twd is required")
	}
	return joinErrors(errs)
}
