package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/services/pipelines"
	"context"
)

// ------------------------------
// EventAccountCreated
// ------------------------------

type eventAccountCreatedProjector struct {
}

func (e *eventAccountCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.AccountCreatePayload]) error {
	var errs []string

	p := ct.Payload
	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "account name is required")
	}

	return joinErrors(errs)
}

func NewEventAccountCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.AccountCreatePayload] {
	return pipelines.NewTypeWithNoState[payload.AccountCreatePayload](&eventAccountCreatedProjector{})
}

// ------------------------------
// EventLedgerAccountCreated
// ------------------------------

type eventLedgerAccountCreatedProjector struct {
}

func (e *eventLedgerAccountCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.LedgerAccountCreatePayload]) error {
	var errs []string

	p := ct.Payload
	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "ledger name is required")
	}

	if p.Currency == "" {
		errs = append(errs, "ledger currency is required")
	}

	if p.Institution == "" {
		errs = append(errs, "ledger institution is required")
	}

	return joinErrors(errs)
}

func NewEventLedgerAccountCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.LedgerAccountCreatePayload] {
	return pipelines.NewTypeWithNoState[payload.LedgerAccountCreatePayload](&eventLedgerAccountCreatedProjector{})
}

// ------------------------------
// EventInvestmentCreated
// ------------------------------

type eventInvestmentCreatedProjector struct {
}

func (e *eventInvestmentCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.InvestmentCreatedPayload]) error {
	var errs []string

	p := ct.Payload
	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}
	if p.Symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	return joinErrors(errs)
}

func NewEventInvestmentCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.InvestmentCreatedPayload] {
	return pipelines.NewTypeWithNoState[payload.InvestmentCreatedPayload](&eventInvestmentCreatedProjector{})
}

// ------------------------------
// EventInvestmentUpdated
// ------------------------------

type eventInvestmentUpdatedProjector struct {
}

func (e *eventInvestmentUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.InvestmentUpdatedPayload]) error {
	var errs []string
	p := ct.Payload
	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}
	if p.Symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	return joinErrors(errs)
}

func NewEventInvestmentUpdatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.InvestmentUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.InvestmentUpdatedPayload](&eventInvestmentUpdatedProjector{})
}
