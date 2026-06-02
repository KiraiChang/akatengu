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
)

// ─────────────────────────────────────────
// PrepaidProjectionService
// ─────────────────────────────────────────

type PrepaidProjectionService struct{}

func (s *PrepaidProjectionService) Name() string { return enums.ProjectionTypePrepaid.String() }

func (s *PrepaidProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventPrepaidCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventPrepaidCreatedWithInstallment:
		return s.applyCreatedWithInstallment(ctx, tx, ct)
	case event_types.EventPrepaidAmortized:
		return s.applyAmortized(ctx, tx, ct)
	case event_types.EventPrepaidDisposed:
		return s.applyDisposed(ctx, tx, ct)
	}
	return nil
}

func (s *PrepaidProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidCreatedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidCreatedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	prepaidID, err := tx.Projection.PrepaidRepo.InsertPrepaid(ctx, sqlcdb.InsertPrepaidParams{
		MerchantID:       ct.MerchantID,
		PrepaidUuid:      ct.Event.EventUuid,
		AccountID:        st.Category.AccountID,
		ExpenseAccountID: st.Category.ExpenseAccountID,
		Name:             p.Name,
		TotalAmount:      p.TotalAmount,
		Periods:          p.Periods,
		StartDate:        p.StartDate,
		UpdatedBy:        updatedBy,
	})
	if err != nil {
		return err
	}

	st.PrepaidID = prepaidID
	st.PrepaidUUID = ct.Event.EventUuid
	return nil
}

func (s *PrepaidProjectionService) applyCreatedWithInstallment(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidCreatedWithInstallmentPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidCreatedWithInstallmentState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	prepaidID, err := tx.Projection.PrepaidRepo.InsertPrepaid(ctx, sqlcdb.InsertPrepaidParams{
		MerchantID:       ct.MerchantID,
		PrepaidUuid:      ct.Event.EventUuid,
		AccountID:        st.Category.AccountID,
		ExpenseAccountID: st.Category.ExpenseAccountID,
		Name:             p.Name,
		TotalAmount:      p.TotalAmount,
		Periods:          p.Periods,
		StartDate:        p.StartDate,
		UpdatedBy:        updatedBy,
	})
	if err != nil {
		return err
	}

	st.PrepaidID = prepaidID
	st.PrepaidUUID = ct.Event.EventUuid
	return nil
}

func (s *PrepaidProjectionService) applyAmortized(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidAmortizedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidAmortizedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	amortAmount := payload.AmortizationAmount(st.Prepaid.TotalAmount, st.Prepaid.Periods, st.Prepaid.AmortizedPeriods, st.Prepaid.AmortizedAmount)

	newAmortized := st.Prepaid.AmortizedAmount.Add(amortAmount)
	isLast := st.Prepaid.AmortizedPeriods+1 >= st.Prepaid.Periods
	newStatus := enums.PrepaidStatusActive.Enum()
	if isLast || newAmortized.GreaterThanOrEqual(st.Prepaid.TotalAmount) {
		newStatus = enums.PrepaidStatusCompleted.Enum()
	}

	if err := tx.Projection.PrepaidRepo.UpdatePrepaidAmortization(ctx, st.Prepaid.ID, ct.MerchantID, amortAmount, newStatus, updatedBy); err != nil {
		return err
	}

	_ = p
	return nil
}

func (s *PrepaidProjectionService) applyDisposed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidDisposedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidDisposedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.PrepaidRepo.UpdatePrepaidDisposed(ctx, st.Prepaid.ID, ct.MerchantID, updatedBy); err != nil {
		return err
	}

	_ = p
	return nil
}

