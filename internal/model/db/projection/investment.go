package projection

import (
	"akatengu/internal/model/enums"

	"github.com/shopspring/decimal"
)

// Investment 投資項目內容
type Investment struct {
	InvestmentId int64            `db:"investment_id"`
	AccountId    string           `db:"account_id"`
	AssetType    enums.AssetType  `db:"asset_type"`
	Currency     string           `db:"currency"`
	Symbol       string           `db:"symbol"`
	Name         string           `db:"name"`
	CostMethod   enums.CostMethod `db:"cost_method"`
	IsActive     bool             `db:"is_active"`
	Version      int64            `db:"version"`
}

// InvestmentLot 投資庫存表
type InvestmentLot struct {
	LotId         int64           `db:"lot_id"`
	InvestmentId  int64           `db:"investment_id"`
	AcquiredDate  string          `db:"acquired_date"`
	TransactionId int64           `db:"txn_id"`
	Quantity      decimal.Decimal `db:"quantity"`
	UnitCost      decimal.Decimal `db:"unit_cost"`
	UnitCostTWD   decimal.Decimal `db:"unit_cost_twd"`
	RemainingQty  decimal.Decimal `db:"remaining_qty"`
	Status        enums.LotStatus `db:"status"`
}

// InvestmentMovement 投資異動表
type InvestmentMovement struct {
	MovementId      int64              `db:"movement_id"`
	InvestmentId    int64              `db:"investment_id"`
	TransactionId   int64              `db:"txn_id"`
	MovementType    enums.MovementType `db:"movement_type"`
	MovementDate    string             `db:"movement_date"`
	Quantity        decimal.Decimal    `db:"quantity"`
	UnitPrice       decimal.Decimal    `db:"unit_price"`
	UnitPriceTWD    decimal.Decimal    `db:"unit_price_twd"`
	ExchangeRate    decimal.Decimal    `db:"exchange_rate"`
	FeeTWD          decimal.Decimal    `db:"fee_twd"`
	TaxTWD          decimal.Decimal    `db:"tax_twd"`
	RealizedGainTWD *decimal.Decimal   `db:"realized_gain_twd"`
	CostBasisTWD    *decimal.Decimal   `db:"cost_basis_twd"`
}

// InvestmentSummary 庫存摘要（從 v_investment_summary）
type InvestmentSummary struct {
	InvestmentId int64            `db:"investment_id"`
	Symbol       string           `db:"symbol"`
	Name         string           `db:"name"`
	AssetType    enums.AssetType  `db:"asset_type"`
	Currency     string           `db:"currency"`
	CostMethod   enums.CostMethod `db:"cost_method"`
	TotalQty     decimal.Decimal  `db:"total_qty"`
	AvgCostTWD   decimal.Decimal  `db:"avg_cost_twd"`
	TotalCostTWD decimal.Decimal  `db:"total_cost_twd"`
}

type ExchangeRate struct {
	RateId   int64   `db:"rate_id"`
	Currency string  `db:"currency"`
	RateDate string  `db:"rate_date"`
	RateTWD  float64 `db:"rate_twd"`
	Source   string  `db:"source"`
}
