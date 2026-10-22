# R2 實作計畫書：主檔、設定與期間開帳改走 BFS

| 項目 | 內容 |
|---|---|
| 狀態 | ⏳ 待 review |
| 建立日期 | 2026-10-22 |
| 最後修訂 | 2026-10-22：依 review 修訂兩處。D12 取消「舊 Replay 讀模型相同」的承諾，只保留 payload JSON 與 UUID 推導的契約測試，完整重建的驗收留給 R3；review 以工作區副本確認整數 `ledger_id` 會讓舊 Replay 失敗，登錄為 ISSUE-031。D6 的既有類別資料改為兩步：F8 以 migration 把已不同步的讀模型版本對齊 aggregate 版本，段 4 只補從未修改、沒有版本紀錄的類別；兩處都補上「舊版已修改一次 → 修正／遷移 → 再次修改」情境 |
| 分支 | `refactor`（基準 commit `3f811e6`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md`：5.3 的 R2 |
| 前置 | F5、F6、F7、R0、R1（皆已完成）；本計畫盤點時發現的 F8（ISSUE-029）需在段 4 之前完成，F9（ISSUE-030）建議在段 5 之前完成 |
| 關聯 | ADR-023、ADR-024（A7）、ADR-027；ISSUE-018、ISSUE-028、ISSUE-029、ISSUE-030、ISSUE-031 |

> R2 是重構，依 `AGENTS.md`「變更幅度與重構」，只做本計畫列出的行為變更（第 5 節）；盤點時發現的 bug 另開 F8、F9，不併入 R2。
> 文中行號以基準 commit 為準。

---

## 1. 範圍

### 1.1 遷移的事件類型

依上層計畫第 3 節 B 類與 A 類的期間開帳，共 16 種。原則 9 的盤點：每種事件類型在 `services/projection/` 只有一個處理者（以事件類型常數搜尋全部非測試程式確認），餘額類 projection 都不處理這些事件。

| 事件類型 | 目前的寫入端 | 舊 pipeline 的驗證（`services/pipelines/factory/`） | 舊 projection |
|---|---|---|---|
| `account.created` | 通用 API（`frontside/src/api/account.ts:25`） | payload | `AccountProjectionService`（`services/projection/account.go`）：INSERT、version 1 |
| `account.updated` | 通用 API（`account.ts:37`） | payload | 同上：帶版本條件的 UPDATE（F6） |
| `account.bulk_imported` | 種子（`bootstrap/event_seed.go:115`） | payload | 同上：逐筆 UPSERT |
| `ledger_account.created` | 通用 API（`ledger.ts:33`） | payload | 同上：INSERT，`ledger_uuid` 為事件 UUID |
| `ledger_account.updated` | 通用 API（`ledger.ts:47`） | payload | 同上：帶版本條件的 UPDATE（F6） |
| `sys_config.sys_account_updated` | `POST /api/sys/account`（`handler/sys.go:52`）、種子（`event_seed.go:164`） | payload、科目存在 | `ConfigProjectionService`（`config.go`）：UPSERT |
| `account_config.ledger_type_updated` | `PUT /api/setting/ledger-account-type/{type}`（`handler/setting.go:110`）、種子（`event_seed.go:215`） | payload、科目存在 | 同上 |
| `account_config.asset_type_updated` | `PUT /api/setting/asset-type/{type}`（`setting.go:161`）、種子（`event_seed.go:264`） | payload、各科目存在 | 同上 |
| `prepaid_category.created`／`updated`／`deleted` | 專屬 API（`handler/prepaid_category.go`） | payload；修改與刪除另查類別存在且啟用 | `PrepaidCategoryProjectionService`：建立時 `category_uuid` 為事件 UUID |
| `fixed_asset_category.created`／`updated`／`deleted` | 專屬 API（`handler/fixed_asset_category.go`） | 同上 | `FixedAssetCategoryProjectionService`：同上 |
| `period.month_started` | 通用 API（`period.ts:41`） | payload | `PeriodProjectionService.applyMonthStarted`：INSERT `period_closings`，`closing_uuid` 為事件 UUID |
| `period.annual_started` | 通用 API（`period.ts:79`） | payload | `applyAnnualStarted`：同上 |

前端經通用 API 只送出其中 6 種（科目、明細帳戶、期間開帳）；類別、設定與 `account.bulk_imported` 沒有前端經通用 API 的呼叫（`frontside/src/` 搜尋事件字串確認）。

### 1.2 盤點時發現的問題

以容器內的工作區副本執行一次性的探測測試（不留在 repo），經 `EventStoreService.Append` 沿用各寫入端的封套：

1. **ISSUE-029：類別的第二次修改或刪除必定失敗。** 建立事件的 aggregate 是 handler 隨機產生的 UUID（`handler/prepaid_category.go:52`），`category_uuid` 卻是事件 UUID（`services/projection/prepaid_category.go:31` 的 `applyCreated`），修改與刪除寫到另一個 aggregate；讀模型的 `version` 從不遞增（`database/queries/prepaid_category.sql:12-26`），畫面一律送 1。第一次修改時該 aggregate 沒有版本紀錄，舊路徑以「送出版本 + 1」新建紀錄；第二次仍送 1，比對不符後再新建，撞上唯一鍵。探測結果（預付類別）：第一次修改成功，aggregate 版本 2、讀模型版本 1；第二次回傳 `UNIQUE constraint failed: aggregate_versions...`。固定資產類別的程式相同，依程式碼推導。以 **F8** 修正（第 1.3 節）。
2. **ISSUE-030：不同商戶無法開立同一期間。** `period_closings` 的唯一鍵是 `UNIQUE(period_type, period_start)`（`database/migrations/20260310001_init_table.sql:215`），之後加入 `merchant_id`（`20260514002_add_merchant_id.sql:16`）時沒有更新唯一鍵。探測結果：商戶 1 開立 2026-10 後，商戶 2 開立同月回傳 `UNIQUE constraint failed: period_closings.period_type, period_closings.period_start`。傳票、預付等 pipeline 都要求月份為 OPEN，因此商戶 2 在該月份無法記帳；也違反 R3 的驗收條件「重建商戶 A 不影響商戶 B」的前提。以 **F9** 修正。
3. **不是問題：同一商戶重複開立同一期間。** 舊 pipeline 只驗證 payload，但上述唯一鍵會擋下（探測：第二次回傳唯一鍵錯誤，只有一筆期間）。段 5 的「交易內」版本策略依賴這個唯一鍵保護「同一期間只開立一次」（ADR-027 規則 4 的代價）；F9 修正後唯一鍵改為每個商戶各自唯一，仍有相同保護。

### 1.3 新增的 bug fix（不屬於 R2 的實作範圍）

| 項目 | 內容 | 依賴 |
|---|---|---|
| F8 | ISSUE-029：(1) 預付與固定資產類別的修改、刪除 projection 遞增讀模型的 `version`，讓畫面送出的版本與 aggregate 版本一致；(2) migration 修正 ISSUE-029 已留下的不一致：`category_uuid` 的 aggregate 已有版本紀錄時，讀模型的 `version` 改為該紀錄的版本（例如修改過一次的類別：讀模型 1、aggregate 2 → 讀模型 2）。補測試重現「連續修改兩次」與「舊版已修改一次 → migration → 再次修改」 | 無；需在 R2 段 4 之前完成 |
| F9 | ISSUE-030：以 migration 把 `period_closings` 的唯一鍵改為 `(merchant_id, period_type, period_start)`；補測試重現「兩個商戶開立同一期間」 | 無；建議在 R2 段 5 之前完成，R3 之前必須完成 |

兩項各自依 `ISSUES.md` 以小幅修改實作與 review，實作計畫另行確認。

---

## 2. 設計決策

| # | 問題 | 建議 | 理由 |
|---|---|---|---|
| D1 | payload 型別放在哪裡 | 16 種 payload 從 `model/payload` **複製**到 `domain/event`（含 `Validate`），JSON tag 完全相同；舊 pipeline、舊 projection 與舊 replay 繼續使用 `model/payload`，到 R10 移除 | F1（ISSUE-020）的教訓：兩種型別混用時斷言失敗，註冊表只解碼 `domain/event` 的型別。舊 replay 仍以舊 pipeline 重播 BFS 寫入的事件，JSON 必須一致，因此每段以測試比對新舊型別的 JSON。替代方案「在 `model/payload` 加上 `EventType()` 直接沿用」會讓註冊表依賴 model 層，違反 ADR-027 規則 1 的依賴限制，不採用。明細帳戶的 `credit_limit` 使用 `decimal.NullDecimal`，規則 1 的依賴清單補上 `shopspring/decimal` |
| D2 | `Spec.Aggregate` 的輸入 | 改為接收事件：`Aggregate func(evt kevent.Event) (enums.AggregateType, string, error)`；`newSpec` 的推導函式同時取得事件與型別化的 payload。cf 的登錄跟著調整 | 類別建立事件的身分是事件 UUID（D6），payload 沒有可用的欄位；其餘事件類型仍只看 payload |
| D3 | 根事件的版本策略如何落到寫入（ADR-027 規則 4） | kernel 事件新增 `RootVersion`（`json:"-"`，做法同 R1 的 `ChildAggregate`）：`RootVersionExpected`（零值，現行行為：`Store.Append` 與 `ExpectedVersion`）、`RootVersionNew`（`Store.Create`：已有任何版本紀錄，包含版本 0，都視為衝突）、`RootVersionCurrent`（`Store` 新增 `AppendCurrent`：在交易內讀取目前版本後寫入下一版，沒有紀錄時建立版本 1）。只影響根事件，子事件仍依 R1 的規則 | 版本策略由註冊表依事件類型決定，不看前端送了什麼；「新建」不能用 `Append` 加版本 0 表達（R1 review 的 P2）；「交易內」必須在 `Run` 的交易內讀取版本才有意義。handler 不自行設定版本（過渡規則） |
| D4 | handler、專屬 API 與種子如何共用寫入流程 | 新增 `internal/application/command`（對應 `TODO.md`「預計專案結構」的 application 層）：`NewEngine(db)` 組裝 BFS（mediator、projector、middleware、UoW），`NewMux` 與 bootstrap 共用；`Build` 依註冊表推導 aggregate、設定版本模式與封套欄位（`MerchantID`、`UpdatedBy` 取自 ctx）；`Append` 執行 `Run` 後以新增的 `query.Event.GetByUUID` 回讀 `event_store` 的紀錄；`DispatchLegacy` 處理通用 API 的派送（ADR-027 規則 2、3）：未登錄 → 交回舊路徑；已登錄但沒有舊映射 → 拒絕；送出的 aggregate 不等於舊映射 → 拒絕；其餘以 `Decode` 解碼後 `Build`、`Append` | bootstrap 不能依賴 handler（handler 已依賴 bootstrap），組裝放在兩者都能依賴的位置，避免兩處的 middleware 與 projector 不一致。這個套件可以依賴 `domain/event`、engine、process、persistence 與 query，反方向不得依賴它 |
| D5 | 錯誤與狀態碼 | 通用 API：拒絕與 BFS 失敗都回 500 與 problem 格式（ADR-027 規則 7），BFS 失敗不改走舊路徑；專屬 API 維持各自目前的狀態碼（設定、類別、現金流分類都是 400） | 規則 7 要求錯誤格式不變；專屬 API 屬於「API 回應不變」 |
| D6 | 類別建立事件的身分（A7） | aggregate ID 為**建立事件的 UUID**；`category_uuid` 的推導不變（仍是事件 UUID），修改與刪除以 `payload.category_uuid` 為 aggregate | 舊 projection 以事件 UUID 作為 `category_uuid`；若改用 handler 產生的 UUID 並讓 projector 讀 aggregate ID，舊 replay 會產生不同的 `category_uuid`，引用類別的預付與固定資產會斷裂。既有開發資料分兩步，完成後每個既有類別的讀模型版本都等於 `category_uuid` aggregate 的版本：(1) F8 的 migration 把已有版本紀錄、但讀模型版本落後的類別（ISSUE-029 修改過的類別）對齊 aggregate 版本；(2) 段 4 的 migration 為仍沒有版本紀錄的類別（從未修改）補上 `(類型, category_uuid, merchant_id, 讀模型 version)`，否則 BFS 的 `Append`（沒有紀錄且送出版本不是 0 時視為衝突，`persistence/repos/store.go:47`）會拒絕修改。不採用「前置條件為清空開發 DB」，因為兩步都只是單一 SQL，且 F8 本來就要修正它造成的資料 |
| D7 | 明細帳戶的 aggregate | 沿用舊映射 `('ACCOUNT', payload.account_id)`，即所屬科目的 aggregate | 修改事件以整數 `ledger_id` 指定明細帳戶，payload 沒有 `ledger_uuid`，不改 payload 就無法推導明細帳戶自己的 aggregate。代價：同時修改同一科目與其明細帳戶時會互相衝突 |
| D8 | 每個商戶一個的 aggregate | 期間改用新類型 `PERIOD`（A7），aggregate ID 沿用設定 aggregate 的慣例 `"1"`（`handler/setting.go:25` 的 `configAggregateID`），常數移到 `domain/event` 供設定、種子、期間共用；種子與設定沿用目前的 aggregate：`('ACCOUNT', "1")`、`('SYS_CONFIG', "1")`、`('ACCOUNT_CONFIG', "1")`，種子的冪等檢查不變 | aggregate 的鍵包含 `merchant_id`，`"1"` 是常數而不是商戶 ID；沿用既有慣例，種子與設定的事件歷史不必換 aggregate。`PERIOD` 需修改 `event_store`、`aggregate_versions` 的 `chk_aggregate_type`（做法同 `database/migrations/20260614001_bank_pdf_multiledger.sql:108-178`） |
| D9 | 業務版本（ADR-027 規則 5、F6） | `account.updated`、`ledger_account.updated` 的 mediator handler 以交易內查詢比對 `payload.version` 與讀模型的版本，不符時回傳衝突；projector 保留 F6 帶版本條件的 UPDATE，影響 0 筆時回傳錯誤。衝突訊息與 F6 相同（`projection_repo.ErrVersionConflict` 的文字），前端顯示不變 | 規則 5 要求在 handler 檢查；保留 projector 的條件是第二道保護，重建（R3）時依事件順序重播，版本會一致 |
| D10 | projector 與寫入方法 | `persistence/handle` 新增科目、設定、類別（預付、固定資產）、期間的 projector；`persistence/repos/projection` 新增呼叫相同 sqlc 查詢的方法。舊 projection 保留給舊 replay。期間只接手兩種開帳事件，結帳與重開仍由舊 `PeriodProjectionService` 處理（R5） | 沿用 F1 的 cf projector 做法；SQL 相同，讀模型的結果相同 |
| D11 | 種子 | `BootstrapEventSeeds` 改經 D4 寫入：每筆種子一次 `Run`（與目前一筆一個 `Append` 相同），版本模式為 `RootVersionExpected`，沿用目前送出的版本；`Source` 設為 `"seed"` | 冪等檢查與事件順序不變 |
| D12 | 舊 replay 相容 | R2 **不承諾**舊 `Replay` 重建的讀模型與寫入時相同；只保證舊 pipeline 能讀取 BFS 寫入的事件所需的契約：payload JSON 與新舊型別一致（D1）、`category_uuid`、`ledger_uuid`、`closing_uuid` 仍由建立事件的 UUID 推導（D6、D10）。每段以契約測試確認，完整重建的驗收留給 R3 | 舊 replay 是 Wontfix（ISSUE-021、022），且已知無法完整重建：review 以工作區副本確認，另一商戶保留 `ledger_id` 2 時，商戶 1 的明細帳戶重播後由 1 變成 3，之後的修改事件仍引用 1，重播失敗（ISSUE-031）。這不是 R2 造成的，R2 不承諾就不需要前置修正 |

## 3. 各事件類型的身分、版本策略與舊映射

| 事件類型 | 推導的 aggregate | 版本策略 | 舊映射（通用 API） | 業務版本 |
|---|---|---|---|---|
| `account.created` | `('ACCOUNT', payload.account_id)` | 新建 | 同推導 | — |
| `account.updated` | `('ACCOUNT', payload.account_id)` | 驗證 | 同推導 | `payload.version` |
| `account.bulk_imported` | `('ACCOUNT', "1")` | 驗證（種子送出的版本） | 無（拒絕） | — |
| `ledger_account.created` | `('ACCOUNT', payload.account_id)` | 驗證 | 同推導 | — |
| `ledger_account.updated` | `('ACCOUNT', payload.account_id)` | 驗證 | 同推導 | `payload.version` |
| `sys_config.sys_account_updated` | `('SYS_CONFIG', "1")` | 驗證（API 讀取目前版本，同 F5） | 無（拒絕） | — |
| `account_config.ledger_type_updated`、`asset_type_updated` | `('ACCOUNT_CONFIG', "1")` | 驗證（同上） | 無（拒絕） | — |
| `prepaid_category.created` | `('PREPAID_CATEGORY', 事件 UUID)` | 新建 | 無（拒絕） | — |
| `prepaid_category.updated`、`deleted` | `('PREPAID_CATEGORY', payload.category_uuid)` | 驗證（畫面版本，規則 6） | 無（拒絕） | — |
| `fixed_asset_category.*` | 同預付類別，類型為 `FIXED_ASSET_CATEGORY` | 同預付類別 | 無（拒絕） | — |
| `period.month_started`、`annual_started` | `('PERIOD', "1")` | 交易內 | `('ACCOUNT', '')` | — |

- 「新建」的 `account.created`：前端送出的版本（`getAggregateVersion` 取得，通常為 0）不使用；aggregate 已存在時回傳衝突。種子匯入的科目寫在 `('ACCOUNT', "1")`，不會讓個別科目的 aggregate 存在，重複的科目編號由讀模型的唯一鍵擋下（F6）。
- 「交易內」的期間開帳：前端送出的版本指向 `('ACCOUNT', '')`，不使用；同一期間只開立一次由 1.2 第 3 點的唯一鍵保護。

## 4. 分段

每段連同自己的測試完成，依 1 → 5 各自 review、commit；開始實作時在 `doc/STEP.md` 加上 R2 子階段。依 F7 的連帶要求，**每段登錄的事件類型，它的所有寫入端（API、種子、測試）在同一段改走 BFS**，否則會被 `EventStoreService.Append` 拒絕。

### 段 1：寫入基礎（不遷移任何事件類型）

修改：

- `domain/event/registry.go`：D2；`registry_test.go` 跟著調整。
- `kernel/event/envelop.go`：`RootVersion`（D3）。
- `runtime/engine/bfs_engine.go` 的 `write`：根事件依 `RootVersion` 選擇寫入方式。
- `persistence/repos/store.go`：`AppendCurrent`。
- `database/queries/event_store.sql`、`repos/query/event.go`：`GetByUUID`。
- `application/command`（新增）：D4 的 `NewEngine`、`Build`、`Append`、`DispatchLegacy`。
- `handler/handler.go`：改用 `command.NewEngine`；`handler/event.go`：`Append` 先經 `DispatchLegacy`，未登錄的事件類型才交給 `EventStoreService.Append`。

測試情境：

| # | 情境 | 預期 |
|---|---|---|
| 1 | `RootVersionNew`，aggregate 不存在 | 寫入版本 1 |
| 2 | `RootVersionNew`，aggregate 已在版本 0 或版本 2 | 衝突，事件與版本紀錄不變 |
| 3 | `RootVersionCurrent`，aggregate 不存在／已在版本 3（`ExpectedVersion` 送其他值） | 分別寫入版本 1／4 |
| 4 | `RootVersionExpected` | 既有的 engine 情境全數通過，行為不變 |
| 5 | `command.Build` 建立 cf 事件 | aggregate 為 `('TRANSACTION', txn_uuid)`，商戶與使用者取自 ctx，版本模式為 `RootVersionExpected` |
| 6 | `DispatchLegacy` 收到未登錄的事件類型 | 交回舊路徑，不寫入 |
| 7 | `DispatchLegacy` 收到 cf（沒有舊映射） | 拒絕，不寫入；通用 API 回 500（F7 的 handler 情境改為驗證新的拒絕訊息） |

反向檢查：`RootVersionNew` 改用 `Append` → 情境 2 的版本 0 失敗；`RootVersionCurrent` 改用 `ExpectedVersion` → 情境 3 失敗。

### 段 2：科目與明細帳戶

登錄 `account.created`、`account.updated`、`account.bulk_imported`、`ledger_account.created`、`ledger_account.updated`。

修改：`domain/event` 的 payload（D1）、`process` 的 mediator handler（payload 驗證；D9 的業務版本）、`persistence/handle` 的科目 projector（D10）、註冊表；`bootstrap/event_seed.go` 的 `bootstrapAccounts` 與 `bootstrap/db.go` 改經 D4。

寫入端：通用 API（前端 4 種）、種子（`account.bulk_imported`）。

既有測試的調整：

- `services/update_version_test.go`：科目與明細帳戶的情境改經通用 API 的 BFS 派送，情境與預期不變；投資情境（R9 才遷移）保留在 service 測試，fixture 中的科目改以種子或 SQL 建立。
- F7 的 `event_append_test.go`、`event_scenarios_test.go`：對照組改用仍走舊路徑的事件類型（實作時選定，例如 `investment.created`）；原本的 `account.created` 對照改為 BFS 派送的成功情境。

新增測試情境：

| # | 情境 | 預期 |
|---|---|---|
| 1 | 通用 API 送出 `account.created` | 200，回應格式與舊路徑相同；事件寫在 `('ACCOUNT', account_id)` 版本 1，封套欄位完整；讀模型新增科目，version 1 |
| 2 | 同上，aggregate 已存在 | 衝突，不寫入 |
| 3 | 通用 API 送出的 aggregate 不等於舊映射（例如 aggregate ID 與 `payload.account_id` 不同） | 拒絕（500），不寫入 |
| 4 | F6 的科目與明細帳戶情境（過期版本、版本相符、不存在的資料） | 結果與 F6 相同，衝突訊息相同 |
| 5 | 明細帳戶建立後修改 | 兩筆事件都在所屬科目的 aggregate，`ledger_uuid` 為建立事件的 UUID |
| 6 | 種子執行兩次 | 只有一筆 `account.bulk_imported`，寫在 `('ACCOUNT', "1")` |
| 7 | 新舊 payload 的 JSON（D12 的契約） | 5 種事件以同一份 fixture 分別序列化，結果相同；舊型別能解碼 BFS 寫入的 payload |

反向檢查：移除業務版本檢查（handler 與 projector 都移除）→ 情境 4 的過期版本失敗；舊映射檢查改為一律通過 → 情境 3 失敗。

### 段 3：系統科目與科目設定

登錄 `sys_config.sys_account_updated`、`account_config.ledger_type_updated`、`account_config.asset_type_updated`。

修改：payload、mediator handler（payload 驗證與科目存在，改用交易內查詢）、設定 projector、註冊表；`handler/setting.go` 的 `appendConfigEvent`（設定與系統科目 API 共用）改經 D4，仍先讀取目前版本（F5 的做法）；種子的 `bootstrapSysAccounts`、`bootstrapAccountConfigs` 改經 D4。

測試情境：F5 的設定 handler 測試（`handler/setting_test.go`）全數通過且不修改預期；通用 API 送出這 3 種事件時拒絕；種子執行兩次不重複；新舊 payload 的 JSON 一致（D12 的契約）。

反向檢查：handler 移除科目存在檢查 → 設定測試中「科目不存在」的情境失敗。

### 段 4：預付與固定資產類別（需先完成 F8）

登錄兩種類別各 3 種事件。

修改：payload、mediator handler（payload 驗證；修改與刪除以交易內查詢確認類別存在且啟用）、兩個類別 projector（含 F8 的版本遞增）、註冊表；兩個類別的專屬 handler 改經 D4：建立以 `RootVersionNew`、aggregate 為事件 UUID（D6），修改與刪除以 `RootVersionExpected` 帶畫面版本；migration 補上既有類別的版本紀錄（D6）。

既有測試的調整：`services/category_test.go` 的情境改經專屬 API 執行，情境與預期不變。

新增測試情境：

| # | 情境 | 預期 |
|---|---|---|
| 1 | 經專屬 API 建立類別 | 事件的 aggregate 等於 `category_uuid`，且 `category_uuid` 為建立事件的 UUID（與舊 projection 的推導相同，D12 的契約） |
| 2 | 連續修改兩次；修改後刪除 | 都成功，讀模型版本與 aggregate 版本一致 |
| 3 | 送出過期的畫面版本 | 衝突，資料不變 |
| 4 | 通用 API 送出類別事件 | 拒絕，不寫入 |
| 5 | 既有類別從未修改（沒有版本紀錄，讀模型 1）→ 段 4 migration → 經 BFS 修改 | 成功，aggregate 與讀模型都是版本 2 |
| 6 | 既有類別在舊版修改過一次（讀模型 1、aggregate 2）→ F8 migration（讀模型 2）→ 段 4 migration（不變）→ 經 BFS 修改 | 成功，aggregate 與讀模型都是版本 3 |
| 7 | 新舊 payload 的 JSON | 一致（D12 的契約） |

情境 5、6 驗證 migration 對既有資料的效果：以 goose 先遷移到該 migration 的前一版、寫入 fixture，再遷移到最新版本（沿用 F9 建立的 `database.MigrateUpTo`，F8 也以同樣方式測試）。

反向檢查：建立事件改回隨機 aggregate → aggregate 等於 `category_uuid` 的情境與連續修改失敗。

### 段 5：期間開帳（建議先完成 F9）

登錄 `period.month_started`、`period.annual_started`。

修改：migration 新增 `PERIOD` aggregate 類型（D8）、`enums.AggregateType`；payload、mediator handler（payload 驗證）、期間 projector（只處理兩種開帳事件）、註冊表（舊映射 `('ACCOUNT', '')`，版本策略「交易內」）。

測試情境：

| # | 情境 | 預期 |
|---|---|---|
| 1 | 通用 API 送出 `period.month_started`，前端版本為任意值 | 200；事件寫在 `('PERIOD', "1")`，版本為交易內的目前版本 + 1；新增 OPEN 的月期間 |
| 2 | 送出的 aggregate 不是 `('ACCOUNT', '')` | 拒絕，不寫入 |
| 3 | 同一商戶重複開立同一月份 | 唯一鍵錯誤，事件與版本紀錄都回復 |
| 4 | BFS 開立的月份以舊路徑結帳（`period.month_closed`） | 結帳成功（R5 前的跨路徑情境） |
| 5 | `period.annual_started` | 同情境 1，期間為年度 |
| 6 | 新舊 payload 的 JSON；`closing_uuid` 的推導 | JSON 一致；`closing_uuid` 為開帳事件的 UUID（D12 的契約） |

反向檢查：版本策略改為「驗證」→ 情境 1 在前端送出 `('ACCOUNT', '')` 的版本時失敗。

## 5. 預期的行為變更

1. 16 種事件改由 BFS 寫入：`event_store` 的 `metadata` 有執行資訊；種子事件的 `updated_by` 由 NULL 變為空字串（BFS `Store` 一律寫入，`persistence/repos/store.go:124`）。
2. aggregate（A7）：期間開帳改寫在 `('PERIOD', "1")`；類別建立改寫在 `('PREPAID_CATEGORY'／'FIXED_ASSET_CATEGORY', 事件 UUID)`。
3. 版本策略（ADR-027 規則 4）：`account.created` 已存在時衝突、不使用前端版本；期間開帳不使用前端版本。
4. 通用 API 不再接受類別、設定與 `account.bulk_imported`（前端沒有這些呼叫），也拒絕不符合舊映射的 aggregate；兩者都回 500。
5. projection checkpoint：BFS projector 以自己的名稱記錄 checkpoint（R3 會使用）。
6. 錯誤訊息的文字可能因 BFS 的錯誤包裝而不同；狀態碼與 problem 格式不變，業務版本衝突的訊息與 F6 相同（D9）。

維持不變：payload 的 JSON；主檔、設定、類別與期間 API 的請求與成功回應；未遷移事件的寫入；期間結帳與重開（R5）；前端（ADR-013）。

## 6. 驗證

每段依 `doc/NOTES.md` 第 3 點的 Docker 流程（`golang:1.26.9`、900 MB 記憶體上限）：`go generate ./...`、`go build ./...`、修改套件的 `go vet`、`go test -vet=off ./...`、修改與新增檔案的 `gofmt -l`、`git diff --check`；反向檢查在工作區副本中修改，不留在 repo。

## 7. 文件

- 本次（計畫書）：ISSUE-029、ISSUE-030 登錄於 `doc/ISSUES.md`；上層計畫 5.2、`doc/STEP.md`、`doc/TODO.md` 加上 F8、F9；`doc/TODO.md` 另記「盤點其他資料表的唯一鍵是否缺少 `merchant_id`」。
- 實作時：ADR-027 規則 1（登錄的事件類型、`Spec.Aggregate` 的輸入、依賴清單補上 decimal）與規則 4（`RootVersion`）；ADR-024 A7（`PERIOD` 與 aggregate ID `"1"`）；`doc/STEP.md` 的 R2 子階段；上層計畫 R2 的狀態。

## 8. 風險與範圍外

- **整數 `ledger_id`（ISSUE-031）**：`ledger_account.updated` 與傳票、投資、銀行對帳的 payload 以整數 `ledger_id` 引用明細帳戶。單商戶重建時，其他商戶仍保留的明細帳戶會讓重新新增的 `ledger_id` 改變，之後的事件引用舊 ID 而失敗；review 已以工作區副本確認舊 `Replay` 會因此失敗（ISSUE-018 的同類問題，當時只補了 UUID 欄位，payload 仍是整數）。R2 不改 payload，也不承諾舊 `Replay`（D12）；R3 的計畫必須處理（例如重建時保留整數 ID，或事件改以 `ledger_uuid` 引用），並列入 R3 的驗收。
- **`('ACCOUNT', "1")`**：使用者若建立編號為 `1` 的科目，`account.created` 的 aggregate 會撞上種子的 aggregate 而回傳衝突。目前種子科目編號至少 3 碼，前端也沒有這樣的科目；記錄為已知限制。
- **跨路徑的期間**：R2～R5 之間，開帳寫在 `PERIOD`，結帳與重開仍寫在 `('ACCOUNT', '')`，兩者沒有共用的樂觀鎖；同一期間的開帳與結帳有前後依賴，由讀模型的狀態檢查與唯一鍵保護，R5 遷移後回到同一個 aggregate。
- **範圍外**：舊 `Replay` 與事件匯入的重建結果（D12；程式不修改，也不在 R2 驗收）；期間結帳與重開（R5）；銀行範本（R6，CSV 範本的讀模型版本是否遞增需在 R6 確認）；重建機制（R3）；錯誤狀態碼的整體調整；`ledger_id` 改為 UUID 引用。
