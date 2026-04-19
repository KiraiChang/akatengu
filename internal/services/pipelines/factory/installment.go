package factory

import (
	"akatengu/internal/enums/sys_codes"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos"
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
	sys   repos.SysRepo
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
	code, err := GetSysAccountCode(ctx, e.sys, sys_codes.SysAccountAssetPrepaidInterest.Enum())
	if err != nil {
		return err
	}

	c.SysAccountAssetPrepaidInterest = code

	return nil
}

func NewEventInstallmentCreatedPipeline(query *query.Repo, sys repos.SysRepo) *pipelines.TypedPipeline[state.InstallmentCreatedState, payload.InstallmentCreatedPayload] {
	return pipelines.NewType[state.InstallmentCreatedState, payload.InstallmentCreatedPayload](&eventInstallmentCreatedProjector{query, sys}, func() *state.InstallmentCreatedState {
		return &state.InstallmentCreatedState{}
	})
}

// ------------------------------
// EventInstallmentPeriodPaid
// ------------------------------

type eventInstallmentPeriodPaidProjector struct {
	query *query.Repo
	sys   repos.SysRepo
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

	code, err := GetSysAccountCode(ctx, e.sys, sys_codes.SysAccountAssetPrepaidInterest.Enum())
	if err != nil {
		return err
	}
	c.SysAccountAssetPrepaidInterest = code

	code, err = GetSysAccountCode(ctx, e.sys, sys_codes.SysAccountExpenseInterestExpense.Enum())
	if err != nil {
		return err
	}
	c.SysAccountExpenseInterestExpense = code

	return nil
}

func NewEventInstallmentPeriodPaidPipeline(query *query.Repo, sys repos.SysRepo) *pipelines.TypedPipeline[state.InstallmentPeriodPaidState, payload.InstallmentPeriodPaidPayload] {
	return pipelines.NewType[state.InstallmentPeriodPaidState, payload.InstallmentPeriodPaidPayload](&eventInstallmentPeriodPaidProjector{query, sys}, func() *state.InstallmentPeriodPaidState {
		return &state.InstallmentPeriodPaidState{}
	})
}

func GetSysAccountCode(ctx context.Context, sys repos.SysRepo, enum sys_codes.SysAccount) (string, error) {
	codes, err := sys.GetSysAccount(ctx)
	if err != nil {
		return "", err
	}

	for _, code := range codes {
		if code.SysCode == enum.String() {
			return code.AccountId, nil
		}
	}
	return "", fmt.Errorf("sys account not found")
}
