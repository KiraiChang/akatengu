package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventPrepaidCategoryCreated
// ------------------------------

type eventPrepaidCategoryCreatedProjector struct{}

func (e *eventPrepaidCategoryCreatedProjector) Project(_ context.Context, ct *pipelines.Context[pipelines.NoState, payload.PrepaidCategoryCreatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventPrepaidCategoryCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.PrepaidCategoryCreatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.PrepaidCategoryCreatedPayload](&eventPrepaidCategoryCreatedProjector{}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ------------------------------
// EventPrepaidCategoryUpdated
// ------------------------------

type eventPrepaidCategoryUpdatedProjector struct {
	query *query.Repo
}

func (e *eventPrepaidCategoryUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.PrepaidCategoryUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	cat, err := e.query.PrepaidCategory.GetPrepaidCategoryByUUID(ctx, ct.Payload.CategoryUUID)
	if err != nil {
		return err
	}
	if cat == nil {
		return fmt.Errorf("prepaid category not found")
	}
	if !cat.IsActive {
		return fmt.Errorf("prepaid category is inactive")
	}
	return nil
}

func NewEventPrepaidCategoryUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.PrepaidCategoryUpdatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.PrepaidCategoryUpdatedPayload](&eventPrepaidCategoryUpdatedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ------------------------------
// EventPrepaidCategoryDeleted
// ------------------------------

type eventPrepaidCategoryDeletedProjector struct {
	query *query.Repo
}

func (e *eventPrepaidCategoryDeletedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.PrepaidCategoryDeletedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	cat, err := e.query.PrepaidCategory.GetPrepaidCategoryByUUID(ctx, ct.Payload.CategoryUUID)
	if err != nil {
		return err
	}
	if cat == nil {
		return fmt.Errorf("prepaid category not found")
	}
	if !cat.IsActive {
		return fmt.Errorf("prepaid category is already inactive")
	}
	return nil
}

func NewEventPrepaidCategoryDeletedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.PrepaidCategoryDeletedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.PrepaidCategoryDeletedPayload](&eventPrepaidCategoryDeletedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}
