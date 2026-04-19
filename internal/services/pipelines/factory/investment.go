package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ------------------------------
// EventRateUpdated
// ------------------------------

type eventRateUpdatedProjector struct {
	query *query.Repo
}

func (e eventRateUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.RateUpdatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventRateUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.RateUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.RateUpdatedPayload](&eventRateUpdatedProjector{query})
}

// ------------------------------
// EventInvestmentBought
// ------------------------------

type eventInvestmentBoughtProjector struct {
	query *query.Repo
}

func (e eventInvestmentBoughtProjector) Project(ctx context.Context, ct *pipelines.Context[state.InvestmentBoughtState, payload.InvestmentBoughtPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	inv, err := e.query.Investment.GetByID(ctx, p.InvestmentId)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %d not found", p.InvestmentId)
	}
	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}

	ct.State.Ledger = *ledger
	ct.State.Investment = *inv
	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: p.InvestmentId,
		MovementType: enums.MovementTypeBuy.Enum(),
		MovementDate: p.Date,
		Quantity:     p.Quantity,
		UnitPrice:    p.UnitPrice,
		UnitPriceTWD: p.UnitPrice.Mul(p.ExchangeRate),
		ExchangeRate: p.ExchangeRate,
		Fee:          p.Fee,
		Tax:          p.Tax,
	}
	if inv.CostMethod.Is(enums.CostMethodAvg) {
		ct.State.Position = projection.InvestmentPosition{
			InvestmentId:  p.InvestmentId,
			TotalQuantity: p.Quantity,
			TotalCost:     p.UnitPrice.Mul(p.Quantity),
		}
	} else {
		ct.State.Lot = projection.InvestmentLot{
			InvestmentId: p.InvestmentId,
			AcquiredDate: p.Date,
			Quantity:     p.Quantity,
			UnitCost:     p.UnitPrice.Mul(p.ExchangeRate),
			TotalCost:    p.UnitPrice.Mul(p.Quantity),
			RemainingQty: p.Quantity,
			Status:       enums.LotStatusOpen.Enum(),
		}
	}

	return nil
}

func NewEventInvestmentBoughtPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InvestmentBoughtState, payload.InvestmentBoughtPayload] {
	return pipelines.NewType[state.InvestmentBoughtState, payload.InvestmentBoughtPayload](&eventInvestmentBoughtProjector{query}, func() *state.InvestmentBoughtState {
		return &state.InvestmentBoughtState{}
	})
}

// ------------------------------
// EventInvestmentSold
// ------------------------------

type eventInvestmentSoldProjector struct {
	query *query.Repo
}

func (e eventInvestmentSoldProjector) Project(ctx context.Context, ct *pipelines.Context[state.InvestmentSoldState, payload.InvestmentSoldPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	inv, err := e.query.Investment.GetByID(ctx, p.InvestmentId)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %d not found", p.InvestmentId)
	}
	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}

	ct.State.Ledger = *ledger
	ct.State.Investment = *inv

	unitPrice := p.UnitPrice.Mul(p.ExchangeRate)
	sellAmount := p.Quantity.Mul(unitPrice)

	// 1. 計算成本和損益
	switch inv.CostMethod.Val() {
	case enums.CostMethodAvg:
		{
			position, err := e.query.Investment.GetPosition(ctx, p.InvestmentId)
			if err != nil {
				return fmt.Errorf("calc avg cost fail: %w", err)
			}
			ct.State.Position = projection.InvestmentPosition{
				InvestmentId:  p.InvestmentId,
				TotalQuantity: p.Quantity,
				TotalCost:     position.AvgCost.Mul(p.Quantity),
			}
		}
	case enums.CostMethodFIFO:
		{
			ct.State.CostBasis, ct.State.LotDisposals, err = e.calcFIFOCostBasis(ctx, p)
			if err != nil {
				return fmt.Errorf("calc fifo cost fail: %w", err)
			}
		}
	}

	ct.State.RealizedGain = sellAmount.Sub(ct.State.CostBasis).Sub(p.Fee).Sub(p.Tax)
	ct.State.NetProceeds = sellAmount.Sub(p.Fee).Sub(p.Tax)

	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: p.InvestmentId,
		MovementType: enums.MovementTypeSell.Enum(),
		MovementDate: p.Date,
		Quantity:     p.Quantity.Neg(),
		UnitPrice:    p.UnitPrice,
		UnitPriceTWD: p.UnitPrice.Mul(p.ExchangeRate),
		ExchangeRate: p.ExchangeRate,
		Fee:          p.Fee,
		Tax:          p.Tax,
		RealizedGain: &ct.State.RealizedGain,
		CostBasis:    &ct.State.CostBasis,
	}

	return nil
}

