package services_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/testutil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─────────────────────────────────────────
// Prepaid
// ─────────────────────────────────────────

var _ = Describe("EventPrepaidCreated", func() {
	It("GIVEN 有效預付費用 payload\n  WHEN 執行 EventPrepaidCreated\n  THEN 建立預付記錄、寫入 2 筆分錄（借方標記 OPERATING）並更新即時餘額", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertPeriodOpen(t, db, 1, "2026-05-01")
		insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "prepaid-1")
		a.do(event_types.EventPrepaidCreated.Enum(), payload.PrepaidCreatedPayload{
			AccountID:        "1104-01",
			ExpenseAccountID: "5201-01",
			LedgerUUID:       testLedgerUUID(1),
			Name:             "人壽保險費 2026",
			TotalAmount:      dec("12000"),
			Periods:          12,
			StartDate:        "2026-05-01",
		})

		txn := queryLastTxn(t, db)
		Expect(txn.Date).To(Equal("2026-05-01"))
		Expect(txn.Status).To(Equal("ACTIVE"))

		entries := queryEntries(t, db, txn.TxnID)
		Expect(entries).To(HaveLen(2))
		assertEntry(t, entries, 0, "1104-01", nil, dec("12000"), dec("0"), cfPtr("OPERATING"))
		assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("0"), dec("12000"), nil)

		assertAccountRB(t, db, "1104-01", dec("12000"), dec("0"))
		assertLedgerRB(t, db, 1, dec("0"), dec("12000"))
	})
})

var _ = Describe("EventPrepaidAmortized", func() {
	It("GIVEN 已建立的預付費用\n  WHEN 執行 EventPrepaidAmortized\n  THEN 寫入第一期攤提分錄（貸方標記 OPERATING）並更新即時餘額", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertPeriodOpen(t, db, 1, "2026-05-01")
		insertPeriodOpen(t, db, 2, "2026-06-01")
		insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "prepaid-amort-1")
		a.do(event_types.EventPrepaidCreated.Enum(), payload.PrepaidCreatedPayload{
			AccountID:        "1104-01",
			ExpenseAccountID: "5201-01",
			LedgerUUID:       testLedgerUUID(1),
			Name:             "人壽保險費 2026",
			TotalAmount:      dec("12000"),
			Periods:          12,
			StartDate:        "2026-05-01",
		})
		prepaidUUID := queryLastPrepaidUUID(t, db)

		a.do(event_types.EventPrepaidAmortized.Enum(), payload.PrepaidAmortizedPayload{
			PrepaidUUID: prepaidUUID,
			PeriodDate:  "2026-06",
		})

		// AmortizationAmount(12000, 12, 0, 0) = 12000/12 truncate(6) = 1000
		wantAmt := dec("1000")
		txn := queryLastTxn(t, db)
		Expect(txn.Date).To(Equal("2026-06-01"))

		entries := queryEntries(t, db, txn.TxnID)
		Expect(entries).To(HaveLen(2))
		// Dr 5201-01 1000, Cr 1104-01 1000 [OPERATING]
		assertEntry(t, entries, 0, "5201-01", nil, wantAmt, dec("0"), nil)
		assertEntry(t, entries, 1, "1104-01", nil, dec("0"), wantAmt, cfPtr("OPERATING"))

		assertAccountRB(t, db, "1104-01", dec("12000"), wantAmt)
		assertAccountRB(t, db, "5201-01", wantAmt, dec("0"))
	})
})

