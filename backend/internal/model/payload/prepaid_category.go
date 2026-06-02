package payload

// PrepaidCategoryCreatedPayload 新增預付費用類別
type PrepaidCategoryCreatedPayload struct {
	Name             string `json:"name"`
	AccountID        string `json:"account_id"`
	ExpenseAccountID string `json:"expense_account_id"`
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

// PrepaidCategoryUpdatedPayload 修改預付費用類別
type PrepaidCategoryUpdatedPayload struct {
	CategoryUUID     string `json:"category_uuid"`
	Name             string `json:"name"`
	AccountID        string `json:"account_id"`
	ExpenseAccountID string `json:"expense_account_id"`
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

// PrepaidCategoryDeletedPayload 刪除預付費用類別（軟刪除）
type PrepaidCategoryDeletedPayload struct {
	CategoryUUID string `json:"category_uuid"`
}

func (p PrepaidCategoryDeletedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	return joinErrors(errs)
}
