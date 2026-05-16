package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
)

// ------------------------------
// EventTransactionCreated
// ------------------------------
type eventTransactionCreatedProjector struct {
	query *query.Repo
}

func (e *eventTransactionCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.TransactionCreatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	return validPeriodMonthlyStatus(ctx, p.TransactionDate, e.query, enums.PeriodTypeStatusOpen.Enum())
}

func NewEventTransactionCreatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.TransactionCreatedPayload] {
	return pipelines.NewTypeWithNoState[payload.TransactionCreatedPayload](&eventTransactionCreatedProjector{query})
}

// ------------------------------
// EventTransactionCorrected
// ------------------------------
type eventTransactionCorrectedProjector struct {
	query *query.Repo
}

func (e *eventTransactionCorrectedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.TransactionCorrectedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	origin, err := e.query.Transaction.GetByID(ctx, p.OriginalTransactionId)
	if err != nil {
		return err
	}

	return validPeriodMonthlyStatus(ctx, origin.TransactionDate, e.query, enums.PeriodTypeStatusOpen.Enum())
}

func NewEventTransactionCorrectedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.TransactionCorrectedPayload] {
	return pipelines.NewTypeWithNoState[payload.TransactionCorrectedPayload](&eventTransactionCorrectedProjector{query})
}

// ------------------------------
// EventTransactionVoided
// ------------------------------
type eventTransactionVoidedProjector struct {
	query *query.Repo
}

func (e eventTransactionVoidedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.TransactionVoidedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	origin, err := e.query.Transaction.GetByID(ctx, p.TransactionId)
	if err != nil {
		return err
	}

	return validPeriodMonthlyStatus(ctx, origin.TransactionDate, e.query, enums.PeriodTypeStatusOpen.Enum())
}

func NewEventTransactionVoidedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.TransactionVoidedPayload] {
	return pipelines.NewTypeWithNoState[payload.TransactionVoidedPayload](&eventTransactionVoidedProjector{query})
}
