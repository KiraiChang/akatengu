package factory

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// EventFixedAssetCategoryCreated
// ------------------------------

type eventFixedAssetCategoryCreatedProjector struct{}

func (e *eventFixedAssetCategoryCreatedProjector) Project(_ context.Context, ct *pipelines.Context[pipelines.NoState, payload.FixedAssetCategoryCreatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventFixedAssetCategoryCreatedPipeline() *pipelines.TypedPipeline[pipelines.NoState, payload.FixedAssetCategoryCreatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.FixedAssetCategoryCreatedPayload](&eventFixedAssetCategoryCreatedProjector{}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ------------------------------
// EventFixedAssetCategoryUpdated
// ------------------------------

type eventFixedAssetCategoryUpdatedProjector struct {
	query *query.Repo
}

func (e *eventFixedAssetCategoryUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.FixedAssetCategoryUpdatedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	cat, err := e.query.FixedAssetCategory.GetFixedAssetCategoryByUUID(ctx, ct.Payload.CategoryUUID)
	if err != nil {
		return err
	}
	if cat == nil {
		return fmt.Errorf("fixed asset category not found")
	}
	if !cat.IsActive {
		return fmt.Errorf("fixed asset category is inactive")
	}
	return nil
}

func NewEventFixedAssetCategoryUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.FixedAssetCategoryUpdatedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.FixedAssetCategoryUpdatedPayload](&eventFixedAssetCategoryUpdatedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}

// ------------------------------
// EventFixedAssetCategoryDeleted
// ------------------------------

type eventFixedAssetCategoryDeletedProjector struct {
	query *query.Repo
}

func (e *eventFixedAssetCategoryDeletedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.FixedAssetCategoryDeletedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	cat, err := e.query.FixedAssetCategory.GetFixedAssetCategoryByUUID(ctx, ct.Payload.CategoryUUID)
	if err != nil {
		return err
	}
	if cat == nil {
		return fmt.Errorf("fixed asset category not found")
	}
	if !cat.IsActive {
		return fmt.Errorf("fixed asset category is already inactive")
	}
	return nil
}

func NewEventFixedAssetCategoryDeletedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.FixedAssetCategoryDeletedPayload] {
	return pipelines.NewType[pipelines.NoState, payload.FixedAssetCategoryDeletedPayload](&eventFixedAssetCategoryDeletedProjector{query}, func() *pipelines.NoState {
		return &pipelines.NoState{}
	})
}