var _ = Describe("EventPrepaidDisposed", func() {
	It("GIVEN 已攤提一期的預付費用\n  WHEN 執行 EventPrepaidDisposed\n  THEN 一次認列剩餘 11000 並使科目歸零", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertPeriodOpen(t, db, 1, "2026-05-01")
		insertPeriodOpen(t, db, 2, "2026-06-01")
		insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "prepaid-disp-1")
		a.do(event_types.EventPrepaidCreated.Enum(), payload.PrepaidCreatedPayload{
			AccountID:        "1104-01",
			ExpenseAccountID: "5201-01",
			LedgerUUID:       testLedgerUUID(1),
			Name:             "人壽保險費 2026",
			TotalAmount:      dec("12000"),
			Periods:          12,
			StartDate:        "2026-05-01",
		})
		prepaidUUID := queryLastPrepaidUUID(t, db)

		a.do(event_types.EventPrepaidAmortized.Enum(), payload.PrepaidAmortizedPayload{
			PrepaidUUID: prepaidUUID,
			PeriodDate:  "2026-06",
		})
		a.do(event_types.EventPrepaidDisposed.Enum(), payload.PrepaidDisposedPayload{
			PrepaidUUID:  prepaidUUID,
			DisposalDate: "2026-06-15",
		})

		remaining := dec("11000")
		txn := queryLastTxn(t, db)
		Expect(txn.Date).To(Equal("2026-06-15"))

		entries := queryEntries(t, db, txn.TxnID)
		Expect(entries).To(HaveLen(2))
		// Dr 5201-01 11000, Cr 1104-01 11000 [OPERATING]
		assertEntry(t, entries, 0, "5201-01", nil, remaining, dec("0"), nil)
		assertEntry(t, entries, 1, "1104-01", nil, dec("0"), remaining, cfPtr("OPERATING"))

		assertAccountRB(t, db, "1104-01", dec("12000"), dec("12000"))
		assertAccountRB(t, db, "5201-01", dec("12000"), dec("0"))
	})
})

// ─────────────────────────────────────────
// Fixed Asset
// ─────────────────────────────────────────

var _ = Describe("EventAssetPurchased", func() {
	Context("CASH 付款", func() {
		It("GIVEN CASH 付款 payload\n  WHEN 執行 EventAssetPurchased\n  THEN 寫入 2 筆分錄（借方標記 INVESTING）並更新即時餘額", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertPeriodOpen(t, db, 1, "2026-05-01")
			insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "asset-cash-1")
			ledgerUUID := testLedgerUUID(1)
			a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
				Name:                         "辦公電腦",
				AssetAccountID:               "1201-04",
				AccumDepreciationAccountID:   "1201-99",
				DepreciationExpenseAccountID: "5501-03",
				Cost:                         dec("100000"),
				ResidualValue:                dec("0"),
				UsefulLifeMonths:             60,
				PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
				LedgerUUID:                   &ledgerUUID,
				PurchaseDate:                 "2026-05-01",
			})

			txn := queryLastTxn(t, db)
			Expect(txn.Date).To(Equal("2026-05-01"))

			entries := queryEntries(t, db, txn.TxnID)
			Expect(entries).To(HaveLen(2))
			// Dr 1201-04 100000 [INVESTING], Cr 1101-02 100000 (ledger=1)
			assertEntry(t, entries, 0, "1201-04", nil, dec("100000"), dec("0"), cfPtr("INVESTING"))
			assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("0"), dec("100000"), nil)

			assertAccountRB(t, db, "1201-04", dec("100000"), dec("0"))
			assertLedgerRB(t, db, 1, dec("0"), dec("100000"))
		})
	})

	Context("LEASE 付款", func() {
		It("GIVEN LEASE 付款 payload\n  WHEN 執行 EventAssetPurchased\n  THEN 寫入 2 筆分錄（無 CF 標記）並更新即時餘額", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertPeriodOpen(t, db, 1, "2026-05-01")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "asset-lease-1")
			a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
				Name:                         "租賃住宅",
				AssetAccountID:               "1202-01",
				AccumDepreciationAccountID:   "1201-99",
				DepreciationExpenseAccountID: "5501-03",
				Cost:                         dec("240000"),
				ResidualValue:                dec("0"),
				UsefulLifeMonths:             24,
				PaymentType:                  enums.AssetPaymentTypeLease.Enum(),
				LiabilityAccountID:           "2202-02",
				PurchaseDate:                 "2026-05-01",
			})

			txn := queryLastTxn(t, db)
			Expect(txn.Date).To(Equal("2026-05-01"))

			entries := queryEntries(t, db, txn.TxnID)
			Expect(entries).To(HaveLen(2))
			// Dr 1202-01 240000 (no CF), Cr 2202-02 240000 (no CF)
			assertEntry(t, entries, 0, "1202-01", nil, dec("240000"), dec("0"), nil)
			assertEntry(t, entries, 1, "2202-02", nil, dec("0"), dec("240000"), nil)

			assertAccountRB(t, db, "1202-01", dec("240000"), dec("0"))
			assertAccountRB(t, db, "2202-02", dec("0"), dec("240000"))
		})
	})
})

