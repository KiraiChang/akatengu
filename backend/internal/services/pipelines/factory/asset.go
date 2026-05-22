package factory

import (
	"akatengu/internal/enums"
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
		ledger, err := e.query.Account.GetLedger(ctx, *p.LedgerID)
		if err != nil {
			return err
		}
		if ledger == nil {
			return fmt.Errorf("ledger not found")
		}
		c.Ledger = ledger
	}

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

	asset, err := e.query.FixedAsset.GetFixedAssetByID(ctx, p.AssetID)
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

	asset, err := e.query.FixedAsset.GetFixedAssetByID(ctx, p.AssetID)
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

	if p.Proceeds.GreaterThan(decimal.Zero) && p.ProceedsLedgerID != nil {
		ledger, err := e.query.Account.GetLedger(ctx, *p.ProceedsLedgerID)
		if err != nil {
			return err
		}
		if ledger == nil {
			return fmt.Errorf("proceeds ledger not found")
		}
		c.ProceedsLedger = ledger
	}

	return nil
}

func NewEventAssetDisposedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.AssetDisposedState, payload.AssetDisposedPayload] {
	return pipelines.NewType[state.AssetDisposedState, payload.AssetDisposedPayload](&eventAssetDisposedProjector{query}, func() *state.AssetDisposedState {
		return &state.AssetDisposedState{}
	})
}
