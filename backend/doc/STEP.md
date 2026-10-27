# BFS Engine 遷移與事件重建 — 實作進度

> 計畫書：`doc/plan/BFS_ENGINE_MIGRATION.md`；決策：ADR-023～027（前端 ADR-013）。
> 依 `AGENTS.md`「變更幅度與重構」，修正類（F）與重構類（R）分開實作、分開 review。本檔只追蹤進度；各項的內容、預期的行為變更與驗收條件以計畫書第 5 節為準。
> R Phase 依序執行。單一 Phase 若本身達到 `backend/CLAUDE.md` 的大型標準，開始前在本檔該 Phase 下方展開子階段。
> 前一份進度（F1：BFS 現金流分類修改修正，ISSUE-020）已完成，紀錄見 commit `8fbea2c` 的本檔。
> 驗證環境（Docker、2 GB 記憶體、`go test -vet=off`）見 `doc/NOTES.md`。

## 重構類：BFS 遷移

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 決策：寫成 ADR-023～027 與前端 ADR-013（ADR-021 由 ADR-026 分階段取代）；`AGENTS.md`、`backend/CLAUDE.md` 加入 BFS 遷移過渡規則；建立 STEP.md | ✅ done（2026-10-21） |
| R0 | 交易內查詢：讓 `query.Repo` 可綁定 `sqlx.Tx` | ✅ done（2026-10-21） |
| R1 | engine 基礎：事件圖一個交易（A1）、子事件封套與版本規則（A2）、handler 改用 `Run` 建立的交易內查詢、移除 projector 的 `state` 參數、建立事件註冊表 | ✅ done（2026-10-21） |
| R2 | 主檔與設定（含 `BootstrapEventSeeds`）、期間開帳改走 BFS；`POST /api/event/append` 依註冊表派送 | ✅ done（2026-10-22） |
| R3 | BFS 重建機制：只跑 projector、以 `Reset()` 宣告清空範圍、單商戶原子重建（A4）；尚不對外 | ⬜ pending |
| R4 | 傳票 `transaction.*` 與 `journal.posted` 子事件；每筆交易各自一個 aggregate（A5） | ⬜ pending |
| R5 | 期間結帳與重開（月結、年結） | ⬜ pending |
| R6 | 銀行範本、匯入改走 BFS；比對與忽略事件化（ADR-026）；更新過渡規則中的 ADR-021 例外 | ⬜ pending |
| R7 | 銀行對帳核准原子化（ISSUE-023 的根本修正）；ADR-021 改為 Superseded by ADR-026 | ⬜ pending |
| R8 | 預付費用、固定資產、分期 | ⬜ pending |
| R9 | 投資；`exchange_rates` 改為商戶資料（A6） | ⬜ pending |
| R10 | 移除舊路徑；replay 與匯入切換到新機制；移除過渡規則；經使用者確認後清空開發 DB | ⬜ pending |

## 修正類：相關 bug fix

不屬於重構的實作範圍，各自依 `ISSUES.md` 小幅修改、各自 review；時程限制是它們與 R Phase 的前後依賴。

| 項目 | 說明 | 時程限制 | 狀態 |
|------|------|----------|------|
| F1 | ISSUE-020：BFS 現金流分類修改無法運作 | — | ✅ done（2026-10-19） |
| F3 | ISSUE-023：銀行對帳核准冪等止血 | — | ✅ done（2026-10-19；原子化待 R7） |
| F4 | ISSUE-025：safety middleware 的狀態改為每次 `Run` 各自獨立 | R1 之前 | ✅ done（2026-10-21） |
| F5 | ISSUE-026：重新提供帳戶類型、資產類型設定與系統科目對應的更新 API | 建議在 R2 之前 | ✅ done（2026-10-21） |
| F6 | ISSUE-027：科目、帳戶與投資修改的業務版本檢查 | R2、R9 之前 | ✅ done（2026-10-21） |
| F7 | ISSUE-028：通用寫入 API 拒絕已遷移到 BFS 的事件類型 | R2 之前 | ✅ done（2026-10-22） |
| F8 | ISSUE-029：預付與固定資產類別的第二次修改或刪除必定失敗 | R2 段 4 之前 | ✅ done（2026-10-22） |
| F9 | ISSUE-030：不同商戶無法開立同一期間 | 建議在 R2 段 5 之前，R3 之前必須 | ✅ done（2026-10-22） |
| F10 | ISSUE-034：不同商戶無法建立同一代號與幣別的投資 | R9 之前 | ⬜ pending |