func (e *eventInvestmentSoldProjector) calcFIFOCostBasis(
	ctx context.Context,
	p payload.InvestmentSoldPayload,
) (decimal.Decimal, []projection.InvestmentLotDisposals, error) {
	lots, err := e.query.Investment.GetOpenLots(ctx, p.InvestmentId)
	if err != nil {
		return decimal.Zero, nil, err
	}

	remaining := p.Quantity
	costBasis := decimal.Zero
	var updates []projection.InvestmentLotDisposals

	for _, lot := range lots {
		if remaining.IsZero() {
			break
		}
		use := decimal.Min(remaining, lot.RemainingQty)
		costBasis = costBasis.Add(use.Mul(lot.UnitCost))
		remaining = remaining.Sub(use)
		day, err := e.holdingDays(lot.AcquiredDate, p.Date)
		if err != nil {
			return decimal.Zero, nil, err
		}
		updates = append(updates, projection.InvestmentLotDisposals{
			LotId:             lot.LotId,
			Quantity:          use,
			CostBasis:         lot.UnitCost.Mul(use),
			SaleProceeds:      p.UnitPrice.Mul(use),
			CapitalGain:       p.UnitPrice.Mul(use).Sub(lot.UnitCost.Mul(use)),
			DisposalDate:      p.Date,
			HoldingPeriodDays: day,
		})
	}

	if remaining.IsPositive() {
		return decimal.Zero, nil, fmt.Errorf("insufficient inventory: need %.4f more", remaining)
	}
	return costBasis, updates, nil
}

func (e *eventInvestmentSoldProjector) holdingDays(buyStr, sellStr string) (int, error) {
	layout := "2006-01-02"

	buy, err := time.Parse(layout, buyStr)
	if err != nil {
		return 0, err
	}

	sell, err := time.Parse(layout, sellStr)
	if err != nil {
		return 0, err
	}

	return int(sell.Sub(buy).Hours() / 24), nil
}

func NewEventInvestmentSoldPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InvestmentSoldState, payload.InvestmentSoldPayload] {
	return pipelines.NewType[state.InvestmentSoldState, payload.InvestmentSoldPayload](&eventInvestmentSoldProjector{query}, func() *state.InvestmentSoldState {
		return &state.InvestmentSoldState{}
	})
}

// ------------------------------
// EventStockSplit
// ------------------------------

type eventStockSplitProjector struct {
	query *query.Repo
}

func (e eventStockSplitProjector) Project(ctx context.Context, ct *pipelines.Context[state.StockSplitState, payload.StockSplitPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	inv, err := e.query.Investment.GetByID(ctx, p.InvestmentId)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %d not found", p.InvestmentId)
	}
	ct.State.Investment = *inv
	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: p.InvestmentId,
		MovementType: enums.MovementTypeSplit.Enum(),
		MovementDate: p.Date,
		SplitRatio:   &p.Ratio,
	}

	return nil
}

func NewEventStockSplitPipeline(query *query.Repo) *pipelines.TypedPipeline[state.StockSplitState, payload.StockSplitPayload] {
	return pipelines.NewType[state.StockSplitState, payload.StockSplitPayload](&eventStockSplitProjector{query}, func() *state.StockSplitState {
		return &state.StockSplitState{}
	})
}

// ------------------------------
// EventDividendReceived
// ------------------------------

type eventDividendReceivedProjector struct {
	query *query.Repo
}

func (e eventDividendReceivedProjector) Project(ctx context.Context, ct *pipelines.Context[state.DividendReceivedState, payload.DividendReceivedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	inv, err := e.query.Investment.GetByID(ctx, p.InvestmentId)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %d not found", p.InvestmentId)
	}

	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}

	ct.State.Investment = *inv
	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: p.InvestmentId,
		MovementType: enums.MovementTypeDividend.Enum(),
		MovementDate: p.Date,
		ExchangeRate: p.ExchangeRate,
		SplitRatio:   &p.Ratio,
	}
	amountTWD := decimal.Zero
	netAmountTWD := decimal.Zero
	if p.Amount.GreaterThan(decimal.Zero) {

		amountTWD = p.Amount.Mul(p.ExchangeRate)
		netAmountTWD = amountTWD.Sub(p.WithholdingTax)
		ct.State.Movement.GrossAmount = &amountTWD
		entries := []payload.TransactionEntryPayload{{AccountId: "4210", Debit: decimal.Zero, Credit: amountTWD}}
		if p.WithholdingTax.GreaterThan(decimal.Zero) {
			ct.State.Movement.WithholdingTax = &p.WithholdingTax
			entries = append(entries, payload.TransactionEntryPayload{AccountId: "5920", Debit: p.WithholdingTax, Credit: decimal.Zero})
		}
		ct.State.Movement.NetAmount = &netAmountTWD
		entries = append(entries, payload.TransactionEntryPayload{AccountId: ledger.AccountId, LedgerId: &p.LedgerId, Debit: netAmountTWD, Credit: decimal.Zero})
		ct.State.Transaction = &payload.TransactionCreatedPayload{
			TransactionDate: p.Date,
			Description:     fmt.Sprintf("%s 配息", inv.Name),
			TotalAmount:     amountTWD,
			Currency:        "TWD",
			Entries:         entries,
		}
	}

	return nil
}

func NewEventDividendReceivedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.DividendReceivedState, payload.DividendReceivedPayload] {
	return pipelines.NewType[state.DividendReceivedState, payload.DividendReceivedPayload](&eventDividendReceivedProjector{query}, func() *state.DividendReceivedState {
		return &state.DividendReceivedState{}
	})
}
