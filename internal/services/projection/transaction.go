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

	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// TransactionProjectionService
// ─────────────────────────────────────────

type TransactionProjectionService struct{}

func (s *TransactionProjectionService) Name() string { return string(enums.AggregateTransaction) }

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
	}
	return nil
}

func (s *TransactionProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCreatedPayload](ct)
	if err != nil {
		return err
	}

	if _, err := s.applyTransaction(ctx, tx, *p, enums.TransactionStatusActive.Enum()); err != nil {
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

	closingTxnId, err := s.applyTransaction(ctx, tx, state.ClosedTxn, enums.TransactionStatusActive.Enum())
	if err != nil {
		return err
	}

	openingTxnId, err := s.applyTransaction(ctx, tx, state.OpenedTxn, enums.TransactionStatusActive.Enum())
	if err != nil {
		return err
	}

	err = tx.Projection.PeriodCloseRepo.UpdatePeriodCloseTxnId(ctx, p.ClosingId, &closingTxnId, &openingTxnId)
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

	reverseClosedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseClosedTxn, enums.TransactionStatusVoidRef.Enum())
	if err != nil {
		return err
	}

	reverseOpenedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseOpenedTxn, enums.TransactionStatusVoidRef.Enum())
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *state.ReverseClosedTxn.RefTxnId, &reverseClosedTxnId, enums.TransactionStatusVoided.Enum())
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *state.ReverseOpenedTxn.RefTxnId, &reverseOpenedTxnId, enums.TransactionStatusVoided.Enum())
	if err != nil {
		return err
	}

	err = tx.Projection.PeriodCloseRepo.UpdatePeriodCloseTxnId(ctx, p.ClosingId, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyInvestmentBought(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.InvestmentBoughtPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentBoughtState](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	payload := payload.TransactionCreatedPayload{
		TransactionDate: p.Date,
		Description:     fmt.Sprintf("買入 %s", st.Investment.Name),
		Currency:        "TWD",
		Entries:         s.buyEntries(st, p),
	}

	txnId, err := s.applyTransaction(ctx, tx, payload, enums.TransactionStatusActive.Enum())
	if err != nil {
		return err
	}

	err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.Is(enums.CostMethodFIFO) {
		err = tx.Projection.InvestmentRepo.UpdateLot(ctx, st.Investment.InvestmentId, txnId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionProjectionService) applyInvestmentSold(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.InvestmentSoldPayload](ct)
	if err != nil {
		return err
	}
	st, err := checkAndGetState[state.InvestmentSoldState](ct)
	if err != nil {
		return err
	}

	// 業務邏輯：組裝 proj model
	payload := payload.TransactionCreatedPayload{
		TransactionDate: p.Date,
		Description:     fmt.Sprintf("賣出 %s", st.Investment.Name),
		Currency:        "TWD",
		Entries:         s.sellEntries(*p, *st),
	}

	txnId, err := s.applyTransaction(ctx, tx, payload, enums.TransactionStatusActive.Enum())
	if err != nil {
		return err
	}

	err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId)
	if err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) buyEntries(st *state.InvestmentBoughtState, p *payload.InvestmentBoughtPayload) []payload.TransactionEntryPayload {
	cost := p.Quantity.Mul(p.UnitPrice).Mul(p.ExchangeRate)
	totalCost := cost.Add(p.Fee).Add(p.Tax)
	entries := []payload.TransactionEntryPayload{
		// 資產增加
		{AccountId: st.Investment.AccountId, Debit: cost, Credit: decimal.Zero},
		// 扣款帳戶
		{AccountId: st.Ledger.AccountId, LedgerId: &p.LedgerId, Debit: decimal.Zero, Credit: totalCost},
	}
	if p.Fee.IsPositive() {
		entries = append(entries,
			payload.TransactionEntryPayload{AccountId: "5940", Debit: p.Fee, Credit: decimal.Zero},
		)
	}
	if p.Tax.IsPositive() {
		entries = append(entries,
			payload.TransactionEntryPayload{AccountId: "5950", Debit: p.Tax, Credit: decimal.Zero},
		)
	}
	return entries
}

func (s *TransactionProjectionService) applyTransaction(ctx context.Context, tx event_store.EventStoreRepositories, p payload.TransactionCreatedPayload, status enums.TransactionStatus) (int64, error) {
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
		TransactionDate: p.TransactionDate,
		Description:     p.Description,
		TotalAmount:     totalDebit,
		Currency:        coalesce(p.Currency, "TWD"),
		Status:          status,
		ReceiptNo:       p.ReceiptNo,
		Note:            p.Note,
		RefTxnId:        p.RefTxnId,
	})
	if err != nil {
		return 0, err
	}

	entries := make([]projection.Entry, len(p.Entries))
	for i, e := range p.Entries {
		entries[i] = projection.Entry{
			TransactionId: txnId,
			LedgerId:      e.LedgerId,
			AccountId:     e.AccountId,
			Debit:         e.Debit,
			Credit:        e.Credit,
		}
	}

	if err := tx.Projection.TransactionRepo.UpsertJournalEntries(ctx, entries); err != nil {
		return 0, err
	}
	return txnId, nil
}

func (s *TransactionProjectionService) sellEntries(
	p payload.InvestmentSoldPayload,
	ct state.InvestmentSoldState,
) []payload.TransactionEntryPayload {
	entries := []payload.TransactionEntryPayload{
		// 入帳
		{AccountId: ct.Ledger.AccountId, LedgerId: &p.LedgerId, Debit: ct.NetProceeds, Credit: decimal.Zero},
		// 成本沖銷
		{AccountId: ct.Investment.AccountId, Debit: decimal.Zero, Credit: ct.CostBasis},
	}
	if p.Fee.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "5940", Debit: p.Fee, Credit: decimal.Zero})
	}
	if p.Tax.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "5950", Debit: p.Tax, Credit: decimal.Zero})
	}
	// 已實現損益（正=利，負=損）
	if ct.RealizedGain.GreaterThanOrEqual(decimal.Zero) {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "4230", Debit: decimal.Zero, Credit: ct.RealizedGain})
	} else {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "4230", Debit: ct.RealizedGain.Abs(), Credit: decimal.Zero})
	}
	return entries
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

	if st.Transaction != nil {
		txnId, err := s.applyTransaction(ctx, tx, *st.Transaction, enums.TransactionStatusActive.Enum())
		if err != nil {
			return err
		}
		err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId)
		if err != nil {
			return err
		}
	}
	return nil
}
