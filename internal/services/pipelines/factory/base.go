package factory

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
	"strings"
)

// ------------------------------
// Helper
// ------------------------------

func joinErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
}

// ------------------------------
// factory
// ------------------------------

func NewPipelineRegistry(query *query.Repo) pipelines.PipelineRegistry {
	p := &registry{
		rules: make(map[event_types.EventType]pipelines.PipelineRunner),
	}
	p.register(query)
	return p
}

type registry struct {
	rules map[event_types.EventType]pipelines.PipelineRunner
}

func (p *registry) Dispatch(ctx context.Context, cmd cmd.AppendCmd) (*pipelines.Result, error) {
	result, ok := p.rules[cmd.EventType]
	if !ok {
		return nil, fmt.Errorf("registry %s not found", cmd.EventType.String())
	}
	return result.Run(ctx, cmd)
}

func (p *registry) register(query *query.Repo) {
	// Accounts
	p.rules[event_types.EventAccountCreated.Enum()] = NewEventAccountCreatedPipeline()
	p.rules[event_types.EventLedgerAccountCreated.Enum()] = NewEventLedgerAccountCreatedPipeline()

	// Transaction
	p.rules[event_types.EventTransactionCreated.Enum()] = NewEventTransactionCreatedPipeline(query)
	p.rules[event_types.EventTransactionVoided.Enum()] = NewEventTransactionVoidedPipeline(query)
	p.rules[event_types.EventTransactionCorrected.Enum()] = NewEventTransactionCorrectedPipeline(query)

	p.rules[event_types.EventPeriodMonthStarted.Enum()] = NewEventPeriodMonthStartedPipeline(query)
	p.rules[event_types.EventPeriodMonthClosed.Enum()] = NewEventPeriodMonthClosedPipeline(query)
	p.rules[event_types.EventPeriodMonthReopened.Enum()] = NewEventPeriodMonthReopenedPipeline(query)

	p.rules[event_types.EventPeriodAnnualStarted.Enum()] = NewEventPeriodAnnualStartedPipeline(query)
	p.rules[event_types.EventPeriodAnnualClosed.Enum()] = NewEventPeriodAnnualClosedPipeline(query)
	p.rules[event_types.EventPeriodAnnualReopened.Enum()] = NewEventPeriodAnnualReopenedPipeline(query)

	// Investment(Account Agg)
	p.rules[event_types.EventInvestmentCreated.Enum()] = NewEventInvestmentCreatedPipeline()
	p.rules[event_types.EventInvestmentUpdated.Enum()] = NewEventInvestmentUpdatedPipeline()

	// Investment(Transaction Agg)
	p.rules[event_types.EventInvestmentBought.Enum()] = NewEventInvestmentBoughtPipeline(query)
	p.rules[event_types.EventInvestmentSold.Enum()] = NewEventInvestmentSoldPipeline(query)
	p.rules[event_types.EventStockSplit.Enum()] = NewEventStockSplitPipeline(query)
	p.rules[event_types.EventDividendReceived.Enum()] = NewEventDividendReceivedPipeline(query)

	p.rules[event_types.EventRateUpdated.Enum()] = NewEventRateUpdatedPipeline(query)
}
