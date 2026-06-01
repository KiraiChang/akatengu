package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// TransactionProjectionService
// ─────────────────────────────────────────

type TransactionProjectionService struct{}

func (s *TransactionProjectionService) Name() string { return enums.ProjectionTypeTransaction.String() }

func (s *TransactionProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventTransactionCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventTransactionCorrected:
		return s.applyCorrected(ctx, tx, ct)
	case event_types.EventTransactionVoided:
		return s.applyVoided(ctx, tx, ct)
	case event_types.EventPeriodAnnualClosed:
		return s.applyPeriodAnnualClosed(ctx, tx, ct)
	case event_types.EventPeriodAnnualReopened:
		return s.applyPeriodAnnualReopened(ctx, tx, ct)

	case event_types.EventInvestmentBought:
		return s.applyInvestmentBought(ctx, tx, ct)
	case event_types.EventInvestmentSold:
		return s.applyInvestmentSold(ctx, tx, ct)
	case event_types.EventDividendReceived:
		return s.applyDevidendReceived(ctx, tx, ct)

	case event_types.EventInstallmentCreated:
		return s.applyInstallmentCreated(ctx, tx, ct)
	case event_types.EventInstallmentPeriodPaid:
		return s.applyInstallmentPeriodPaid(ctx, tx, ct)

	case event_types.EventPrepaidCreated:
		return s.applyPrepaidCreated(ctx, tx, ct)
	case event_types.EventPrepaidAmortized:
		return s.applyPrepaidAmortized(ctx, tx, ct)
	case event_types.EventPrepaidDisposed:
		return s.applyPrepaidDisposed(ctx, tx, ct)

	case event_types.EventAssetPurchased:
		return s.applyAssetPurchased(ctx, tx, ct)
	case event_types.EventAssetPurchasedWithInstallment:
		return s.applyAssetPurchasedWithInstallment(ctx, tx, ct)
	case event_types.EventAssetDepreciated:
		return s.applyAssetDepreciated(ctx, tx, ct)
	case event_types.EventAssetDisposed:
		return s.applyAssetDisposed(ctx, tx, ct)

	case event_types.EventPrepaidCreatedWithInstallment:
		return s.applyPrepaidCreatedWithInstallment(ctx, tx, ct)
	}
	return nil
}

