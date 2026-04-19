package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
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

	// 1. 轉換開張日期
	periodStart, err := periodStartDate(p.TransactionDate)
	if err != nil {
		return fmt.Errorf("parse date %s, fail: %w", p.TransactionDate, err)
	}

	// 2. 檢查是否已結帳
	existing, err := e.query.Period.GetByPeriod(ctx, enums.PeriodMonthly.Enum(), periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if !existing.Status.Is(enums.PeriodTypeStatusOpen) {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}
	return nil
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

	// 1. 轉換開張日期
	periodStart, err := periodStartDate(origin.TransactionDate)
	if err != nil {
		return fmt.Errorf("parse date %s, fail: %w", origin.TransactionDate, err)
	}

	// 2. 檢查是否已結帳
	existing, err := e.query.Period.GetByPeriod(ctx, enums.PeriodMonthly.Enum(), periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if !existing.Status.Is(enums.PeriodTypeStatusOpen) {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}
	return nil
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

	// 1. 轉換開張日期
	periodStart, err := periodStartDate(origin.TransactionDate)
	if err != nil {
		return fmt.Errorf("parse date %s, fail: %w", origin.TransactionDate, err)
	}

	// 2. 檢查是否已結帳
	existing, err := e.query.Period.GetByPeriod(ctx, enums.PeriodMonthly.Enum(), periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if !existing.Status.Is(enums.PeriodTypeStatusOpen) {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}
	return nil
}

func NewEventTransactionVoidedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.TransactionVoidedPayload] {
	return pipelines.NewTypeWithNoState[payload.TransactionVoidedPayload](&eventTransactionVoidedProjector{query})
}
