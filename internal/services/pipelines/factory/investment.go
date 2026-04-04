package factory

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// ------------------------------
// EventRateUpdated
// ------------------------------

type eventRateUpdatedProjector struct {
	query *query.Repo
}

func (e eventRateUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.RateUpdatedPayload]) error {
	p := ct.Payload
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
	if len(errs) > 0 {
		return joinErrors(errs)
	}

	return nil
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
	p := ct.Payload
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

	if len(errs) > 0 {
		return joinErrors(errs)
	}

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
		MovementType: enums.MovementTypeBuy,
		MovementDate: p.Date,
		Quantity:     p.Quantity,
		UnitPrice:    p.UnitPrice,
		UnitPriceTWD: p.UnitPrice.Mul(p.ExchangeRate),
		ExchangeRate: p.ExchangeRate,
		FeeTWD:       p.Fee,
		TaxTWD:       p.Tax,
	}
	if inv.CostMethod.String() == enums.CostMethodAvg.String() {
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
			Status:       enums.LotStatusOpen,
		}
	}

	return nil
}

func NewEventInvestmentBoughtPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InvestmentBoughtState, payload.InvestmentBoughtPayload] {
	return pipelines.NewType[state.InvestmentBoughtState, payload.InvestmentBoughtPayload](&eventInvestmentBoughtProjector{query}, func() *state.InvestmentBoughtState {
		return &state.InvestmentBoughtState{}
	})
}
