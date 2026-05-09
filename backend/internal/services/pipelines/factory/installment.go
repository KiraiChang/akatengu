package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/sys_codes"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventInstallmentCreated
// ------------------------------

type eventInstallmentCreatedProjector struct {
	query *query.Repo
}

func (e *eventInstallmentCreatedProjector) Project(ctx context.Context, ct *pipelines.Context[state.InstallmentCreatedState, payload.InstallmentCreatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}
	if ledger == nil {
		return fmt.Errorf("ledger not found")
	}
	c.Ledger = ledger
	inst, err := p.CreateInstallment()
	if err != nil {
		return err
	}
	c.Installment = inst
	payments, err := p.CreatePayments()
	if err != nil {
		return err
	}
	c.InstallmentPayments = payments
	code, err := getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountAssetPrepaidInterest.Enum())
	if err != nil {
		return err
	}

	c.SysAccountAssetPrepaidInterest = code

	return nil
}

func NewEventInstallmentCreatedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InstallmentCreatedState, payload.InstallmentCreatedPayload] {
	return pipelines.NewType[state.InstallmentCreatedState, payload.InstallmentCreatedPayload](&eventInstallmentCreatedProjector{query}, func() *state.InstallmentCreatedState {
		return &state.InstallmentCreatedState{}
	})
}

// ------------------------------
// EventInstallmentPeriodPaid
// ------------------------------

type eventInstallmentPeriodPaidProjector struct {
	query *query.Repo
}

func (e *eventInstallmentPeriodPaidProjector) Project(ctx context.Context, ct *pipelines.Context[state.InstallmentPeriodPaidState, payload.InstallmentPeriodPaidPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	ledger, err := e.query.Account.GetLedger(ctx, p.PaidLedgerId)
	if err != nil {
		return err
	}
	if ledger == nil {
		return fmt.Errorf("ledger not found")
	}
	c.PaidLedger = ledger

	c.Installment, err = e.query.Installment.GetInstallment(ctx, p.InstallmentId)
	if err != nil {
		return err
	}

	if c.Installment == nil {
		return fmt.Errorf("installment not found")
	}

	c.InstallmentPayments, err = e.query.Installment.GetPayment(ctx, p.InstallmentId, p.Period)
	if err != nil {
		return err
	}

	if c.InstallmentPayments == nil {
		return fmt.Errorf("payment not found")
	}

	ledger, err = e.query.Account.GetLedger(ctx, c.Installment.LedgerId)
	if err != nil {
		return err
	}
	if ledger == nil {
		return fmt.Errorf("ledger not found")
	}
	c.Ledger = ledger

	code, err := getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountAssetPrepaidInterest.Enum())
	if err != nil {
		return err
	}
	c.SysAccountAssetPrepaidInterest = code

	switch c.Ledger.Type.Val() {
	case enums.LedgerAccountTypeLoan:
		code, err = getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountExpenseLoanInterestExpense.Enum())
		if err != nil {
			return err
		}
	case enums.LedgerAccountTypeCreditCard:
		code, err = getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountExpenseCreditCardInterestExpense.Enum())
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid ledger type")
	}

	c.SysAccountExpenseInterestExpense = code

	return nil
}

func NewEventInstallmentPeriodPaidPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InstallmentPeriodPaidState, payload.InstallmentPeriodPaidPayload] {
	return pipelines.NewType[state.InstallmentPeriodPaidState, payload.InstallmentPeriodPaidPayload](&eventInstallmentPeriodPaidProjector{query}, func() *state.InstallmentPeriodPaidState {
		return &state.InstallmentPeriodPaidState{}
	})
}