var _ = Describe("EventAssetDepreciated", func() {
	It("GIVEN 已現金購入的固定資產\n  WHEN 執行 EventAssetDepreciated\n  THEN 寫入折舊分錄（貸方標記 OPERATING）並更新即時餘額", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertPeriodOpen(t, db, 1, "2026-05-01")
		insertPeriodOpen(t, db, 2, "2026-06-01")
		insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "asset-depr-1")
		ledgerUUID := testLedgerUUID(1)
		a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
			Name:                         "辦公電腦",
			AssetAccountID:               "1201-04",
			AccumDepreciationAccountID:   "1201-99",
			DepreciationExpenseAccountID: "5501-03",
			Cost:                         dec("120000"),
			ResidualValue:                dec("0"),
			UsefulLifeMonths:             60,
			PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
			LedgerUUID:                   &ledgerUUID,
			PurchaseDate:                 "2026-05-01",
		})
		assetUUID := queryLastAssetUUID(t, db)

		a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{
			AssetUUID:  assetUUID,
			PeriodDate: "2026-06",
		})

		// DepreciationAmount(120000, 0, 60, 0, 0) = 120000/60 = 2000
		wantAmt := dec("2000")
		txn := queryLastTxn(t, db)
		Expect(txn.Date).To(Equal("2026-06-01"))

		entries := queryEntries(t, db, txn.TxnID)
		Expect(entries).To(HaveLen(2))
		// Dr 5501-03, Cr 1201-99 [OPERATING]
		assertEntry(t, entries, 0, "5501-03", nil, wantAmt, dec("0"), nil)
		assertEntry(t, entries, 1, "1201-99", nil, dec("0"), wantAmt, cfPtr("OPERATING"))

		assertAccountRB(t, db, "1201-99", dec("0"), wantAmt)
		assertAccountRB(t, db, "5501-03", wantAmt, dec("0"))
	})
})

