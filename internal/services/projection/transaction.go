package projection

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
	"encoding/json"
)

// ─────────────────────────────────────────
// TransactionProjectionService
// ─────────────────────────────────────────

type TransactionProjectionService struct{}

func (s *TransactionProjectionService) Name() string { return enums.AggregateTransaction.String() }

func (s *TransactionProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	switch event.EventType.String() {
	case event_types.EventTransactionCreated.String():
		return s.applyCreated(ctx, tx, event)
	case event_types.EventTransactionCorrected.String():
		return s.applyCorrected(ctx, tx, event)
	case event_types.EventTransactionVoided.String():
		return s.applyVoided(ctx, tx, event)
	case event_types.EventPeriodAnnualClosed.String():
		return s.applyPeriodAnnualClosed(ctx, tx, event)
	case event_types.EventPeriodAnnualReopened.String():
		return s.applyPeriodAnnualReopened(ctx, tx, event)
	}
	return nil
}

func (s *TransactionProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.TransactionCreatedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if _, err := s.applyTransaction(ctx, tx, p, enums.TransactionStatusActive); err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyCorrected(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.TransactionCorrectedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if err := tx.Projection.TransactionRepo.UpdateTxnStatus(ctx, p.OriginalTransactionId,
		enums.TransactionStatusCorrected,
		event.AggregateVersion,
	); err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyVoided(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.TransactionVoidedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if err := tx.Projection.TransactionRepo.UpdateTxnStatus(ctx,
		p.TransactionId,
		enums.TransactionStatusVoided,
		event.AggregateVersion,
	); err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyPeriodAnnualClosed(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.PeriodAnnualClosedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	closingTxnId, err := s.applyTransaction(ctx, tx, p.ClosedTransaction, enums.TransactionStatusActive)
	if err != nil {
		return err
	}

	openingTxnId, err := s.applyTransaction(ctx, tx, p.OpenedTransaction, enums.TransactionStatusActive)
	if err != nil {
		return err
	}

	err = tx.Projection.PeriodCloseRepo.UpdatePeriodCloseTxnId(ctx, p.ClosingId, &closingTxnId, &openingTxnId)
	if err != nil {
		return err
	}

	return nil
}

func (s *TransactionProjectionService) applyPeriodAnnualReopened(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var p payload.PeriodAnnualReopenedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	reverseClosedTxnId, err := s.applyTransaction(ctx, tx, p.ReverseClosedTxn, enums.TransactionStatusVoidRef)
	if err != nil {
		return err
	}

	reverseOpenedTxnId, err := s.applyTransaction(ctx, tx, p.ReverseOpenedTxn, enums.TransactionStatusVoidRef)
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *p.ReverseClosedTxn.RefTxnId, &reverseClosedTxnId, enums.TransactionStatusVoided)
	if err != nil {
		return err
	}

	err = tx.Projection.TransactionRepo.SysUpdateTxnStatus(ctx, *p.ReverseOpenedTxn.RefTxnId, &reverseOpenedTxnId, enums.TransactionStatusVoided)
	if err != nil {
		return err
	}

	err = tx.Projection.PeriodCloseRepo.UpdatePeriodCloseTxnId(ctx, p.ClosingId, nil, nil)
	if err != nil {
		return err
	}

	return nil
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
