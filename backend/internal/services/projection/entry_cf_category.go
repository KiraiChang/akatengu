package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"strconv"
)

// ─────────────────────────────────────────
// CashFlowCategoryProjection
// ─────────────────────────────────────────

type CashFlowCategoryProjection struct{}

func (s *CashFlowCategoryProjection) Name() string {
	return enums.ProjectionTypeCashFlowCategory.String()
}

func (s *CashFlowCategoryProjection) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventTransactionCreated:
		return s.applyCreated(ctx, tx, ct)

	case event_types.EventInvestmentBought:
		st, err := checkAndGetState[state.InvestmentBoughtState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventInvestmentSold:
		st, err := checkAndGetState[state.InvestmentSoldState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventDividendReceived:
		st, err := checkAndGetState[state.DividendReceivedState](ct)
		if err != nil {
			return err
		}
		if st.Transaction != nil {
			return s.applyTxnEntries(ctx, tx, ct.MerchantID, *st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)
		}

	case event_types.EventInstallmentCreated:
		// Installment creation is non-cash (asset vs liability commitment); CF is classified at period-paid events.
		return nil

	case event_types.EventInstallmentPeriodPaid:
		st, err := checkAndGetState[state.InstallmentPeriodPaidState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventPrepaidCreated:
		st, err := checkAndGetState[state.PrepaidCreatedState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventPrepaidAmortized:
		st, err := checkAndGetState[state.PrepaidAmortizedState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventPrepaidDisposed:
		st, err := checkAndGetState[state.PrepaidDisposedState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventPrepaidCreatedWithInstallment:
		st, err := checkAndGetState[state.PrepaidCreatedWithInstallmentState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventAssetPurchased:
		st, err := checkAndGetState[state.AssetPurchasedState](ct)
		if err != nil {
			return err
		}
		if st.Ledger == nil {
			// LEASE purchase: non-cash (right-of-use asset vs lease liability); no CF classification.
			return nil
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventAssetPurchasedWithInstallment:
		st, err := checkAndGetState[state.AssetPurchasedWithInstallmentState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventAssetDepreciated:
		st, err := checkAndGetState[state.AssetDepreciatedState](ct)
		if err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy)

	case event_types.EventAssetDisposed:
		st, err := checkAndGetState[state.AssetDisposedState](ct)
		if err != nil {
			return err
		}
		// All non-cash entries in a disposal are INVESTING regardless of their account's default CF category.
		// (AccumDepr account is OPERATING for depreciation but INVESTING when cleared in a disposal.)
		return s.applyTxnEntriesOverrideAll(ctx, tx, ct.MerchantID, st.Transaction, ct.Event.EventUuid, "", ct.UpdatedBy, enums.CashFlowCategoryInvesting.Enum())

	case event_types.EventPeriodAnnualClosed:
		st, err := checkAndGetState[state.PeriodAnnualClosedState](ct)
		if err != nil {
			return err
		}
		if err := s.applyTxnEntries(ctx, tx, ct.MerchantID, st.ClosedTxn, ct.Event.EventUuid, "closing", ct.UpdatedBy); err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.OpenedTxn, ct.Event.EventUuid, "opening", ct.UpdatedBy)

	case event_types.EventPeriodAnnualReopened:
		st, err := checkAndGetState[state.PeriodAnnualReopenedState](ct)
		if err != nil {
			return err
		}
		if err := s.applyTxnEntries(ctx, tx, ct.MerchantID, st.ReverseClosedTxn, ct.Event.EventUuid, "reverse_closing", ct.UpdatedBy); err != nil {
			return err
		}
		return s.applyTxnEntries(ctx, tx, ct.MerchantID, st.ReverseOpenedTxn, ct.Event.EventUuid, "reverse_opening", ct.UpdatedBy)

	case event_types.EventTransactionCFCategoryUpdated:
		return s.applyUserUpdate(ctx, tx, ct)
	}
	return nil
}

func (s *CashFlowCategoryProjection) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCreatedPayload](ct)
	if err != nil {
		return err
	}
	return s.applyTxnEntries(ctx, tx, ct.MerchantID, *p, ct.Event.EventUuid, "", ct.UpdatedBy)
}

func (s *CashFlowCategoryProjection) applyTxnEntries(
	ctx context.Context,
	tx event_store.EventStoreRepositories,
	merchantID int64,
	p payload.TransactionCreatedPayload,
	eventUUID string,
	txnQualifier string,
	updatedBy string,
) error {
	var updBy *string
	if updatedBy != "" {
		updBy = &updatedBy
	}

	for i, e := range p.Entries {
		entryUUID := uuidx.NewFromEvent(eventUUID, "entry:"+txnQualifier+":"+strconv.Itoa(i))

		var cfCategory enums.CashFlowCategory

		acctCF, err := tx.Projection.AccountRepo.GetAccountCFCategory(ctx, e.AccountId, merchantID)
		if err != nil {
			return err
		}
		cfCategory = acctCF

		// only write if we have a valid entry CF category (OPERATING/INVESTING/FINANCING)
		// CASH is an account-level marker identifying cash accounts, not a valid entry category
		if !cfCategory.In(enums.CashFlowCategoryOperating, enums.CashFlowCategoryInvesting, enums.CashFlowCategoryFinancing) {
			continue
		}

		if err := tx.Projection.EntryCFCategoryRepo.UpsertEntryCFCategory(ctx, sqlcdb.UpsertEntryCFCategoryParams{
			EntryUuid:   entryUUID,
			MerchantID:  merchantID,
			CfCategory:  cfCategory,
			IsConfirmed: false,
			UpdatedBy:   updBy,
		}); err != nil {
			return err
		}
	}
	return nil
}

// applyTxnEntriesOverrideAll tags every entry whose account has a classifiable CF category (OPERATING/INVESTING/FINANCING)
// with overrideCategory, ignoring the account's own CF setting. CASH and NULL accounts are skipped as usual.
func (s *CashFlowCategoryProjection) applyTxnEntriesOverrideAll(
	ctx context.Context,
	tx event_store.EventStoreRepositories,
	merchantID int64,
	p payload.TransactionCreatedPayload,
	eventUUID string,
	txnQualifier string,
	updatedBy string,
	overrideCategory enums.CashFlowCategory,
) error {
	var updBy *string
	if updatedBy != "" {
		updBy = &updatedBy
	}

	for i, e := range p.Entries {
		entryUUID := uuidx.NewFromEvent(eventUUID, "entry:"+txnQualifier+":"+strconv.Itoa(i))

		acctCF, err := tx.Projection.AccountRepo.GetAccountCFCategory(ctx, e.AccountId, merchantID)
		if err != nil {
			return err
		}
		if !acctCF.In(enums.CashFlowCategoryOperating, enums.CashFlowCategoryInvesting, enums.CashFlowCategoryFinancing) {
			continue
		}

		if err := tx.Projection.EntryCFCategoryRepo.UpsertEntryCFCategory(ctx, sqlcdb.UpsertEntryCFCategoryParams{
			EntryUuid:   entryUUID,
			MerchantID:  merchantID,
			CfCategory:  overrideCategory,
			IsConfirmed: false,
			UpdatedBy:   updBy,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *CashFlowCategoryProjection) applyUserUpdate(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCFCategoryUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updBy := toUpdatedBy(ct.UpdatedBy)
	for _, item := range p.Entries {
		if err := tx.Projection.EntryCFCategoryRepo.UpsertEntryCFCategory(ctx, sqlcdb.UpsertEntryCFCategoryParams{
			EntryUuid:   item.EntryUUID,
			MerchantID:  ct.MerchantID,
			CfCategory:  item.CFCategory,
			IsConfirmed: true,
			UpdatedBy:   updBy,
		}); err != nil {
			return err
		}
	}
	return nil
}