F7 驗證（2026-10-22，Docker `golang:1.26.9`，計畫書 `doc/plan/BFS_PREREQ_F7.md`，一段完成）：`go generate ./...`、`go build ./...`、`go vet`（`services`、`handler`）通過，修改與新增的 4 個檔案 `gofmt -l` 無輸出；`TestAppend_MigratedEventType` 的 4 個情境與 `NewMux POST /api/event/append` 的 2 個 spec 通過；`go test -vet=off ./...` 全數通過。反向檢查（計畫書第 5 節）：移除檢查 → service 情境 1、2、4 與 handler 的已遷移情境失敗；拒絕所有事件類型 → 兩個對照情境失敗；檢查移到 pipeline 之後 → service 情境 4 失敗（回傳 pipeline 的驗證錯誤）。

F9 驗證（2026-10-22，Docker `golang:1.26.9`，計畫書 `doc/plan/BFS_PREREQ_F9.md`，一段完成）：
- **建置與格式**：`go generate ./...`、`go build ./...`、`go vet`（`database`、`services`）通過；修改與新增的 3 個 Go 檔 `gofmt -l` 無輸出。
- **測試**：`TestPeriodStarted_PerMerchant` 的 3 個情境與 `TestPeriodClosingMerchantUnique_Migration` 的 2 個情境（外鍵關閉、開啟）通過；`go test -vet=off ./...` 全數通過。
- **反向檢查**（計畫書第 7 節）：
  - 唯一鍵改回舊鍵 → 寫入情境 1、2 與 migration 的 2 個情境失敗。
  - 移除唯一鍵 → 寫入情境 3 與 migration 的 2 個情境失敗。
  - 移除 `defer_foreign_keys` → migration 的外鍵開啟情境失敗（`DROP TABLE` 時 `FOREIGN KEY constraint failed`）。
  - 不還原計數 → migration 的 2 個情境失敗（下一筆 `closing_id` 為 4，預期 6）。

F8 驗證（2026-10-22，Docker `golang:1.26.9`，計畫書 `doc/plan/BFS_PREREQ_F8.md`，一段完成）：
- **建置與格式**：`go generate ./...`、`go build ./...`、`go vet`（`services`）通過；新增的測試檔 `gofmt -l` 無輸出。
- **測試**：`TestCategoryVersion_Write` 的 3 個情境與 `TestCategoryVersion_Migration` 的 4 個情境，兩種類別各一次，共 14 個子測試通過；`go test -vet=off ./...` 全數通過。
- **反向檢查**（計畫書第 7 節，修改查詢後重新 `go generate`）：
  - 修改的查詢不遞增 `version` → 寫入情境 1～3 失敗（第一次修改後讀模型 1、aggregate 2），migration 情境 4、6 失敗。
  - 停用的查詢不遞增 `version` → 寫入情境 2 失敗。
  - migration 不更新 → migration 情境 4、5 失敗。
  - migration 不比對 `merchant_id` → migration 情境 7 失敗（讀模型被改成其他商戶的版本 5）。

### F5 子階段

計畫書：`doc/plan/BFS_PREREQ_F5_F6_R0.md` 第 2 節。

| 段 | 說明 | 狀態 |
|----|------|------|
| 1 | 程式：`internal/handler/setting.go`、`sys.go`、`handler.go`，補回三個寫入 API 並註冊路由 | ✅ done（2026-10-21） |
| 2 | 測試：`internal/handler/setting_scenarios_test.go`、`setting_test.go`；反向檢查（移除路由後 `NewMux` 情境失敗） | ✅ done（2026-10-21） |

驗證（2026-10-21，Docker `golang:1.26`）：`go generate ./...`、`go build ./...`、`go vet ./internal/handler/` 通過，修改的 5 個檔案 `gofmt -l` 無輸出；`go test -vet=off -count=1 ./internal/handler/` 34 個 spec 全數通過（新增 16 個）；`go test -vet=off ./...` 全數通過。反向檢查：移除 `handler.go` 新增的三條路由後，三個 `NewMux` 情境都回 200（SPA 的 `index.html`），並在事件筆數的斷言失敗。review 修正（路徑的類型優先於 body 中不存在的類型值）：還原為修正前的 `setting.go` 時，兩個回歸情境都回 400（enumx invalid value）。

