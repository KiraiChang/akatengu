package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventPrepaidCreated
// ------------------------------

type eventPrepaidCreatedProjector struct {
	query *query.Repo
}

func (e *eventPrepaidCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[state.PrepaidCreatedState, payload.PrepaidCreatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	if err := validPeriodMonthlyStatus(ctx, p.StartDate, e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	ledger, err := e.query.Account.GetLedgerByUuid(ctx, p.LedgerUUID)
	if err != nil {
		return err
	}
	if ledger == nil {
		return fmt.Errorf("ledger not found")
	}
	c.Ledger = ledger
	c.Transaction = payload.BuildPrepaidCreatedTransaction(p, c.Ledger)

	return nil
}

func NewEventPrepaidCreatedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.PrepaidCreatedState, payload.PrepaidCreatedPayload] {
	return pipelines.NewType[state.PrepaidCreatedState, payload.PrepaidCreatedPayload](&eventPrepaidCreatedProjector{query}, func() *state.PrepaidCreatedState {
		return &state.PrepaidCreatedState{}
	})
}

// ------------------------------
// EventPrepaidAmortized
// ------------------------------

type eventPrepaidAmortizedProjector struct {
	query *query.Repo
}

func (e *eventPrepaidAmortizedProjector) Project(ctx context.Context, ct *pipelines.Context[state.PrepaidAmortizedState, payload.PrepaidAmortizedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	if err := validPeriodMonthlyStatus(ctx, p.PeriodDate+"-01", e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	prepaid, err := e.query.Prepaid.GetPrepaidByUUID(ctx, p.PrepaidUUID)
	if err != nil {
		return err
	}
	if prepaid == nil {
		return fmt.Errorf("prepaid not found")
	}
	if !prepaid.Status.Is(enums.PrepaidStatusActive) {
		return fmt.Errorf("prepaid is not active")
	}
	if prepaid.AmortizedPeriods >= prepaid.Periods {
		return fmt.Errorf("prepaid is already fully amortized")
	}
	c.Prepaid = prepaid
	c.Transaction = payload.BuildPrepaidAmortizedTransaction(p, c.Prepaid)

	return nil
}

func NewEventPrepaidAmortizedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.PrepaidAmortizedState, payload.PrepaidAmortizedPayload] {
	return pipelines.NewType[state.PrepaidAmortizedState, payload.PrepaidAmortizedPayload](&eventPrepaidAmortizedProjector{query}, func() *state.PrepaidAmortizedState {
		return &state.PrepaidAmortizedState{}
	})
}

// ------------------------------
// EventPrepaidDisposed
// ------------------------------

type eventPrepaidDisposedProjector struct {
	query *query.Repo
}

func (e *eventPrepaidDisposedProjector) Project(ctx context.Context, ct *pipelines.Context[state.PrepaidDisposedState, payload.PrepaidDisposedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	prepaid, err := e.query.Prepaid.GetPrepaidByUUID(ctx, p.PrepaidUUID)
	if err != nil {
		return err
	}
	if prepaid == nil {
		return fmt.Errorf("prepaid not found")
	}
	if !prepaid.Status.Is(enums.PrepaidStatusActive) {
		return fmt.Errorf("prepaid is not active")
	}
	c.Prepaid = prepaid
	c.Transaction = payload.BuildPrepaidDisposedTransaction(p, c.Prepaid)

	return nil
}

func NewEventPrepaidDisposedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.PrepaidDisposedState, payload.PrepaidDisposedPayload] {
	return pipelines.NewType[state.PrepaidDisposedState, payload.PrepaidDisposedPayload](&eventPrepaidDisposedProjector{query}, func() *state.PrepaidDisposedState {
		return &state.PrepaidDisposedState{}
	})
}
