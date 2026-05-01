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
)

type InvestmentProjectionService struct{}

func (s *InvestmentProjectionService) Name() string { return string(enums.AggregateAccount) }

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
	if err := tx.Projection.InvestmentRepo.CreateInvestment(ctx, projection.Investment{
		AccountId:  p.AccountId,
		AssetType:  p.AssetType,
		Currency:   coalesce(p.Currency, "TWD"),
		Symbol:     p.Symbol,
		Name:       p.Name,
		CostMethod: p.CostMethod,
		IsActive:   p.IsActive,
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

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.InvestmentRepo.UpdateInvestment(ctx, projection.Investment{
		InvestmentId: p.InvestmentId,
		AccountId:    p.AccountId,
		AssetType:    p.AssetType,
		Currency:     coalesce(p.Currency, "TWD"),
		Symbol:       p.Symbol,
		Name:         p.Name,
		CostMethod:   p.CostMethod,
		IsActive:     p.IsActive,
		Version:      p.Version,
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
	st.Movement.EventId = ct.Event.EventId
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}
	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		err = tx.Projection.InvestmentRepo.UpsertPosition(ctx, st.Position)
	} else {
		st.Lot.MovementId = st.Movement.MovementId
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
	st.Movement.EventId = ct.Event.EventId
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		err = tx.Projection.InvestmentRepo.UpdateInvestmentPositionSold(ctx, st.Position)
		if err != nil {
			return err
		}
	} else {
		for _, lots := range st.LotDisposals {
			lots.MovementId = st.Movement.MovementId
			err := tx.Projection.InvestmentRepo.InsertLotDisposals(ctx, lots)
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
	st.Movement.EventId = ct.Event.EventId
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		err = tx.Projection.InvestmentRepo.PositionSplit(ctx, p.InvestmentId, p.Ratio)
		if err != nil {
			return err
		}
	} else {
		err := tx.Projection.InvestmentRepo.LotSplit(ctx, p.InvestmentId, p.Ratio)
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
	st.Movement.EventId = ct.Event.EventId
	st.Movement.MovementId, err = tx.Projection.InvestmentRepo.InsertMovement(ctx, st.Movement)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodAvg) {
		err = tx.Projection.InvestmentRepo.PositionSplit(ctx, p.InvestmentId, p.Ratio)
		if err != nil {
			return err
		}
	} else {
		err := tx.Projection.InvestmentRepo.LotSplit(ctx, p.InvestmentId, p.Ratio)
		if err != nil {
			return err
		}
	}
	return nil
}