### F6 子階段

計畫書：`doc/plan/BFS_PREREQ_F5_F6_R0.md` 第 3 節。

| 段 | 說明 | 狀態 |
|----|------|------|
| 1 | 程式：`queries/account.sql`、`queries/investment.sql`、`projection_repo/account.go`、`projection_repo/investment.go`、`services/projection/account.go`，新增 migration `20261021001_account_initial_version.sql` | ✅ done（2026-10-21） |
| 2 | 測試：`internal/services/update_version_test.go`；反向檢查（還原段 1 的程式後，版本衝突與重複新增的情境失敗） | ✅ done（2026-10-21） |

驗證（2026-10-21，Docker `golang:1.26`）：`go generate ./...`、`go build ./...`、`go vet ./internal/services/` 通過，測試檔 `gofmt -l` 無輸出；新增的 11 個情境全數通過；`go test -vet=off ./...` 全數通過。反向檢查：段 1 的 5 個程式檔還原為修正前（只保留 `ErrVersionConflict` 宣告）後，11 個情境中 9 個失敗，只有實體帳戶與投資以正確版本修改的 2 個情境通過（修正前本來就正常）。失敗原因：科目修改後 version 不遞增；過期版本、重複新增、修改不存在的科目或 ledger_id 都沒有回傳錯誤。

### R0 子階段

計畫書：`doc/plan/BFS_PREREQ_F5_F6_R0.md` 第 4 節。

| 段 | 說明 | 狀態 |
|----|------|------|
| 1 | 程式：`internal/repos/query/` 的 `base.go`、`account.go`、`config.go`、`dashboard.go`、`investment.go`、`report.go`，新增 `tx.go` | ✅ done（2026-10-21） |
| 2 | 測試：`internal/repos/query/tx_test.go` | ✅ done（2026-10-21） |

驗證（2026-10-21，Docker `golang:1.26`）：`go generate ./...`、`go build ./...`、`go vet ./internal/repos/query/` 通過，`tx_test.go`、`tx.go`、`base.go` 的 `gofmt -l` 無輸出；新增的 3 個測試（18 個子測試）全數通過；`go test -vet=off ./...` 全數通過。本項沒有修正既有行為，不做反向檢查；另以變異確認寫入清單的效果：刪除 `tx.go` 的 `Merchant.Deactivate`、`Config.UpsertLedgerAccountTypeConfig` 兩個覆寫後，只有這兩個子測試失敗。

### R1 子階段

計畫書：`doc/plan/BFS_R1_ENGINE_FOUNDATION.md`。各段連同自己的測試完成，依 1 → 2 → 3 各自 review、commit。

| 段 | 說明 | 狀態 |
|----|------|------|
| 1 | 移除 projector 的 `state` 參數（計畫書第 3 節） | ✅ done（2026-10-21） |
| 2 | 事件圖交易邊界、Run-local 交易內查詢、子事件封套與版本（A1、A2，計畫書第 4 節） | ✅ done（2026-10-21） |

段 2 驗證（2026-10-21，Docker `golang:1.26`）：`go generate ./...`、`go build ./...` 通過，修改與新增的檔案 `gofmt -l` 無輸出；`go vet` 只回報既有的 `db_transaction.go` 未具名欄位（HEAD 已存在，記入 `TODO.md`）；`runtime/engine` 42 個 spec、`kernel/event` 測試全數通過；`go test -vet=off ./...` 全數通過。`-race`：接 DB 後 sqlite 無法在記憶體上限內以 `-race` 編譯，改在容器副本移除兩個接 DB 的測試檔後執行 `./internal/runtime/...`，通過（`doc/NOTES.md` 測試第 9 點）。反向檢查（計畫書 4.4）：每事件各自提交 → 9～12 與 7a、7b、8、16 失敗；移除 ctx 檢查 → 13a；全域 Repo → 14；移除版本紀錄 → 2～4、6；衍生改用 Append → 7a；移除封套覆寫 → 15。計畫外的修改：`shared/utils/test/db.go` 新增 `InitEnums()`、engine suite 改呼叫它，避免與測試 DB 重複初始化 enum 而 panic（`doc/NOTES.md` 測試第 8 點）。
| 3 | 事件註冊表（計畫書第 5 節） | ✅ done（2026-10-21） |

