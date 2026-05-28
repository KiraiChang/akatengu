package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventLedgerAccountTypeConfigUpdated
// ------------------------------

type eventLedgerAccountTypeConfigUpdatedProjector struct {
	query *query.Repo
}

func (e *eventLedgerAccountTypeConfigUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.LedgerAccountTypeConfigUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	acct, err := e.query.Account.GetAccount(ctx, ct.Payload.AccountID)
	if err != nil {
		return err
	}
	if acct == nil {
		return fmt.Errorf("account %s not found", ct.Payload.AccountID)
	}
	return nil
}

func NewEventLedgerAccountTypeConfigUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.LedgerAccountTypeConfigUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.LedgerAccountTypeConfigUpdatedPayload](&eventLedgerAccountTypeConfigUpdatedProjector{query: query})
}

// ------------------------------
// EventAssetTypeAccountConfigUpdated
// ------------------------------

type eventAssetTypeAccountConfigUpdatedProjector struct {
	query *query.Repo
}

func (e *eventAssetTypeAccountConfigUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.AssetTypeAccountConfigUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	p := ct.Payload
	requiredIDs := []string{
		p.RealizedGainAccountID,
		p.RealizedLossAccountID,
		p.UnrealizedGainAccountID,
		p.UnrealizedLossAccountID,
		p.FeeAccountID,
		p.TaxAccountID,
	}
	for _, id := range requiredIDs {
		acct, err := e.query.Account.GetAccount(ctx, id)
		if err != nil {
			return err
		}
		if acct == nil {
			return fmt.Errorf("account %s not found", id)
		}
	}
	for _, optID := range []*string{p.OCIAccountID, p.AccountID} {
		if optID == nil {
			continue
		}
		acct, err := e.query.Account.GetAccount(ctx, *optID)
		if err != nil {
			return err
		}
		if acct == nil {
			return fmt.Errorf("account %s not found", *optID)
		}
	}
	return nil
}

func NewEventAssetTypeAccountConfigUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.AssetTypeAccountConfigUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.AssetTypeAccountConfigUpdatedPayload](&eventAssetTypeAccountConfigUpdatedProjector{query: query})
}
