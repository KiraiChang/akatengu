package process

import (
	event_payload "akatengu/internal/domain/event"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// accountRefHandle validates a config payload and checks that every account it
// refers to exists, as the legacy pipelines do
// (services/pipelines/factory/sysconfig.go, accountconfig.go).
func accountRefHandle[T interface{ Validate() error }](accountIDs func(T) []string) mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[T](evt)
		if err != nil {
			return result, err
		}
		if err := (*p).Validate(); err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, err)
		}
		for _, id := range accountIDs(*p) {
			if _, err := repo.Account.GetAccount(ctx, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("account %s not found", id))
				}
				return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("get account %s: %w", id, err))
			}
		}
		return result, nil
	}
}

func EventSysAccountUpdatedHandle() mediator.HandlerFunc {
	return accountRefHandle(func(p event_payload.SysAccountUpdatedPayload) []string {
		return []string{p.AccountID}
	})
}

func EventLedgerAccountTypeConfigUpdatedHandle() mediator.HandlerFunc {
	return accountRefHandle(func(p event_payload.LedgerAccountTypeConfigUpdatedPayload) []string {
		return []string{p.AccountID}
	})
}

// EventAssetTypeAccountConfigUpdatedHandle checks the required accounts first,
// then the optional ones that are set.
func EventAssetTypeAccountConfigUpdatedHandle() mediator.HandlerFunc {
	return accountRefHandle(func(p event_payload.AssetTypeAccountConfigUpdatedPayload) []string {
		ids := []string{
			p.RealizedGainAccountID,
			p.RealizedLossAccountID,
			p.UnrealizedGainAccountID,
			p.UnrealizedLossAccountID,
			p.FeeAccountID,
			p.TaxAccountID,
		}
		for _, opt := range []*string{p.OCIAccountID, p.AccountID} {
			if opt != nil {
				ids = append(ids, *opt)
			}
		}
		return ids
	})
}
