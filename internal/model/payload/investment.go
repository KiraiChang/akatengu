package payload

import (
	"akatengu/internal/model/enums"

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
	// 計算結果，由 service 填入後存進 payload
	UnitPriceTWD decimal.Decimal `json:"unit_price_twd"`
	TotalCostTWD decimal.Decimal `json:"total_cost_twd"`
}

type InvestmentSoldPayload struct {
	InvestmentId    int64           `json:"investment_id"`
	Date            string          `json:"date"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	ExchangeRate    decimal.Decimal `json:"exchange_rate"`
	Fee             decimal.Decimal `json:"fee"`
	Tax             decimal.Decimal `json:"tax"`
	LedgerId        int64           `json:"ledger_id"`      // 入到哪個帳戶
	CostBasisTWD    decimal.Decimal `json:"cost_basis_twd"` // 由 service 計算後填入
	RealizedGainTWD decimal.Decimal `json:"realized_gain_twd"`
	LotUpdates      []LotUpdate     `json:"lot_updates"` // FIFO 批次更新明細
}

type LotUpdate struct {
	LotId        int64           `json:"lot_id"`
	RemainingQty decimal.Decimal `json:"remaining_qty"`
	Status       string          `json:"status"`
}

type DividendReceivedPayload struct {
	InvestmentId   int64           `json:"investment_id"`
	Date           string          `json:"date"`
	Amount         decimal.Decimal `json:"amount"`
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`
	WithholdingTax decimal.Decimal `json:"withholding_tax"`
	LedgerId       int64           `json:"ledger_id"` // 入到哪個帳戶
	AmountTWD      decimal.Decimal `json:"amount_twd"`
}

type StockSplitPayload struct {
	InvestmentId int64            `json:"investment_id"`
	Date         string           `json:"date"`
	Ratio        decimal.Decimal  `json:"ratio"`
	LotUpdates   []SplitLotUpdate `json:"lot_updates"`
}

type SplitLotUpdate struct {
	LotId          int64           `json:"lot_id"`
	OldQty         decimal.Decimal `json:"old_qty"`
	NewQty         decimal.Decimal `json:"new_qty"`
	OldUnitCostTWD decimal.Decimal `json:"old_unit_cost_twd"`
	NewUnitCostTWD decimal.Decimal `json:"new_unit_cost_twd"`
}

type FxBoughtPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Amount       decimal.Decimal `json:"amount"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	Fee          decimal.Decimal `json:"fee"`
	LedgerID     int64           `json:"ledger_id"` // 由哪個帳戶扣款
	AmountTWD    decimal.Decimal `json:"amount_twd"`
}

type FxSoldPayload struct {
	InvestmentId int64           `json:"investment_id"`
	Date         string          `json:"date"`
	Amount       decimal.Decimal `json:"amount"`
	ExchangeRate decimal.Decimal `json:"exchange_rate"`
	Fee          decimal.Decimal `json:"fee"`
	LedgerId     int64           `json:"ledger_id"` // 入哪個帳戶
	CostBasisTWD decimal.Decimal `json:"cost_basis_twd"`
	FxGainTWD    decimal.Decimal `json:"fx_gain_twd"`
	LotUpdates   []LotUpdate     `json:"lot_updates"`
}

type UnrealizedMarkedPayload struct {
	InvestmentId      int64           `json:"investment_id"`
	Date              string          `json:"date"`
	MarketPrice       decimal.Decimal `json:"market_price"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	TotalQty          decimal.Decimal `json:"total_qty"`
	MarketValueTWD    decimal.Decimal `json:"market_value_twd"`
	CostTWD           decimal.Decimal `json:"cost_twd"`
	UnrealizedGainTWD decimal.Decimal `json:"unrealized_gain_twd"`
}

type RateUpdatedPayload struct {
	Currency string          `json:"currency"`
	Date     string          `json:"date"`
	RateTWD  decimal.Decimal `json:"rate_twd"`
}
