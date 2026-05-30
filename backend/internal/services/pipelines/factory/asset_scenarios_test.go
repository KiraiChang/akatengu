package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"errors"

	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// Scenario types
// ─────────────────────────────────────────

type assetPurchasedScenario struct {
	given   string
	when    string
	then    string
	payload payload.AssetPurchasedPayload

	period    *projection.PeriodClosing
	periodErr error
	ledger    *projection.LedgerAccount
	ledgerErr error

	wantErrContain string
	checkState     func(*state.AssetPurchasedState)
}

type assetDepreciatedScenario struct {
	given   string
	when    string
	then    string
	payload payload.AssetDepreciatedPayload

	period    *projection.PeriodClosing
	periodErr error
	asset     *projection.FixedAsset
	assetErr  error

	wantErrContain string
	checkState     func(*state.AssetDepreciatedState)
}

type assetDisposedScenario struct {
	given   string
	when    string
	then    string
	payload payload.AssetDisposedPayload

	asset          *projection.FixedAsset
	assetErr       error
	proceedsLedger *projection.LedgerAccount
	ledgerErr      error

	wantErrContain string
	checkState     func(*state.AssetDisposedState)
}

// ─────────────────────────────────────────
// Scenario builders（在 Describe 內呼叫，避免 package init 觸發 enum）
// ─────────────────────────────────────────

func buildAssetPurchasedScenarios() []assetPurchasedScenario {
	return []assetPurchasedScenario{
		{
			given:          "payload 缺少必填欄位（name 為空）",
			when:           "執行 AssetPurchased pipeline",
			then:           "回傳包含 name is required 的驗證錯誤",
			payload:        payload.AssetPurchasedPayload{},
			wantErrContain: "name is required",
		},
		{
			given:          "period 不存在（GetByPeriod 回傳 nil）",
			when:           "執行 AssetPurchased pipeline",
			then:           "回傳 period is not exists 錯誤",
			payload:        validPurchasedCashPayload(),
			period:         nil,
			wantErrContain: "is not exists",
		},
		{
			given:          "period 狀態非 OPEN（已結帳）",
			when:           "執行 AssetPurchased pipeline",
			then:           "回傳 is already closed 錯誤",
			payload:        validPurchasedCashPayload(),
			period:         closedPeriod(),
			wantErrContain: "is already closed",
		},
		{
			given:          "CASH 付款，GetLedgerByUuid 回傳錯誤",
			when:           "執行 AssetPurchased pipeline",
			then:           "回傳 ledger query 錯誤",
			payload:        validPurchasedCashPayload(),
			period:         openPeriod(),
			ledgerErr:      errors.New("db error"),
			wantErrContain: "db error",
		},
		{
			given:          "CASH 付款，GetLedgerByUuid 回傳 nil",
			when:           "執行 AssetPurchased pipeline",
			then:           "回傳 ledger not found 錯誤",
			payload:        validPurchasedCashPayload(),
			period:         openPeriod(),
			ledger:         nil,
			wantErrContain: "ledger not found",
		},
		{
			given:   "CASH 付款且 Ledger 正常",
			when:    "執行 AssetPurchased pipeline",
			then:    "State.Ledger 已設定，Transaction 不為空",
			payload: validPurchasedCashPayload(),
			period:  openPeriod(),
			ledger:  testLedger(),
			checkState: func(st *state.AssetPurchasedState) {
				Expect(st.Ledger).NotTo(BeNil())
				Expect(st.Transaction.TransactionDate).To(Equal("2026-05-01"))
				Expect(len(st.Transaction.Entries)).To(Equal(2))
			},
		},
		{
			given:   "LEASE 付款（不需查詢 Ledger）",
			when:    "執行 AssetPurchased pipeline",
			then:    "State.Ledger 為 nil，Transaction 包含資產科目與負債科目",
			payload: validPurchasedLeasePayload(),
			period:  openPeriod(),
			checkState: func(st *state.AssetPurchasedState) {
				Expect(st.Ledger).To(BeNil())
				Expect(len(st.Transaction.Entries)).To(Equal(2))
			},
		},
	}
}

