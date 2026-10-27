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
| `[Fix]` | 已知 bug 修正（對應 `ISSUES.md`），與重構分開進行 |
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

**BFS Engine 遷移與事件重建**

計畫書：`doc/plan/BFS_ENGINE_MIGRATION.md`（2026-10-19 review 通過；2026-10-21 Phase 0 完成，決策見 ADR-023～027，進度見 `doc/STEP.md`）。修正類與重構類分開進行、分開 review。
2026-10-19 決定：目前系統沒有需要保留的資料，舊 replay 不再修正（ISSUE-021、022 改為 Wontfix），由計畫書 R3 的新重建機制取代。

- [x] [Fix] BFS 現金流分類修改無法運作 — 計畫書 F1，完成日期：2026-10-19，關聯：ISSUE-020
- [x] [Fix] 銀行對帳核准必定失敗並留下未關聯的交易，重按會產生重複會計交易（冪等止血） — 計畫書 F3（S4 已確定採 a），完成日期：2026-10-19，關聯：ISSUE-023（原子化待 R7）
- [x] [Fix] BFS safety middleware 的狀態由所有請求共用，並行請求可能讓後端程式終止 — 計畫書 F4，完成日期：2026-10-21，關聯：ISSUE-025
- [x] [Fix] 前端更新帳戶類型、資產類型設定與系統科目對應時，呼叫已移除的 API — 計畫書 F5，完成日期：2026-10-21，關聯：ISSUE-026
- [x] [Fix] 科目、帳戶與投資修改的畫面版本檢查沒有作用；一併修正科目初始版本為 0、重複新增會覆寫既有科目 — 計畫書 F6，完成日期：2026-10-21，關聯：ISSUE-027
- [x] [Fix] 通用寫入 API 仍可經舊路徑寫入已遷移到 BFS 的現金流分類事件 — 計畫書 F7，完成日期：2026-10-22，關聯：ISSUE-028
- [x] [Fix] 預付與固定資產類別的第二次修改或刪除必定失敗 — 計畫書 F8，完成日期：2026-10-22，關聯：ISSUE-029
- [x] [Fix] 不同商戶無法開立同一期間 — 計畫書 F9，完成日期：2026-10-22，關聯：ISSUE-030
- [ ] [Fix] 單商戶重建後整數 `ledger_id` 改變，引用它的事件重播失敗 — 由 R3 的計畫處理並列入驗收（舊 replay 不修正；做法見 `doc/plan/BFS_R3_REBUILD.md` 的 D3），關聯：ISSUE-031
- [ ] [Fix] 不同商戶無法建立同一代號與幣別的投資 — 計畫書 F10，R9 之前必須完成，關聯：ISSUE-034
- [ ] [Fix] SQL 種子直接寫入讀模型，種子檔之後新增的科目沒有事件，由事件重建時遺失 — R10 之前處理，關聯：ISSUE-035
- [ ] [Refactor] 其他被事件 payload 引用的整數 ID，依 R3 計畫書 D3 的規則由建立事件記錄：R4 交易 ID、R5 `closing_id`（開帳事件記錄）、R6 銀行對帳的 `bank_txn_id`、`matched_entry_id`、`created_txn_id`、`template_id`。各 Phase 開始前在計畫書列出（`doc/plan/BFS_R3_REBUILD.md` 1.2 第 8 點）
- [ ] [Refactor] 重大交易全面改走 BFS engine，並建立只跑 projector 的事件重建機制；新重建機制（R3）的驗收條件需涵蓋 ISSUE-021、ISSUE-022 的要求 — 計畫書 R0～R10，關聯：ADR-023～027

**Replay 穩定性**

