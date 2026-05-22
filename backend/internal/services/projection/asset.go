package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"

	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// FixedAssetProjectionService
// ─────────────────────────────────────────

type FixedAssetProjectionService struct{}

func (s *FixedAssetProjectionService) Name() string { return enums.ProjectionTypeFixedAsset.String() }

func (s *FixedAssetProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventAssetPurchased:
		return s.applyPurchased(ctx, tx, ct)
	case event_types.EventAssetDepreciated:
		return s.applyDepreciated(ctx, tx, ct)
	case event_types.EventAssetDisposed:
		return s.applyDisposed(ctx, tx, ct)
	}
	return nil
}

func (s *FixedAssetProjectionService) applyPurchased(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AssetPurchasedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.AssetPurchasedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	assetID, err := tx.Projection.FixedAssetRepo.InsertFixedAsset(ctx, sqlcdb.InsertFixedAssetParams{
		MerchantID:                   ct.MerchantID,
		Name:                         p.Name,
		AssetAccountID:               p.AssetAccountID,
		AccumDepreciationAccountID:   p.AccumDepreciationAccountID,
		DepreciationExpenseAccountID: p.DepreciationExpenseAccountID,
		Cost:                         p.Cost,
		ResidualValue:                p.ResidualValue,
		UsefulLifeMonths:             p.UsefulLifeMonths,
		DepreciationMethod:           enums.DepreciationMethodStraightLine.Enum(),
		PaymentType:                  p.PaymentType,
		PurchaseDate:                 p.PurchaseDate,
		UpdatedBy:                    updatedBy,
	})
	if err != nil {
		return err
	}
	st.AssetID = assetID
	return nil
}

func (s *FixedAssetProjectionService) applyDepreciated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	if _, err := checkAndGetPayload[payload.AssetDepreciatedPayload](ct); err != nil {
		return err
	}
	st, err := checkAndGetState[state.AssetDepreciatedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	deprAmount := thisDepreciationAmount(st.Asset.Cost, st.Asset.ResidualValue, st.Asset.UsefulLifeMonths, st.Asset.DepreciatedPeriods, st.Asset.TotalDepreciated)

	if err := tx.Projection.FixedAssetRepo.UpdateFixedAssetDepreciation(ctx, st.Asset.ID, ct.MerchantID, deprAmount, updatedBy); err != nil {
		return err
	}

	return nil
}

func (s *FixedAssetProjectionService) applyDisposed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AssetDisposedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.AssetDisposedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.FixedAssetRepo.UpdateFixedAssetDisposed(ctx, st.Asset.ID, ct.MerchantID, p.DisposalDate, updatedBy); err != nil {
		return err
	}

	return nil
}

// thisDepreciationAmount calculates this period's straight-line depreciation.
// Last period gets the remainder to avoid decimal drift.
func thisDepreciationAmount(cost, residualValue decimal.Decimal, usefulLifeMonths, depreciatedPeriods int64, totalDepreciated decimal.Decimal) decimal.Decimal {
	depreciableAmount := cost.Sub(residualValue)
	remaining := usefulLifeMonths - depreciatedPeriods
	if remaining <= 1 {
		return depreciableAmount.Sub(totalDepreciated)
	}
	base := depreciableAmount.Div(decimal.NewFromInt(usefulLifeMonths)).Truncate(6)
	return base
}
