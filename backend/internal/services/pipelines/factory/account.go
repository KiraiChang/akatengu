package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventAccountBulkImported
// ------------------------------

type eventAccountBulkImportedProjector struct{}

func (e *eventAccountBulkImportedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.AccountBulkImportedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventAccountBulkImportedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.AccountBulkImportedPayload] {
	return pipelines.NewTypeWithNoState[payload.AccountBulkImportedPayload](&eventAccountBulkImportedProjector{})
}

// ------------------------------
// EventAccountCreated
// ------------------------------

type eventAccountCreatedProjector struct {
}

func (e *eventAccountCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.AccountCreatePayload]) error {
	return ct.Payload.Validate()
}

func NewEventAccountCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.AccountCreatePayload] {
	return pipelines.NewTypeWithNoState[payload.AccountCreatePayload](&eventAccountCreatedProjector{})
}

// ------------------------------
// EventAccountUpdated
// ------------------------------

type eventAccountUpdatedProjector struct {
}

func (e *eventAccountUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.AccountUpdatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventAccountUpdatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.AccountUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.AccountUpdatedPayload](&eventAccountUpdatedProjector{})
}

// ------------------------------
// EventLedgerAccountCreated
// ------------------------------

type eventLedgerAccountCreatedProjector struct {
}

func (e *eventLedgerAccountCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.LedgerAccountCreatePayload]) error {
	return ct.Payload.Validate()
}

func NewEventLedgerAccountCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.LedgerAccountCreatePayload] {
	return pipelines.NewTypeWithNoState[payload.LedgerAccountCreatePayload](&eventLedgerAccountCreatedProjector{})
}

// ------------------------------
// EventLedgerAccountUpdate
// ------------------------------

type eventLedgerAccountUpdatedProjector struct {
}

func (e *eventLedgerAccountUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.LedgerAccountUpdatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventLedgerAccountUpdatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.LedgerAccountUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.LedgerAccountUpdatedPayload](&eventLedgerAccountUpdatedProjector{})
}

// ------------------------------
// EventInvestmentCreated
// ------------------------------

type eventInvestmentCreatedProjector struct {
}

func (e *eventInvestmentCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.InvestmentCreatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventInvestmentCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.InvestmentCreatedPayload] {
	return pipelines.NewTypeWithNoState[payload.InvestmentCreatedPayload](&eventInvestmentCreatedProjector{})
}

// ------------------------------
// EventInvestmentUpdated
// ------------------------------

type eventInvestmentUpdatedProjector struct {
	query *query.Repo
}

func (e *eventInvestmentUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[state.InvestmentUpdatedState, payload.InvestmentUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	inv, err := e.query.Investment.GetByCreationEventUuid(ctx, ct.Payload.InvestmentUUID)
	if err != nil {
		return err
	}
	if inv == nil {
		return fmt.Errorf("investment %s not found", ct.Payload.InvestmentUUID)
	}
	ct.State.InvestmentId = inv.InvestmentId
	return nil
}

func NewEventInvestmentUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InvestmentUpdatedState, payload.InvestmentUpdatedPayload] {
	return pipelines.NewType[state.InvestmentUpdatedState, payload.InvestmentUpdatedPayload](&eventInvestmentUpdatedProjector{query: query}, func() *state.InvestmentUpdatedState {
		return &state.InvestmentUpdatedState{}
	})
}
