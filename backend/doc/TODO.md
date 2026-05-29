# TODO

計畫中的待辦事項。這裡記錄「知道要做但現在不做」的工作，包含功能擴充、技術債、效能優化與文件補充。
已確認本次要做的需求請直接進入需求處理流程，不需先記錄於此。

---

## 狀態說明

| 狀態 | 說明 |
|------|------|
| ⬜ Backlog | 計畫中，尚未排程 |
| 🔵 Planned | 已排入近期開發計畫 |
| 🟡 In Progress | 進行中 |
| ✅ Done | 已完成 |
| ❌ Cancelled | 取消，附理由 |

## 分類說明

| 分類 | 說明 |
|------|------|
| `[Feature]` | 新功能或功能擴充 |
| `[Refactor]` | 程式碼重構、架構調整 |
| `[Perf]` | 效能優化 |
| `[Test]` | 測試補充或改善 |
| `[Docs]` | 文件補充 |
| `[DX]` | 開發體驗改善（工具、腳本、設定） |
| `[Security]` | 安全性強化 |

---

## 待辦清單

<!-- 新增時插入對應優先順序區塊，格式如下 -->

<!--
- [ ] [分類] 標題 — 簡短說明，關聯：ISSUE-XXX 或 ADR-XXX（可省略）
-->

### 高優先（P1）

> 影響核心功能正確性或穩定性，應在近期處理。

**Replay 穩定性**

- [x] [Feature] 投資事件以 `creation_event_uuid` 取代 AUTOINCREMENT `investment_id` 作為跨事件引用鍵，完成日期：2026-05-28，關聯：ISSUE-018
- [x] [Feature] 全體 event sourcing projection 資料表加入 UUID 欄位（含跨表 UUID FK），確保 replay 不因 AUTOINCREMENT 計數器改變而斷裂 FK 關聯，完成日期：2026-05-29，關聯：ADR-018
- [ ] [Feature] `event_seed.go` bootstrapInvestmentIdUpcast — 將既有 EventInvestmentBought/Sold 等事件的 `investment_id` payload 欄位替換成對應的 `creation_event_uuid`，使舊資料亦可正確 replay（現版本「不考慮既存資料」，啟用前須先執行 upcast）

**Running Balance 系統性缺口（AccountBalanceRealtimeProjection）**

- [ ] [Feature] `EventInvestmentMarked` 的 running balance 更新 — `UnrealizedMarkedState.Transaction` 已存在，但 `TransactionProjectionService` 未處理此事件（分錄也未寫入），需先確認 MARK 分錄應由哪個 Projection 負責再一起修正，關聯：ISSUE-014
- [x] [Feature] Installment 事件 running balance 更新 — `EventInstallmentCreated` / `EventInstallmentPeriodPaid`，完成日期：2026-05-27，關聯：ISSUE-014，ADR-017
- [x] [Feature] Prepaid 事件 running balance 更新 — `EventPrepaidCreated` / `EventPrepaidAmortized` / `EventPrepaidDisposed`，完成日期：2026-05-27，關聯：ISSUE-014，ADR-017
- [x] [Feature] FixedAsset 事件 running balance 更新 — `EventAssetPurchased` / `EventAssetDepreciated` / `EventAssetDisposed`，完成日期：2026-05-27，關聯：ISSUE-014，ADR-017
- [x] [Feature] PeriodClose 事件 running balance 更新 — `EventPeriodAnnualClosed` / `EventPeriodAnnualReopened`（各建立 2 筆 txn），完成日期：2026-05-27，關聯：ISSUE-014，ADR-017

**多租戶事件核心隔離（Merchant ID in Event Core）**

- [x] [Feature] Migration：event_store、aggregate_versions、snapshots 新增 `merchant_id` 欄位（含 Migration 20260526002）— 完成日期：2026-05-26
- [x] [Feature] UoW repos（event.go、version.go、snapshot.go）改從 ctx 取得 merchantID 並傳入 sqlcdb 查詢 — 完成日期：2026-05-26
- [x] [Feature] query/event.go 的所有 sqlcdb 呼叫加入 MerchantID 參數 — 完成日期：2026-05-26
- [x] [Security] ISSUE-013：TruncateProjections 加入 merchant_id 隔離，補入遺漏表並移除 sqlite_sequence 重置 — 完成日期：2026-05-26，關聯：ISSUE-013

**預付費用 & 固定資產（Prepaid & Fixed Asset）**

