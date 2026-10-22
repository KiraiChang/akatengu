package event

import "akatengu/internal/enums/event_types"

// The category payloads are copies of model/payload/prepaid_category.go and
// fixed_asset_category.go (BFS R2 D1); the JSON tags must stay identical until
// R10. A category is named after its creation event's UUID (D6), so the
// created payloads carry no identity.

// PrepaidCategoryCreatedPayload creates a prepaid expense category.
type PrepaidCategoryCreatedPayload struct {
	Name             string `json:"name"`
	AccountID        string `json:"account_id"`
	ExpenseAccountID string `json:"expense_account_id"`
}

func (p PrepaidCategoryCreatedPayload) EventType() event_types.EventType {
	return event_types.EventPrepaidCategoryCreated.Enum()
}

func (p PrepaidCategoryCreatedPayload) Validate() error {
	var errs []string
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.AccountID == "" {
		errs = append(errs, "account_id is required")
	}
	if p.ExpenseAccountID == "" {
		errs = append(errs, "expense_account_id is required")
	}
	return joinErrors(errs)
}

// PrepaidCategoryUpdatedPayload updates a prepaid expense category.
type PrepaidCategoryUpdatedPayload struct {
	CategoryUUID     string `json:"category_uuid"`
	Name             string `json:"name"`
	AccountID        string `json:"account_id"`
	ExpenseAccountID string `json:"expense_account_id"`
}

func (p PrepaidCategoryUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventPrepaidCategoryUpdated.Enum()
}

func (p PrepaidCategoryUpdatedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.AccountID == "" {
		errs = append(errs, "account_id is required")
	}
	if p.ExpenseAccountID == "" {
		errs = append(errs, "expense_account_id is required")
	}
	return joinErrors(errs)
}

// PrepaidCategoryDeletedPayload soft-deletes a prepaid expense category.
type PrepaidCategoryDeletedPayload struct {
	CategoryUUID string `json:"category_uuid"`
}

func (p PrepaidCategoryDeletedPayload) EventType() event_types.EventType {
	return event_types.EventPrepaidCategoryDeleted.Enum()
}

func (p PrepaidCategoryDeletedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	return joinErrors(errs)
}

// FixedAssetCategoryCreatedPayload creates a fixed asset category.
type FixedAssetCategoryCreatedPayload struct {
	Name                         string `json:"name"`
	AssetAccountID               string `json:"asset_account_id"`
	AccumDepreciationAccountID   string `json:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string `json:"depreciation_expense_account_id"`
}

func (p FixedAssetCategoryCreatedPayload) EventType() event_types.EventType {
	return event_types.EventFixedAssetCategoryCreated.Enum()
}

func (p FixedAssetCategoryCreatedPayload) Validate() error {
	var errs []string
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.AssetAccountID == "" {
		errs = append(errs, "asset_account_id is required")
	}
	if p.AccumDepreciationAccountID == "" {
		errs = append(errs, "accum_depreciation_account_id is required")
	}
	if p.DepreciationExpenseAccountID == "" {
		errs = append(errs, "depreciation_expense_account_id is required")
	}
	return joinErrors(errs)
}

// FixedAssetCategoryUpdatedPayload updates a fixed asset category.
type FixedAssetCategoryUpdatedPayload struct {
	CategoryUUID                 string `json:"category_uuid"`
	Name                         string `json:"name"`
	AssetAccountID               string `json:"asset_account_id"`
	AccumDepreciationAccountID   string `json:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string `json:"depreciation_expense_account_id"`
}

func (p FixedAssetCategoryUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventFixedAssetCategoryUpdated.Enum()
}

func (p FixedAssetCategoryUpdatedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.AssetAccountID == "" {
		errs = append(errs, "asset_account_id is required")
	}
	if p.AccumDepreciationAccountID == "" {
		errs = append(errs, "accum_depreciation_account_id is required")
	}
	if p.DepreciationExpenseAccountID == "" {
		errs = append(errs, "depreciation_expense_account_id is required")
	}
	return joinErrors(errs)
}

// FixedAssetCategoryDeletedPayload soft-deletes a fixed asset category.
type FixedAssetCategoryDeletedPayload struct {
	CategoryUUID string `json:"category_uuid"`
}

func (p FixedAssetCategoryDeletedPayload) EventType() event_types.EventType {
	return event_types.EventFixedAssetCategoryDeleted.Enum()
}

func (p FixedAssetCategoryDeletedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	return joinErrors(errs)
}
