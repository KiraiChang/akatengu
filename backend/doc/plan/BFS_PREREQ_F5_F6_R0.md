# F5、F6、R0 實作計畫書

| 項目 | 內容 |
|---|---|
| 狀態 | ⏳ 待 review（2026-10-21：需求與範圍已確認，見 1.3） |
| 建立日期 | 2026-10-21 |
| 最後修訂 | 2026-10-21：依 review 修正兩處。F6 的 `account.created` 改用 INSERT，重複新增時直接失敗，避免版本被重設；R0 綁定交易的 Repo 盤點出全部 10 個寫入方法，一律拒絕 |
| 分支 | `refactor`（基準 commit `cedfff7`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md`：5.2 的 F5、F6，5.3 的 R0 |
| 關聯 | ISSUE-024、ISSUE-026、ISSUE-027；ADR-023、ADR-027 |

> 本文件細化上層計畫書的三個項目。F5、F6 是 bug fix，R0 是重構；依 `AGENTS.md`「變更幅度與重構」，三項各自實作、各自 review、各自 commit。
> 文中行號以基準 commit 為準。

---

## 1. 共通事項

### 1.1 執行順序

F5 → F6 → R0。三項彼此不依賴；每一項完成並經 review、commit 後，才開始下一項。

### 1.2 分段

三項的修改都達到 5 個檔案以上，依 `backend/CLAUDE.md` 拆成兩段：段 1 是程式，段 2 是測試。開始某一項時，在 `doc/STEP.md` 該項下方加上兩段的子進度；每段完成後執行 `go build ./...`，再更新狀態。

### 1.3 已確認的決定（2026-10-21）

1. F5 一併補回 `POST /api/sys/account`。
2. F5 沿用種子事件的 aggregate，`ExpectedVersion` 由後端讀取目前版本。
3. F6 一併修正科目的初始版本，並以 migration 修正既有資料。
4. 執行順序與 commit 方式如 1.1。

### 1.4 驗證流程

以 Docker（`golang:1.26`）執行，環境限制見 `doc/NOTES.md`：

1. `go generate ./...`、`go build ./...`。
2. `go test -vet=off ./...`（`go test` 附帶的 vet 分析 modernc sqlite 時會 OOM），另對修改的套件執行 `go vet`。
3. 修改與新增的檔案執行 `gofmt -l`，不得有輸出。
4. 反向檢查：在容器內的副本把修正還原（方式見各項），確認新測試會失敗，也就是測試確實能攔截原本的問題。
5. 回報時附上每個步驟的實際輸出。

---

## 2. F5：補回設定與系統科目的寫入 API（ISSUE-026）

### 2.1 問題

commit `9e3fcfc` 把三種設定改為事件溯源，同時移除了寫入 API，但前端仍在呼叫：

| 前端畫面 | 呼叫位置 | API |
|---|---|---|
| 帳戶類型科目設定 | `LedgerTypeConfig.svelte:62` → `api/setting.ts:23` | `PUT /api/setting/ledger-account-type/{type}` |
| 資產類型科目設定 | `AssetTypeConfig.svelte:116` → `api/setting.ts:47` | `PUT /api/setting/asset-type/{type}` |
| 系統科目對應 | `SysAccountConfig.svelte:52` → `api/sys.ts:14` | `POST /api/sys/account`（ISSUE-026 漏列） |

ISSUE-026 寫的「更新一定失敗」需要更正：沒有對應路由時，請求會落到 `mux.Handle("/", web.SPAHandler(...))`（`handler/handler.go:86`），而 SPA handler 對沒有副檔名的路徑一律回 `index.html`（`web/embed.go:46`）。所以前端收到 200、當作儲存成功，實際上沒有寫入任何資料。以上依程式碼推導，段 2 的反向檢查會確認。

### 2.2 API 與事件對應

| API | 前端送出 | 事件與 payload | aggregate |
|---|---|---|---|
| `PUT /api/setting/ledger-account-type/{type}` | `{account_id}` | `account_config.ledger_type_updated`：`type` 取自路徑，`account_id` 取自 body | `('ACCOUNT_CONFIG', "1")` |
| `PUT /api/setting/asset-type/{type}` | `AssetTypeAccountConfigUpdatedPayload` 去掉 `asset_type` 的其餘欄位（`types/setting.ts` 的 `UpdateAssetTypePayload`） | `account_config.asset_type_updated`：`asset_type` 取自路徑，其餘欄位取自 body | `('ACCOUNT_CONFIG', "1")` |
| `POST /api/sys/account` | `{sys_code, description, account_id}` | `sys_config.sys_account_updated`：欄位與前端送出的完全相同 | `('SYS_CONFIG', "1")` |

### 2.3 設計

1. **aggregate**：沿用 `BootstrapEventSeeds` 寫入種子事件時的 aggregate（`bootstrap/event_seed.go:21`、`:130`、`:180`）。`aggregate_versions` 的主鍵與 `event_store` 的唯一鍵都包含 `merchant_id`（`migrations/20260614001_bank_pdf_multiledger.sql:119`、`:149`），因此每個商戶各有一個。handler 以常數 `configAggregateID = "1"` 表示，並註明與種子事件相同；R2 會依 ADR-024 重新定義。
2. **ExpectedVersion**：Append 前以 `query.AggerateRepo.GetVersion` 讀取目前版本；沒有版本紀錄（`sql.ErrNoRows`）時視為 0，例如種子事件只為商戶 1 建立設定。保護程度與前端其他寫入用 `getAggregateVersion` 取最新版本相同，只擋得下「讀版本到 Append」之間的並行寫入。
3. **驗證**：handler 只做 payload 的 `Validate()`。科目是否存在由既有的 pipeline 檢查（`pipelines/factory/accountconfig.go`、`sysconfig.go`），讀模型由既有的 `ConfigProjectionService` 寫入（`services/projection/config.go`），兩者都不修改。
4. **路徑參數優先**：body 若也帶了 `type`／`asset_type`，以路徑的值為準。
5. **回應**：
   - 成功回 200，body 為 `null`，與被移除的設定 API 相同。
   - 請求解析、驗證、讀取版本或 Append 失敗時，回 400 與 problem 格式，沿用 `fixed_asset_category.go` 的做法。
   - 被移除的 sys API 失敗時回 401（commit `9e3fcfc` 刪除的 `handler/sys.go`），恢復時改為 400。
   - 前端只判斷 `res.ok` 與錯誤的 `detail`／`title`。

### 2.4 修改內容

**段 1：程式（3 個檔案）**

| 檔案 | 修改 |
|---|---|
| `internal/handler/setting.go` | `settingHandler` 加入 `es *services.EventStoreService` 與 `versions query.AggerateRepo` 欄位，建構函式改為 `newSettingHandler(db, es, logger)`。新增 `UpdateLedgerAccountTypeConfig`、`UpdateAssetTypeAccountConfig`。新增套件內共用的 `appendConfigEvent(ctx, es, versions, aggType, eventType, payload)`：marshal、讀取版本、呼叫 `es.Append` |
| `internal/handler/sys.go` | `sysHandler` 加入同樣的兩個欄位，建構函式改為 `newSysHandler(db, es, logger)`。新增 `UpdateSysAccount`，使用同一個 helper |
| `internal/handler/handler.go` | 兩個 handler 改為傳入 `eventService`；在既有的 GET 路由旁（`:158`、`:160`、`:161`）註冊 `POST /sys/account`、`PUT /setting/ledger-account-type/{type}`、`PUT /setting/asset-type/{type}` |

**段 2：測試（2 個檔案）**

- `internal/handler/setting_scenarios_test.go`：情境表。
- `internal/handler/setting_test.go`：fixture、helper 與 `Describe`。

兩個檔案共用既有的 Ginkgo suite（`merchant_suite_test.go`）。經 `NewMux` 的情境沿用 `entry_cf_category_test.go` 的 `insertCFAuthFixture` 與 JWT 常數。

### 2.5 測試情境

**測試資料**：

- 使用 `test.NewTestDBGinkgo()`（migration 加上 `accounts.sql` 種子），設為單一連線。
- 科目取自 `seeds/accounts.sql`；資產類型的科目值取自 `seeds/account_type_configs.sql` 的 STOCK 列，系統科目取自 `seeds/sys_accounts.sql`。
- 測試 DB 沒有設定表與 `sys_accounts` 的種子，第一次寫入時由 projection 新增。

**直接呼叫 handler**（ctx 帶商戶 1 與使用者，路徑參數以 `SetPathValue` 設定）：

| # | 情境 | 預期 |
|---|---|---|
| 1 | `PUT ledger-account-type/BANK_ACCOUNT`，body `{"account_id":"1101-02"}`，`ACCOUNT_CONFIG` 還沒有版本紀錄 | 200；寫入一筆 `account_config.ledger_type_updated`：aggregate `('ACCOUNT_CONFIG', "1")`、版本 1、payload `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`、`updated_by` 為使用者；`ledger_account_type_config` 的 BANK_ACCOUNT 為 `1101-02` |
| 2 | 同 1，但 `ACCOUNT_CONFIG` 已在版本 3（模擬已有種子事件的商戶） | 200；事件版本為 4 |
| 3 | `PUT ledger-account-type/SAVINGS` | 400；沒有事件 |
| 4 | `PUT ledger-account-type/LOAN`，body `{"account_id":"9999-99"}`（不存在的科目） | 400；沒有事件，設定表沒有 LOAN |
| 5 | `PUT ledger-account-type/LOAN`，body `{}` | 400；沒有事件 |
| 6 | `PUT asset-type/STOCK`，body 為種子的 STOCK 科目，`oci_account_id` 與 `account_id` 為 null | 200；一筆 `account_config.asset_type_updated`，版本 1，payload 的 `asset_type` 為 STOCK；`asset_type_account_config` 的 STOCK 列與送出的值相同 |
| 7 | `PUT asset-type/STOCK`，缺 `fee_account_id` | 400；沒有事件 |
| 8 | `PUT asset-type/BOND` | 400；沒有事件 |
| 9 | `POST sys/account`，body `{"sys_code":"SYS:EQUITY:EQUITY_OPENING","description":"期初權益","account_id":"3101-01"}` | 200；一筆 `sys_config.sys_account_updated`：aggregate `('SYS_CONFIG', "1")`、版本 1；`sys_accounts` 寫入該列 |
| 10 | `POST sys/account`，缺 `account_id` | 400；沒有事件 |
| 11 | 先更新帳戶類型，再更新資產類型 | 兩筆事件都屬於 `('ACCOUNT_CONFIG', "1")`，版本依序為 1、2 |

**經正式 `NewMux`**（含 JWT 與商戶驗證）：三條路由各送一筆成功請求，斷言與情境 1、6、9 相同。修正前這三個請求也會回 200（`index.html`），所以斷言以事件與讀模型為準，不能只看狀態碼。

**反向檢查**：在容器內的副本刪除 `handler.go` 新增的三條路由，確認三個 `NewMux` 情境都失敗。

### 2.6 預期的行為變更

- 三個 API 恢復可用：寫入對應的事件，並更新讀模型。
- 同一商戶同時更新設定時，後到的請求會因版本衝突失敗。舊路徑的版本衝突以唯一鍵錯誤呈現（ISSUE-024），使用者重送即可；被移除的舊 API 直接 UPSERT，沒有這個限制。

### 2.7 文件

- ISSUE-026：補上 sys API；把「一定失敗」更正為「回 200 但沒有寫入」；改為 Resolved，並記錄解決方式。
- `TODO.md`：F5 標為完成；新增 2.8 第 1 點的 P3 項目。
- 上層計畫書 5.2 的 F5：補上 sys API，更新狀態。
- `doc/STEP.md`：F5 的子進度與狀態。

### 2.8 風險與範圍外

1. `sys_config.sys_account_updated` 的 pipeline 不檢查 `sys_code` 是否存在，送出不存在的代碼時，projection 會新增一筆系統科目；被移除的舊 API 是 UPDATE，不會新增。前端只會送既有的代碼，本項不修改，記入 `TODO.md`。
2. `model/request/setting.go`、`model/request/sys.go` 的請求型別在 `9e3fcfc` 之後就沒有使用。本項改為直接解析成事件 payload，不刪除這些型別（無關清理）。
3. 前端不修改。

---

## 3. F6：讓業務版本檢查真正生效（ISSUE-027）

### 3.1 問題

ISSUE-027 列出的三點：

1. `account.updated` 的 projection 呼叫 `UpsertAccount`，衝突時以 `version = excluded.version` 覆寫（`queries/account.sql:128`、`:141`），沒有版本條件，版本也不會遞增。
2. `ledger_account.updated`、`investment.updated` 的 `UPDATE ... AND version = ?`（`queries/account.sql:110`、`queries/investment.sql:100`）是 `:exec`，projection repository 不檢查影響筆數。版本不符時更新 0 筆，但事件照樣寫入，API 也回應成功。

盤點時另外發現：**目前所有科目都無法修改**。以下依程式碼推導，段 2 會以測試確認。

1. `account.created`、`account.bulk_imported` 的 projection 以 `Version: 0` 寫入（`services/projection/account.go:63`、`:91`）。
2. 啟動時的種子事件 `account.bulk_imported`（`bootstrap/event_seed.go:119`）經 `UpsertAccount` 衝突覆寫，把種子 SQL 設定的 version 1 改成 0。
3. 前端修改科目時送出畫面上的 version（`Accounts.svelte:145`），而 `AccountUpdatedPayload.Validate` 要求 version 大於 0（`model/payload/account.go:105`）。
4. 結果：由種子事件建立的科目與使用者新增的科目，都會在驗證時被拒絕。

review 另外指出：`account.created` 的 projection 也呼叫 `UpsertAccount`（`services/projection/account.go:78`），而 pipeline 只驗證 payload（`pipelines/factory/account.go:33`），不檢查科目是否已存在。所以用同一個編號重複新增時，會覆寫既有的科目，並把版本重設為初始值。只修正初始版本的話，會出現以下情況：

1. 使用者 A 打開畫面，拿到 version 1。
2. 使用者 B 修改，version 變成 2。
3. 有人用同一個編號重複新增科目，資料被覆寫，version 重設為 1。
4. A 依過期畫面送出的修改，又能通過版本檢查。

### 3.2 設計

**比對放在 projection 的 UPDATE 條件**：

- 舊路徑在 `UnitOfWork.Do` 內依序更新版本紀錄、寫入事件、套用 projection（`services/event.go:72`、`:118`）；projection 回傳錯誤時，整個交易回滾。
- 因此把業務版本的比對放在 projection 的 `UPDATE ... WHERE version = ?`，影響 0 筆就回傳錯誤。比對與事件寫入在同一個交易內，版本不符時，事件、版本紀錄與讀模型都不會改變。

**新增科目改用 INSERT**：單筆的 `account.created` 改用既有的 `CreateAccount`（`queries/account.sql:84`）。`CreateAccount` 是 INSERT，並固定寫入 version 1。重複的編號由 `accounts` 的唯一鍵 `(account_id, merchant_id)` 擋下，事件與版本紀錄一併回滾。種子匯入 `account.bulk_imported` 仍保留 UPSERT（見 3.7）。

**不在 pipeline 比對**：pipeline 在交易外查詢（`services/event.go:55`），比對完到寫入之前，資料仍可能被其他請求修改。檢查科目是否已存在也是同樣的道理，所以交給交易內的唯一鍵處理，不在 pipeline 另外檢查。

**錯誤**：

- `projection_repo` 新增 `ErrVersionConflict`，訊息說明資料不存在或已被修改、需重新整理；回傳時以 `%w` 包上資料識別與送出的版本。
- projection 的錯誤以 `%w` 往上傳（`services/event.go:119`），呼叫端可以用 `errors.Is` 判斷。
- 經 `POST /api/event/append` 回傳 500 與 problem 格式（`handler/event.go:103`），前端把 `detail` 顯示在 `saveError`。

**遷移時**：R2、R9 遷移到 BFS 時，依 ADR-027 規則第 5 點保留同樣的行為。

### 3.3 修改內容

**段 1：程式（6 個檔案）**

| 檔案 | 修改 |
|---|---|
| `internal/database/queries/account.sql` | `UpdateAccount`（`:89`）、`UpdateLedgerAccount`（`:110`）由 `:exec` 改為 `:execrows` |
| `internal/database/queries/investment.sql` | `UpdateInvestment`（`:100`）由 `:exec` 改為 `:execrows` |
| `internal/repos/unit_of_work/event_store/projection_repo/account.go` | 新增 `ErrVersionConflict`；`UpdateAccount`、`UpdateLedgerAccount` 影響 0 筆時回傳它 |
| `internal/repos/unit_of_work/event_store/projection_repo/investment.go` | `UpdateInvestment` 影響 0 筆時回傳 `ErrVersionConflict` |
| `internal/services/projection/account.go` | `applyAccountUpdated` 由 `UpsertAccount` 改用 `UpdateAccount`，更新的欄位相同，並帶版本條件與 `version = version + 1`。`applyAccountCreate` 由 `UpsertAccount` 改用 `CreateAccount`，SQL 固定寫入 version 1，因此移除不再使用的 `Version: 0`。`applyAccountBulkImported` 保留 `UpsertAccount`，初始版本改為 1。三者都與實體帳戶、投資的初始版本一致 |
| `internal/database/migrations/20261021001_account_initial_version.sql` | 新增 goose migration。Up：`UPDATE accounts SET version = 1 WHERE version = 0`。Down 不還原，因為無法分辨原本就是 0 的資料 |

補充：

- `go generate ./...` 會重新產生 sqlc 程式碼；不讀取、也不手改 `sqlcdb/`。
- `projection_repo/interface.go` 的方法簽章不變。
- `projection_repo` 的 `CreateAccount`、`UpdateAccount` 目前沒有呼叫端，本項開始分別由 `applyAccountCreate`、`applyAccountUpdated` 使用。
- `accounts.updated_at` 有預設值 `datetime('now')`（`migrations/20260519001_add_audit_fields.sql:8`），`CreateAccount` 沒有寫入這個欄位也沒有影響。
- 新 DB 執行 migration 時 `accounts` 表還是空的（種子在 migration 之後才寫入），所以不受影響。

**段 2：測試（1 個檔案）**

`internal/services/update_version_test.go`：package `services_test`，標準 table-driven。

- 一律經 `EventStoreService.Append` 寫入。
- 事件封套沿用前端使用的 aggregate：科目與實體帳戶是 `('ACCOUNT', account_id)`，投資是 `('TRANSACTION', '')`。
- 每次 Append 前讀取該 aggregate 的目前版本，與前端的 `getAggregateVersion` 相同。

### 3.4 測試情境

**測試資料**：使用 `test.NewTestDB(t)`，設為單一連線，先以 Append 建立：

- 科目 `1101-90`：`account.created`，parent 為 `1101`，ASSET／DEBIT。
- 科目 `1101-91`：`account.bulk_imported`。
- 實體帳戶：`ledger_account.created`，科目 `1101-02`，BANK_ACCOUNT。
- 投資：`investment.created`，科目 `1203-02`，STOCK／FIFO／FVOCI。

`1101`、`1101-02`、`1203-02` 取自種子；`1101-90`、`1101-91` 不在種子中。

| # | 情境 | 預期 |
|---|---|---|
| 1 | `account.created` 之後，以 version 1 修改名稱 | 成功；名稱更新、version 變成 2；`event_store` 多一筆 |
| 2 | `account.bulk_imported` 之後，以 version 1 修改 | 成功；version 變成 2 |
| 3 | 科目已在 version 2，再以 version 1 修改（過期的畫面） | `errors.Is(err, projection_repo.ErrVersionConflict)`；名稱、version、`event_store` 筆數與 aggregate 版本都不變 |
| 4 | 科目 `1101-90` 已在 version 2，再以同一個編號送出 `account.created`（名稱不同） | 回傳錯誤（唯一鍵）；名稱、version 2、`event_store` 筆數與 aggregate 版本都不變 |
| 5 | 接續情境 4，再以過期的 version 1 修改 | 仍然回傳 `ErrVersionConflict`；資料不變 |
| 6 | 修改不存在的科目 `9999-99` | `ErrVersionConflict`；不會新增科目 |
| 7 | 實體帳戶以 version 1 修改 | 成功；version 變成 2 |
| 8 | 實體帳戶以過期的 version 修改 | `ErrVersionConflict`；資料與 `event_store` 都不變 |
| 9 | 修改不存在的 `ledger_id` | `ErrVersionConflict` |
| 10 | 投資以 version 1 修改 | 成功；version 變成 2 |
| 11 | 投資以過期的 version 修改 | `ErrVersionConflict`；資料與 `event_store` 都不變 |

migration 只有一行資料修正，不另寫測試；套用到開發 DB 前後，可用 `SELECT version, COUNT(*) FROM accounts GROUP BY version` 確認。

**反向檢查**：在容器內的副本保留 `ErrVersionConflict` 的宣告，其餘 5 個程式檔還原為 HEAD 的寫法，預期：

- 情境 1、2 因為 version 0 而在驗證時失敗。
- 情境 3、8、11 不會回傳錯誤。
- 情境 4 的重複新增會成功，並覆寫既有的科目。

### 3.5 預期的行為變更

- 依過期畫面送出的修改改為回傳錯誤。原本科目會被覆蓋，實體帳戶與投資則是無聲地不生效。
- 科目修改後 version 會遞增；原本會被寫回畫面上的值。
- 新科目的初始版本改為 1；migration 把既有 version 0 的科目改為 1。
- 修改不存在的科目改為回傳錯誤；原本 `UpsertAccount` 會新增一筆。
- 以已存在的編號新增科目改為回傳錯誤，錯誤訊息是資料庫的唯一鍵錯誤；原本會覆寫既有的科目，並重設版本。
- 前端不修改：三個畫面失敗時都會顯示 `saveError`，成功後都會重新載入（`Accounts.svelte:176-178`、`Ledger.svelte:218-220`、`Investment.svelte:577-579`）。

### 3.6 文件

- ISSUE-027：補上科目初始版本的發現；改為 Resolved，並記錄解決方式。
- `TODO.md`：F6 標為完成。
- 上層計畫書 5.2 的 F6：補上初始版本的修正，更新狀態。
- `doc/STEP.md`：F6 的子進度與狀態。

### 3.7 風險

1. 舊 replay 重播 `account.updated` 時，會拿事件內的版本比對。修正前寫入開發 DB 的事件，版本不會連續（修正前 version 不遞增），重播會失敗。舊 replay 已經 Wontfix，R10 之前不使用（上層計畫書第 6 節）。
2. `account.bulk_imported` 仍是 UPSERT，會把既有科目的 version 重設為 1。目前只有種子事件會送出它，而且只在 `ACCOUNT` aggregate 還沒有事件時送一次，不會在使用者修改之後發生。

---

## 4. R0：讓查詢層可以綁定交易

### 4.1 目標

讓 `query.Repo` 可以綁定 `*sqlx.Tx`，供 R1 的 handler 在事件圖交易內查詢（ADR-023 A1 第 2 點：交易內的查詢 repository 由 `Run` 建立，屬於單次 `Run`）。

本項只提供這個能力，不改任何呼叫端。預期的行為變更：無。

### 4.2 現況

- `NewQueryRepository(db *sqlx.DB)`（`repos/query/base.go:37`）以 `sqlcdb.New(db)` 建立 sqlc 查詢。`sqlcdb.New` 也接受 `*sqlx.Tx`，兩個 UnitOfWork 都是這樣用的（`unit_of_work/event_store/implement.go:28`、`persistence/repos/unit_of_work.go:32`）。
- 6 個 repo 另外持有 `*sqlx.DB`：

  | repo | 如何使用 |
  |---|---|
  | investment | `QueryRowxContext`、`SelectContext`（`investment.go:223`、`:239`） |
  | report | `Rebind`、`SelectContext`、`GetContext`（動態查詢） |
  | account、config、dashboard | 持有但沒有使用 |
  | template | `CreateTemplate`、`UpdateTemplate` 以 `r.db.DB.BeginTx` 自己開交易寫入（`template.go:114`、`:150`）；`DeleteTemplate` 以 `r.q` 寫入 |

- report 使用的 `sqlx.Named` 是套件函式，不需要連線。
- `query.Repo` 雖然定位為唯讀查詢層，實際上有 10 個寫入方法。盤點方式：比對 `queries/*.sql` 中所有寫入類的查詢名稱，與 `repos/query/` 實際呼叫的查詢；`repos/query/` 沒有直接以 sqlx 執行的寫入。

  | repo | 寫入方法 | 目前的呼叫端 |
  |---|---|---|
  | Config | `UpsertLedgerAccountTypeConfig`（`config.go:70`）、`UpsertAssetTypeAccountConfig`（`:114`） | 無（`9e3fcfc` 起設定改由 projection 寫入） |
  | Merchant | `Create`（`merchant.go:28`）、`AddUser`（`:63`）、`Update`（`:71`）、`Deactivate`（`:80`） | 商戶管理 |
  | User | `Create`（`user.go:40`） | 使用者註冊 |
  | Template | `CreateTemplate`（`template.go:109`）、`UpdateTemplate`（`:145`）、`DeleteTemplate`（`:183`） | 交易範本 |

### 4.3 設計

1. `base.go` 新增未匯出的介面 `sqlxQuerier`，包含 `Rebind`、`GetContext`、`SelectContext`、`QueryRowxContext`；`*sqlx.DB` 與 `*sqlx.Tx` 都滿足它。
2. 建構函式：
   - `NewQueryRepository(db)` 對外不變，內部改為呼叫 `newQueryRepository(sqlcdb.New(db), db, db)`。
   - 新增 `NewQueryRepositoryTx(tx *sqlx.Tx) *Repo`，內部呼叫 `newQueryRepository(sqlcdb.New(tx), tx, nil)`。
3. investment、report、account、config、dashboard 的 `db` 欄位與 `newXxxRepo` 參數改為 `sqlxQuerier`。公開的 `NewXxxRepo(db *sqlx.DB)` 與各方法的內容都不變。
4. **綁定交易的 Repo 只供讀取**：4.2 盤點的 10 個寫入方法，在綁定模式下一律回傳匯出的 `ErrTxBoundWrite`，不寫入任何資料，也不開巢狀交易。
   - 實作方式：新檔 `tx.go` 為 Config、Merchant、User、Template 各定義一個包裝型別。包裝型別內嵌原本的介面，讀取方法直接委派，只覆寫寫入方法。
   - `NewQueryRepositoryTx` 用這些包裝型別取代對應的欄位。原本的 repo 檔案（`config.go` 的寫入方法、`merchant.go`、`user.go`、`template.go`）不修改，一般模式的行為不變。
   - template 仍持有 `*sqlx.DB`；綁定模式沒有 `*sqlx.DB`，所以傳入 nil。寫入方法已被包裝型別攔下，不會用到這個 nil。
   - `tx.go` 的註解列出這份清單，說明在這 4 個 repo 新增寫入方法時，必須同步加入包裝型別；段 2 的測試逐一涵蓋清單中的方法。
5. `NewQueryRepositoryTx` 的註解說明：綁定的 Repo 只在交易有效期間可用，生命週期由呼叫端管理（R1 的 `Run`）。
6. 不修改 mediator（`runtime/mediator/mediator.go:29`）、handler 與 engine；R1 才改用 `NewQueryRepositoryTx`。

### 4.4 修改內容

**段 1：程式（7 個檔案）**

`internal/repos/query/` 下的 7 個檔案：

- 修改：`base.go`、`account.go`、`config.go`、`dashboard.go`、`investment.go`、`report.go`。其中 `config.go` 只改 `db` 欄位的型別。
- 新增：`tx.go`，包含 `NewQueryRepositoryTx`、`ErrTxBoundWrite` 與 4 個包裝型別。

`template.go`、`merchant.go`、`user.go` 不修改。

**段 2：測試（1 個檔案）**

`internal/repos/query/tx_test.go`：package `query_test`，標準 table-driven，沿用 `account_test.go` 的 `testCtx`。

### 4.5 測試情境

**測試資料**：使用 `test.NewTestDB(t)`，設為單一連線。開啟交易後，以 SQL 寫入：

- 科目 `1101-92`（parent `1101`）。
- 一筆投資：科目 `1203-02`，uuid 為固定值。

其餘資料取自種子。

| # | 情境 | 預期 |
|---|---|---|
| 1 | 交易內寫入後，以 `NewQueryRepositoryTx(tx).Account.GetAccount` 查詢（sqlc 路徑） | 查得到 `1101-92` |
| 2 | 以綁定交易的 `Investment.GetAllSummaries`（`SelectContext`）與 `GetSummary`（`QueryRowxContext`）查詢 | 查得到該筆投資 |
| 3 | 交易進行中，以一般的 `NewQueryRepository(db)` 查詢，context 設 200ms timeout | 回傳 `context.DeadlineExceeded`。說明為什麼需要 R0：唯一的連線被交易占用 |
| 4 | 交易回滾後，以一般的 Repo 查詢 | 查不到 `1101-92` 與該筆投資 |
| 5 | 以綁定交易的 Repo 逐一呼叫 4.2 的 10 個寫入方法（table-driven，每列一個方法） | 都回傳 `ErrTxBoundWrite`；交易內對應資料表（`ledger_account_type_config`、`asset_type_account_config`、`merchants`、`user_merchants`、`users`、`transaction_templates`）的內容不變 |
| 6 | 以一般的 `NewQueryRepository(db)` 呼叫 `Config.UpsertLedgerAccountTypeConfig` | 照常寫入，一般模式不受包裝型別影響 |
| 7 | 既有測試 | 全數通過，`NewQueryRepository` 的行為不變 |

report 的動態查詢與 investment 一樣改用 `sqlxQuerier`，能否綁定交易由編譯期的型別檢查保證，不另寫情境；既有的 report 測試涵蓋一般路徑。

**反向檢查**：本項沒有修正既有行為。情境 3 說明了問題本身：修改前，在交易內查詢只能透過一般的 Repo，會等不到連線。

### 4.6 文件

- 上層計畫書 5.3 的 R0：更新狀態。
- `doc/STEP.md`：R0 的子進度與狀態。
- `backend/CLAUDE.md`「查詢層 `query.Repo`」：補一句，說明可以用 `NewQueryRepositoryTx` 綁定交易，供 R1 起的 BFS handler 使用。
- `TODO.md` 新增 P3：`query.Repo` 定位為唯讀查詢層，卻有 Config、Merchant、User、Template 的寫入方法；其中 Config 的兩個 UPSERT 從 `9e3fcfc` 起就沒有呼叫端。是否移出查詢層，另行處理。

### 4.7 風險與範圍外

1. 交易結束後再使用綁定的 Repo，會回傳 `sql.ErrTxDone`；由 R1 的 `Run` 管理生命週期。
2. account、config、dashboard 的 `db` 欄位目前沒有使用；本項只改型別，不刪除（無關清理）。
3. 包裝型別內嵌原本的介面。之後如果在這 4 個 repo 新增寫入方法，卻忘了加進包裝型別，綁定模式就會放行該寫入。對策是 `tx.go` 的註解與段 2 的清單測試；review 新增的寫入方法時，一併檢查包裝型別。

---

## 5. 各項的 commit 範圍

| 項目 | 程式與測試 | 文件 |
|---|---|---|
| F5 | 2.4 的 3 個程式檔與 2 個測試檔 | `ISSUES.md`、`TODO.md`、上層計畫書、`STEP.md` |
| F6 | 3.3 的 6 個檔案與 1 個測試檔 | `ISSUES.md`、`TODO.md`、上層計畫書、`STEP.md` |
| R0 | 4.4 的 7 個程式檔（含新增的 `tx.go`）與 1 個測試檔 | 上層計畫書、`STEP.md`、`backend/CLAUDE.md`、`TODO.md` |

本計畫書 review 通過後，先單獨 commit。
