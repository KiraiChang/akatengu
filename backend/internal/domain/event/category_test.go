package event_test

import (
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/model/payload"
	"testing"
)

// contractCategoryUUID names the category of the contract fixtures; the
// accounts follow services/category_test.go.
const contractCategoryUUID = "0192f3f2-0000-7000-8000-000000000901"

// TestCategoryPayloads_LegacyJSON is the contract of BFS R2 D12 for the
// category events; see TestAccountPayloads_LegacyJSON.
func TestCategoryPayloads_LegacyJSON(t *testing.T) {
	t.Run("prepaid_category.created", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.PrepaidCategoryCreatedPayload{Name: "保險費", AccountID: "1104-01", ExpenseAccountID: "5201-01"},
			payload.PrepaidCategoryCreatedPayload{Name: "保險費", AccountID: "1104-01", ExpenseAccountID: "5201-01"})
	})
	t.Run("prepaid_category.updated", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.PrepaidCategoryUpdatedPayload{CategoryUUID: contractCategoryUUID, Name: "保險費", AccountID: "1104-01", ExpenseAccountID: "5201-01"},
			payload.PrepaidCategoryUpdatedPayload{CategoryUUID: contractCategoryUUID, Name: "保險費", AccountID: "1104-01", ExpenseAccountID: "5201-01"})
	})
	t.Run("prepaid_category.deleted", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.PrepaidCategoryDeletedPayload{CategoryUUID: contractCategoryUUID},
			payload.PrepaidCategoryDeletedPayload{CategoryUUID: contractCategoryUUID})
	})
	t.Run("fixed_asset_category.created", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.FixedAssetCategoryCreatedPayload{Name: "電腦設備",
				AssetAccountID: "1601-01", AccumDepreciationAccountID: "1601-02", DepreciationExpenseAccountID: "5601-01"},
			payload.FixedAssetCategoryCreatedPayload{Name: "電腦設備",
				AssetAccountID: "1601-01", AccumDepreciationAccountID: "1601-02", DepreciationExpenseAccountID: "5601-01"})
	})
	t.Run("fixed_asset_category.updated", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.FixedAssetCategoryUpdatedPayload{CategoryUUID: contractCategoryUUID, Name: "電腦設備",
				AssetAccountID: "1601-01", AccumDepreciationAccountID: "1601-02", DepreciationExpenseAccountID: "5601-01"},
			payload.FixedAssetCategoryUpdatedPayload{CategoryUUID: contractCategoryUUID, Name: "電腦設備",
				AssetAccountID: "1601-01", AccumDepreciationAccountID: "1601-02", DepreciationExpenseAccountID: "5601-01"})
	})
	t.Run("fixed_asset_category.deleted", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.FixedAssetCategoryDeletedPayload{CategoryUUID: contractCategoryUUID},
			payload.FixedAssetCategoryDeletedPayload{CategoryUUID: contractCategoryUUID})
	})
}
