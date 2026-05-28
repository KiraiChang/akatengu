package state

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"

	"github.com/shopspring/decimal"
)

type InvestmentUpdatedState struct {
	InvestmentId int64
}

type InvestmentBoughtState struct {
	Investment  projection.Investment             `json:"investment"`
	Movement    projection.InvestmentMovement     `json:"movement"`
	Lot         projection.InvestmentLot          `json:"lot"`
	Position    projection.InvestmentPosition     `json:"position"`
	Ledger      projection.LedgerAccount          `json:"ledger"`
	Transaction payload.TransactionCreatedPayload `json:"transaction"`
}

type InvestmentSoldState struct {
	Investment               projection.Investment               `json:"investment"`
	Movement                 projection.InvestmentMovement       `json:"movement"`
	LotDisposals             []projection.InvestmentLotDisposals `json:"lot_disposals"`              // FIFO 批次更新明細
	Position                 projection.InvestmentPosition       `json:"position"`                   // AVG 使用
	Ledger                   projection.LedgerAccount            `json:"ledger"`
	CostBasis                decimal.Decimal                     `json:"cost_basis"`                 // 帳面沖銷金額（FVTPL/FVOCI 為 FV）
	RealizedGain             decimal.Decimal                     `json:"realized_gain"`              // 已實現損益（vs 原始成本）
	NetProceeds              decimal.Decimal                     `json:"net_proceeds"`
	AccumulatedUnrealizedTWD decimal.Decimal                     `json:"accumulated_unrealized_twd"` // 累積未實現（需沖回）
	Transaction              payload.TransactionCreatedPayload   `json:"transaction"`
}

type StockSplitState struct {
	Investment projection.Investment         `json:"investment"`
	Movement   projection.InvestmentMovement `json:"movement"`
}

type DividendReceivedState struct {
	Investment  projection.Investment              `json:"investment"`
	Movement    projection.InvestmentMovement      `json:"movement"`
	Transaction *payload.TransactionCreatedPayload `json:"transaction"`
}

// LotUnrealizedUpdate 每批次未實現損益更新（FIFO MARK 使用）
type LotUnrealizedUpdate struct {
	LotId             int64
	UnrealizedUnitTWD decimal.Decimal
}

type UnrealizedMarkedState struct {
	Investment          projection.Investment              `json:"investment"`
	TotalQty            decimal.Decimal                   `json:"total_qty"`
	CarryingAmountTWD   decimal.Decimal                   `json:"carrying_amount_twd"`  // 前期帳面值
	NewMarketValueTWD   decimal.Decimal                   `json:"new_market_value_twd"`
	AdjustmentTWD       decimal.Decimal                   `json:"adjustment_twd"`       // 本期增量（可正可負）
	NewMarketPriceTWD   decimal.Decimal                   `json:"new_market_price_twd"` // 每單位公允價值
	LotUnrealizedUpdates []LotUnrealizedUpdate             `json:"lot_unrealized_updates"` // FIFO: 各批次未實現更新
	Movement            projection.InvestmentMovement     `json:"movement"`
	Transaction         *payload.TransactionCreatedPayload `json:"transaction"`
}