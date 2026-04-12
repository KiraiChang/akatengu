package projection

import (
	"akatengu/internal/enums"

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
	MovementId    int64           `db:"movement_id"`
	AcquiredDate  string          `db:"acquired_date"`
	TransactionId *int64          `db:"txn_id"`
	Quantity      decimal.Decimal `db:"quantity"`
	UnitCost      decimal.Decimal `db:"unit_cost"`
	TotalCost     decimal.Decimal `db:"total_cost"`
	RemainingQty  decimal.Decimal `db:"remaining_qty"`
	Status        enums.LotStatus `db:"status"`
}

type InvestmentLotDisposals struct {
	Id                int64           `db:"id"`
	LotId             int64           `db:"lot_id"`
	MovementId        int64           `db:"movement_id"`
	Quantity          decimal.Decimal `db:"quantity"`
	CostBasis         decimal.Decimal `db:"cost_basis"`
	SaleProceeds      decimal.Decimal `db:"sale_proceeds"`
	CapitalGain       decimal.Decimal `db:"capital_gain"`
	HoldingPeriodDays int             `db:"holding_period_days"`
	DisposalDate      string          `db:"disposal_date"`
}

type InvestmentPosition struct {
	Id            int64           `db:"id"`
	InvestmentId  int64           `db:"investment_id"`
	TotalQuantity decimal.Decimal `db:"total_quantity"`
	TotalCost     decimal.Decimal `db:"total_cost"`
	AvgCost       decimal.Decimal `db:"avg_cost"`
}

// InvestmentMovement 投資異動表
type InvestmentMovement struct {
	MovementId    int64              `db:"movement_id"`
	InvestmentId  int64              `db:"investment_id"`
	EventId       int64              `db:"event_id"`
	TransactionId *int64             `db:"txn_id"`
	MovementType  enums.MovementType `db:"movement_type"`
	MovementDate  string             `db:"movement_date"`
	Quantity      decimal.Decimal    `db:"quantity"`
	UnitPrice     decimal.Decimal    `db:"unit_price"`
	UnitPriceTWD  decimal.Decimal    `db:"unit_price_twd"`
	ExchangeRate  decimal.Decimal    `db:"exchange_rate"`
	Fee           decimal.Decimal    `db:"fee"`
	Tax           decimal.Decimal    `db:"tax"`
	RealizedGain  *decimal.Decimal   `db:"realized_gain"`
	CostBasis     *decimal.Decimal   `db:"cost_basis"`

	// --- Income (Cash Dividend) ---
	GrossAmount    *decimal.Decimal `db:"gross_amount"`    // 含稅
	NetAmount      *decimal.Decimal `db:"net_amount"`      // 入帳金額
	WithholdingTax *decimal.Decimal `db:"withholding_tax"` // 預扣稅

	// --- Corporate Action ---
	SplitRatio *decimal.Decimal `db:"split_ratio"`
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
	RateId   int64            `db:"rate_id"`
	Currency string           `db:"currency"`
	RateDate string           `db:"rate_date"`
	RateTWD  decimal.Decimal  `db:"rate_twd"`
	Source   enums.RateSource `db:"source"`
}
