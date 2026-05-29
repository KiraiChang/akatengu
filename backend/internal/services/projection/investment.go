package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

type InvestmentProjectionService struct{}

func (s *InvestmentProjectionService) Name() string { return enums.ProjectionTypeInvestment.String() }

func (s *InvestmentProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	// Account
	case event_types.EventInvestmentCreated:
		return s.applyInvestmentCreated(ctx, tx, ct)
	case event_types.EventInvestmentUpdated:
		return s.applyInvestmentUpdate(ctx, tx, ct)

	// Transaction
	case event_types.EventRateUpdated:
		return s.applyRateUpdated(ctx, tx, ct)
	case event_types.EventInvestmentBought:
		return s.applyInvestmentBought(ctx, tx, ct)
	case event_types.EventInvestmentSold:
		return s.applyInvestmentSold(ctx, tx, ct)
	case event_types.EventStockSplit:
		return s.applyStockSplit(ctx, tx, ct)
	case event_types.EventDividendReceived:
		return s.applyDividendReceived(ctx, tx, ct)
	case event_types.EventUnrealizedMarked:
		return s.applyUnrealizedMarked(ctx, tx, ct)
	}
	return nil
}

// ─────────────────────────────────────────
// Account
// ─────────────────────────────────────────

func (s *InvestmentProjectionService) applyInvestmentCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.InvestmentCreatedPayload](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	ifrsCategory := p.IFRSCategory
	if ifrsCategory.IsZero() {
		ifrsCategory = enums.IFRSCategoryFVTPL.Enum()
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.InvestmentRepo.CreateInvestment(ctx, projection.Investment{
		MerchantID:   ct.MerchantID,
		AccountId:    p.AccountId,
		AssetType:    p.AssetType,
		Currency:     coalesce(p.Currency, "TWD"),
		Symbol:       p.Symbol,
		Name:         p.Name,
		CostMethod:   p.CostMethod,
		IFRSCategory: ifrsCategory,
		IsActive:     p.IsActive,
		UpdatedBy:    updatedBy,
		Uuid:         ct.Event.EventUuid,
	}); err != nil {
		return err
	}

	return nil
}

func (s *InvestmentProjectionService) applyInvestmentUpdate(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.InvestmentUpdatedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentUpdatedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.InvestmentRepo.UpdateInvestment(ctx, projection.Investment{
		MerchantID:   ct.MerchantID,
		InvestmentId: st.InvestmentId,
		AccountId:    p.AccountId,
		AssetType:    p.AssetType,
		Currency:     coalesce(p.Currency, "TWD"),
		Symbol:       p.Symbol,
		Name:         p.Name,
		CostMethod:   p.CostMethod,
		IsActive:     p.IsActive,
		Version:      p.Version,
		UpdatedBy:    updatedBy,
	}); err != nil {
		return err
	}

	return nil
}

// ─────────────────────────────────────────
// Transaction
// ─────────────────────────────────────────

func (s *InvestmentProjectionService) applyRateUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.RateUpdatedPayload](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.InvestmentRepo.UpsertExchangeRate(ctx, projection.ExchangeRate{
		RateDate: p.Date,
		Currency: coalesce(p.Currency, "TWD"),
		RateTWD:  p.RateTWD,
		Source:   enums.RateSourceManual.Enum(),
	}); err != nil {
		return err
	}

	return nil
}

func (s *InvestmentProjectionService) applyInvestmentBought(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.InvestmentBoughtPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentBoughtState](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	st.Movement.MerchantID = ct.MerchantID
	st.Movement.EventId = ct.Event.EventId
	st.Movement.UpdatedBy = updatedBy
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}
	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		st.Position.MerchantID = ct.MerchantID
		st.Position.UpdatedBy = updatedBy
		err = tx.Projection.InvestmentRepo.UpsertPosition(ctx, st.Position)
	} else {
		st.Lot.MerchantID = ct.MerchantID
		st.Lot.MovementId = st.Movement.MovementId
		st.Lot.UpdatedBy = updatedBy
		st.Lot.LotId, err = tx.Projection.InvestmentRepo.InsertLot(ctx, st.Lot)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *InvestmentProjectionService) applyInvestmentSold(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.InvestmentSoldPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentSoldState](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	st.Movement.MerchantID = ct.MerchantID
	st.Movement.EventId = ct.Event.EventId
	st.Movement.UpdatedBy = updatedBy
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		st.Position.UpdatedBy = updatedBy
		err = tx.Projection.InvestmentRepo.UpdateInvestmentPositionSold(ctx, st.Position)
		if err != nil {
			return err
		}
	} else {
		for index, lots := range st.LotDisposals {
			lots.MerchantID = ct.MerchantID
			lots.MovementId = st.Movement.MovementId
			st.LotDisposals[index].Id, err = tx.Projection.InvestmentRepo.InsertLotDisposals(ctx, lots)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *InvestmentProjectionService) applyStockSplit(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.StockSplitPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.StockSplitState](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	st.Movement.MerchantID = ct.MerchantID
	st.Movement.EventId = ct.Event.EventId
	st.Movement.UpdatedBy = updatedBy
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		err = tx.Projection.InvestmentRepo.PositionSplit(ctx, st.Investment.InvestmentId, p.Ratio, updatedBy)
		if err != nil {
			return err
		}
	} else {
		err := tx.Projection.InvestmentRepo.LotSplit(ctx, st.Investment.InvestmentId, p.Ratio, updatedBy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InvestmentProjectionService) applyDividendReceived(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.DividendReceivedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.DividendReceivedState](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	st.Movement.MerchantID = ct.MerchantID
	st.Movement.EventId = ct.Event.EventId
	st.Movement.UpdatedBy = updatedBy
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		err = tx.Projection.InvestmentRepo.PositionSplit(ctx, st.Investment.InvestmentId, p.Ratio, updatedBy)
		if err != nil {
			return err
		}
	} else {
		err := tx.Projection.InvestmentRepo.LotSplit(ctx, st.Investment.InvestmentId, p.Ratio, updatedBy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InvestmentProjectionService) applyUnrealizedMarked(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.UnrealizedMarkedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.UnrealizedMarkedState](ct)
	if err != nil {
		return err
	}

	// 1. 插入 movement（MARK）
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	st.Movement.MerchantID = ct.MerchantID
	st.Movement.EventId = ct.Event.EventId
	st.Movement.UpdatedBy = updatedBy
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return fmt.Errorf("insert mark movement: %w", err)
	}

	// 2. 更新公允價值：AVG 更新 position，FIFO 更新各批次 unrealized_unit_twd
	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		if err := tx.Projection.InvestmentRepo.UpdatePositionFairValue(ctx, st.Investment.InvestmentId, st.NewMarketPriceTWD, updatedBy); err != nil {
			return fmt.Errorf("update position fair value: %w", err)
		}
	} else {
		for _, u := range st.LotUnrealizedUpdates {
			if err := tx.Projection.InvestmentRepo.UpdateLotUnrealizedUnit(ctx, u.LotId, u.UnrealizedUnitTWD, updatedBy); err != nil {
				return fmt.Errorf("update lot unrealized (lot %d): %w", u.LotId, err)
			}
		}
	}

	return nil
}
