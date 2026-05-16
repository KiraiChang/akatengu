package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
)

type LedgerAccountBalanceSnapshotProjection struct{}

func (s *LedgerAccountBalanceSnapshotProjection) Name() string {
	return enums.ProjectionTypeLedgerAccountBalanceSnapshot.String()
}

func (s *LedgerAccountBalanceSnapshotProjection) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventPeriodMonthClosed:
		p, err := checkAndGetPayload[payload.PeriodMonthClosedPayload](ct)
		if err != nil {
			return err
		}
		return tx.Projection.LedgerAccountBalanceSnapshotRepo.BulkInsert(ctx, ct.MerchantID, p.ClosingId)
	case event_types.EventPeriodAnnualClosed:
		p, err := checkAndGetPayload[payload.PeriodAnnualClosedPayload](ct)
		if err != nil {
			return err
		}
		return tx.Projection.LedgerAccountBalanceSnapshotRepo.BulkInsert(ctx, ct.MerchantID, p.ClosingId)
	case event_types.EventPeriodMonthReopened:
		p, err := checkAndGetPayload[payload.PeriodMonthReopenedPayload](ct)
		if err != nil {
			return err
		}
		return tx.Projection.LedgerAccountBalanceSnapshotRepo.DeleteByClosingId(ctx, ct.MerchantID, p.ClosingId)
	case event_types.EventPeriodAnnualReopened:
		p, err := checkAndGetPayload[payload.PeriodAnnualReopenedPayload](ct)
		if err != nil {
			return err
		}
		return tx.Projection.LedgerAccountBalanceSnapshotRepo.DeleteByClosingId(ctx, ct.MerchantID, p.ClosingId)
	}
	return nil
}