段 3 驗證（2026-10-21，Docker `golang:1.26.9`）：`go generate ./...`、`go build ./...`、`go vet`（`domain/event`、`persistence/handle`）通過，新增的 3 個檔案 `gofmt -l` 無輸出；`domain/event` 的 3 個測試與 `persistence/handle` 的解碼後套用測試通過；`go test -vet=off ./...` 全數通過。反向檢查（計畫書 5.4）：`Decode` 改回傳指標 → 解碼情境與解碼後套用（projector 的 payload type error）失敗；移除 cf 的登錄 → Lookup、解碼、註冊表檢查與解碼後套用都失敗。

### R2 子階段

計畫書：`doc/plan/BFS_R2_MASTER_DATA.md`。各段連同自己的測試完成，依 1 → 5 各自 review、commit。段 1～5 修改的檔案超過 5 個，各自再分為三個 phase（1a～1c、2a～2c、3a～3c、4a～4c、5a～5c），每個 phase 完成後執行 `go build ./...`，同一段的三個 phase 完成後一起 review、commit。

| 段 | 說明 | 狀態 |
|----|------|------|
| 1a | 註冊表的 `Spec.Aggregate` 改為接收事件（D2）；kernel 事件的 `RootVersion`、engine 依它寫入根事件、`Store.AppendCurrent`（D3）；測試情境 1～4 | ✅ done（2026-10-22） |
| 1b | `query.Event.GetByUUID`；`application/command` 的 `NewEngine`、`Build`、`Append`、`DispatchLegacy`（D4）；測試情境 5、6 | ✅ done（2026-10-22） |
| 1c | `NewMux` 改用 `command.NewEngine`；通用 API 先經 `DispatchLegacy`；測試情境 7（F7 的 handler 情境改為驗證新的拒絕訊息） | ✅ done（2026-10-22） |
| 2a | 科目與明細帳戶：`domain/event` 的 payload（D1）與 `MerchantAggregateID`（D8）、`process` 的 mediator handler（含 D9 的業務版本）；不登錄；測試情境 7（新舊 payload 的 JSON） | ✅ done（2026-10-22） |
| 2b | 科目與明細帳戶：`persistence/repos/projection` 的寫入方法、`persistence/handle` 的科目 projector（D10）、`command.NewEngine` 註冊；不登錄 | ✅ done（2026-10-22） |
| 2c | 科目與明細帳戶：註冊表登錄 5 種事件、種子的 `bootstrapAccounts` 改經 D4（D11）；既有測試的調整與測試情境 1～6 | ✅ done（2026-10-22） |
| 3a | 系統科目與科目設定：`domain/event` 的 payload（D1）、`process` 的 mediator handler（payload 驗證與科目存在）；不登錄；新舊 payload 的 JSON 契約測試 | ✅ done（2026-10-22） |
| 3b | 系統科目與科目設定：`persistence/repos/projection` 的寫入方法、`persistence/handle` 的設定 projector（D10）、`command.NewEngine` 註冊；不登錄 | ✅ done（2026-10-22） |
| 3c | 系統科目與科目設定：註冊表登錄 3 種事件；`handler/setting.go` 的 `appendConfigEvent`（設定與系統科目 API）與種子的 `bootstrapSysAccounts`、`bootstrapAccountConfigs` 改經 D4；F5 測試只換建構參數；通用 API 拒絕與種子兩次的測試 | ✅ done（2026-10-22） |
| 4a | 預付與固定資產類別：`domain/event` 的 payload（D1）、`process` 的 mediator handler（payload 驗證；修改與刪除確認類別存在且啟用）；不登錄；新舊 payload 的 JSON 契約測試（情境 7） | ✅ done（2026-10-22） |
| 4b | 預付與固定資產類別：`persistence/repos/projection` 的寫入方法、兩個類別 projector（D10，含 F8 的版本遞增）、`command.NewEngine` 註冊；不登錄 | ✅ done（2026-10-22） |
| 4c | 預付與固定資產類別：註冊表登錄 6 種事件（D6）；兩個專屬 handler 改經 D4；migration 補上既有類別的版本紀錄（D6）；`category_test.go`、F8 的 `category_version_test.go` 改經 BFS；情境 1～6 | ✅ done（2026-10-22） |
| 5a | 期間開帳：`domain/event` 的 payload（D1）、`process` 的 mediator handler（payload 驗證）；不登錄；新舊 payload 的 JSON 契約測試 | ✅ done（2026-10-22） |
| 5b | 期間開帳：`enums.AggregatePeriod` 與 migration（`chk_aggregate_type` 加入 `PERIOD`，D8）及其測試；只處理兩種開帳事件的期間 projector（D10）、`command.NewEngine` 註冊；不登錄 | ✅ done（2026-10-22） |
| 5c | 期間開帳：註冊表登錄 2 種事件（`('PERIOD', "1")`、版本策略「交易內」、舊映射 `('ACCOUNT', '')`）；F9 的 `period_unique_test.go` 改經 BFS 派送；情境 1～6 | ✅ done（2026-10-22） |

