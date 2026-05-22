package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"

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
	case event_types.EventAssetDepreciated:
		return s.applyAssetDepreciated(ctx, tx, ct)
	case event_types.EventAssetDisposed:
		return s.applyAssetDisposed(ctx, tx, ct)
	}
	return nil
}

func (s *TransactionProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCreatedPayload](ct)
	if err != nil {
		return err
	}

	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if _, err := s.applyTransaction(ctx, tx, *p, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy); err != nil {
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
	closingTxnId, err := s.applyTransaction(ctx, tx, state.ClosedTxn, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}

	openingTxnId, err := s.applyTransaction(ctx, tx, state.OpenedTxn, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
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
	reverseClosedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseClosedTxn, enums.TransactionStatusVoidRef.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}

	reverseOpenedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseOpenedTxn, enums.TransactionStatusVoidRef.Enum(), ct.MerchantID, updatedBy)
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
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
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
	txnId, err := s.applyTransaction(ctx, tx, st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}

	err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId, updatedBy)
	if err != nil {
		return err
	}

	if st.LotDisposals != nil {
		for _, lot := range st.LotDisposals {
			err = tx.Projection.InvestmentRepo.UpdateDisposalTxn(ctx, lot.LotId, txnId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TransactionProjectionService) applyTransaction(ctx context.Context, tx event_store.EventStoreRepositories, p payload.TransactionCreatedPayload, status enums.TransactionStatus, merchantID int64, updatedBy *string) (int64, error) {
	totalDebit, totalCredit := decimal.Zero, decimal.Zero
	for _, e := range p.Entries {
		totalDebit = totalDebit.Add(e.Debit)
		totalCredit = totalCredit.Add(e.Credit)
	}
	if !(totalDebit.Sub(totalCredit)).IsZero() {
		return 0, fmt.Errorf("debit != credit")
	}
	// 業務邏輯：組裝 proj model
	txnId, err := tx.Projection.TransactionRepo.InsertTxn(ctx, projection.Transaction{
		MerchantID:      merchantID,
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
		txnId, err := s.applyTransaction(ctx, tx, *st.Transaction, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
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
	p, err := checkAndGetPayload[payload.InstallmentCreatedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InstallmentCreatedState](ct)
	if err != nil {
		return err
	}

	var entries []payload.TransactionEntryPayload
	switch p.InterestType.Val() {
	case enums.InterestTypeFree:
		entries = []payload.TransactionEntryPayload{
			// 獲得資產，或支付費用
			{AccountId: p.AccountId, Debit: p.Amount, Credit: decimal.Zero},
			// 應付帳款
			{AccountId: st.Ledger.AccountId, LedgerId: &p.LedgerId, Debit: decimal.Zero, Credit: p.Amount},
		}
	case enums.InterestTypeFixedRate:
		interest := decimal.Zero
		for _, p := range st.InstallmentPayments {
			interest = interest.Add(p.Interest)
		}
		entries = []payload.TransactionEntryPayload{
			// 獲得資產，或支付費用
			{AccountId: p.AccountId, Debit: p.Amount, Credit: decimal.Zero},
			// 預付利息
			{AccountId: st.SysAccountAssetPrepaidInterest, Debit: p.Amount, Credit: decimal.Zero},
			// 應付帳款
			{AccountId: st.Ledger.AccountId, LedgerId: &p.LedgerId, Debit: decimal.Zero, Credit: p.Amount.Add(interest)},
		}
	default:
		return fmt.Errorf("invalid interest type: %s", p.InterestType.Val())
	}

	payload := payload.TransactionCreatedPayload{
		TransactionDate: p.StartDate,
		Description:     fmt.Sprintf("分期付款 %s", p.Memo),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, payload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}
	return tx.Projection.InstallmentRepo.UpdateInstallmentTxn(ctx, st.Installment.InstallmentId, txnId, updatedBy)
}

func (s *TransactionProjectionService) applyInstallmentPeriodPaid(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.InstallmentPeriodPaidPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InstallmentPeriodPaidState](ct)
	if err != nil {
		return err
	}

	cfFin := enums.CashFlowCategoryFinancing.Enum()
	var entries []payload.TransactionEntryPayload
	switch st.Installment.InterestType.Val() {
	case enums.InterestTypeFree:
		entries = []payload.TransactionEntryPayload{
			// Dr. 應付帳款（融資活動：還款減少負債）
			{AccountId: st.Ledger.AccountId, LedgerId: &st.Ledger.LedgerId, Debit: st.InstallmentPayments.Amount, Credit: decimal.Zero, CashFlowCategory: &cfFin},
			// Cr. 銀行/信用卡帳單（bank，CF 以期初期末餘額差計算）
			{AccountId: st.PaidLedger.AccountId, LedgerId: &st.PaidLedger.LedgerId, Debit: decimal.Zero, Credit: st.InstallmentPayments.Amount},
		}
	case enums.InterestTypeFixedRate:
		totalAmount := st.InstallmentPayments.Amount.Add(st.InstallmentPayments.Interest)
		entries = []payload.TransactionEntryPayload{
			// Dr. 應付帳款本金（融資活動：還款減少負債）
			{AccountId: st.Ledger.AccountId, LedgerId: &st.Ledger.LedgerId, Debit: st.InstallmentPayments.Amount, Credit: decimal.Zero, CashFlowCategory: &cfFin},
			// Cr. 預付利息
			{AccountId: st.SysAccountAssetPrepaidInterest, Debit: decimal.Zero, Credit: st.InstallmentPayments.Interest},
			// Dr. 利息費用（留在 NI，間接法中 NI 已含此費用）
			{AccountId: st.SysAccountExpenseInterestExpense, Debit: st.InstallmentPayments.Interest, Credit: decimal.Zero},
			// Cr. 銀行/信用卡帳單（bank，CF 以期初期末餘額差計算）
			{AccountId: st.PaidLedger.AccountId, LedgerId: &st.PaidLedger.LedgerId, Debit: decimal.Zero, Credit: totalAmount},
		}
	default:
		return fmt.Errorf("invalid interest type: %s", st.Installment.InterestType.Val())
	}

	payload := payload.TransactionCreatedPayload{
		TransactionDate: p.PaidDate,
		Description:     fmt.Sprintf("每期還款 %s", st.Installment.Description),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, payload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}
	return tx.Projection.InstallmentRepo.UpdatePaymentTxn(ctx, st.Installment.InstallmentId, txnId, updatedBy)
}

func (s *TransactionProjectionService) applyPrepaidCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidCreatedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidCreatedState](ct)
	if err != nil {
		return err
	}

	cfOperating := enums.CashFlowCategoryOperating.Enum()
	entries := []payload.TransactionEntryPayload{
		// Dr. 預付科目 [OPERATING]（營業活動：現金用於取得預付費用）
		{AccountId: p.AccountID, Debit: p.TotalAmount, Credit: decimal.Zero, CashFlowCategory: &cfOperating},
		// Cr. 付款帳戶（CASH）
		{AccountId: st.Ledger.AccountId, LedgerId: &st.Ledger.LedgerId, Debit: decimal.Zero, Credit: p.TotalAmount},
	}
	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: p.StartDate,
		Description:     fmt.Sprintf("預付費用 %s", p.Name),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, txnPayload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
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

	cfOperating := enums.CashFlowCategoryOperating.Enum()
	amortAmount := thisAmortizationAmount(st.Prepaid.TotalAmount, st.Prepaid.Periods, st.Prepaid.AmortizedPeriods, st.Prepaid.AmortizedAmount)
	entries := []payload.TransactionEntryPayload{
		// Dr. 費用科目（攤提費用，計入 NI）
		{AccountId: st.Prepaid.ExpenseAccountID, Debit: amortAmount, Credit: decimal.Zero},
		// Cr. 預付科目 [OPERATING]（非現金：加回 NI 中已認列費用）
		{AccountId: st.Prepaid.AccountID, Debit: decimal.Zero, Credit: amortAmount, CashFlowCategory: &cfOperating},
	}
	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: p.PeriodDate + "-01",
		Description:     fmt.Sprintf("預付費用攤提 %s %s", st.Prepaid.Name, p.PeriodDate),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, txnPayload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}
	return tx.Projection.PrepaidRepo.InsertPrepaidAmortization(ctx, sqlcdb.InsertPrepaidAmortizationParams{
		MerchantID: ct.MerchantID,
		PrepaidID:  st.Prepaid.ID,
		TxnID:      txnId,
		PeriodDate: p.PeriodDate,
		Amount:     amortAmount,
		UpdatedBy:  updatedBy,
	})
}

func (s *TransactionProjectionService) applyPrepaidDisposed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PrepaidDisposedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.PrepaidDisposedState](ct)
	if err != nil {
		return err
	}

	cfOperating := enums.CashFlowCategoryOperating.Enum()
	remaining := st.Prepaid.TotalAmount.Sub(st.Prepaid.AmortizedAmount)
	entries := []payload.TransactionEntryPayload{
		// Dr. 費用科目（一次認列剩餘預付費用）
		{AccountId: st.Prepaid.ExpenseAccountID, Debit: remaining, Credit: decimal.Zero},
		// Cr. 預付科目 [OPERATING]（非現金：加回 NI 中已認列費用）
		{AccountId: st.Prepaid.AccountID, Debit: decimal.Zero, Credit: remaining, CashFlowCategory: &cfOperating},
	}
	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: p.DisposalDate,
		Description:     fmt.Sprintf("預付費用提前終止 %s", st.Prepaid.Name),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	_, err = s.applyTransaction(ctx, tx, txnPayload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	return err
}

