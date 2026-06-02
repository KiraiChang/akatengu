package payload

// FixedAssetCategoryCreatedPayload 新增固定資產類別
type FixedAssetCategoryCreatedPayload struct {
	Name                         string `json:"name"`
	AssetAccountID               string `json:"asset_account_id"`
	AccumDepreciationAccountID   string `json:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string `json:"depreciation_expense_account_id"`
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

// FixedAssetCategoryUpdatedPayload 修改固定資產類別
type FixedAssetCategoryUpdatedPayload struct {
	CategoryUUID                 string `json:"category_uuid"`
	Name                         string `json:"name"`
	AssetAccountID               string `json:"asset_account_id"`
	AccumDepreciationAccountID   string `json:"accum_depreciation_account_id"`
	DepreciationExpenseAccountID string `json:"depreciation_expense_account_id"`
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

// FixedAssetCategoryDeletedPayload 刪除固定資產類別（軟刪除）
type FixedAssetCategoryDeletedPayload struct {
	CategoryUUID string `json:"category_uuid"`
}

func (p FixedAssetCategoryDeletedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	return joinErrors(errs)
}