段 1 驗證（2026-10-22，Docker `golang:1.26.9`）：
- **建置與格式**：`go generate ./...`、`go build ./...` 通過；`go vet`（`kernel`、`domain/event`、`runtime/engine`、`repos/query`、`application`、`handler`）只回報既有的 `db_transaction.go` 未具名欄位；修改與新增的 16 個 Go 檔 `gofmt -l` 無輸出（`handler/event.go` 原本缺少檔尾換行，依 `doc/NOTES.md` 第 7 點對該檔執行 `gofmt`）。
- **測試**：
  - `runtime/engine` 47 個 spec 通過，含新增的 5 個根事件版本情境（情境 1～3；既有 42 個即情境 4）。
  - `application/command` 4 個測試、6 個子測試通過（情境 5～7，另含 `Build` 的拒絕與 `Writer.Append` 的成功、過期版本）。
  - `NewMux POST /api/event/append` 與現金流分類的 handler 情境通過。
  - `go test -vet=off ./...` 全數通過。
- **反向檢查**（計畫書段 1）：
  - `RootVersionNew` 改用 `Append` → 「已有版本 0」與「已在版本 2」兩個情境失敗。
  - `RootVersionCurrent` 改用 `Append` → 情境 3 的兩個情境失敗。
  - 計畫外另加：handler 一律走舊路徑 → 通用 API 的 cf 情境失敗（detail 變成 F7 的訊息）。
- **計畫外的修改**：
  - `kernel/errors/code.go` 新增 `ErrRootEnvelopeInvalid`，供未知的 `RootVersion` 使用。
  - `runtime/engine/bfs_engine_mocks_test.go` 的 `mockStore` 補上 `AppendCurrent`。
  - `registry_test.go` 補上「payload 為指標時拒絕」的檢查。

段 2 驗證（2026-10-22，Docker `golang:1.26.9`）：
- **建置與格式**：`go generate ./...`、`go build ./...` 通過（2a、2b、2c 各自執行）；`go vet`（`domain/event`、`process`、`persistence/handle`、`persistence/repos/projection`、`application/command`、`bootstrap`、`services`、`handler`）通過，`./internal/persistence/...` 只回報既有的 `db_transaction.go` 未具名欄位；修改與新增的 19 個 Go 檔 `gofmt -l` 無輸出；`git diff --check` 無輸出。
- **測試**：
  - 情境 7：`TestAccountPayloads_LegacyJSON` 的 5 種事件通過（新舊型別的 JSON 相同，且可互相解碼）。
  - 情境 1～3：`TestDispatchLegacy_AccountCreated`，以及 `TestDispatchLegacy_AccountCreatedRejected` 的 4 個子測試（aggregate 已存在、種子科目撞唯一鍵、aggregate ID 不符、aggregate 類型不符）通過。
  - 情境 4：F6 的科目與明細帳戶情境（9 個）與投資情境通過。
  - 情境 5：`TestLedgerAccountEvents_AccountAggregate` 通過。
  - 情境 6：`TestBootstrapEventSeeds_AccountsTwice` 通過。
  - F7：`TestAppend_MigratedEventType` 的 5 個情境通過；handler 38 個 spec 通過，其中 `NewMux POST /api/event/append` 有 4 個。
  - `go test -vet=off ./...` 全數通過。