var _ = Describe("EventAssetDisposed", func() {
	Context("處分利得（收益 > 帳面價值）", func() {
		It("GIVEN 折舊 2 期後處分，收益 124000 帳面 116000\n  WHEN 執行 EventAssetDisposed\n  THEN 寫入 4 筆分錄並認列處分利得 8000", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertPeriodOpen(t, db, 1, "2026-05-01")
			insertPeriodOpen(t, db, 2, "2026-06-01")
			insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "asset-gain-1")
			ledgerUUID := testLedgerUUID(1)
			a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
				Name:                         "辦公電腦",
				AssetAccountID:               "1201-04",
				AccumDepreciationAccountID:   "1201-99",
				DepreciationExpenseAccountID: "5501-03",
				Cost:                         dec("120000"),
				ResidualValue:                dec("0"),
				UsefulLifeMonths:             60,
				PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
				LedgerUUID:                   &ledgerUUID,
				PurchaseDate:                 "2026-05-01",
			})
			assetUUID := queryLastAssetUUID(t, db)

			// 2 depreciation periods → totalDepreciated = 4000
			a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetUUID: assetUUID, PeriodDate: "2026-05"})
			a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetUUID: assetUUID, PeriodDate: "2026-06"})

			// bookValue = 120000 - 4000 = 116000; proceeds = 124000; gain = 8000
			a.do(event_types.EventAssetDisposed.Enum(), payload.AssetDisposedPayload{
				AssetUUID:          assetUUID,
				DisposalDate:       "2026-06-30",
				Proceeds:           dec("124000"),
				ProceedsLedgerUUID: &ledgerUUID,
				GainAccountID:      "4205",
				LossAccountID:      "5601",
			})

			txn := queryLastTxn(t, db)
			Expect(txn.Date).To(Equal("2026-06-30"))

			entries := queryEntries(t, db, txn.TxnID)
			Expect(entries).To(HaveLen(4))
			// Dr 1201-99 4000 [INVESTING], Cr 1201-04 120000 [INVESTING]
			// Dr 1101-02 124000 (ledger=1), Cr 4205 8000 [INVESTING]
			assertEntry(t, entries, 0, "1201-99", nil, dec("4000"), dec("0"), cfPtr("INVESTING"))
			assertEntry(t, entries, 1, "1201-04", nil, dec("0"), dec("120000"), cfPtr("INVESTING"))
			assertEntry(t, entries, 2, "1101-02", lidPtr(1), dec("124000"), dec("0"), nil)
			assertEntry(t, entries, 3, "4205", nil, dec("0"), dec("8000"), cfPtr("INVESTING"))

			assertAccountRB(t, db, "1201-04", dec("120000"), dec("120000"))
			assertAccountRB(t, db, "1201-99", dec("4000"), dec("4000"))
			assertAccountRB(t, db, "4205", dec("0"), dec("8000"))
			assertLedgerRB(t, db, 1, dec("124000"), dec("120000"))
		})
	})

	Context("處分損失（收益 < 帳面價值）", func() {
		It("GIVEN 折舊 2 期後處分，收益 100000 帳面 116000\n  WHEN 執行 EventAssetDisposed\n  THEN 寫入 4 筆分錄並認列處分損失 16000", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertPeriodOpen(t, db, 1, "2026-05-01")
			insertPeriodOpen(t, db, 2, "2026-06-01")
			insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "asset-loss-1")
			ledgerUUID := testLedgerUUID(1)
			a.do(event_types.EventAssetPurchased.Enum(), payload.AssetPurchasedPayload{
				Name:                         "辦公電腦",
				AssetAccountID:               "1201-04",
				AccumDepreciationAccountID:   "1201-99",
				DepreciationExpenseAccountID: "5501-03",
				Cost:                         dec("120000"),
				ResidualValue:                dec("0"),
				UsefulLifeMonths:             60,
				PaymentType:                  enums.AssetPaymentTypeCash.Enum(),
				LedgerUUID:                   &ledgerUUID,
				PurchaseDate:                 "2026-05-01",
			})
			assetUUID := queryLastAssetUUID(t, db)

			a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetUUID: assetUUID, PeriodDate: "2026-05"})
			a.do(event_types.EventAssetDepreciated.Enum(), payload.AssetDepreciatedPayload{AssetUUID: assetUUID, PeriodDate: "2026-06"})

			// bookValue = 116000; proceeds = 100000; loss = 16000
			a.do(event_types.EventAssetDisposed.Enum(), payload.AssetDisposedPayload{
				AssetUUID:          assetUUID,
				DisposalDate:       "2026-06-30",
				Proceeds:           dec("100000"),
				ProceedsLedgerUUID: &ledgerUUID,
				GainAccountID:      "4205",
				LossAccountID:      "5601",
			})

			entries := queryEntries(t, db, queryLastTxn(t, db).TxnID)
			Expect(entries).To(HaveLen(4))
			// Dr 1201-99 4000 [INVESTING], Cr 1201-04 120000 [INVESTING]
			// Dr 1101-02 100000 (ledger=1), Dr 5601 16000 [INVESTING]
			assertEntry(t, entries, 0, "1201-99", nil, dec("4000"), dec("0"), cfPtr("INVESTING"))
			assertEntry(t, entries, 1, "1201-04", nil, dec("0"), dec("120000"), cfPtr("INVESTING"))
			assertEntry(t, entries, 2, "1101-02", lidPtr(1), dec("100000"), dec("0"), nil)
			assertEntry(t, entries, 3, "5601", nil, dec("16000"), dec("0"), cfPtr("INVESTING"))

			assertAccountRB(t, db, "5601", dec("16000"), dec("0"))
			assertLedgerRB(t, db, 1, dec("100000"), dec("120000"))
		})
	})
})

