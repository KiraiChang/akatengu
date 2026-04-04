package factory

import (
	"akatengu/internal/model/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ------------------------------
// EventTransactionCreated
// ------------------------------
type eventTransactionCreatedProjector struct {
	query *query.Repo
}

func (e *eventTransactionCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.TransactionCreatedPayload]) error {
	p := ct.Payload

	var errs []string

	if p.TransactionDate == "" {
		errs = append(errs, "txn_date is required")
	} else if _, err := time.Parse("2006-01-02", p.TransactionDate); err != nil {
		errs = append(errs, "txn_date must be YYYY-MM-DD")
	}
	if p.Description == "" {
		errs = append(errs, "description is required")
	}
	if !p.TotalAmount.IsPositive() {
		errs = append(errs, "total_amount must be > 0")
	}
	if len(p.Entries) < 2 {
		errs = append(errs, "entries must have at least 2 items")
	}

	// 借貸必須平衡
	var totalDebit, totalCredit decimal.Decimal
	for i, e := range p.Entries {
		if e.AccountId == "" {
			errs = append(errs, fmt.Sprintf("entries[%d].account_id is required", i))
		}
		totalDebit = totalDebit.Add(e.Debit)
		totalCredit = totalCredit.Add(e.Credit)
	}
	if !totalDebit.Sub(totalCredit).IsZero() {
		errs = append(errs, fmt.Sprintf(
			"entries not balanced: debit=%.2f credit=%.2f", totalDebit, totalCredit,
		))
	}
	if len(errs) > 0 {
		return joinErrors(errs)
	}

	// 1. 轉換開張日期
	periodStart, err := periodStartDate(p.TransactionDate)
	if err != nil {
		return fmt.Errorf("parse date %s, fail: %w", p.TransactionDate, err)
	}

	// 2. 檢查是否已結帳
	existing, err := e.query.Period.GetByPeriod(ctx, enums.PeriodMonthly, periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if existing.Status.String() != enums.PeriodTypeStatusOpen.String() {
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
	p := ct.Payload

	var errs []string
	if p.OriginalTransactionId <= 0 {
		errs = append(errs, "original_txn_id is required")
	}
	if p.VoidTransactionId <= 0 {
		errs = append(errs, "void_txn_id is required")
	}
	if p.CorrectedTransactionId <= 0 {
		errs = append(errs, "correction_txn_id is required")
	}
	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}
	if len(errs) > 0 {
		return joinErrors(errs)
	}

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
	existing, err := e.query.Period.GetByPeriod(ctx, enums.PeriodMonthly, periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if existing.Status.String() != enums.PeriodTypeStatusOpen.String() {
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
	p := ct.Payload
	var errs []string
	if p.TransactionId <= 0 {
		errs = append(errs, "txn_id is required")
	}
	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}
	if len(errs) > 0 {
		return joinErrors(errs)
	}
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
	existing, err := e.query.Period.GetByPeriod(ctx, enums.PeriodMonthly, periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if existing.Status.String() != enums.PeriodTypeStatusOpen.String() {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}
	return nil
}

func NewEventTransactionVoidedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.TransactionVoidedPayload] {
	return pipelines.NewTypeWithNoState[payload.TransactionVoidedPayload](&eventTransactionVoidedProjector{query})
}