- [x] [Feature] 投資事件以 `creation_event_uuid` 取代 AUTOINCREMENT `investment_id` 作為跨事件引用鍵，完成日期：2026-05-28，關聯：ISSUE-018
- [x] [Feature] 全體 event sourcing projection 資料表加入 UUID 欄位（含跨表 UUID FK），確保 replay 不因 AUTOINCREMENT 計數器改變而斷裂 FK 關聯，完成日期：2026-05-29，關聯：ADR-018
- [ ] ❌ [Feature] `event_seed.go` bootstrapInvestmentIdUpcast — 將既有 EventInvestmentBought/Sold 等事件的 `investment_id` payload 欄位替換成對應的 `creation_event_uuid`，使舊資料亦可正確 replay（現版本「不考慮既存資料」，啟用前須先執行 upcast）— Cancelled（2026-10-19）：目前系統沒有需要保留的資料，不需為既有事件做 upcast

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
- [x] [Fix] 盤點其他資料表的唯一鍵是否缺少 `merchant_id`（`period_closings` 已發現，見 ISSUE-030）；需在 R3（單商戶重建）之前完成，完成日期：2026-10-27（R3 計畫書 1.2 第 5 點：只有 `investments(symbol, currency)` 有問題，登錄為 ISSUE-034、F10），關聯：ISSUE-030、ISSUE-034

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

**商戶管理 API（Merchant Management）**

- [x] [Feature] `PUT /api/merchant/{merchant_id}` — 修改商戶 name / display_name / currency（ownership 驗證 via GetUserRole，403 / 422 / 500），完成日期：2026-06-03
- [x] [Feature] `DELETE /api/merchant/{merchant_id}` — 將商戶 status 設為 INACTIVE（ownership 驗證，403 / 500），完成日期：2026-06-03
- [x] [Test] merchant handler BDD 測試（10 個場景：update 成功/400×3/422/403/500，deactivate 成功/400/403/500），完成日期：2026-06-03

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

- [x] [Refactor] 現金流量分類抽離為獨立 `entry_cf_categories` Projection，移除 `journal_entries.cash_flow_category` 欄位 — 完成日期：2026-06-15，關聯：ADR-022
- [ ] [Feature] CF 分類審視 UI — `GET /cf-categories`（未分類清單）與 `PUT /txn/{uuid}/cf-category`（使用者手動覆蓋）API 已建立（Phase 5），前端審視流程尚未實作

**直接法現金流量表（Direct Method）**

- [ ] [Feature] 直接法混合分類交易精準拆分 — 目前 `queryDirectOperatingCash` 對含 INCOME/EXPENSE 分錄的交易整筆歸入 Operating，不按比例拆分。評估個人財務場景是否真實需要；若需可在 `journal_entries` 加入比例欄位，或要求使用者拆成兩筆交易
- [ ] [Feature] 直接法明細行（Line Items）— 目前直接法只提供 CashReceived / CashPaid 兩個加總數字，若需「收到薪資 3000 / 支付房租 1500」明細清單，需新增 `queryDirectOperatingCashItems`（依科目彙總 CASH 移動明細）

**其他應評估科目類型**

- [ ] [Feature] 應收 / 應付帳款變動 — 確認前端預設值規則是否與科目 `accounts.cash_flow_category` 對應正確

**伺服器設定（Server Config）**

- [ ] [Refactor] `cmd/main.go` 套用 `cfg.Server` — `bootstrap.LoadConfig()` 已讀取 `SERVER_ADDR`、`SERVER_READ_TIMEOUT`（預設 5s）、`SERVER_WRITE_TIMEOUT`（預設 10s），但 `http.Server` 仍寫死 `Addr: ":8080"` 且未設定 timeout。套用前須先確認預設 timeout 不會中斷 PDF 上傳、`POST /api/audit/replay` 等長請求
- [ ] [Fix] 引用 `accounts(account_id)` 的 7 個外鍵改以 `(account_id, merchant_id)` 引用或移除（需重建資料表）；須在開啟外鍵（ISSUE-032）之前完成，關聯：ISSUE-033
- [ ] [Fix] 預設 `DB_PARAM` 改用 modernc 的 `_pragma` 格式，讓外鍵、busy timeout、WAL 生效；開啟外鍵前先完成 ISSUE-033、檢查既有資料的孤兒引用（開啟外鍵不會追溯檢查），並在外鍵開啟下執行全部測試與 migration，關聯：ISSUE-032、ISSUE-033

### 低優先（P3）

> Nice-to-have，不影響現有功能，有空再做。

**程式碼格式**