// ─────────────────────────────────────────
// Installment
// ─────────────────────────────────────────

var _ = Describe("EventInstallmentCreated", func() {
	Context("無息分期", func() {
		It("GIVEN 無息分期 payload\n  WHEN 執行 EventInstallmentCreated\n  THEN 寫入 2 筆分錄（無 CF 標記）並更新即時餘額", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertSysAccounts(t, db)
			insertPeriodOpen(t, db, 1, "2026-05-01")
			insertLedger(t, db, 1, "LOAN", "2201-01")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "install-1")
			a.do(event_types.EventInstallmentCreated.Enum(), payload.InstallmentCreatedPayload{
				Amount:           dec("10000"),
				InstallmentCount: 12,
				StartDate:        "2026-05-01",
				InterestType:     enums.InterestTypeFree.Enum(),
				AccountId:        "1201-04",
				LedgerUuid:       testLedgerUUID(1),
				Memo:             "辦公電腦分期",
			})

			txn := queryLastTxn(t, db)
			Expect(txn.Date).To(Equal("2026-05-01"))

			entries := queryEntries(t, db, txn.TxnID)
			Expect(entries).To(HaveLen(2))
			// FREE: Dr 1201-04 10000, Cr 2201-01 10000 (ledger=1)
			assertEntry(t, entries, 0, "1201-04", nil, dec("10000"), dec("0"), nil)
			assertEntry(t, entries, 1, "2201-01", lidPtr(1), dec("0"), dec("10000"), nil)

			assertAccountRB(t, db, "1201-04", dec("10000"), dec("0"))
			assertLedgerRB(t, db, 1, dec("0"), dec("10000"))
		})
	})
})

var _ = Describe("EventInstallmentPeriodPaid", func() {
	Context("無息分期第一期還款", func() {
		It("GIVEN 已建立的無息分期\n  WHEN 執行 EventInstallmentPeriodPaid（第 1 期）\n  THEN 寫入 2 筆分錄（借方標記 FINANCING）並更新即時餘額", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertSysAccounts(t, db)
			insertPeriodOpen(t, db, 1, "2026-05-01")
			insertPeriodOpen(t, db, 2, "2026-06-01")
			insertLedger(t, db, 1, "LOAN", "2201-01")
			insertLedger(t, db, 2, "BANK_ACCOUNT", "1101-02")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "install-paid-1")
			a.do(event_types.EventInstallmentCreated.Enum(), payload.InstallmentCreatedPayload{
				Amount:           dec("10000"),
				InstallmentCount: 12,
				StartDate:        "2026-05-01",
				InterestType:     enums.InterestTypeFree.Enum(),
				AccountId:        "1201-04",
				LedgerUuid:       testLedgerUUID(1),
				Memo:             "辦公電腦分期",
			})
			installmentUUID := queryLastInstallmentUUID(t, db)

			a.do(event_types.EventInstallmentPeriodPaid.Enum(), payload.InstallmentPeriodPaidPayload{
				InstallmentUuid: installmentUUID,
				Period:          1,
				PaidDate:        "2026-06-01",
				PaidLedgerUuid:  testLedgerUUID(2),
			})

			// Period 1: 10000/12 truncate(6) = 833.333333
			wantAmt := dec("833.333333")
			txn := queryLastTxn(t, db)
			Expect(txn.Date).To(Equal("2026-06-01"))

			entries := queryEntries(t, db, txn.TxnID)
			Expect(entries).To(HaveLen(2))
			// Dr 2201-01 833.333333 (ledger=1) [FINANCING], Cr 1101-02 833.333333 (ledger=2)
			assertEntry(t, entries, 0, "2201-01", lidPtr(1), wantAmt, dec("0"), cfPtr("FINANCING"))
			assertEntry(t, entries, 1, "1101-02", lidPtr(2), dec("0"), wantAmt, nil)

			assertLedgerRB(t, db, 1, wantAmt, dec("10000"))
			assertLedgerRB(t, db, 2, dec("0"), wantAmt)
		})
	})
})

