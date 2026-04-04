package state

import (
	"akatengu/internal/model/db/projection"

	"github.com/shopspring/decimal"
)

type InvestmentBoughtState struct {
	Investment projection.Investment         `json:"investment"`
	Movement   projection.InvestmentMovement `json:"movement"`
	Lot        projection.InvestmentLot      `json:"lot"`
	Position   projection.InvestmentPosition `json:"position"`
	Ledger     projection.LedgerAccount      `json:"ledger"`
}

type InvestmentSoldState struct {
	CostBasisTWD    decimal.Decimal `json:"cost_basis_twd"` // 由 service 計算後填入
	RealizedGainTWD decimal.Decimal `json:"realized_gain_twd"`
	LotUpdates      []LotUpdate     `json:"lot_updates"` // FIFO 批次更新明細
}

type LotUpdate struct {
	LotId        int64           `json:"lot_id"`
	RemainingQty decimal.Decimal `json:"remaining_qty"`
	Status       string          `json:"status"`
}

type DividendReceivedState struct {
	AmountTWD decimal.Decimal `json:"amount_twd"`
}

type StockSplitState struct {
	LotUpdates []SplitLotUpdate `json:"lot_updates"`
}

type SplitLotUpdate struct {
	LotId          int64           `json:"lot_id"`
	OldQty         decimal.Decimal `json:"old_qty"`
	NewQty         decimal.Decimal `json:"new_qty"`
	OldUnitCostTWD decimal.Decimal `json:"old_unit_cost_twd"`
	NewUnitCostTWD decimal.Decimal `json:"new_unit_cost_twd"`
}

type FxBoughtState struct {
	AmountTWD decimal.Decimal `json:"amount_twd"`
}

type FxSoldState struct {
	CostBasisTWD decimal.Decimal `json:"cost_basis_twd"`
	FxGainTWD    decimal.Decimal `json:"fx_gain_twd"`
	LotUpdates   []LotUpdate     `json:"lot_updates"`
}

type UnrealizedMarkedState struct {
	TotalQty          decimal.Decimal `json:"total_qty"`
	MarketValueTWD    decimal.Decimal `json:"market_value_twd"`
	CostTWD           decimal.Decimal `json:"cost_twd"`
	UnrealizedGainTWD decimal.Decimal `json:"unrealized_gain_twd"`
}