func buildAssetDepreciatedScenarios() []assetDepreciatedScenario {
	return []assetDepreciatedScenario{
		{
			given:          "payload 缺少必填欄位（AssetUUID 為空）",
			when:           "執行 AssetDepreciated pipeline",
			then:           "回傳 asset_uuid is required 驗證錯誤",
			payload:        payload.AssetDepreciatedPayload{},
			wantErrContain: "asset_uuid is required",
		},
		{
			given:          "period 不存在",
			when:           "執行 AssetDepreciated pipeline",
			then:           "回傳 period is not exists 錯誤",
			payload:        payload.AssetDepreciatedPayload{AssetUUID: "uuid-1", PeriodDate: "2026-05"},
			period:         nil,
			wantErrContain: "is not exists",
		},
		{
			given:          "GetFixedAssetByUUID 回傳 nil（資產不存在）",
			when:           "執行 AssetDepreciated pipeline",
			then:           "回傳 fixed asset not found 錯誤",
			payload:        payload.AssetDepreciatedPayload{AssetUUID: "uuid-1", PeriodDate: "2026-05"},
			period:         openPeriod(),
			asset:          nil,
			wantErrContain: "fixed asset not found",
		},
		{
			given:          "資產狀態非 ACTIVE",
			when:           "執行 AssetDepreciated pipeline",
			then:           "回傳 fixed asset is not active 錯誤",
			payload:        payload.AssetDepreciatedPayload{AssetUUID: "uuid-1", PeriodDate: "2026-05"},
			period:         openPeriod(),
			asset:          disposedAsset(),
			wantErrContain: "fixed asset is not active",
		},
		{
			given:          "資產已完全折舊（DepreciatedPeriods >= UsefulLifeMonths）",
			when:           "執行 AssetDepreciated pipeline",
			then:           "回傳 fixed asset is already fully depreciated 錯誤",
			payload:        payload.AssetDepreciatedPayload{AssetUUID: "uuid-1", PeriodDate: "2026-05"},
			period:         openPeriod(),
			asset:          fullyDepreciatedAsset(),
			wantErrContain: "fixed asset is already fully depreciated",
		},
		{
			given:   "所有條件正常",
			when:    "執行 AssetDepreciated pipeline",
			then:    "State.Asset 已設定，Transaction 不為空",
			payload: payload.AssetDepreciatedPayload{AssetUUID: "uuid-1", PeriodDate: "2026-05"},
			period:  openPeriod(),
			asset:   activeAsset(),
			checkState: func(st *state.AssetDepreciatedState) {
				Expect(st.Asset).NotTo(BeNil())
				Expect(st.Transaction.TransactionDate).To(Equal("2026-05-01"))
				Expect(len(st.Transaction.Entries)).To(Equal(2))
			},
		},
	}
}

func buildAssetDisposedScenarios() []assetDisposedScenario {
	proceedsLedgerUUID := "ledger-uuid-1"
	return []assetDisposedScenario{
		{
			given:          "payload 缺少必填欄位（AssetUUID 為空）",
			when:           "執行 AssetDisposed pipeline",
			then:           "回傳 asset_uuid is required 驗證錯誤",
			payload:        payload.AssetDisposedPayload{},
			wantErrContain: "asset_uuid is required",
		},
		{
			given:          "GetFixedAssetByUUID 回傳 nil（資產不存在）",
			when:           "執行 AssetDisposed pipeline",
			then:           "回傳 fixed asset not found 錯誤",
			payload:        validDisposedPayloadNoProceeds(),
			asset:          nil,
			wantErrContain: "fixed asset not found",
		},
		{
			given:          "資產狀態非 ACTIVE",
			when:           "執行 AssetDisposed pipeline",
			then:           "回傳 fixed asset is not active 錯誤",
			payload:        validDisposedPayloadNoProceeds(),
			asset:          disposedAsset(),
			wantErrContain: "fixed asset is not active",
		},
		{
			given: "有處分收益且 ProceedsLedgerUUID 已設定，但 GetLedgerByUuid 回傳錯誤",
			when:  "執行 AssetDisposed pipeline",
			then:  "回傳 ledger query 錯誤",
			payload: payload.AssetDisposedPayload{
				AssetUUID:          "asset-uuid-1",
				DisposalDate:       "2026-06-30",
				Proceeds:           decimal.NewFromInt(50000),
				ProceedsLedgerUUID: &proceedsLedgerUUID,
				GainAccountID:      "4205",
				LossAccountID:      "5601",
			},
			asset:          activeAsset(),
			ledgerErr:      errors.New("db error"),
			wantErrContain: "db error",
		},
		{
			given: "有處分收益且 ProceedsLedgerUUID 已設定，但 Ledger 為 nil",
			when:  "執行 AssetDisposed pipeline",
			then:  "回傳 proceeds ledger not found 錯誤",
			payload: payload.AssetDisposedPayload{
				AssetUUID:          "asset-uuid-1",
				DisposalDate:       "2026-06-30",
				Proceeds:           decimal.NewFromInt(50000),
				ProceedsLedgerUUID: &proceedsLedgerUUID,
				GainAccountID:      "4205",
				LossAccountID:      "5601",
			},
			asset:          activeAsset(),
			proceedsLedger: nil,
			wantErrContain: "proceeds ledger not found",
		},
		{
			given: "有處分收益且 Ledger 正常",
			when:  "執行 AssetDisposed pipeline",
			then:  "State.ProceedsLedger 已設定，Transaction 含收益科目分錄",
			payload: payload.AssetDisposedPayload{
				AssetUUID:          "asset-uuid-1",
				DisposalDate:       "2026-06-30",
				Proceeds:           decimal.NewFromInt(50000),
				ProceedsLedgerUUID: &proceedsLedgerUUID,
				GainAccountID:      "4205",
				LossAccountID:      "5601",
			},
			asset:          activeAsset(),
			proceedsLedger: testLedger(),
			checkState: func(st *state.AssetDisposedState) {
				Expect(st.ProceedsLedger).NotTo(BeNil())
				Expect(len(st.Transaction.Entries)).To(BeNumerically(">=", 3))
			},
		},
		{
			given:   "無處分收益（Proceeds = 0，不查詢 Ledger）",
			when:    "執行 AssetDisposed pipeline",
			then:    "State.ProceedsLedger 為 nil，Transaction 不含收款帳戶分錄",
			payload: validDisposedPayloadNoProceeds(),
			asset:   activeAsset(),
			checkState: func(st *state.AssetDisposedState) {
				Expect(st.ProceedsLedger).To(BeNil())
				Expect(len(st.Transaction.Entries)).To(BeNumerically(">=", 2))
				for _, e := range st.Transaction.Entries {
					Expect(e.LedgerId).To(BeNil())
				}
			},
		},
	}
}

