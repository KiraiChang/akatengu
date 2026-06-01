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

	"github.com/shopspring/decimal"
)

// ------------------------------
// EventAssetPurchased
// ------------------------------

type eventAssetPurchasedProjector struct {
	query *query.Repo
}

func (e *eventAssetPurchasedProjector) Project(ctx context.Context, ct *pipelines.Context[state.AssetPurchasedState, payload.AssetPurchasedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	if err := validPeriodMonthlyStatus(ctx, p.PurchaseDate, e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	if p.PaymentType == enums.AssetPaymentTypeCash.Enum() {
		ledger, err := e.query.Account.GetLedgerByUuid(ctx, *p.LedgerUUID)
		if err != nil {
			return err
		}
		if ledger == nil {
			return fmt.Errorf("ledger not found")
		}
		c.Ledger = ledger
	}

	txn, err := payload.BuildAssetPurchasedTransaction(p, c.Ledger)
	if err != nil {
		return err
	}
	c.Transaction = txn

	return nil
}

func NewEventAssetPurchasedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.AssetPurchasedState, payload.AssetPurchasedPayload] {
	return pipelines.NewType[state.AssetPurchasedState, payload.AssetPurchasedPayload](&eventAssetPurchasedProjector{query}, func() *state.AssetPurchasedState {
		return &state.AssetPurchasedState{}
	})
}

// ------------------------------
// EventAssetDepreciated
// ------------------------------

type eventAssetDepreciatedProjector struct {
	query *query.Repo
}

func (e *eventAssetDepreciatedProjector) Project(ctx context.Context, ct *pipelines.Context[state.AssetDepreciatedState, payload.AssetDepreciatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	if err := validPeriodMonthlyStatus(ctx, p.PeriodDate+"-01", e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	asset, err := e.query.FixedAsset.GetFixedAssetByUUID(ctx, p.AssetUUID)
	if err != nil {
		return err
	}
	if asset == nil {
		return fmt.Errorf("fixed asset not found")
	}
	if !asset.Status.Is(enums.FixedAssetStatusActive) {
		return fmt.Errorf("fixed asset is not active")
	}
	if asset.DepreciatedPeriods >= asset.UsefulLifeMonths {
		return fmt.Errorf("fixed asset is already fully depreciated")
	}
	c.Asset = asset
	c.Transaction = payload.BuildAssetDepreciatedTransaction(p, c.Asset)

	return nil
}

func NewEventAssetDepreciatedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.AssetDepreciatedState, payload.AssetDepreciatedPayload] {
	return pipelines.NewType[state.AssetDepreciatedState, payload.AssetDepreciatedPayload](&eventAssetDepreciatedProjector{query}, func() *state.AssetDepreciatedState {
		return &state.AssetDepreciatedState{}
	})
}

// ------------------------------
// EventAssetDisposed
// ------------------------------

type eventAssetDisposedProjector struct {
	query *query.Repo
}

func (e *eventAssetDisposedProjector) Project(ctx context.Context, ct *pipelines.Context[state.AssetDisposedState, payload.AssetDisposedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	asset, err := e.query.FixedAsset.GetFixedAssetByUUID(ctx, p.AssetUUID)
	if err != nil {
		return err
	}
	if asset == nil {
		return fmt.Errorf("fixed asset not found")
	}
	if !asset.Status.Is(enums.FixedAssetStatusActive) {
		return fmt.Errorf("fixed asset is not active")
	}
	c.Asset = asset

	if p.Proceeds.GreaterThan(decimal.Zero) && p.ProceedsLedgerUUID != nil {
		ledger, err := e.query.Account.GetLedgerByUuid(ctx, *p.ProceedsLedgerUUID)
		if err != nil {
			return err
		}
		if ledger == nil {
			return fmt.Errorf("proceeds ledger not found")
		}
		c.ProceedsLedger = ledger
	}

	c.Transaction = payload.BuildAssetDisposedTransaction(p, c.Asset, c.ProceedsLedger)

	return nil
}

func NewEventAssetDisposedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.AssetDisposedState, payload.AssetDisposedPayload] {
	return pipelines.NewType[state.AssetDisposedState, payload.AssetDisposedPayload](&eventAssetDisposedProjector{query}, func() *state.AssetDisposedState {
		return &state.AssetDisposedState{}
	})
}

// ------------------------------
// EventAssetPurchasedWithInstallment
// ------------------------------

type eventAssetPurchasedWithInstallmentProjector struct {
	query *query.Repo
}

func (e *eventAssetPurchasedWithInstallmentProjector) Project(ctx context.Context, ct *pipelines.Context[state.AssetPurchasedWithInstallmentState, payload.AssetPurchasedWithInstallmentPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	c := ct.State

	if err := validPeriodMonthlyStatus(ctx, p.PurchaseDate, e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	ledger, err := e.query.Account.GetLedgerByUuid(ctx, p.Installment.LedgerUuid)
	if err != nil {
		return err
	}
	if ledger == nil {
		return fmt.Errorf("ledger not found")
	}

	ip := p.ToInstallmentPayload()
	inst, err := ip.CreateInstallment()
	if err != nil {
		return err
	}
	c.Installment = inst

	payments, err := ip.CreatePayments()
	if err != nil {
		return err
	}
	c.InstallmentPayments = payments

	code, err := getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountAssetPrepaidInterest.Enum())
	if err != nil {
		return err
	}
	c.SysAccountAssetPrepaidInterest = code

	txn, err := payload.BuildAssetPurchasedWithInstallmentTransaction(p, ledger, c.SysAccountAssetPrepaidInterest, c.InstallmentPayments)
	if err != nil {
		return err
	}
	c.Transaction = txn

	return nil
}

func NewEventAssetPurchasedWithInstallmentPipeline(query *query.Repo) *pipelines.TypedPipeline[state.AssetPurchasedWithInstallmentState, payload.AssetPurchasedWithInstallmentPayload] {
	return pipelines.NewType[state.AssetPurchasedWithInstallmentState, payload.AssetPurchasedWithInstallmentPayload](&eventAssetPurchasedWithInstallmentProjector{query}, func() *state.AssetPurchasedWithInstallmentState {
		return &state.AssetPurchasedWithInstallmentState{}
	})
}
