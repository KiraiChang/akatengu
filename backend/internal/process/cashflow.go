package process

import (
	"akatengu/internal/enums"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"context"
	"fmt"
)

// ----------------------------------
// EventTransactionCFCategoryUpdated
// ----------------------------------

func EventTransactionCFCategoryUpdatedHandle() mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[payload.TransactionCFCategoryUpdatedPayload](evt)
		if err != nil {
			return result, err
		}
		if err := p.Validate(); err != nil {
			return result, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, err)
		}

		txn, err := repo.Transaction.GetByUUID(ctx, p.TxnUUID)
		if err != nil {
			return result, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, fmt.Errorf("get transaction: %w", err))
		}
		if txn == nil {
			return result, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, fmt.Errorf("transaction %s not found", p.TxnUUID))
		}
		if !txn.Status.Is(enums.TransactionStatusActive) {
			return result, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, fmt.Errorf("transaction %s is not active", p.TxnUUID))
		}

		entries, err := repo.Transaction.GetEntries(ctx, txn.TransactionId)
		if err != nil {
			return result, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, fmt.Errorf("get entries: %w", err))
		}

		validUUIDs := make(map[string]bool, len(entries))
		for _, en := range entries {
			validUUIDs[en.EntryUuid] = true
		}

		for i, item := range p.Entries {
			if !validUUIDs[item.EntryUUID] {
				return result, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, fmt.Errorf("entries[%d].entry_uuid %s does not belong to transaction %s", i, item.EntryUUID, p.TxnUUID))
			}
		}
		return result, nil
	}
}