- **反向檢查**（計畫書段 2，在容器副本中修改）：
  - handler 與 projector 都移除業務版本檢查 → 科目的過期版本、重複新增後過期版本、不存在的科目，以及明細帳戶的過期版本、不存在的 ledger_id，共 5 個情境失敗。
  - 舊映射檢查改為一律通過 → 情境 3 的 2 個子測試與 handler 的不符情境失敗。
  - 計畫外另加：
    - `account.created` 的版本策略改為「驗證」→ 情境 2 與 F6 的 2 個重複新增情境失敗。
    - 種子改回 `EventStoreService.Append` → 情境 6 失敗（F7 的 `ErrEventTypeMigrated`）。
    - BFS payload 的一個 JSON tag 改名 → 情境 7 的 `ledger_account.created` 失敗。
- **與計畫書的差異**：
  - F6 的兩個「重複新增」步驟，預期錯誤由 `UNIQUE constraint failed` 改為 aggregate 版本衝突（計畫書第 5 節第 3 點）；唯一鍵改由情境 2b（種子科目）覆蓋。
  - 科目與明細帳戶的衝突斷言改為比對 F6 的訊息文字：`kerrors.EventError` 沒有 `Unwrap`，`errors.Is` 無法穿透。BFS 端沿用 `projection_repo.ErrVersionConflict`，已記入 `TODO.md`。
  - 段 1 的 `command_test.go`、F7 的 service 測試與 handler 情境，未登錄的對照改用 `investment.created`；F7 原本的 `account.created` 對照改為拒絕情境。
  - D8 的 `MerchantAggregateID` 在本段建立，種子的 `seedAggregateID` 改為引用它；`handler/setting.go` 的 `configAggregateID` 留到段 3。

段 3 驗證（2026-10-22，Docker `golang:1.26.9`）：
- **建置與格式**：
  - `go generate ./...`、`go build ./...` 通過（3a、3b、3c 各自執行）。
  - `go vet`（`domain/event`、`process`、`persistence/handle`、`persistence/repos/projection`、`application/command`、`bootstrap`、`services`、`handler`）通過。
  - 修改與新增的 17 個 Go 檔 `gofmt -l` 無輸出；`git diff --check` 無輸出。
- **測試**：
  - F5 的設定 handler 情境 13 個、`NewMux` 情境 3 個，預期都沒有修改，全數通過；`setting_test.go` 只把 handler 的建構參數從 `EventStoreService` 換成 `command.Writer`。
  - 通用 API 拒絕：`TestDispatchLegacy_ConfigRejected` 的 3 種事件通過。
  - 種子兩次：`TestBootstrapEventSeeds_ConfigsTwice` 通過，使用真實的 `sys_accounts.sql`、`account_type_configs.sql`，驗證版本序列與 Source 為 seed；段 2 的 `TestBootstrapEventSeeds_AccountsTwice` 仍通過。
  - JSON 契約：`TestConfigPayloads_LegacyJSON` 的 4 個子測試通過，含選填科目為 null 的情境。
  - `go test -vet=off ./...` 全數通過。
- **反向檢查**（在容器副本中修改）：
  - mediator handler 略過科目存在檢查 → F5 的「不存在的科目 9999-99」情境失敗（計畫書段 3）。
  - 計畫外另加：
    - 給 `sys_config.sys_account_updated` 加上舊映射 → 拒絕測試的該子測試失敗。
    - 種子的帳戶類型設定不累計版本 → 種子測試失敗（aggregate_versions 唯一鍵）。
    - BFS payload 的 `oci_account_id` 改名 → 2 個資產類型子測試失敗。
    - 拿掉 `account_id` 的 `omitempty` → 選填為 null 的子測試失敗。
- **與計畫書的差異**：
  - 科目不存在時的錯誤文字由 `sql: no rows in result set` 改為 `account <id> not found`。舊 pipeline 的 `GetAccount` 查不到時回傳錯誤、不是 nil，原本的 not found 分支永遠跑不到；狀態碼仍為 400（計畫書第 5 節第 6 點）。
  - F5 測試的情境與預期不變，但 handler 的建構參數必須改。
  - `BootstrapEventSeeds` 移除 `es` 參數，`bootstrap/db.go` 不再建立舊的 `EventStoreService`。段 2 的 `bootstrapAccounts` 也改用共用的 `appendSeedEvent`，行為不變。

