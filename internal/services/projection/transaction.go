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

func (s *TransactionProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, prevResult Result) (Result, error) {
	switch event.EventType.String() {
	case event_types.EventTransactionCreated.String():
		return s.applyCreated(ctx, tx, event, prevResult)
	case event_types.EventTransactionCorrected.String():
		return s.applyCorrected(ctx, tx, event, prevResult)
	case event_types.EventTransactionVoided.String():
		return s.applyVoided(ctx, tx, event, prevResult)
	}
	return prevResult, nil
}

func (s *TransactionProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.TransactionCreatedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	// 業務邏輯：組裝 proj model
	if err := tx.Projection.UpsertTransaction(ctx, projection.Transaction{
		TransactionId:   event.EventId,
		TransactionDate: p.TransactionDate,
		Description:     p.Description,
		TotalAmount:     p.TotalAmount,
		Currency:        coalesce(p.Currency, "TWD"),
		Status:          enums.TransactionStatusActive,
		ReceiptNo:       p.ReceiptNo,
		Note:            p.Note,
		Version:         event.AggregateVersion,
	}); err != nil {
		return result, err
	}

	entries := make([]projection.Entry, len(p.Entries))
	for i, e := range p.Entries {
		entries[i] = projection.Entry{
			TransactionId: event.EventId,
			LedgerId:      e.LedgerId,
			AccountId:     e.AccountId,
			Debit:         e.Debit,
			Credit:        e.Credit,
		}
	}

	if err := tx.Projection.UpsertJournalEntries(ctx, entries); err != nil {
		return result, err
	}

	return result, nil
}

func (s *TransactionProjectionService) applyCorrected(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.TransactionCorrectedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	if err := tx.Projection.UpsertTransaction(ctx, projection.Transaction{
		TransactionId: p.OriginalTransactionId,
		Status:        enums.TransactionStatusCorrected,
		Version:       event.AggregateVersion,
	}); err != nil {
		return result, err
	}

	return result, nil
}

func (s *TransactionProjectionService) applyVoided(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var p payload.TransactionVoidedPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return result, err
	}

	if err := tx.Projection.UpsertTransaction(ctx, projection.Transaction{
		TransactionId: p.TransactionId,
		Status:        enums.TransactionStatusVoided,
		Version:       event.AggregateVersion,
	}); err != nil {
		return result, err
	}

	return result, nil
}
