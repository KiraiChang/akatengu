package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	dbprojection "akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
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
	case event_types.EventAssetPurchasedWithInstallment:
		return s.applyCreatedFromAssetState(ctx, tx, ct)
	case event_types.EventPrepaidCreatedWithInstallment:
		return s.applyCreatedFromPrepaidState(ctx, tx, ct)
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

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	installmentUuid := ct.Event.EventUuid
	st.Installment.MerchantID = ct.MerchantID
	st.Installment.InstallmentUuid = installmentUuid
	st.Installment.UpdatedBy = updatedBy
	st.Installment.InstallmentId, err = tx.Projection.InstallmentRepo.InsertInstallment(ctx, st.Installment)
	if err != nil {
		return err
	}
	for i, r := range st.InstallmentPayments {
		r.MerchantID = ct.MerchantID
		r.PaymentUuid = uuidx.NewFromEvent(ct.Event.EventUuid, fmt.Sprintf("payment:%d", i))
		r.InstallmentUuid = installmentUuid
		r.InstallmentId = st.Installment.InstallmentId
		r.UpdatedBy = updatedBy
		if _, err := tx.Projection.InstallmentRepo.InsertInstallmentPayment(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

func (s *InstallmentProjectionService) applyCreatedFromAssetState(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.AssetPurchasedWithInstallmentState](ct)
	if err != nil {
		return err
	}
	return s.insertInstallmentAndPayments(ctx, tx, ct, st.Installment, st.InstallmentPayments)
}

func (s *InstallmentProjectionService) applyCreatedFromPrepaidState(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.PrepaidCreatedWithInstallmentState](ct)
	if err != nil {
		return err
	}
	return s.insertInstallmentAndPayments(ctx, tx, ct, st.Installment, st.InstallmentPayments)
}

// insertInstallmentAndPayments 寫入 Installment 主檔與各期明細。
// Installment UUID 由 event UUID 衍生，避免與主體（Asset/Prepaid）UUID 衝突。
func (s *InstallmentProjectionService) insertInstallmentAndPayments(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result, inst *dbprojection.Installment, payments []*dbprojection.InstallmentPayment) error {
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	installmentUuid := uuidx.NewFromEvent(ct.Event.EventUuid, "installment")
	inst.MerchantID = ct.MerchantID
	inst.InstallmentUuid = installmentUuid
	inst.UpdatedBy = updatedBy
	var err error
	inst.InstallmentId, err = tx.Projection.InstallmentRepo.InsertInstallment(ctx, inst)
	if err != nil {
		return err
	}
	for i, r := range payments {
		r.MerchantID = ct.MerchantID
		r.PaymentUuid = uuidx.NewFromEvent(ct.Event.EventUuid, fmt.Sprintf("payment:%d", i))
		r.InstallmentUuid = installmentUuid
		r.InstallmentId = inst.InstallmentId
		r.UpdatedBy = updatedBy
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

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	err = tx.Projection.InstallmentRepo.PaidInstallmentPayment(ctx, st.InstallmentPayments.PaymentId, p.PaidDate, updatedBy)
	if err != nil {
		return err
	}

	if st.InstallmentPayments.Period == st.Installment.PaidPeriods {
		err := tx.Projection.InstallmentRepo.UpdateInstallmentStatus(ctx, st.Installment.InstallmentId, enums.InstallmentStatusCompleted.Enum(), updatedBy)
		if err != nil {
			return err
		}
	}

	return nil
}