- [x] [Feature] 設計確認 — 完成日期：2026-05-22（計畫書已確認，關聯：ADR-014）
- [x] [Feature] 實作 Migration + Seeds（4 張資料表 + 處分損益科目）— 完成日期：2026-05-22
- [x] [Feature] 實作 Enums（DepreciationMethod、AssetPaymentType、6 個 EventType）— 完成日期：2026-05-22
- [x] [Feature] 實作 SQL Queries + sqlc.yaml overrides + go generate — 完成日期：2026-05-22
- [x] [Feature] 實作 Payload、State（6 組）— 完成日期：2026-05-22
- [x] [Feature] 實作 Pipeline（factory/prepaid.go、factory/asset.go，含 CF 標籤）— 完成日期：2026-05-22
- [x] [Feature] 實作 UoW Projection Repos（PrepaidRepo、FixedAssetRepo）— 完成日期：2026-05-22
- [x] [Feature] 實作 projection/transaction.go（6 個 apply 函式 + UpdatePrepaidTxn/UpdateFixedAssetTxn）+ base.go 登記 — 完成日期：2026-05-22
- [x] [Feature] 實作 query/prepaid.go、query/asset.go + query/base.go 掛載 — 完成日期：2026-05-22
- [x] [Feature] 實作 Service（PrepaidService、FixedAssetService）— 完成日期：2026-05-22
- [x] [Feature] 實作 Handler（prepaid.go、asset.go）+ handler.go 路由 — 完成日期：2026-05-22
- [x] [Test] 補充 CF 一致性測試（6 個測試：預付創建/攤提、折舊、資產購入/處分利得/處分損失）— 完成日期：2026-05-22

### 中優先（P2）

**儀表板 API（Dashboard）**

- [x] [Feature] `GET /api/dashboard/summary` — 回傳本月收入/支出、淨資產、現金餘額（四個 stat card 數字），完成日期：2026-05-26
- [x] [Feature] `GET /api/dashboard/monthly-trend?months=12` — 過去 N 個月每月收入 / 支出 / 淨額，完成日期：2026-05-26
- [x] [Feature] `GET /api/ledger/balances` — 各 LedgerAccount 目前餘額含 name/institution/type，完成日期：2026-05-26
- [x] [Test] dashboard API 整合測試 — 完成日期：2026-05-26（`dashboard_test.go`，涵蓋 summary 資產/現金餘額、monthly-trend 月份邊界、ledger balance 借貸淨額計算）

**稽核查詢 API（Audit）**

- [x] [Feature] `GET /api/audit/aggregate-version` — 查詢當前商戶所有 aggregate 版本，完成日期：2026-05-26
- [x] [Feature] `GET /api/audit/event?page=&page_size=` — 事件日誌分頁查詢（依 event_id 降冪），完成日期：2026-05-26
- [x] [Feature] `GET /api/audit/checkpoint` — 查詢所有投影機 checkpoint，完成日期：2026-05-26
- [x] [Feature] `GET /api/audit/snapshot` — 查詢所有快照，完成日期：2026-05-26
- [x] [Feature] `GET /api/exchange-rate?currency=` — 查詢匯率（全域資料，無商戶過濾），完成日期：2026-05-26
- [x] [Feature] `POST /api/audit/replay` — 全量或部分重建 projection（`from_event_id` + `aggregate_type` 可選），完成日期：2026-05-26

> 提升系統品質或開發效率，可排入下一個迭代。

**股息接收科目 Hardcode**

- [ ] [Feature] DividendReceived factory 硬編碼帳戶 ID — `factory/investment.go` 的股息接收 pipeline 直接用字串 `"4210"`（INCOME）與 `"5920"`（EXPENSE）查詢帳戶。這兩個科目不在 `seeds/accounts.sql` 中，測試必須手動插入才能通過。應改為從 `asset_type_account_config`（或另一個設定表）動態查詢，消除硬編碼依賴

**現金流量表（Cash Flow Statement）**

- [ ] [Feature] 系統自動分錄補上 `cash_flow_category` — 年度結帳（Period Annual Close / Reopen）結帳分錄為科目間內部軋轉，應保持 `cash_flow_category = NULL`（投資買賣、分期付款已於 2026-05-22 完成）

**直接法現金流量表（Direct Method）**

- [ ] [Feature] 直接法混合分類交易精準拆分 — 目前 `queryDirectOperatingCash` 對含 INCOME/EXPENSE 分錄的交易整筆歸入 Operating，不按比例拆分。評估個人財務場景是否真實需要；若需可在 `journal_entries` 加入比例欄位，或要求使用者拆成兩筆交易
- [ ] [Feature] 直接法明細行（Line Items）— 目前直接法只提供 CashReceived / CashPaid 兩個加總數字，若需「收到薪資 3000 / 支付房租 1500」明細清單，需新增 `queryDirectOperatingCashItems`（依科目彙總 CASH 移動明細）