func (s *TransactionProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCreatedPayload](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if _, err := s.applyTransaction(ctx, tx, *p, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, ""); err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyCorrected(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	//var p payload.TransactionCorrectedPayload
	//if err := json.Unmarshal(event.Payload, &p); err != nil {
	//	return err
	//}
	//
	//if err := tx.Projection.TransactionRepo.UpdateTxnStatus(ctx, p.OriginalTransactionId,
	//	enums.TransactionStatusCorrected,
	//	event.AggregateVersion,
	//); err != nil {
	//	return err
	//}

	return nil
}

func (s *TransactionProjectionService) applyVoided(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	//var p payload.TransactionVoidedPayload
	//if err := json.Unmarshal(event.Payload, &p); err != nil {
	//	return err
	//}
	//
	//if err := tx.Projection.TransactionRepo.UpdateTxnStatus(ctx,
	//	p.TransactionId,
	//	enums.TransactionStatusVoided,
	//	event.AggregateVersion,
	//); err != nil {
	//	return err
	//}

	return nil
}

func (s *TransactionProjectionService) applyPeriodAnnualClosed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodAnnualClosedPayload](ct)
	if err != nil {
		return err
	}

	state, err := checkAndGetState[state.PeriodAnnualClosedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	closingTxnId, err := s.applyTransaction(ctx, tx, state.ClosedTxn, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "closing")
	if err != nil {
		return err
	}

	openingTxnId, err := s.applyTransaction(ctx, tx, state.OpenedTxn, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "opening")
	if err != nil {
		return err
	}

	err = tx.Projection.PeriodCloseRepo.UpdatePeriodCloseTxnId(ctx, p.ClosingId, &closingTxnId, &openingTxnId, updatedBy)
	if err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyPeriodAnnualReopened(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodAnnualReopenedPayload](ct)
	if err != nil {
		return err
	}

	state, err := checkAndGetState[state.PeriodAnnualReopenedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	reverseClosedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseClosedTxn, enums.TransactionStatusVoidRef.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "reverse_closing")
	if err != nil {
		return err
	}

	reverseOpenedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseOpenedTxn, enums.TransactionStatusVoidRef.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "reverse_opening")
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *state.ReverseClosedTxn.RefTxnId, &reverseClosedTxnId, enums.TransactionStatusVoided.Enum(), updatedBy)
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *state.ReverseOpenedTxn.RefTxnId, &reverseOpenedTxnId, enums.TransactionStatusVoided.Enum(), updatedBy)
	if err != nil {
		return err
	}

	err = tx.Projection.PeriodCloseRepo.UpdatePeriodCloseTxnId(ctx, p.ClosingId, nil, nil, updatedBy)
	if err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyInvestmentBought(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.InvestmentBoughtPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentBoughtState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}

	err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId, updatedBy)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodFIFO) {
		err = tx.Projection.InvestmentRepo.UpdateLot(ctx, st.Investment.InvestmentId, txnId, updatedBy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionProjectionService) applyInvestmentSold(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.InvestmentSoldPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentSoldState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}

	err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId, updatedBy)
	if err != nil {
		return err
	}

	if st.LotDisposals != nil {
		for _, lot := range st.LotDisposals {
			err = tx.Projection.InvestmentRepo.UpdateDisposalTxn(ctx, lot.Id, txnId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TransactionProjectionService) applyTransaction(ctx context.Context, tx event_store.EventStoreRepositories, p payload.TransactionCreatedPayload, status enums.TransactionStatus, merchantID int64, updatedBy *string, eventUUID string, txnQualifier string) (int64, error) {
	totalDebit, totalCredit := decimal.Zero, decimal.Zero
	for _, e := range p.Entries {
		totalDebit = totalDebit.Add(e.Debit)
		totalCredit = totalCredit.Add(e.Credit)
	}
	if !(totalDebit.Sub(totalCredit)).IsZero() {
		return 0, fmt.Errorf("debit != credit")
	}

	txnUuid := uuidx.NewFromEvent(eventUUID, "txn:"+txnQualifier)

	txnId, err := tx.Projection.TransactionRepo.InsertTxn(ctx, projection.Transaction{
		MerchantID:      merchantID,
		TxnUuid:         txnUuid,
		TransactionDate: p.TransactionDate,
		Description:     p.Description,
		TotalAmount:     totalDebit,
		Currency:        coalesce(p.Currency, "TWD"),
		Status:          status,
		ReceiptNo:       p.ReceiptNo,
		Note:            p.Note,
		RefTxnId:        p.RefTxnId,
		UpdatedBy:       updatedBy,
	})
	if err != nil {
		return 0, err
	}

	entries := make([]projection.Entry, len(p.Entries))
	for i, e := range p.Entries {
		var cf enums.CashFlowCategory
		if e.CashFlowCategory != nil {
			cf = *e.CashFlowCategory
		}
		entries[i] = projection.Entry{
			MerchantID:       merchantID,
			EntryUuid:        uuidx.NewFromEvent(eventUUID, "entry:"+txnQualifier+":"+strconv.Itoa(i)),
			TxnUuid:          txnUuid,
			TransactionId:    txnId,
			LedgerId:         e.LedgerId,
			AccountId:        e.AccountId,
			Debit:            e.Debit,
			Credit:           e.Credit,
			CashFlowCategory: cf,
			UpdatedBy:        updatedBy,
		}
	}

	if err := tx.Projection.TransactionRepo.UpsertJournalEntries(ctx, entries); err != nil {
		return 0, err
	}
	return txnId, nil
}

func (s *TransactionProjectionService) applyDevidendReceived(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	_, err := checkAndGetPayload[payload.DividendReceivedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.DividendReceivedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if st.Transaction != nil {
		txnId, err := s.applyTransaction(ctx, tx, *st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
		if err != nil {
			return err
		}
		err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId, updatedBy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionProjectionService) applyInstallmentCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.InstallmentCreatedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	return tx.Projection.InstallmentRepo.UpdateInstallmentTxn(ctx, st.Installment.InstallmentId, txnId, updatedBy)
}

func (s *TransactionProjectionService) applyInstallmentPeriodPaid(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.InstallmentPeriodPaidState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	return tx.Projection.InstallmentRepo.UpdatePaymentTxn(ctx, st.InstallmentPayments.PaymentId, txnId, updatedBy)
}

func (s *TransactionProjectionService) applyPrepaidCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.PrepaidCreatedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	return tx.Projection.PrepaidRepo.UpdatePrepaidTxn(ctx, st.PrepaidID, ct.MerchantID, txnId)
}

func (s *TransactionProjectionService) applyPrepaidAmortized(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidAmortizedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidAmortizedState](ct)
	if err != nil {
		return err
	}

	amortAmount := payload.AmortizationAmount(st.Prepaid.TotalAmount, st.Prepaid.Periods, st.Prepaid.AmortizedPeriods, st.Prepaid.AmortizedAmount)
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	return tx.Projection.PrepaidRepo.InsertPrepaidAmortization(ctx, sqlcdb.InsertPrepaidAmortizationParams{
		MerchantID:       ct.MerchantID,
		AmortizationUuid: uuidx.NewFromEvent(ct.Event.EventUuid, "amortization"),
		PrepaidUuid:      st.Prepaid.PrepaidUuid,
		PrepaidID:        st.Prepaid.ID,
		TxnID:            txnId,
		PeriodDate:       p.PeriodDate,
		Amount:           amortAmount,
		UpdatedBy:        updatedBy,
	})
}

func (s *TransactionProjectionService) applyPrepaidDisposed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.PrepaidDisposedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	_, err = s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	return err
}

func (s *TransactionProjectionService) applyAssetPurchased(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.AssetPurchasedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	return tx.Projection.FixedAssetRepo.UpdateFixedAssetTxn(ctx, st.AssetID, ct.MerchantID, txnId)
}

func (s *TransactionProjectionService) applyAssetDepreciated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AssetDepreciatedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.AssetDepreciatedState](ct)
	if err != nil {
		return err
	}

	deprAmount := payload.DepreciationAmount(st.Asset.Cost, st.Asset.ResidualValue, st.Asset.UsefulLifeMonths, st.Asset.DepreciatedPeriods, st.Asset.TotalDepreciated)
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	return tx.Projection.FixedAssetRepo.InsertFixedAssetDepreciation(ctx, sqlcdb.InsertFixedAssetDepreciationParams{
		MerchantID:       ct.MerchantID,
		DepreciationUuid: uuidx.NewFromEvent(ct.Event.EventUuid, "depreciation"),
		AssetUuid:        st.Asset.AssetUuid,
		AssetID:          st.Asset.ID,
		TxnID:            txnId,
		PeriodDate:       p.PeriodDate,
		Amount:           deprAmount,
		UpdatedBy:        updatedBy,
	})
}

func (s *TransactionProjectionService) applyAssetDisposed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.AssetDisposedState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	_, err = s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	return err
}

func (s *TransactionProjectionService) applyAssetPurchasedWithInstallment(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.AssetPurchasedWithInstallmentState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	if err := tx.Projection.FixedAssetRepo.UpdateFixedAssetTxn(ctx, st.AssetID, ct.MerchantID, txnId); err != nil {
		return err
	}
	return tx.Projection.InstallmentRepo.UpdateInstallmentTxn(ctx, st.Installment.InstallmentId, txnId, updatedBy)
}

func (s *TransactionProjectionService) applyPrepaidCreatedWithInstallment(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.PrepaidCreatedWithInstallmentState](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy, ct.Event.EventUuid, "")
	if err != nil {
		return err
	}
	if err := tx.Projection.PrepaidRepo.UpdatePrepaidTxn(ctx, st.PrepaidID, ct.MerchantID, txnId); err != nil {
		return err
	}
	return tx.Projection.InstallmentRepo.UpdateInstallmentTxn(ctx, st.Installment.InstallmentId, txnId, updatedBy)
}
