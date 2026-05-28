package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventSysAccountUpdated
// ------------------------------

type eventSysAccountUpdatedProjector struct {
	query *query.Repo
}

func (e *eventSysAccountUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.SysAccountUpdatedPayload]) error {
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

func NewEventSysAccountUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.SysAccountUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.SysAccountUpdatedPayload](&eventSysAccountUpdatedProjector{query: query})
}