- [ ] [DX] `internal/persistence/repos/db_transaction.go` 的 `sqlcdb.UpsertCheckpointParams` 以未具名欄位建立，`go vet ./internal/persistence/repos/` 回報 composites（`struct literal uses unkeyed fields`）。2026-10-21 於 R1 段 2 驗證時發現，HEAD 已存在，非 R1 引入；以具名欄位改寫即可，另以小幅修改處理
- [ ] [DX] 全專案 52 個已追蹤的 Go 檔案未通過 `gofmt`（2026-10-19 於 F3 驗證時發現，例如 `internal/services/bank_statement_review.go` 的 struct 欄位未對齊）。建議以只調整格式的獨立 commit 執行 `gofmt -w`，不混入功能修改，方便 review

**測試規範一致性**

- [ ] [Test] `internal/services/**` 既有測試使用 Ginkgo，與 ADR-019／`doc/TESTING.md`（Services 層用標準 Go testing、table-driven）不一致：`services/journal_entry_suite_test.go`（含同 suite 的 `journal_entry_test.go`、`category_test.go`）、`services/pipelines/factory/asset_*_test.go`、`services/projection/asset_*_test.go`（2026-10-21 於 F3 review 時發現；F3 新增的 `bank_statement_review_test.go` 已改用標準 testing）。需先決定改寫既有測試或修訂規範，再以獨立修改處理

**寫入版本檢查（BFS 遷移後）**

- [ ] [Feature] 經 `POST /api/event/append` 的寫入改帶畫面版本，或改為專屬 API — 目前前端在送出前才以 `getAggregateVersion` 取得最新版本，只擋得下「取得版本到送出」之間的並行寫入；BFS 遷移期間採「交易內」版本策略的事件連這段衝突也不再偵測（ADR-027 規則第 4 點）。做法：讀取 API 回傳版本、前端保存後送回，或改為各功能的專屬 API；同時移除共用 aggregate 已無作用的 `getAggregateVersion` 呼叫。屬於跨端新需求，在計畫書 R10 之後另行規劃，關聯：ADR-027 規則第 8 點、前端 ADR-013

**查詢層的寫入方法（query.Repo）**

- [ ] [Refactor] `query.Repo` 定位為唯讀查詢層，卻有 Config、Merchant、User、Template 共 10 個寫入方法；其中 Config 的 `UpsertLedgerAccountTypeConfig`、`UpsertAssetTypeAccountConfig` 從 `9e3fcfc` 起就沒有呼叫端（設定改由 projection 寫入）。R0 只讓綁定交易的 Repo 拒絕這些寫入（`ErrTxBoundWrite`），沒有移動它們；是否移出查詢層，另行規劃（2026-10-21，計畫書 `doc/plan/BFS_PREREQ_F5_F6_R0.md` 4.6）

**系統科目對應（Sys Account）**

- [ ] [Fix] `sys_config.sys_account_updated` 的 pipeline（`internal/services/pipelines/factory/sysconfig.go`）不檢查 `sys_code` 是否存在，送出不存在的代碼時，projection 的 UPSERT 會新增一筆系統科目；`9e3fcfc` 移除的舊 API 是 UPDATE，不會新增。前端只會送既有的代碼，F5 不修改（2026-10-21，計畫書 `doc/plan/BFS_PREREQ_F5_F6_R0.md` 2.8 第 1 點），關聯：ISSUE-026

**銀行對帳單（Bank Statement Reconciliation）**

- [x] [Feature] CSV 欄位對應範本 CRUD（bank_csv_templates）— 完成日期：2026-06-03
- [x] [Feature] CSV 匯入與解析（bank_statement_imports / bank_statement_txns）— 完成日期：2026-06-03
- [x] [Feature] 自動比對（EXACT 精確 / FUZZY 模糊 ±3 天，可重跑）— 完成日期：2026-06-03
- [x] [Feature] 人工審核（建議分錄預覽、同意建立 transaction.created、忽略）— 完成日期：2026-06-03
- [x] [Feature] PDF 匯入框架 — bank_pdf_templates CRUD、pdfparser interface/factory、ImportPDF handler、多帳本多 ledger UUID 支援，完成日期：2026-06-13
- [x] [Feature] 永豐銀行 PDF 解析實作 — `internal/pkg/pdfparser/sinopac/parser.go`，以 pdftotext -layout 萃取文字、欄位 rune 邊界法解析多帳戶對帳單，完成日期：2026-06-13
- [ ] [Feature] 其他銀行 PDF 解析 — 新銀行在 `internal/pkg/pdfparser/{bank}/parser.go` 實作並在 `internal/pkg/pdfparserfactory/factory.go` 中登記
- [x] [Feature] BankType enum 新增 SINOPAC — 完成日期：2026-06-13
- [x] [Feature] `GET /api/bank-pdf-template/{uuid}/ledgers` — 回傳指定 PDF 範本的帳本對應清單（`bank_pdf_template_ledgers`），供前端編輯 Modal 預填現有 ledger 資料，完成日期：2026-06-15

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