// ─────────────────────────────────────────
// Investment
// ─────────────────────────────────────────

var _ = Describe("EventInvestmentBought", func() {
	It("GIVEN 股票買入 payload（數量 10、單價 100、手續費 5）\n  WHEN 執行 EventInvestmentBought\n  THEN 寫入 3 筆分錄（投資與費用標記 INVESTING）並更新即時餘額", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertAssetTypeConfig(t, db)
		insertPeriodOpen(t, db, 1, "2026-05-01")
		insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")
		insertInvestment(t, db, 1, "1102-01", "STOCK", "AVG", "FVTPL")

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "inv-buy-1")
		a.do(event_types.EventInvestmentBought.Enum(), payload.InvestmentBoughtPayload{
			InvestmentUUID: testInvestmentUUID(1),
			Date:           "2026-05-01",
			Quantity:       dec("10"),
			UnitPrice:      dec("100"),
			ExchangeRate:   dec("1"),
			Fee:            dec("5"),
			Tax:            dec("0"),
			LedgerId:       1,
		})

		// cost = 10×100×1 = 1000; totalCost = 1005 (fee=5)
		txn := queryLastTxn(t, db)
		Expect(txn.Date).To(Equal("2026-05-01"))

		entries := queryEntries(t, db, txn.TxnID)
		Expect(entries).To(HaveLen(3))
		// Dr 1102-01 1000 [INVESTING], Cr 1101-02 1005 (ledger=1), Dr 5402-01 5 [INVESTING]
		assertEntry(t, entries, 0, "1102-01", nil, dec("1000"), dec("0"), cfPtr("INVESTING"))
		assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("0"), dec("1005"), nil)
		assertEntry(t, entries, 2, "5402-01", nil, dec("5"), dec("0"), cfPtr("INVESTING"))

		assertAccountRB(t, db, "1102-01", dec("1000"), dec("0"))
		assertAccountRB(t, db, "5402-01", dec("5"), dec("0"))
		assertLedgerRB(t, db, 1, dec("0"), dec("1005"))
	})
})

var _ = Describe("EventInvestmentSold", func() {
	Context("處分利得（賣出價高於成本）", func() {
		It("GIVEN 以 100 買入後以 120 賣出（手續費 5）\n  WHEN 執行 EventInvestmentSold\n  THEN 寫入 4 筆分錄並認列利得 200（標記 INVESTING）", func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			t := GinkgoT()

			insertAssetTypeConfig(t, db)
			insertPeriodOpen(t, db, 1, "2026-05-01")
			insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")
			insertInvestment(t, db, 1, "1102-01", "STOCK", "AVG", "FVTPL")

			svc := newSvc(db)
			a := newAppender(t, svc, testCtx(), "inv-sell-1")

			// Buy 10 @ 100 (no fee)
			a.do(event_types.EventInvestmentBought.Enum(), payload.InvestmentBoughtPayload{
				InvestmentUUID: testInvestmentUUID(1),
				Date:           "2026-05-01",
				Quantity:       dec("10"),
				UnitPrice:      dec("100"),
				ExchangeRate:   dec("1"),
				Fee:            dec("0"),
				Tax:            dec("0"),
				LedgerId:       1,
			})

			// Sell 10 @ 120; fee=5; netProceeds = 10×120-5 = 1195; costBasis = 1000; gain = 200
			a.do(event_types.EventInvestmentSold.Enum(), payload.InvestmentSoldPayload{
				InvestmentUUID: testInvestmentUUID(1),
				Date:           "2026-05-15",
				Quantity:       dec("10"),
				UnitPrice:      dec("120"),
				ExchangeRate:   dec("1"),
				Fee:            dec("5"),
				Tax:            dec("0"),
				LedgerId:       1,
			})

			txn := queryLastTxn(t, db)
			Expect(txn.Date).To(Equal("2026-05-15"))

			entries := queryEntries(t, db, txn.TxnID)
			Expect(entries).To(HaveLen(4))
			// Dr 1101-02 1195 (ledger=1), Cr 1102-01 1000 [INVESTING]
			// Dr 5402-01 5 [INVESTING], Cr 4203-01 200 [INVESTING]
			assertEntry(t, entries, 0, "1101-02", lidPtr(1), dec("1195"), dec("0"), nil)
			assertEntry(t, entries, 1, "1102-01", nil, dec("0"), dec("1000"), cfPtr("INVESTING"))
			assertEntry(t, entries, 2, "5402-01", nil, dec("5"), dec("0"), cfPtr("INVESTING"))
			assertEntry(t, entries, 3, "4203-01", nil, dec("0"), dec("200"), cfPtr("INVESTING"))

			assertAccountRB(t, db, "1102-01", dec("1000"), dec("1000"))
			assertAccountRB(t, db, "4203-01", dec("0"), dec("200"))
			assertLedgerRB(t, db, 1, dec("1195"), dec("1000"))
		})
	})
})

