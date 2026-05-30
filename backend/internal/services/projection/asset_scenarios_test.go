package projection_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"

	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// Scenario types
// ─────────────────────────────────────────

type assetProjPurchasedScenario struct {
	given       string
	when        string
	then        string
	paymentType string // "CASH" | "LEASE"
	checkDB     func(db projTestDB)
}

type assetProjDepreciatedScenario struct {
	given          string
	when           string
	then           string
	wantErrContain string
	checkDB        func(db projTestDB)
}

type assetProjDisposedScenario struct {
	given          string
	when           string
	then           string
	wantErrContain string
	checkDB        func(db projTestDB)
}

// ─────────────────────────────────────────
// Scenario builders
// ─────────────────────────────────────────

func buildPurchasedScenarios() []assetProjPurchasedScenario {
	return []assetProjPurchasedScenario{
		{
			given:       "LEASE 付款方式，所需帳戶與期間已建立",
			when:        "Append EventAssetPurchased",
			then:        "DB 新增 fixed_assets 記錄，payment_type=LEASE，status=ACTIVE",
			paymentType: "LEASE",
			checkDB: func(db projTestDB) {
				asset := db.queryLastAsset()
				Expect(asset.paymentType).To(Equal("LEASE"))
				Expect(asset.status).To(Equal("ACTIVE"))
				Expect(asset.cost.Equal(dec("120000"))).To(BeTrue())
			},
		},
		{
			given:       "CASH 付款方式，Ledger 已建立",
			when:        "Append EventAssetPurchased",
			then:        "DB 新增 fixed_assets 記錄，payment_type=CASH，status=ACTIVE",
			paymentType: "CASH",
			checkDB: func(db projTestDB) {
				asset := db.queryLastAsset()
				Expect(asset.paymentType).To(Equal("CASH"))
				Expect(asset.status).To(Equal("ACTIVE"))
				Expect(asset.cost.Equal(dec("120000"))).To(BeTrue())
			},
		},
	}
}

func buildDepreciatedScenarios() []assetProjDepreciatedScenario {
	return []assetProjDepreciatedScenario{
		{
			given: "有效 active 資產，期間 OPEN",
			when:  "Append EventAssetDepreciated",
			then:  "depreciated_periods +1，total_depreciated 增加正確金額",
			checkDB: func(db projTestDB) {
				asset := db.queryLastAsset()
				Expect(asset.depreciatedPeriods).To(Equal(int64(1)))
				Expect(asset.totalDepreciated.Equal(dec("2000"))).To(BeTrue())
			},
		},
		{
			given:          "已完全折舊的資產（DepreciatedPeriods >= UsefulLifeMonths）",
			when:           "Append EventAssetDepreciated",
			then:           "回傳 already fully depreciated 錯誤，DB 無異動",
			wantErrContain: "fully depreciated",
		},
	}
}

func buildDisposedScenarios() []assetProjDisposedScenario {
	return []assetProjDisposedScenario{
		{
			given: "有效 active 資產",
			when:  "Append EventAssetDisposed",
			then:  "DB 中資產 status=DISPOSED，disposal_date 設定正確",
			checkDB: func(db projTestDB) {
				asset := db.queryLastAsset()
				Expect(asset.status).To(Equal("DISPOSED"))
				Expect(asset.disposalDate).To(Equal("2026-06-30"))
			},
		},
		{
			given:          "不存在的 AssetUUID",
			when:           "Append EventAssetDisposed",
			then:           "回傳 fixed asset not found 錯誤",
			wantErrContain: "fixed asset not found",
		},
	}
}

// ─────────────────────────────────────────
// Payload helpers（在 Describe 時呼叫，enum registry 已就緒）
// ─────────────────────────────────────────

func leasePurchasedPayload() payload.AssetPurchasedPayload {
	return payload.AssetPurchasedPayload{
		Name:                         "辦公設備",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("120000"),
		ResidualValue:                decimal.Zero,
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeLease.Enum(),
		LiabilityAccountID:           "2202-02",
		PurchaseDate:                 "2026-05-01",
	}
}

func cashPurchasedPayload(ledgerUUID string) payload.AssetPurchasedPayload {
	return payload.AssetPurchasedPayload{
		Name:                         "辦公設備",
		AssetAccountID:               "1201-04",
		AccumDepreciationAccountID:   "1201-99",
		DepreciationExpenseAccountID: "5501-03",
		Cost:                         dec("120000"),
		ResidualValue:                decimal.Zero,
		UsefulLifeMonths:             60,
		PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
		LedgerUUID:                   &ledgerUUID,
		PurchaseDate:                 "2026-05-01",
	}
}

func depreciatedPayload(assetUUID string) payload.AssetDepreciatedPayload {
	return payload.AssetDepreciatedPayload{AssetUUID: assetUUID, PeriodDate: "2026-06"}
}

func disposedPayload(assetUUID string) payload.AssetDisposedPayload {
	return payload.AssetDisposedPayload{
		AssetUUID:     assetUUID,
		DisposalDate:  "2026-06-30",
		Proceeds:      decimal.Zero,
		GainAccountID: "4205",
		LossAccountID: "5601",
	}
}

func dec(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