段 4 驗證（2026-10-22，Docker `golang:1.26.9`）：
- **建置與格式**：
  - `go generate ./...`、`go build ./...` 通過（4a、4b、4c 各自執行）。
  - `go vet`（`domain/event`、`process`、`persistence/handle`、`persistence/repos/projection`、`application/command`、`handler`、`services`）通過。
  - 修改與新增的 18 個 Go 檔 `gofmt -l` 無輸出；`git diff --check` 無輸出。
- **測試**（計畫書段 4 的情境）：
  - 情境 1：`NewMux category write routes` 的 2 個 spec（兩種類別各一）。經正式路由建立、修改、刪除，3 筆事件都寫在 `(類型, category_uuid)`，版本 1～3，`category_uuid` 為建立事件的 UUID；F8 的寫入測試另檢查建立後的 aggregate。
  - 情境 2、3：F8 的 `TestCategoryVersion_Write` 6 個子測試通過。過期的畫面版本回傳版本衝突，資料不變。
  - 情境 4：`TestDispatchLegacy_CategoryRejected` 的 6 種事件通過。
  - 情境 5、6：F8 的 `TestCategoryVersion_Migration` 通過，含 review 後新增的情境共 14 個子測試。新增 migration 後的 aggregate 版本斷言：從未修改的類別補上版本 1，已修改的維持 2，其他商戶的紀錄不變；之後經 BFS 修改分別到版本 2、3。
  - 情境 7：`TestCategoryPayloads_LegacyJSON` 的 6 種事件通過。
  - `services/category_test.go` 的 8 個 spec 通過，情境與預期不變。
  - `go test -vet=off ./...` 全數通過。
- **反向檢查**（在容器副本中修改）：
  - 建立事件改回其他 aggregate → F8 寫入測試 6 個子測試、`category_test.go` 6 個 spec、`NewMux` 的 2 個 spec 失敗（計畫書段 4）。
  - 計畫外另加：
    - 移除補版本紀錄的 migration → migration 測試中「從未修改」與「只在商戶 2 有紀錄」共 4 個子測試失敗。
    - 修改與刪除略過啟用檢查 → `category_test.go` 的 2 個重複刪除 spec 失敗。
    - 給 `prepaid_category.created` 加上舊映射 → 拒絕測試的該子測試失敗。
    - BFS payload 的一個 JSON tag 改名 → 契約測試失敗。
- **review 修正（F8 之後的版本不一致）**：
  - **問題**：F8 之後、段 4 之前，舊路徑在類別第一次修改時以「送出版本 + 1」建立 `category_uuid` 的紀錄，讀模型則變成 2。送出版本不是 1 時（例如 0）會留下「讀模型 2、aggregate 1」。原本的 migration 只補缺少的紀錄，F8 的 migration 也不會重跑，之後以畫面版本修改會回傳版本衝突。
  - **修正**：`20261022003` 先比照 F8，把已有紀錄但不同步的讀模型 version 改為紀錄的版本，再補缺少的紀錄。不調整 `aggregate_versions`，已存事件的 aggregate 版本不會倒退。
  - **測試**：migration 測試新增 3 個從 `20261022002` 開始的情境，兩種類別各一次。
    - 讀模型 2、aggregate 1 → 1。
    - 讀模型 2、aggregate 6 → 6。
    - 已同步的 2、2 不變。
    - 之後都以畫面版本經 BFS 修改成功。
  - **反向檢查**：移除對齊的步驟 → 4 個不同步的子測試失敗。全套 `go test -vet=off ./...` 通過，`git diff --check` 無輸出。
- **與計畫書的差異**：
  - F8 的 `category_version_test.go` 是在計畫書之後才加的，本段也改為經 BFS 寫入。「過期版本」的預期錯誤由舊路徑的 `UNIQUE constraint failed: aggregate_versions.` 改為版本衝突（情境 3）；migration 測試補上 aggregate 版本斷言，涵蓋情境 5、6。
  - `category_test.go` 沒有經 HTTP，而是以與 handler 相同的 `command.Build` 加 `Writer.Append` 寫入（`writeCategoryEvent`），因為 services 測試無法使用 handler 未公開的建構函式；專屬 API 的路由與組裝由新增的 `NewMux` 情境驗證。
  - `category_test.go` 送出的版本改為畫面會送的讀模型版本（1、2）。原本的 0 只是因為 ISSUE-029 才剛好成功。重複刪除的斷言改為比對 `is already inactive`。