var _ = Describe("EventDividendReceived", func() {
	It("GIVEN 持有 10 股後收取股利 1000\n  WHEN 執行 EventDividendReceived\n  THEN 寫入 2 筆分錄並更新即時餘額", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertDividendAccounts(t, db)
		insertAssetTypeConfig(t, db)
		insertPeriodOpen(t, db, 1, "2026-05-01")
		insertLedger(t, db, 1, "BANK_ACCOUNT", "1101-02")
		insertInvestment(t, db, 1, "1102-01", "STOCK", "AVG", "FVTPL")

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "div-1")

		a.do(event_types.EventInvestmentBought.Enum(), payload.InvestmentBoughtPayload{
			InvestmentUUID: testInvestmentUUID(1),
			Date:           "2026-05-01",
			Quantity:       dec("10"),
			UnitPrice:      dec("100"),
			ExchangeRate:   dec("1"),
			Fee:            dec("0"),
			Tax:            dec("0"),
			LedgerId:       1,
		})

		a.do(event_types.EventDividendReceived.Enum(), payload.DividendReceivedPayload{
			InvestmentUUID: testInvestmentUUID(1),
			Date:           "2026-05-15",
			Amount:         dec("1000"),
			ExchangeRate:   dec("1"),
			WithholdingTax: dec("0"),
			Ratio:          dec("0"),
			LedgerId:       1,
		})

		txn := queryLastTxn(t, db)
		Expect(txn.Date).To(Equal("2026-05-15"))

		entries := queryEntries(t, db, txn.TxnID)
		Expect(entries).To(HaveLen(2))
		// Cr 4210 1000; Dr 1101-02 1000
		assertEntry(t, entries, 0, "4210", nil, dec("0"), dec("1000"), nil)
		assertEntry(t, entries, 1, "1101-02", lidPtr(1), dec("1000"), dec("0"), nil)

		assertAccountRB(t, db, "4210", dec("0"), dec("1000"))
		assertLedgerRB(t, db, 1, dec("1000"), dec("1000"))
	})
})

// ─────────────────────────────────────────
// Period
// ─────────────────────────────────────────

