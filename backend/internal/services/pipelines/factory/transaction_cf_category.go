package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventTransactionCFCategoryUpdated
// ------------------------------

type eventTransactionCFCategoryUpdatedProjector struct {
	query *query.Repo
}

func (e *eventTransactionCFCategoryUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.TransactionCFCategoryUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	txn, err := e.query.Transaction.GetByUUID(ctx, ct.Payload.TxnUUID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}
	if txn == nil {
		return fmt.Errorf("transaction %s not found", ct.Payload.TxnUUID)
	}
	if !txn.Status.Is(enums.TransactionStatusActive) {
		return fmt.Errorf("transaction %s is not active", ct.Payload.TxnUUID)
	}

	entries, err := e.query.Transaction.GetEntries(ctx, txn.TransactionId)
	if err != nil {
		return fmt.Errorf("get entries: %w", err)
	}

	validUUIDs := make(map[string]bool, len(entries))
	for _, en := range entries {
		validUUIDs[en.EntryUuid] = true
	}

	for i, item := range ct.Payload.Entries {
		if !validUUIDs[item.EntryUUID] {
			return fmt.Errorf("entries[%d].entry_uuid %s does not belong to transaction %s", i, item.EntryUUID, ct.Payload.TxnUUID)
		}
	}

	return nil
}

func NewEventTransactionCFCategoryUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.TransactionCFCategoryUpdatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.TransactionCFCategoryUpdatedPayload](
		&eventTransactionCFCategoryUpdatedProjector{query},
		func() *pipelines.NoState { return &pipelines.NoState{} },
	)
}
