package projection

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
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

func (s *TransactionProjectionService) Name() string { return enums.AggregateTransaction.String() }

func (s *TransactionProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.String() {
	case event_types.EventTransactionCreated.String():
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventTransactionCorrected.String():
		return s.applyCorrected(ctx, tx, ct)
	case event_types.EventTransactionVoided.String():
		return s.applyVoided(ctx, tx, ct)
	case event_types.EventPeriodAnnualClosed.String():
		return s.applyPeriodAnnualClosed(ctx, tx, ct)
	case event_types.EventPeriodAnnualReopened.String():
		return s.applyPeriodAnnualReopened(ctx, tx, ct)

	case event_types.EventInvestmentBought.String():
		return s.applyInvestmentBought(ctx, tx, ct)
	}
	return nil
}

func (s *TransactionProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCreatedPayload](ct)
	if err != nil {
		return err
	}

	if _, err := s.applyTransaction(ctx, tx, *p, enums.TransactionStatusActive); err != nil {
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

	closingTxnId, err := s.applyTransaction(ctx, tx, state.ClosedTxn, enums.TransactionStatusActive)
	if err != nil {
		return err
	}

	openingTxnId, err := s.applyTransaction(ctx, tx, state.OpenedTxn, enums.TransactionStatusActive)
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

	reverseClosedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseClosedTxn, enums.TransactionStatusVoidRef)
	if err != nil {
		return err
	}

	reverseOpenedTxnId, err := s.applyTransaction(ctx, tx, state.ReverseOpenedTxn, enums.TransactionStatusVoidRef)
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *state.ReverseClosedTxn.RefTxnId, &reverseClosedTxnId, enums.TransactionStatusVoided)
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *state.ReverseOpenedTxn.RefTxnId, &reverseOpenedTxnId, enums.TransactionStatusVoided)
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
		TotalAmount:     st.Movement.UnitPriceTWD.Mul(st.Movement.Quantity),
		Currency:        "TWD",
		Entries:         s.buyEntries(st, p),
	}

	txnId, err := s.applyTransaction(ctx, tx, payload, enums.TransactionStatusActive)
	if err != nil {
		return err
	}

	err = tx.Projection.InvestmentRepo.UpdateMovement(ctx, st.Movement.MovementId, txnId)
	if err != nil {
		return err
	}

	if st.Investment.CostMethod.String() == enums.CostMethodFIFO.String() {
		err = tx.Projection.InvestmentRepo.UpdateLot(ctx, st.Investment.InvestmentId, txnId)
		if err != nil {
			return err
		}
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
	// 業務邏輯：組裝 proj model
	txnId, err := tx.Projection.TransactionRepo.InsertTxn(ctx, projection.Transaction{
		TransactionDate: p.TransactionDate,
		Description:     p.Description,
		TotalAmount:     p.TotalAmount,
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
