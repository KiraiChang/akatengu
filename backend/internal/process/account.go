package process

import (
	event_payload "akatengu/internal/domain/event"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/runtime/mediator"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// validateHandle only validates the payload, as the legacy pipelines of these
// event types do (services/pipelines/factory/account.go).
func validateHandle[T interface{ Validate() error }]() mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[T](evt)
		if err != nil {
			return result, err
		}
		if err := (*p).Validate(); err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, err)
		}
		return result, nil
	}
}

// versionConflict rejects an update whose version is not the read model's
// (ADR-027 rule 5). The message is the one of the legacy projection (F6), so
// the frontend shows the same text.
func versionConflict(evt event.Event, subject string, version int64) error {
	return kerrors.NewBusinessError(kerrors.ErrBusinessRuleFailed, evt,
		fmt.Errorf("%s version %d: %w", subject, version, projection_repo.ErrVersionConflict))
}

// ----------------------------------
// EventAccountUpdated
// ----------------------------------

func EventAccountUpdatedHandle() mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[event_payload.AccountUpdatedPayload](evt)
		if err != nil {
			return result, err
		}
		if err := p.Validate(); err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, err)
		}

		// A missing account is a conflict too, as the legacy versioned UPDATE
		// that matches no row.
		a, err := repo.Account.GetAccount(ctx, p.AccountId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("get account: %w", err))
		}
		if a == nil || a.Version != p.Version {
			return result, versionConflict(evt, "account "+p.AccountId, p.Version)
		}
		return result, nil
	}
}

// ----------------------------------
// EventLedgerAccountCreated
// ----------------------------------

// EventLedgerAccountCreatedHandle allocates the ledger_id inside the Run's
// transaction and records it in the event, replacing any value the client
// sent, so a rebuild writes the ledger account back under the same ID
// (BFS R3 D3).
func EventLedgerAccountCreatedHandle() mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[event_payload.LedgerAccountCreatePayload](evt)
		if err != nil {
			return result, err
		}
		if err := p.Validate(); err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, err)
		}

		id, err := repo.Account.NextLedgerID(ctx)
		if err != nil {
			return result, kerrors.NewInfrastructureError(kerrors.ErrStoreReadFailed, evt, fmt.Errorf("next ledger_id: %w", err))
		}
		p.LedgerId = id
		result.Payload = *p
		return result, nil
	}
}

// ----------------------------------
// EventLedgerAccountUpdated
// ----------------------------------

func EventLedgerAccountUpdatedHandle() mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[event_payload.LedgerAccountUpdatedPayload](evt)
		if err != nil {
			return result, err
		}
		if err := p.Validate(); err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, err)
		}

		la, err := repo.Account.GetLedger(ctx, p.LedgerId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("get ledger account: %w", err))
		}
		if la == nil || la.Version != p.Version {
			return result, versionConflict(evt, fmt.Sprintf("ledger account %d", p.LedgerId), p.Version)
		}
		return result, nil
	}
}
