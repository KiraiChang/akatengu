package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

// Investment 投資項目內容
//
//dbmap:sqlcdb=Investment
//dbmap:sqlcdb=CreateInvestmentParams
//dbmap:sqlcdb=UpdateInvestmentParams
//dbmap:sqlcdb=GetInvestmentPagedRow
type Investment struct {
	InvestmentId int64            `db:"investment_id" json:"investment_id"`
	AccountId    string           `db:"account_id" json:"account_id"`
	AssetType    enums.AssetType  `db:"asset_type" json:"asset_type"`
	Currency     string           `db:"currency" json:"currency"`
	Symbol       string           `db:"symbol" json:"symbol"`
	Name         string           `db:"name" json:"name"`
	CostMethod   enums.CostMethod `db:"cost_method" json:"cost_method"`
	IsActive     bool             `db:"is_active" json:"is_active"`
	Version      int64            `db:"version" json:"version"`
}

// InvestmentLot 投資庫存表
//
//dbmap:sqlcdb=InvestmentLot
//dbmap:sqlcdb=InsertInvestmentLotParams
//dbmap:sqlcdb=GetOpenLotsPagedRow
type InvestmentLot struct {
	LotId         int64           `db:"lot_id" json:"lot_id"`
	InvestmentId  int64           `db:"investment_id" json:"investment_id"`
	MovementId    int64           `db:"movement_id" json:"movement_id"`
	AcquiredDate  string          `db:"acquired_date" json:"acquired_date"`
	TransactionId *int64          `db:"txn_id" json:"txn_id"`
	Quantity      decimal.Decimal `db:"quantity" json:"quantity"`
	UnitCost      decimal.Decimal `db:"unit_cost" json:"unit_cost"`
	TotalCost     decimal.Decimal `db:"total_cost" json:"total_cost"`
	RemainingQty  decimal.Decimal `db:"remaining_qty" json:"remaining_qty"`
	Status        enums.LotStatus `db:"status" json:"status"`
}

//dbmap:sqlcdb=InvestmentLotDisposal
//dbmap:sqlcdb=InsertInvestmentLotDisposalParams
//dbmap:sqlcdb=GetOpenLotDisposalsPagedRow
type InvestmentLotDisposals struct {
	Id                int64           `db:"id" json:"id"`
	LotId             int64           `db:"lot_id" json:"lot_id"`
	MovementId        int64           `db:"movement_id" json:"movement_id"`
	Quantity          decimal.Decimal `db:"quantity" json:"quantity"`
	CostBasis         decimal.Decimal `db:"cost_basis" json:"cost_basis"`
	SaleProceeds      decimal.Decimal `db:"sale_proceeds" json:"sale_proceeds"`
	CapitalGain       decimal.Decimal `db:"capital_gain" json:"capital_gain"`
	HoldingPeriodDays int64           `db:"holding_period_days" json:"holding_period_days"`
	DisposalDate      string          `db:"disposal_date" json:"disposal_date"`
}

//dbmap:sqlcdb=InvestmentPosition
//dbmap:sqlcdb=UpsertInvestmentPositionParams
//dbmap:sqlcdb=UpdateInvestmentPositionSoldParams
type InvestmentPosition struct {
	Id            int64           `db:"id" json:"id"`
	InvestmentId  int64           `db:"investment_id" json:"investment_id"`
	TotalQuantity decimal.Decimal `db:"total_quantity" json:"total_quantity"`
	TotalCost     decimal.Decimal `db:"total_cost" json:"total_cost"`
	AvgCost       decimal.Decimal `db:"avg_cost" json:"avg_cost"`
}

// InvestmentMovement 投資異動表
//
//dbmap:sqlcdb=InvestmentMovement
//dbmap:sqlcdb=InsertInvestmentMovementParams
//dbmap:sqlcdb=GetInvestmentMovementsRow
type InvestmentMovement struct {
	MovementId    int64               `db:"movement_id"`
	InvestmentId  int64               `db:"investment_id"`
	EventId       int64               `db:"event_id"`
	TransactionId *int64              `db:"txn_id"`
	MovementType  enums.MovementType  `db:"movement_type"`
	MovementDate  string              `db:"movement_date"`
	Quantity      decimal.Decimal     `db:"quantity"`
	UnitPrice     decimal.Decimal     `db:"unit_price"`
	UnitPriceTWD  decimal.Decimal     `db:"unit_price_twd"`
	ExchangeRate  decimal.Decimal     `db:"exchange_rate"`
	Fee           decimal.Decimal     `db:"fee"`
	Tax           decimal.Decimal     `db:"tax"`
	RealizedGain  decimal.NullDecimal `db:"realized_gain"`
	CostBasis     decimal.NullDecimal `db:"cost_basis"`

	// --- Income (Cash Dividend) ---
	GrossAmount    decimal.NullDecimal `db:"gross_amount"`    // 含稅
	NetAmount      decimal.NullDecimal `db:"net_amount"`      // 入帳金額
	WithholdingTax decimal.NullDecimal `db:"withholding_tax"` // 預扣稅

	// --- Corporate Action ---
	SplitRatio decimal.NullDecimal `db:"split_ratio"`
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

//dbmap:sqlcdb=ExchangeRate
//dbmap:sqlcdb=UpsertExchangeRateParams
type ExchangeRate struct {
	RateId   int64            `db:"rate_id" json:"rate_id"`
	Currency string           `db:"currency" json:"currency"`
	RateDate string           `db:"rate_date" json:"rate_date"`
	RateTWD  decimal.Decimal  `db:"rate_twd" json:"rate_twd"`
	Source   enums.RateSource `db:"source" json:"source"`
}
