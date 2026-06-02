package services_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/testutil"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─────────────────────────────────────────
// FixedAssetCategory
// ─────────────────────────────────────────

var _ = Describe("EventFixedAssetCategoryCreated", func() {
	It("GIVEN 有效固定資產類別 payload\n  WHEN 執行 EventFixedAssetCategoryCreated\n  THEN 建立類別記錄並可查詢到正確欄位", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		svc := newSvc(db)
		raw, err := json.Marshal(payload.FixedAssetCategoryCreatedPayload{
			Name:                         "電腦設備",
			AssetAccountID:               "1601-01",
			AccumDepreciationAccountID:   "1601-02",
			DepreciationExpenseAccountID: "5601-01",
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     "fa-cat-agg-1",
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryCreated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())

		// CategoryUuid = EventUuid（建立事件的 UUID）
		catUUID := result.EventUuid
		row := queryFixedAssetCategory(t, db, catUUID)
		Expect(row.Name).To(Equal("電腦設備"))
		Expect(row.AssetAccountID).To(Equal("1601-01"))
		Expect(row.AccumDepreciationAccountID).To(Equal("1601-02"))
		Expect(row.DepreciationExpenseAccountID).To(Equal("5601-01"))
		Expect(row.IsActive).To(BeTrue())
	})
})

var _ = Describe("EventFixedAssetCategoryUpdated", func() {
	It("GIVEN 已建立的固定資產類別\n  WHEN 執行 EventFixedAssetCategoryUpdated\n  THEN 類別欄位更新為新值", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		svc := newSvc(db)

		// 先建立類別
		raw, _ := json.Marshal(payload.FixedAssetCategoryCreatedPayload{
			Name:                         "辦公設備",
			AssetAccountID:               "1601-01",
			AccumDepreciationAccountID:   "1601-02",
			DepreciationExpenseAccountID: "5601-01",
		})
		created, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     "fa-cat-agg-2",
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryCreated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())
		catUUID := created.EventUuid

		// 使用 EventUuid 作為後續操作的 AggregateID
		raw, _ = json.Marshal(payload.FixedAssetCategoryUpdatedPayload{
			CategoryUUID:                 catUUID,
			Name:                         "辦公設備（修正）",
			AssetAccountID:               "1602-01",
			AccumDepreciationAccountID:   "1602-02",
			DepreciationExpenseAccountID: "5602-01",
		})
		_, err = svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryUpdated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())

		row := queryFixedAssetCategory(t, db, catUUID)
		Expect(row.Name).To(Equal("辦公設備（修正）"))
		Expect(row.AssetAccountID).To(Equal("1602-01"))
		Expect(row.IsActive).To(BeTrue())
	})
})

var _ = Describe("EventFixedAssetCategoryDeleted", func() {
	It("GIVEN 已建立的固定資產類別\n  WHEN 執行 EventFixedAssetCategoryDeleted\n  THEN 類別軟刪除（is_active = false）", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		svc := newSvc(db)

		raw, _ := json.Marshal(payload.FixedAssetCategoryCreatedPayload{
			Name:                         "機械設備",
			AssetAccountID:               "1601-01",
			AccumDepreciationAccountID:   "1601-02",
			DepreciationExpenseAccountID: "5601-01",
		})
		created, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     "fa-cat-agg-3",
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryCreated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())
		catUUID := created.EventUuid

		raw, _ = json.Marshal(payload.FixedAssetCategoryDeletedPayload{CategoryUUID: catUUID})
		_, err = svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryDeleted.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())

		row := queryFixedAssetCategory(t, db, catUUID)
		Expect(row.IsActive).To(BeFalse())
	})

	It("GIVEN 已軟刪除的固定資產類別\n  WHEN 再次執行 EventFixedAssetCategoryDeleted\n  THEN 回傳錯誤", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)

		svc := newSvc(db)

		raw, _ := json.Marshal(payload.FixedAssetCategoryCreatedPayload{
			Name:                         "車輛",
			AssetAccountID:               "1601-01",
			AccumDepreciationAccountID:   "1601-02",
			DepreciationExpenseAccountID: "5601-01",
		})
		created, _ := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     "fa-cat-agg-4",
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryCreated.Enum(),
			Payload:         raw,
		})
		catUUID := created.EventUuid

		raw, _ = json.Marshal(payload.FixedAssetCategoryDeletedPayload{CategoryUUID: catUUID})
		svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 0,
			EventType:       event_types.EventFixedAssetCategoryDeleted.Enum(),
			Payload:         raw,
		})

		// 再次刪除應失敗（Pipeline 驗證 is_active = false）
		_, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 1,
			EventType:       event_types.EventFixedAssetCategoryDeleted.Enum(),
			Payload:         raw,
		})
		Expect(err).To(HaveOccurred())
	})
})

