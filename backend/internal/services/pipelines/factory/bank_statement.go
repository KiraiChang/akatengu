package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ─────────────────────────────────────────
// EventBankPdfTemplateCreated
// ─────────────────────────────────────────

type eventBankPdfTemplateCreatedProjector struct{}

func (e *eventBankPdfTemplateCreatedProjector) Project(_ context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankPdfTemplateCreatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventBankPdfTemplateCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.BankPdfTemplateCreatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankPdfTemplateCreatedPayload](&eventBankPdfTemplateCreatedProjector{}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankPdfTemplateUpdated
// ─────────────────────────────────────────

type eventBankPdfTemplateUpdatedProjector struct {
	query *query.Repo
}

func (e *eventBankPdfTemplateUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankPdfTemplateUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	tmpl, err := e.query.BankPdfTemplate.GetBankPdfTemplateByUUID(ctx, ct.Payload.TemplateUUID)
	if err != nil {
		return fmt.Errorf("get bank pdf template: %w", err)
	}
	if tmpl == nil {
		return fmt.Errorf("bank pdf template not found")
	}
	if !tmpl.IsActive {
		return fmt.Errorf("bank pdf template is inactive")
	}
	return nil
}

func NewEventBankPdfTemplateUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankPdfTemplateUpdatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankPdfTemplateUpdatedPayload](&eventBankPdfTemplateUpdatedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankPdfTemplateDeactivated
// ─────────────────────────────────────────

type eventBankPdfTemplateDeactivatedProjector struct {
	query *query.Repo
}

func (e *eventBankPdfTemplateDeactivatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankPdfTemplateDeactivatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	tmpl, err := e.query.BankPdfTemplate.GetBankPdfTemplateByUUID(ctx, ct.Payload.TemplateUUID)
	if err != nil {
		return fmt.Errorf("get bank pdf template: %w", err)
	}
	if tmpl == nil {
		return fmt.Errorf("bank pdf template not found")
	}
	if !tmpl.IsActive {
		return fmt.Errorf("bank pdf template is already inactive")
	}
	return nil
}

func NewEventBankPdfTemplateDeactivatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankPdfTemplateDeactivatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankPdfTemplateDeactivatedPayload](&eventBankPdfTemplateDeactivatedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankCsvTemplateCreated
// ─────────────────────────────────────────

type eventBankCsvTemplateCreatedProjector struct{}

func (e *eventBankCsvTemplateCreatedProjector) Project(_ context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankCsvTemplateCreatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventBankCsvTemplateCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.BankCsvTemplateCreatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankCsvTemplateCreatedPayload](&eventBankCsvTemplateCreatedProjector{}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankCsvTemplateUpdated
// ─────────────────────────────────────────

type eventBankCsvTemplateUpdatedProjector struct {
	query *query.Repo
}

func (e *eventBankCsvTemplateUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankCsvTemplateUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	tmpl, err := e.query.BankCsvTemplate.GetBankCsvTemplateByUUID(ctx, ct.Payload.TemplateUUID)
	if err != nil {
		return fmt.Errorf("get bank csv template: %w", err)
	}
	if tmpl == nil {
		return fmt.Errorf("bank csv template not found")
	}
	if !tmpl.IsActive {
		return fmt.Errorf("bank csv template is inactive")
	}
	return nil
}

func NewEventBankCsvTemplateUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankCsvTemplateUpdatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankCsvTemplateUpdatedPayload](&eventBankCsvTemplateUpdatedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankCsvTemplateDeactivated
// ─────────────────────────────────────────

type eventBankCsvTemplateDeactivatedProjector struct {
	query *query.Repo
}

func (e *eventBankCsvTemplateDeactivatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankCsvTemplateDeactivatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	tmpl, err := e.query.BankCsvTemplate.GetBankCsvTemplateByUUID(ctx, ct.Payload.TemplateUUID)
	if err != nil {
		return fmt.Errorf("get bank csv template: %w", err)
	}
	if tmpl == nil {
		return fmt.Errorf("bank csv template not found")
	}
	if !tmpl.IsActive {
		return fmt.Errorf("bank csv template is already inactive")
	}
	return nil
}

func NewEventBankCsvTemplateDeactivatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankCsvTemplateDeactivatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankCsvTemplateDeactivatedPayload](&eventBankCsvTemplateDeactivatedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankStatementImported
// ─────────────────────────────────────────

type eventBankStatementImportedProjector struct {
	query *query.Repo
}

func (e *eventBankStatementImportedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankStatementImportedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	ledger, err := e.query.Account.GetLedger(ctx, ct.Payload.LedgerID)
	if err != nil {
		return fmt.Errorf("get ledger: %w", err)
	}
	if ledger == nil {
		return fmt.Errorf("ledger account not found")
	}
	if !ledger.IsActive {
		return fmt.Errorf("ledger account is inactive")
	}
	return nil
}

func NewEventBankStatementImportedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankStatementImportedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankStatementImportedPayload](&eventBankStatementImportedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankStatementTxnMatched
// ─────────────────────────────────────────

type eventBankStatementTxnMatchedProjector struct {
	query *query.Repo
}

func (e *eventBankStatementTxnMatchedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankStatementTxnMatchedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	txn, err := e.query.BankStatementImport.GetBankStatementTxnByID(ctx, ct.Payload.BankTxnID)
	if err != nil {
		return fmt.Errorf("get bank statement txn: %w", err)
	}
	if txn == nil {
		return fmt.Errorf("bank statement txn not found")
	}
	return nil
}

func NewEventBankStatementTxnMatchedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankStatementTxnMatchedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankStatementTxnMatchedPayload](&eventBankStatementTxnMatchedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankStatementTxnIgnored
// ─────────────────────────────────────────

type eventBankStatementTxnIgnoredProjector struct {
	query *query.Repo
}

func (e *eventBankStatementTxnIgnoredProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankStatementTxnIgnoredPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	txn, err := e.query.BankStatementImport.GetBankStatementTxnByID(ctx, ct.Payload.BankTxnID)
	if err != nil {
		return fmt.Errorf("get bank statement txn: %w", err)
	}
	if txn == nil {
		return fmt.Errorf("bank statement txn not found")
	}
	return nil
}

func NewEventBankStatementTxnIgnoredPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankStatementTxnIgnoredPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankStatementTxnIgnoredPayload](&eventBankStatementTxnIgnoredProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankStatementAdjustmentApproved
// ─────────────────────────────────────────

type eventBankStatementAdjustmentApprovedProjector struct {
	query *query.Repo
}

func (e *eventBankStatementAdjustmentApprovedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.BankStatementAdjustmentApprovedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventBankStatementAdjustmentApprovedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.BankStatementAdjustmentApprovedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.BankStatementAdjustmentApprovedPayload](&eventBankStatementAdjustmentApprovedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ─────────────────────────────────────────
// EventBankStatementCompleted
// ─────────────────────────────────────────

type eventBankStatementCompletedProjector struct {
	query *query.Repo
}

func (e *eventBankStatementCompletedProjector) Project(ctx context.Context, ct *pipelines.Context[state.BankStatementCompletedState, payload.BankStatementCompletedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	imp, err := e.query.BankStatementImport.GetBankStatementImportByUUID(ctx, ct.Payload.ImportUUID)
	if err != nil {
		return fmt.Errorf("get bank statement import: %w", err)
	}
	if imp == nil {
		return fmt.Errorf("bank statement import not found")
	}
	if imp.Status == "COMPLETED" {
		return fmt.Errorf("bank statement import is already completed")
	}
	ct.State.ImportID = imp.ImportID
	return nil
}

func NewEventBankStatementCompletedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.BankStatementCompletedState, payload.BankStatementCompletedPayload] {
	return pipelines.NewType[state.BankStatementCompletedState, payload.BankStatementCompletedPayload](&eventBankStatementCompletedProjector{query}, func() *state.BankStatementCompletedState {
		return &state.BankStatementCompletedState{}
	})
}