**其他應評估科目類型**

- [ ] [Feature] 應收 / 應付帳款變動 — 確認前端預設值規則是否與科目 `accounts.cash_flow_category` 對應正確

### 低優先（P3）

> Nice-to-have，不影響現有功能，有空再做。

**會計分錄範本（Transaction Template）**

- [x] [Feature] 設計確認 — 完成日期：2026-05-25（參考金額、LedgerAccount 關聯、分類標籤、模糊搜尋，後端 API）
- [x] [Feature] 實作 Migration + Schema（2 張資料表）— 完成日期：2026-05-25
- [x] [Feature] 實作 SQL Queries + sqlc.yaml overrides + go generate — 完成日期：2026-05-25
- [x] [Feature] 實作 Projection Model（TransactionTemplate、TransactionTemplateEntry、TransactionTemplateDetail）— 完成日期：2026-05-25
- [x] [Feature] 實作 Repository（TemplateRepo，含 Create/Update 的 DB 交易）— 完成日期：2026-05-25
- [x] [Feature] 實作 Service（TemplateService）— 完成日期：2026-05-25
- [x] [Feature] 實作 Handler（5 個 API：GET /template、POST /template、GET /template/{id}、PUT /template/{id}、DELETE /template/{id}）— 完成日期：2026-05-25

**科目鑽取分析（Account Analysis）**

- [x] [Feature] 設計確認 — 完成日期：2026-05-25（三項 API：子科目餘額、分錄分頁、月度趨勢）
- [x] [Feature] 實作 Migration（journal_entries 索引）— 完成日期：2026-05-25
- [x] [Feature] 實作 SQL Queries（6 個查詢）+ go generate — 完成日期：2026-05-25
- [x] [Feature] 實作 Projection Model（AccountChildBalance、AccountJournalEntryRow、AccountMonthlyBalance 等）— 完成日期：2026-05-25
- [x] [Feature] 實作 Repository（AccountAnalysisRepo）— 完成日期：2026-05-25
- [x] [Feature] 實作 Service（AccountAnalysisService）— 完成日期：2026-05-25
- [x] [Feature] 實作 Handler（3 個 API：GET /account/{id}/children-balance、/entries、/monthly-balance）— 完成日期：2026-05-25
- [x] [Test] 補充測試（5 個測試：子科目餘額含/無 running balance、分錄分頁日期篩選、月度趨勢含快照標記、格式錯誤）— 完成日期：2026-05-25

**技術優化**

- [ ] [DX] `sqlc.yaml` nullable enum override 自動化 — 每新增一個 nullable enum 欄位都需手動宣告 override，考慮以 go:generate 腳本輔助生成（關聯：ADR-005）
- [ ] [Perf] 股東權益變動表快照優化 — 目前 `queryEquityAccounts` 全表掃描 journal_entries，未使用 period_closings 快照。可參考 `queryBalanceSheetWithSnap` 的模式，在有快照時以快照基底 + delta 取代全量掃描（關聯：ADR-006）

---

## 完成紀錄

<!-- 項目完成後從上方清單移至此區，保留歷史 -->

- [x] [Feature] 科目維護 UI 支援設定 `cash_flow_category` — 目前只有 seed 預設值，使用者新增科目後無法從 UI 設定分類，導致該科目不出現在現金流量表
- [x] [Test] 現金流量表 API 端對端測試 — 完成日期：2026-05-15（`report_test.go`，涵蓋期初/期末現金、三大活動分類金額、淨利計入營業小計、NetChange 驗證）
- [x] [Test] 股東權益變動表 API 端對端測試 — 完成日期：2026-05-15（`report_test.go`，涵蓋期初/期間區分、彙總聚合、本期淨利虛擬行、total_* 含淨利合計）
- [x] [Docs] 補充 `cash_flow_category` 各分類的科目對應說明 — 完成日期：2026-05-15（`doc/account/CASH_FLOW_GUIDE.md`，涵蓋四大分類科目列表、分類原則、新增科目決策流程）
- [x] [Test] 現金流量表快照路徑測試 — 完成日期：2026-05-15（`report_test.go`，驗證月結快照存在時期初現金/期間調整仍依日期邊界正確篩選）
- [x] [Test] 股東權益變動表快照路徑測試 — 完成日期：2026-05-15（`report_test.go`，驗證月結快照存在時 begin_balance 與 period_change 正確分離）
- [x] [Feature] 現金流量表支援科目階層彙總 — 完成日期：2026-05-15（`queryCashFlowChanges` 加入 summary_cf UNION ALL；`CashFlowItem` 新增 `is_summary`；補充兩個階層聚合測試）