// ─────────────────────────────────────────
// PrepaidCategory
// ─────────────────────────────────────────

var _ = Describe("EventPrepaidCategoryCreated", func() {
	It("GIVEN 有效預付費用類別 payload\n  WHEN 執行 EventPrepaidCategoryCreated\n  THEN 建立類別記錄並可查詢到正確欄位", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		svc := newSvc(db)
		raw, err := json.Marshal(payload.PrepaidCategoryCreatedPayload{
			Name:             "保險費",
			AccountID:        "1104-01",
			ExpenseAccountID: "5201-01",
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     "prepaid-cat-agg-1",
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryCreated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())

		catUUID := result.EventUuid
		row := queryPrepaidCategory(t, db, catUUID)
		Expect(row.Name).To(Equal("保險費"))
		Expect(row.AccountID).To(Equal("1104-01"))
		Expect(row.ExpenseAccountID).To(Equal("5201-01"))
		Expect(row.IsActive).To(BeTrue())
	})
})

var _ = Describe("EventPrepaidCategoryUpdated", func() {
	It("GIVEN 已建立的預付費用類別\n  WHEN 執行 EventPrepaidCategoryUpdated\n  THEN 類別欄位更新為新值", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		svc := newSvc(db)

		raw, _ := json.Marshal(payload.PrepaidCategoryCreatedPayload{
			Name: "租金", AccountID: "1104-02", ExpenseAccountID: "5202-01",
		})
		created, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     "prepaid-cat-agg-2",
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryCreated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())
		catUUID := created.EventUuid

		raw, _ = json.Marshal(payload.PrepaidCategoryUpdatedPayload{
			CategoryUUID: catUUID, Name: "租金（修正）", AccountID: "1104-03", ExpenseAccountID: "5202-02",
		})
		_, err = svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryUpdated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())

		row := queryPrepaidCategory(t, db, catUUID)
		Expect(row.Name).To(Equal("租金（修正）"))
		Expect(row.AccountID).To(Equal("1104-03"))
		Expect(row.IsActive).To(BeTrue())
	})
})

var _ = Describe("EventPrepaidCategoryDeleted", func() {
	It("GIVEN 已建立的預付費用類別\n  WHEN 執行 EventPrepaidCategoryDeleted\n  THEN 類別軟刪除（is_active = false）", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		svc := newSvc(db)

		raw, _ := json.Marshal(payload.PrepaidCategoryCreatedPayload{
			Name: "訂閱費", AccountID: "1104-04", ExpenseAccountID: "5203-01",
		})
		created, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     "prepaid-cat-agg-3",
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryCreated.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())
		catUUID := created.EventUuid

		raw, _ = json.Marshal(payload.PrepaidCategoryDeletedPayload{CategoryUUID: catUUID})
		_, err = svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryDeleted.Enum(),
			Payload:         raw,
		})
		Expect(err).NotTo(HaveOccurred())

		row := queryPrepaidCategory(t, db, catUUID)
		Expect(row.IsActive).To(BeFalse())
	})

	It("GIVEN 已軟刪除的預付費用類別\n  WHEN 再次執行 EventPrepaidCategoryDeleted\n  THEN 回傳錯誤", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)

		svc := newSvc(db)

		raw, _ := json.Marshal(payload.PrepaidCategoryCreatedPayload{
			Name: "授權費", AccountID: "1104-05", ExpenseAccountID: "5204-01",
		})
		created, _ := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     "prepaid-cat-agg-4",
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryCreated.Enum(),
			Payload:         raw,
		})
		catUUID := created.EventUuid

		raw, _ = json.Marshal(payload.PrepaidCategoryDeletedPayload{CategoryUUID: catUUID})
		svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 0,
			EventType:       event_types.EventPrepaidCategoryDeleted.Enum(),
			Payload:         raw,
		})

		// 再次刪除應失敗（Pipeline 驗證 is_active = false）
		_, err := svc.Append(testCtx(), cmd.AppendCmd{
			AggregateType:   enums.AggregatePrepaidCategory.Enum(),
			AggregateID:     catUUID,
			ExpectedVersion: 1,
			EventType:       event_types.EventPrepaidCategoryDeleted.Enum(),
			Payload:         raw,
		})
		Expect(err).To(HaveOccurred())
	})
})
