package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
)

// ─────────────────────────────────────────
// InstallmentProjectionService
// ─────────────────────────────────────────

type InstallmentProjectionService struct{}

func (s *InstallmentProjectionService) Name() string { return enums.ProjectionTypeInstallment.String() }

func (s *InstallmentProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventInstallmentCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventInstallmentPeriodPaid:
		return s.applyPeriodPaid(ctx, tx, ct)
	}
	return nil
}

func (s *InstallmentProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.InstallmentCreatedPayload](ct)
	if err != nil {
		return err
	}

	st, err := checkAndGetState[state.InstallmentCreatedState](ct)
	if err != nil {
		return err
	}

	st.Installment.InstallmentId, err = tx.Projection.InstallmentRepo.InsertInstallment(ctx, st.Installment)
	if err != nil {
		return err
	}
	for _, r := range st.InstallmentPayments {
		r.InstallmentId = st.Installment.InstallmentId
		if _, err := tx.Projection.InstallmentRepo.InsertInstallmentPayment(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

func (s *InstallmentProjectionService) applyPeriodPaid(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.InstallmentPeriodPaidPayload](ct)
	if err != nil {
		return err
	}

	st, err := checkAndGetState[state.InstallmentPeriodPaidState](ct)
	if err != nil {
		return err
	}

	err = tx.Projection.InstallmentRepo.PaidInstallmentPayment(ctx, st.InstallmentPayments.PaymentId, p.PaidDate)
	if err != nil {
		return err
	}

	if st.InstallmentPayments.Period == st.Installment.PaidPeriods {
		err := tx.Projection.InstallmentRepo.UpdateInstallmentStatus(ctx, st.Installment.InstallmentId, enums.InstallmentStatusCompleted.Enum())
		if err != nil {
			return err
		}
	}

	return nil
}
