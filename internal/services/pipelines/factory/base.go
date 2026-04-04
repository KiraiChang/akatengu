package factory

import (
	"akatengu/internal/model/enums/event_types"
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
		rules: make(map[string]pipelines.PipelineRunner),
	}
	p.register(query)
	return p
}

type registry struct {
	rules map[string]pipelines.PipelineRunner
}

func (p *registry) Dispatch(ctx context.Context, cmd cmd.AppendCmd) (*pipelines.Result, error) {
	result, ok := p.rules[cmd.EventType.String()]
	if !ok {
		return nil, fmt.Errorf("registry %s not found", cmd.EventType.String())
	}
	return result.Run(ctx, cmd)
}

func (p *registry) register(query *query.Repo) {
	// Accounts
	p.rules[event_types.EventAccountCreated.String()] = NewEventAccountCreatedPipeline()
	p.rules[event_types.EventLedgerAccountCreated.String()] = NewEventLedgerAccountCreatedPipeline()

	// Transaction
	p.rules[event_types.EventTransactionCreated.String()] = NewEventTransactionCreatedPipeline(query)
	p.rules[event_types.EventTransactionVoided.String()] = NewEventTransactionVoidedPipeline(query)
	p.rules[event_types.EventTransactionCorrected.String()] = NewEventTransactionCorrectedPipeline(query)

	p.rules[event_types.EventPeriodMonthStarted.String()] = NewEventPeriodMonthStartedPipeline(query)
	p.rules[event_types.EventPeriodMonthClosed.String()] = NewEventPeriodMonthClosedPipeline(query)
	p.rules[event_types.EventPeriodMonthReopened.String()] = NewEventPeriodMonthReopenedPipeline(query)

	p.rules[event_types.EventPeriodAnnualStarted.String()] = NewEventPeriodAnnualStartedPipeline(query)
	p.rules[event_types.EventPeriodAnnualClosed.String()] = NewEventPeriodAnnualClosedPipeline(query)
	p.rules[event_types.EventPeriodAnnualReopened.String()] = NewEventPeriodAnnualReopenedPipeline(query)

	// Investment(Account Agg)
	p.rules[event_types.EventInvestmentCreated.String()] = NewEventInvestmentCreatedPipeline()
	p.rules[event_types.EventInvestmentUpdated.String()] = NewEventInvestmentUpdatedPipeline()

	// Investment(Transaction Agg)
	p.rules[event_types.EventInvestmentBought.String()] = NewEventInvestmentBoughtPipeline(query)
	p.rules[event_types.EventRateUpdated.String()] = NewEventRateUpdatedPipeline(query)
}