// ─────────────────────────────────────────
// Fixture helpers
// ─────────────────────────────────────────

func validPurchasedCashPayload() payload.AssetPurchasedPayload {
	ledgerUUID := "ledger-uuid-1"
	return payload.AssetPurchasedPayload{
		Name:                         "辦公電腦",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         decimal.NewFromInt(100000),
		ResidualValue:                decimal.Zero,
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
		LedgerUUID:                   &ledgerUUID,
		PurchaseDate:                 "2026-05-01",
	}
}

func validPurchasedLeasePayload() payload.AssetPurchasedPayload {
	return payload.AssetPurchasedPayload{
		Name:                         "租賃資產",
		AssetAccountID:               "1202-01",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         decimal.NewFromInt(240000),
		ResidualValue:                decimal.Zero,
		UsefulLifeMonths:             24,
		PaymentType:                  enums.AssetPaymentTypeLease.Enum(),
		LiabilityAccountID:           "2202-02",
		PurchaseDate:                 "2026-05-01",
	}
}

func validDisposedPayloadNoProceeds() payload.AssetDisposedPayload {
	return payload.AssetDisposedPayload{
		AssetUUID:     "asset-uuid-1",
		DisposalDate:  "2026-06-30",
		Proceeds:      decimal.Zero,
		GainAccountID: "4205",
		LossAccountID: "5601",
	}
}

func openPeriod() *projection.PeriodClosing {
	return &projection.PeriodClosing{
		PeriodStart: "2026-05-01",
		PeriodEnd:   "2026-05-31",
		Status:      enums.PeriodTypeStatusOpen.Enum(),
	}
}

func closedPeriod() *projection.PeriodClosing {
	return &projection.PeriodClosing{
		PeriodStart: "2026-05-01",
		PeriodEnd:   "2026-05-31",
		Status:      enums.PeriodTypeStatusClosed.Enum(),
	}
}

func testLedger() *projection.LedgerAccount {
	return &projection.LedgerAccount{
		LedgerId:  1,
		AccountId: "1101-02",
		Name:      "台銀帳戶",
	}
}

func activeAsset() *projection.FixedAsset {
	return &projection.FixedAsset{
		ID:                           1,
		AssetUuid:                    "asset-uuid-1",
		Name:                         "辦公電腦",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         decimal.NewFromInt(120000),
		ResidualValue:                decimal.Zero,
		TotalDepreciated:             decimal.Zero,
		UsefulLifeMonths:             60,
		DepreciatedPeriods:           0,
		Status:                       enums.FixedAssetStatusActive.Enum(),
	}
}

func disposedAsset() *projection.FixedAsset {
	a := activeAsset()
	a.Status = enums.FixedAssetStatusDisposed.Enum()
	return a
}

func fullyDepreciatedAsset() *projection.FixedAsset {
	a := activeAsset()
	a.DepreciatedPeriods = a.UsefulLifeMonths
	return a
}
