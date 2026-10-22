package process

import (
	event_payload "akatengu/internal/domain/event"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"context"
	"fmt"
)

// categoryLookup reports whether a category exists and is active, read with
// the Run's transaction-bound repo.
type categoryLookup func(ctx context.Context, repo *query.Repo, categoryUUID string) (found, active bool, err error)

func prepaidCategoryLookup(ctx context.Context, repo *query.Repo, categoryUUID string) (bool, bool, error) {
	cat, err := repo.PrepaidCategory.GetPrepaidCategoryByUUID(ctx, categoryUUID)
	if err != nil || cat == nil {
		return false, false, err
	}
	return true, cat.IsActive, nil
}

func fixedAssetCategoryLookup(ctx context.Context, repo *query.Repo, categoryUUID string) (bool, bool, error) {
	cat, err := repo.FixedAssetCategory.GetFixedAssetCategoryByUUID(ctx, categoryUUID)
	if err != nil || cat == nil {
		return false, false, err
	}
	return true, cat.IsActive, nil
}

// categoryChangeHandle validates an update or a delete and checks that the
// category exists and is active, with the messages of the legacy pipelines
// (services/pipelines/factory/prepaid_category.go, fixed_asset_category.go):
// "<subject> not found" and "<subject> <inactive>".
func categoryChangeHandle[T interface{ Validate() error }](lookup categoryLookup, subject, inactive string, categoryUUID func(T) string) mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
		result := result.EventResult{}
		p, err := checkAndGetPayload[T](evt)
		if err != nil {
			return result, err
		}
		if err := (*p).Validate(); err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, err)
		}

		found, active, err := lookup(ctx, repo, categoryUUID(*p))
		if err != nil {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("get %s: %w", subject, err))
		}
		if !found {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("%s not found", subject))
		}
		if !active {
			return result, kerrors.NewBusinessError(kerrors.ErrPayloadInvalid, evt, fmt.Errorf("%s %s", subject, inactive))
		}
		return result, nil
	}
}

// ----------------------------------
// PrepaidCategory
// ----------------------------------

func EventPrepaidCategoryUpdatedHandle() mediator.HandlerFunc {
	return categoryChangeHandle(prepaidCategoryLookup, "prepaid category", "is inactive",
		func(p event_payload.PrepaidCategoryUpdatedPayload) string { return p.CategoryUUID })
}

func EventPrepaidCategoryDeletedHandle() mediator.HandlerFunc {
	return categoryChangeHandle(prepaidCategoryLookup, "prepaid category", "is already inactive",
		func(p event_payload.PrepaidCategoryDeletedPayload) string { return p.CategoryUUID })
}

// ----------------------------------
// FixedAssetCategory
// ----------------------------------

func EventFixedAssetCategoryUpdatedHandle() mediator.HandlerFunc {
	return categoryChangeHandle(fixedAssetCategoryLookup, "fixed asset category", "is inactive",
		func(p event_payload.FixedAssetCategoryUpdatedPayload) string { return p.CategoryUUID })
}

func EventFixedAssetCategoryDeletedHandle() mediator.HandlerFunc {
	return categoryChangeHandle(fixedAssetCategoryLookup, "fixed asset category", "is already inactive",
		func(p event_payload.FixedAssetCategoryDeletedPayload) string { return p.CategoryUUID })
}