func (s *TransactionProjectionService) applyAssetPurchased(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AssetPurchasedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.AssetPurchasedState](ct)
	if err != nil {
		return err
	}

	cfInvesting := enums.CashFlowCategoryInvesting.Enum()
	var entries []payload.TransactionEntryPayload

	switch p.PaymentType.Val() {
	case enums.AssetPaymentTypeCash:
		// 現金購買：Dr. 資產帳戶 [INVESTING]；Cr. 付款帳戶 [CASH（無標籤）]
		entries = []payload.TransactionEntryPayload{
			{AccountId: p.AssetAccountID, Debit: p.Cost, Credit: decimal.Zero, CashFlowCategory: &cfInvesting},
			{AccountId: st.Ledger.AccountId, LedgerId: &st.Ledger.LedgerId, Debit: decimal.Zero, Credit: p.Cost},
		}
	case enums.AssetPaymentTypeLease:
		// 租賃取得：Dr. 使用權資產 [NULL]；Cr. 租賃負債科目 [NULL]（非現金交易）
		entries = []payload.TransactionEntryPayload{
			{AccountId: p.AssetAccountID, Debit: p.Cost, Credit: decimal.Zero},
			{AccountId: p.LiabilityAccountID, Debit: decimal.Zero, Credit: p.Cost},
		}
	default:
		return fmt.Errorf("invalid payment type: %s", p.PaymentType.Val())
	}

	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: p.PurchaseDate,
		Description:     fmt.Sprintf("固定資產購入 %s", p.Name),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, txnPayload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
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

	cfOperating := enums.CashFlowCategoryOperating.Enum()
	deprAmount := thisDepreciationAmount(st.Asset.Cost, st.Asset.ResidualValue, st.Asset.UsefulLifeMonths, st.Asset.DepreciatedPeriods, st.Asset.TotalDepreciated)
	entries := []payload.TransactionEntryPayload{
		// Dr. 折舊費用（計入 NI）
		{AccountId: st.Asset.DepreciationExpenseAccountID, Debit: deprAmount, Credit: decimal.Zero},
		// Cr. 累計折舊 [OPERATING]（非現金：加回 NI 中已認列折舊費用）
		{AccountId: st.Asset.AccumDepreciationAccountID, Debit: decimal.Zero, Credit: deprAmount, CashFlowCategory: &cfOperating},
	}
	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: p.PeriodDate + "-01",
		Description:     fmt.Sprintf("固定資產折舊 %s %s", st.Asset.Name, p.PeriodDate),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	txnId, err := s.applyTransaction(ctx, tx, txnPayload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	if err != nil {
		return err
	}
	return tx.Projection.FixedAssetRepo.InsertFixedAssetDepreciation(ctx, sqlcdb.InsertFixedAssetDepreciationParams{
		MerchantID: ct.MerchantID,
		AssetID:    st.Asset.ID,
		TxnID:      txnId,
		PeriodDate: p.PeriodDate,
		Amount:     deprAmount,
		UpdatedBy:  updatedBy,
	})
}