var _ = Describe("EventPeriodAnnualClosed", func() {
	It("GIVEN 2025 年所有月份已關帳且年度期間 OPEN\n  WHEN 執行 EventPeriodAnnualClosed\n  THEN 產生結帳與開帳兩筆交易，並正確轉移淨利至權益", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertSysAccounts(t, db)
		insertIncomeTransaction(t, db, 9001, "2025-06-15", "4101-01", 50000)
		insertAllMonthsClosed(t, db, 2025, 1)
		insertAnnualPeriodOpen(t, db, 13, 2025)

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "period-annual-1")
		a.do(event_types.EventPeriodAnnualClosed.Enum(), payload.PeriodAnnualClosedPayload{
			ClosingId: 13,
			ClosedAt:  "2025-12-31",
		})

		var txns []struct {
			TxnID   int64  `db:"txn_id"`
			TxnDate string `db:"txn_date"`
		}
		err := db.SelectContext(testCtx(), &txns,
			`SELECT txn_id, txn_date FROM transactions
			 WHERE merchant_id=? AND status='ACTIVE' AND txn_id != 9001
			 ORDER BY txn_id`,
			testMID)
		Expect(err).NotTo(HaveOccurred())
		Expect(txns).To(HaveLen(2))

		Expect(txns[0].TxnDate).To(Equal("2025-12-31"))
		Expect(txns[1].TxnDate).To(Equal("2026-01-01"))

		closingTxnID := txns[0].TxnID
		openingTxnID := txns[1].TxnID

		// Closing: Dr 4101-01 50000, Cr 3102-01 50000
		closingEntries := queryEntries(t, db, closingTxnID)
		incomeDrFound, netIncomeCrFound := false, false
		for _, e := range closingEntries {
			if e.AccountID == "4101-01" && e.Debit.Equal(dec("50000")) {
				incomeDrFound = true
			}
			if e.AccountID == "3102-01" && e.Credit.Equal(dec("50000")) {
				netIncomeCrFound = true
			}
		}
		Expect(incomeDrFound).To(BeTrue(), "closing txn: missing Dr 4101-01 50000")
		Expect(netIncomeCrFound).To(BeTrue(), "closing txn: missing Cr 3102-01 50000")

		// Opening: Dr 3102-01 50000, Cr 3101-01 50000
		openingEntries := queryEntries(t, db, openingTxnID)
		netIncomeDrFound, openingCrFound := false, false
		for _, e := range openingEntries {
			if e.AccountID == "3102-01" && e.Debit.Equal(dec("50000")) {
				netIncomeDrFound = true
			}
			if e.AccountID == "3101-01" && e.Credit.Equal(dec("50000")) {
				openingCrFound = true
			}
		}
		Expect(netIncomeDrFound).To(BeTrue(), "opening txn: missing Dr 3102-01 50000")
		Expect(openingCrFound).To(BeTrue(), "opening txn: missing Cr 3101-01 50000")

		assertAccountRB(t, db, "3101-01", dec("0"), dec("50000"))
	})
})

var _ = Describe("EventPeriodAnnualReopened", func() {
	It("GIVEN 已執行年度關帳\n  WHEN 執行 EventPeriodAnnualReopened\n  THEN 產生兩筆反轉交易且各帶 ref_txn_id", func() {
		db, closeDB := testutil.NewTestDBGinkgo()
		DeferCleanup(closeDB)
		t := GinkgoT()

		insertSysAccounts(t, db)
		insertIncomeTransaction(t, db, 9001, "2025-06-15", "4101-01", 50000)
		insertAllMonthsClosed(t, db, 2025, 1)
		insertAnnualPeriodOpen(t, db, 13, 2025)

		svc := newSvc(db)
		a := newAppender(t, svc, testCtx(), "period-reopen-1")

		a.do(event_types.EventPeriodAnnualClosed.Enum(), payload.PeriodAnnualClosedPayload{
			ClosingId: 13,
			ClosedAt:  "2025-12-31",
		})
		a.do(event_types.EventPeriodAnnualReopened.Enum(), payload.PeriodAnnualReopenedPayload{
			ClosingId:  13,
			Reason:     "correction",
			ReopenedAt: "2026-01-05",
		})

		var txns []struct {
			TxnID int64  `db:"txn_id"`
			RefID *int64 `db:"ref_txn_id"`
		}
		err := db.SelectContext(testCtx(), &txns,
			`SELECT txn_id, ref_txn_id FROM transactions
			 WHERE merchant_id=? AND txn_id != 9001 ORDER BY txn_id`,
			testMID)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(txns)).To(BeNumerically(">=", 4))

		// 最後兩筆（reopen 產生的反轉）必須有 ref_txn_id
		reverses := txns[len(txns)-2:]
		for _, r := range reverses {
			Expect(r.RefID).NotTo(BeNil(), "reopen txn %d: expected ref_txn_id", r.TxnID)
		}
	})
})