段 5 驗證（2026-10-22，Docker `golang:1.26.9`）：
- **建置與格式**：
  - `go generate ./...`（含新增的 `enums.AggregatePeriod`）、`go build ./...` 通過（5a、5b、5c 各自執行）。
  - `go vet`（`domain/event`、`process`、`enums`、`database`、`persistence/handle`、`persistence/repos/projection`、`application/command`、`services`、`handler`）通過。
  - 修改與新增的 Go 檔 `gofmt -l` 無輸出；`git diff --check` 無輸出。gofmt 會把 doc comment 中的 `''` 改成 `”`，因此 doc comment 的空 aggregate ID 改寫為 `("ACCOUNT", "")`。
- **測試**（計畫書段 5 的情境）：
  - 情境 1、3、5：F9 的 `TestPeriodStarted_PerMerchant` 3 個子測試，改經通用 API 的 BFS 派送，前端版本送 7。
    - 兩個商戶各自開立同一月份、同一年度。
    - 同一商戶重複開立時唯一鍵擋下，事件與版本紀錄都回復。
    - 版本改讀 `('PERIOD', "1")`。
  - 情境 1 另有：
    - `TestDispatchLegacy_PeriodStarted`：`PERIOD` 已在版本 3、前端送 0 → 寫入版本 4，封套欄位完整。
    - 經正式 `NewMux` 的 `period.month_started` 情境：200、BFS 執行資訊、開立 OPEN 的月份。
  - 情境 2：`TestDispatchLegacy_PeriodStartedRejected` 的 3 個子測試（`('ACCOUNT', "1")`、推導的 `('PERIOD', "1")`、`TRANSACTION`）通過。
  - 情境 4：`TestPeriodStarted_ClosedByLegacy`。BFS 開立 2026-10，以舊路徑的 `period.month_closed` 結帳成功，並開立 2026-11。
  - 情境 6：`TestPeriodPayloads_LegacyJSON` 的 2 種事件通過；`closing_uuid` 為開帳事件的 UUID（情境 1、4 的測試檢查）。
  - migration：`TestPeriodAggregateType_Migration` 通過。
    - 兩張表的資料列、欄位定義、索引與 UNIQUE 不變。
    - CHECK 接受 `PERIOD`、仍拒絕未知類型。
    - 下一筆 `event_id` 為保存的計數 + 1。
  - `go test -vet=off ./...` 全數通過；handler 41 個 spec 通過。
- **反向檢查**（在容器副本中修改）：
  - 版本策略改為「驗證」→ `TestDispatchLegacy_PeriodStarted`、F9 的 3 個子測試與跨路徑結帳測試失敗（計畫書段 5）。
  - 計畫外另加：
    - 舊映射改為一律通過 → 情境 2 的 3 個子測試失敗。
    - migration 不還原計數 → migration 測試失敗（下一筆 `event_id` 為 4）。
    - 拿掉期間 projector 的註冊 → command、F9、跨路徑測試，以及 `NewMux` 情境失敗。`NewMux` 情境原本只檢查事件，攔不到這個問題，因此補上 `wantPeriod`，檢查開立的期間。
    - BFS payload 的 JSON tag 改名 → 契約測試失敗。
- **與計畫書的差異**：
  - F9 的 `period_unique_test.go` 是在計畫書之後才加的，本段改經通用 API 的 BFS 派送。讀取的版本改為 `('PERIOD', "1")`（計畫書第 5 節第 2 點），情境與其餘預期不變。
  - migration 不沿用 `20260614001` 的「`_new` → 改名」，改採 F9 的同名重建並保存 AUTOINCREMENT 計數（`doc/NOTES.md` 資料庫第 3、4 點）。
  - handler 的通用 API 情境新增 `wantPeriod` 欄位。

R2 全部完成（段 1～5）。下一步為 R3（BFS 重建機制），計畫書：`doc/plan/BFS_R3_REBUILD.md`（2026-10-27 撰寫，待 review）；ISSUE-031 依其 D3 處理。子階段在實作開始時展開。

## 最終驗證結果

R10 完成後記錄全套驗證結果。