func (s *TransactionProjectionService) applyAssetDisposed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.AssetDisposedPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.AssetDisposedState](ct)
	if err != nil {
		return err
	}

	cfInvesting := enums.CashFlowCategoryInvesting.Enum()
	bookValue := st.Asset.Cost.Sub(st.Asset.TotalDepreciated)
	gainLoss := p.Proceeds.Sub(bookValue)

	// 處分分錄：
	// Cr. 資產帳戶（認列資產帳面成本移出）[INVESTING]
	// Dr. 累計折舊（移除已累積的折舊）[INVESTING]
	// Dr. 收款帳戶（現金流入，若有）[NULL → CASH 帳戶]
	// 損益分錄（重分類至 INVESTING，從 NI 移出）
	entries := []payload.TransactionEntryPayload{
		{AccountId: st.Asset.AccumDepreciationAccountID, Debit: st.Asset.TotalDepreciated, Credit: decimal.Zero, CashFlowCategory: &cfInvesting},
		{AccountId: st.Asset.AssetAccountID, Debit: decimal.Zero, Credit: st.Asset.Cost, CashFlowCategory: &cfInvesting},
	}

	if p.Proceeds.GreaterThan(decimal.Zero) && st.ProceedsLedger != nil {
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId: st.ProceedsLedger.AccountId,
			LedgerId:  &st.ProceedsLedger.LedgerId,
			Debit:     p.Proceeds,
			Credit:    decimal.Zero,
		})
	}

	if gainLoss.GreaterThan(decimal.Zero) {
		// 處分利得（移至 INVESTING，重分類出 NI）
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId:        p.GainAccountID,
			Credit:           gainLoss,
			Debit:            decimal.Zero,
			CashFlowCategory: &cfInvesting,
		})
	} else if gainLoss.LessThan(decimal.Zero) {
		// 處分損失（移至 INVESTING，重分類出 NI）
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId:        p.LossAccountID,
			Debit:            gainLoss.Abs(),
			Credit:           decimal.Zero,
			CashFlowCategory: &cfInvesting,
		})
	}

	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: p.DisposalDate,
		Description:     fmt.Sprintf("固定資產處分 %s", st.Asset.Name),
		Currency:        "TWD",
		Entries:         entries,
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	_, err = s.applyTransaction(ctx, tx, txnPayload, enums.TransactionStatusActive.Enum(), ct.MerchantID, updatedBy)
	return err
}
