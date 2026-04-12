package state

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"

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
	Investment   projection.Investment               `json:"investment"`
	Movement     projection.InvestmentMovement       `json:"movement"`
	LotDisposals []projection.InvestmentLotDisposals `json:"lot_disposals"` // FIFO 批次更新明細
	Position     projection.InvestmentPosition       `json:"position"`      // AVG 使用
	Ledger       projection.LedgerAccount            `json:"ledger"`
	CostBasis    decimal.Decimal                     `json:"cost_basis"`
	RealizedGain decimal.Decimal                     `json:"realized_gain"`
	NetProceeds  decimal.Decimal                     `json:"net_proceeds"`
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

type UnrealizedMarkedState struct {
	TotalQty          decimal.Decimal `json:"total_qty"`
	MarketValueTWD    decimal.Decimal `json:"market_value_twd"`
	CostTWD           decimal.Decimal `json:"cost_twd"`
	UnrealizedGainTWD decimal.Decimal `json:"unrealized_gain_twd"`
}