- [ ] [Refactor] `Executor.ExecuteBatch`（`internal/runtime/engine/executor.go`）只有測試使用，以 goroutine 並行 dispatch，與 ADR-023 A1「整棵事件圖一個交易、依序處理」不一致。R1 只多一個 `repo` 參數，並在註解寫明不在 `Run` 的執行路徑上、不得共用 `Run` 的交易內 Repo；是否移除另行評估（`doc/plan/BFS_R1_ENGINE_FOUNDATION.md` 的 D3）
- [ ] [Refactor] 現金流分類的專屬 handler（`internal/handler/entry_cf_category.go`）仍自行組根事件的封套，沒有改用 R2 段 1 的 `command.Build`。改用後的 aggregate、版本與封套欄位相同；若同時改用 `command.Writer.Append`，回應會從 kernel 事件變成 `event_store` 的紀錄，需先確認前端（`frontside/src/api/cfCategory.ts`）是否使用回應內容。R2 計畫未列入，另行評估（2026-10-22）
- [ ] [Refactor] BFS 的科目與明細帳戶寫入沿用舊套件的 `projection_repo.ErrVersionConflict`：`internal/process/account.go` 與 `internal/persistence/repos/projection/account.go` 引用 `repos/unit_of_work/event_store/projection_repo`，以保持與 F6 相同的衝突訊息（R2 D9）。R10 移除舊路徑時，把這個 sentinel 移到 BFS 端（例如 `domain/event`），R8、R9 遷移投資等事件時也沿用同一個（2026-10-22）
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


## 暫存

### 預計專案結構
```
/project-root
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── kernal/
│   │   ├── event/
│   │   ├── errors/
│   │   └── result/
|   |
│   │
│   ├── domain/
│   │   ├── event/
│   │   │   ├── journal_posted.go
│   │   │   ├── cash_movement_detected.go
│   │   │   └── cash_flow_fact_created.go
│   │   │
│   │   ├── aggregate/
│   │   │   ├── journal.go
│   │   │   └── account.go
│   │   │
│   │   └── valueobject/
│   │       ├── money.go
│   │       └── id.go
│   │
│   │
│   ├── process/
│   │   ├── journal_pm.go
│   │   ├── cashflow_pm.go
│   │   └── registry.go
│   │
│   │
│   ├── runtime/
│   │   ├── engine/
│   │   │   ├── bfs_engine.go
│   │   │   ├── queue.go
│   │   │   └── executor.go
│   │   │
│   │   ├── mediator/
│   │   │   ├── mediator.go
│   │   │   ├── dispatch.go
│   │   │   └── registry.go
│   │   │
│   │   ├── scheduler/
│   │   │   ├── worker_pool.go
│   │   │   └── partition_router.go
│   │   │
│   │   └── safety/
│   │       ├── cycle_detector.go
│   │       ├── backpressure.go
│   │       └── depth_guard.go
│   │
│   │
│   ├── application/
│   │   ├── api/
│   │   │   ├── command_handler.go
│   │   │   └── routes.go
│   │   │
│   │   └── service/
│   │       ├── journal_service.go
│   │       └── cashflow_service.go
│   │
│   │
│   ├── persistence/
│   │   ├── eventstore/
│   │   │   ├── store.go
│   │   │   ├── append.go
│   │   │   └── snapshot.go
│   │   │
│   │   └── projection/
│   │       ├── projector.go
│   │       ├── ledger_projection.go
│   │       └── cashflow_projection.go
│   │
│   │
│   └── shared/
│       ├── errors/
│       ├── logger/
│       └── utils/
│
└── go.mod
```