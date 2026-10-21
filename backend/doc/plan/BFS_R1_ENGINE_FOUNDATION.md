# R1 實作計畫書：BFS engine 基礎

| 項目 | 內容 |
|---|---|
| 狀態 | ⏳ 待 review |
| 建立日期 | 2026-10-21 |
| 最後修訂 | 2026-10-21：依 review 修訂。衍生 aggregate 改以 `Store.Create` 建立（交易內確認不存在、只新增），補上版本 0 的情況；D1～D5 確定，D2 改為另開 bug fix F7（ISSUE-028）；`ExecuteBatch` 補轉傳測試；註冊表限制依賴；engine 情境補上「衍生 aggregate 的孫事件」與「terminal 完成後判定超量」 |
| 分支 | `refactor`（基準 commit `f6089a9`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md` 5.3 的 R1 |
| 關聯 | ADR-023（原則 2、6、7，A1、A2）、ADR-024（身分映射規則）、ADR-027（規則第 1 點）；R0（`query.NewQueryRepositoryTx`）；ISSUE-028（F7） |

> R1 是重構。依 `AGENTS.md`「變更幅度與重構」，本文件列出範圍、維持不變的行為、分段步驟、驗證方式，以及每一項預期的行為變更；未列出的行為一律維持不變。
> 文中行號以基準 commit 為準。

---

## 1. 共通事項

### 1.1 目標與範圍

上層計畫書的 R1 有五項，本計畫依序對應：

| # | 上層計畫的內容 | 本計畫 |
|---|---|---|
| 1 | 事件圖交易邊界（A1） | 段 2 |
| 2 | 子事件封套與版本規則（A2） | 段 2 |
| 3 | handler 改用 R0 的交易內查詢 | 段 2 |
| 4 | 移除 projector 的 `state` 參數（原則 2） | 段 1 |
| 5 | 建立事件註冊表（ADR-027 規則第 1 點），先登錄 `transaction.cf_category_updated` | 段 3 |

上層計畫的驗收條件：engine 測試涵蓋子事件失敗回復、同 aggregate 連續版本、跨 aggregate 子事件；從 `event_store` 讀出的 JSON 經註冊表解碼後，projector 可以套用。

**不在範圍**：

- 不遷移任何新的事件類型；正式環境的 BFS 事件仍只有 `transaction.cf_category_updated`。
- 不建立 `POST /api/event/append` 的派送（R2），也不建立重建機制（R3）；註冊表在 R1 只由測試使用。
- 不修改舊路徑（`EventStoreService`、pipeline、舊 projection）與前端。
- 舊入口仍可寫入 cf 事件的問題（ISSUE-028）由 bug fix F7 另行處理（D2）。

### 1.2 分段

修改超過 5 個檔案，依 `backend/CLAUDE.md` 分段。R1 的三個部分彼此獨立，所以按功能分段，每段連同自己的測試一起完成、一起驗證，依 1 → 2 → 3 的順序各自 review、各自 commit：

| 段 | 內容 | 行為變更 |
|---|---|---|
| 1 | 移除 projector 的 `state` 參數 | 無 |
| 2 | 事件圖交易邊界、Run-local 交易內查詢、子事件封套與版本（A1、A2） | 有，見 4.5 |
| 3 | 事件註冊表 | 無 |

開始實作時，在 `doc/STEP.md` 的 R1 下方加上三段的子進度；每段完成後執行驗證，再更新狀態。

### 1.3 已確認的決定（2026-10-21 review）

| # | 問題 | 決定 |
|---|---|---|
| D1 | A2 第 4 點「第一次碰到的其他既有 aggregate」：要在交易內讀取目前版本，命令帶有版本前提時還要檢查。目前沒有任何 handler 需要，R1 的驗收條件也沒有列出 | **延到第一個需要的 Phase 再實作**，並在 ADR-023 註明。R1 只支援「沿用父事件的 aggregate」與「衍生新的 aggregate」兩種子事件，其他寫法一律回傳錯誤（4.1 第 3 點）。前提是衍生 aggregate 的新建保護完整（4.1 第 5 點） |
| D2 | `transaction.cf_category_updated` 仍登錄在舊 pipeline（`services/pipelines/factory/base.go:57`），舊 projection 也會處理它（`services/projection/entry_cf_category.go:147`），`POST /api/event/append`（`handler/event.go:100`）仍能以舊路徑寫入。通用 API 接受呼叫端指定的 aggregate，同一筆分類可以寫到另一個 aggregate，避開以交易 UUID 為 aggregate 的版本檢查 | **另開小型 bug fix F7**（ISSUE-028），先在舊入口拒絕寫入已遷移到 BFS 的事件類型，不併入 R1 的重構。F7 與 R1 互不依賴；F7 的計畫另行確認 |
| D3 | `Executor.ExecuteBatch` 以 goroutine 並行 dispatch（`runtime/engine/executor.go:36`），只有測試使用 | **保留，只多一個 `repo` 參數**，並維持在 `Run` 的執行路徑之外：註解寫明 `Run` 不使用它、並行的 handler 不得共用 `Run` 的交易內 Repo。補上非 nil Repo 的轉傳測試（4.2），避免只改簽章卻漏傳。是否移除記入 `TODO.md` |
| D4 | 分段方式 | **1.2 的三段**，各段包含自己的測試 |
| D5 | 註冊表的位置 | **`internal/domain/event/registry.go`**。只負責宣告、解碼與身分規則，不依賴 engine 與 persistence（5.1） |

### 1.4 驗證流程

沿用 F5、F6、R0 的流程，以 Docker（`golang:1.26`）執行，環境限制見 `doc/NOTES.md`：

1. `go generate ./...`、`go build ./...`。
2. `go test -vet=off ./...`；另對修改的套件執行 `go vet`。
3. 修改與新增的 Go 檔案執行 `gofmt -l`，不得有輸出；修改到既有未通過 gofmt 的檔案時，對該檔執行 gofmt（NOTES 第 7 點）。`git diff --check` 不得有輸出。
4. 段 2 另以 `-race` 執行 `./internal/runtime/...`（沿用 F4）。
5. 反向檢查：在容器內的副本把修改還原（方式見各段），確認新測試會失敗。
6. 回報時附上每個步驟的實際輸出。

---

## 2. 現況

| # | 現況 | 位置 |
|---|---|---|
| 1 | **每個事件各自一個交易**：terminal 先 dispatch handler，再各自呼叫 `uow.Do` 寫入事件與套用 projector。根事件 commit 後才處理子事件 | `runtime/engine/bfs_engine.go:27`、`:34` |
| 2 | **handler 的查詢不在交易內**：`Mediator.WithQuery(db)` 建立綁定全域 DB 的 `query.Repo`，所有 `Run` 共用，`Dispatch` 一律傳入它 | `runtime/mediator/mediator.go:26`、`:29`、`:49`；`handler/handler.go:25` |
| 3 | **projector 收到 handler 的即時 state**：`ApplyAll(ctx, tx, ev, r.State)`；`result.EventResult.State` 沒有任何 handler 設定，`cashFlowCategoryProjector` 也沒有使用 | `bfs_engine.go:41`；`kernel/result/result.go:15`；`persistence/handle/projector.go:13`、`:21`、`:38` |
| 4 | `handle/utils.go` 的 `checkAndGetState` 沒有呼叫端，它也是 BFS 的 persistence 套件 import 舊 `services/pipelines` 的唯一原因 | `persistence/handle/utils.go:24` |
| 5 | **子事件封套不完整**：`NewNextEvent` 只帶 aggregate ID、`CorrelationID`、`CausationID`，沒有 `AggregateType`、`MerchantID`、`UpdatedBy`、`RequestID`；`ExpectedVersion` 為 0 | `kernel/event/envelop.go:133` |
| 6 | **`ExpectedVersion` 為 0 不代表 aggregate 不存在**：`Store.Append` 先執行 `UPDATE ... WHERE current_version = ?`，`aggregate_versions.current_version` 的預設值是 0、沒有下限限制。已有版本 0 的紀錄時，以 0 送出會被更新成版本 1，當成既有 aggregate 的下一個版本；版本大於 0 時才會改走新增並撞上主鍵 | `persistence/repos/store.go:33`、`:42`；`migrations/20260614001_bank_pdf_multiledger.sql` 的 `aggregate_versions_new` |
| 7 | 沒有建立「其他 aggregate」子事件的方式 | — |
| 8 | 沒有事件註冊表；projector 以型別斷言取得 payload，`event_store` 讀出的 JSON 無法直接套用 | `persistence/handle/utils.go:16` |
| 9 | engine 的單元測試以 mock UnitOfWork 執行，沒有接 DB；`emitChildren`、`makeTestEvent` 建立的子事件沒有呼叫 `NewNextEvent` | `runtime/engine/bfs_engine_test.go:28`、`:48`；`bfs_engine_mocks_test.go` |
| 10 | 唯一的 mediator handler `transaction.cf_category_updated` 只查詢、不產生子事件；HTTP 回應是 `Run` 回傳的根事件 | `process/cashflow.go:19`；`handler/entry_cf_category.go:93`、`:137` |

---

## 3. 段 1：移除 projector 的 `state` 參數

### 3.1 修改內容

| 檔案 | 修改 |
|---|---|
| `internal/persistence/handle/projector.go` | `handle` 型別、`Projector.Apply`、`Registry.ApplyAll` 移除 `state any` 參數 |
| `internal/persistence/handle/cash_flow_cateory.go` | `Apply`、`applyUserUpdate` 配合移除 `state` |
| `internal/persistence/handle/utils.go` | 刪除沒有呼叫端的 `checkAndGetState`，一併移除對 `services/pipelines` 的 import |
| `internal/kernel/result/result.go` | 刪除 `EventResult.State`：唯一的讀取者是 projector，沒有任何 handler 設定它 |
| `internal/runtime/engine/bfs_engine.go` | `ApplyAll(ctx, tx, ev)` |
| `internal/runtime/engine/executor.go` | `Dispatch` 註解中的「including State」 |
| `internal/runtime/engine/bfs_engine_mocks_test.go` | `mockProjector.Apply` 配合簽章 |

`NewCashFlowCategoryProjector(db *sqlx.DB)` 的 `db` 參數沒有使用，本段不修改（無關清理，見第 7 節第 3 點）。

### 3.2 行為與驗證

- 預期的行為變更：無。
- 驗證：1.4 的流程；既有測試全數通過即可，不新增測試、不做反向檢查。

---

## 4. 段 2：事件圖交易邊界與子事件規則（A1、A2）

### 4.1 設計

**1. 交易邊界（A1 第 1 點）**

- `Run` 只呼叫一次 `uow.Do`，整個 BFS 迴圈在同一個交易內執行：每個事件的 middleware、handler、append、projector、checkpoint 與 snapshot。任一步驟回傳錯誤，`Do` 回復整個交易，`Run` 回傳該錯誤。middleware 在 terminal 完成之後才判定的錯誤（例如 `BackpressureScheduler` 的超量）也一樣回復。
- terminal 不再是 engine 的欄位，改由 `Run` 建立：terminal 以閉包取得這次的交易與版本紀錄，與 F4 的 middleware chain 一樣屬於單次 `Run`。
- `Run` 在處理每個事件前檢查 `ctx.Err()`，已取消時回傳該錯誤（A1 第 4 點）。`BeginTxx` 綁定同一個 `ctx`，取消時 `database/sql` 也會回復交易。
- 沒有設定 UnitOfWork 時（engine 的單元測試）維持現狀：不開交易、不寫入，handler 收到 nil 的查詢 Repo。

**2. Run-local 交易內查詢（A1 第 2 點）**

- `persistence/repos.DbTransaction` 新增 `Query *query.Repo`，由 `sqlxUnitOfWork.Do` 以 `query.NewQueryRepositoryTx(tx)` 建立。每次 `Do` 各自一份，屬於這次 `Run`。
- `Mediator.Dispatch(ctx, repo, evt)` 改由呼叫端傳入 Repo；刪除 `Mediator.repo` 欄位與 `WithQuery`，不再有共用的查詢 Repo。`mediator.Handler` 的簽章不變。
- `Executor.Dispatch(ctx, repo, evt)`、`ExecuteBatch(ctx, repo, batch)` 多一個 `repo` 參數；`ExecuteBatch` 依 D3 維持在 `Run` 的執行路徑之外。
- 綁定交易的 Repo 只供讀取，寫入方法回傳 `query.ErrTxBoundWrite`（R0）；handler 的寫入一律以事件表達。

**3. 子事件的 aggregate（A2、ADR-024）**

`kernel/event.Event` 新增 `ChildAggregate` 欄位（`json:"-"`，不寫入 `event_store`），表示子事件屬於哪個 aggregate：

| 值 | 建立方式 | engine 的處理 |
|---|---|---|
| `ChildAggregateParent`（零值） | `NewNextEvent(parent, payload)`，或測試直接建立的 `event.Event` | 沿用父事件的 `AggregateType` 與 aggregate ID。子事件若已填入不同的 aggregate，回傳錯誤，不會悄悄改成父事件的 aggregate |
| `ChildAggregateDerived` | 新增 `NewDerivedEvent(parent, aggType, purpose, payload)` | aggregate ID 為 `uuidx.NewFromEvent(parent.Uuid, purpose)`，依父事件 UUID 衍生，重建時可以重現（ADR-024 規則 2）；`aggType` 不得為零值。寫入時以 `Store.Create` 建立（第 5 點） |

- 零值選擇「沿用父事件」：與目前 `NewNextEvent` 的行為相同，既有測試的子事件不必修改。
- 錯誤使用新增的 `kernel/errors.ErrChildEnvelopeInvalid`（`RUNTIME_CHILD_ENVELOPE_INVALID`）。
- 「第一次碰到的其他既有 aggregate」依 D1 延後；在那之前，handler 無法表達這種子事件，寫成其他 aggregate 會被上表擋下。

**4. 封套欄位繼承（A2 欄位）**

engine 在子事件加入佇列前，以父事件的值覆寫：`MerchantID`、`UpdatedBy`、`Metadata.Tracing.CorrelationID`、`Metadata.Tracing.RequestID`；`Metadata.Tracing.CausationID` 設為父事件的 UUID。覆寫由 engine 統一處理（原則 7），handler 填錯也不會生效。`Execution`、`Schema`、`Observability` 維持建構函式設定的值。

**5. 版本配置與新建（A2 版本 1～3、5、6）**

`Run` 以 `(MerchantID, AggregateType, AggregateID)` 為 key，記錄這次 `Run` 每個 aggregate 最新的版本（`Store.Append`／`Store.Create` 的回傳值）。terminal 在寫入前才決定寫入方式，因此兄弟子事件寫入同一個 aggregate 時，後者會拿到前者寫入後的版本：

| 事件 | 寫入方式 |
|---|---|
| 根事件 | `Store.Append`，`ExpectedVersion` 為呼叫端提供的值，不得以最新版本取代（A2 第 1 點）；行為與目前相同 |
| 沿用父事件 aggregate 的子事件 | `Store.Append`，`ExpectedVersion` 為這次 `Run` 記錄的最新版本（A2 第 2 點） |
| 衍生 aggregate 的子事件 | `Store.Create`（A2 第 3 點） |

**`Store.Create`**（新增）：只新增、不更新的建立語意。在同一個交易內以既有的 `GetAggregateVersion(aggregate_type, aggregate_id, merchant_id)`（`queries/event_store.sql:94`）確認版本紀錄不存在，再新增版本 1 的紀錄並寫入事件。版本紀錄已存在時，不論版本是 0 或大於 0，都回傳 `ErrVersionConflict`、不寫入任何資料。同一次 `Run` 衍生出相同的 aggregate 時，第二次也會因此衝突。

- 以事件的 `MerchantID` 查詢，不依賴 ctx 的商戶。
- 檢查與新增在同一個交易內；SQLite 同時只有一個寫入者，檢查到新增之間不會被其他交易插入。
- R2 的「新建」版本策略（ADR-027 規則第 4 點）可以沿用 `Store.Create`。

handler 設定的 `ExpectedVersion` 對子事件不生效（A2 第 6 點）。

### 4.2 修改內容

**程式（9 個檔案）**

| 檔案 | 修改 |
|---|---|
| `internal/runtime/engine/bfs_engine.go` | `Run` 依 4.1 第 1、3～5 點改寫；terminal 改為由 `Run` 建立 |
| `internal/runtime/engine/executor.go` | `Dispatch`、`ExecuteBatch` 多一個 `repo` 參數；`ExecuteBatch` 的註解寫明 D3 的限制 |
| `internal/runtime/mediator/mediator.go` | `Dispatch(ctx, repo, evt)`；刪除 `repo` 欄位與 `WithQuery` |
| `internal/persistence/repos/db_transaction.go` | `DbTransaction` 新增 `Query *query.Repo` |
| `internal/persistence/repos/unit_of_work.go` | `Do` 建立 `Query: query.NewQueryRepositoryTx(tx)` |
| `internal/persistence/repos/store.go` | `Store` 介面新增 `Create`；與 `Append` 共用事件寫入的部分 |
| `internal/kernel/event/envelop.go` | `ChildAggregate` 型別與欄位、`NewDerivedEvent` |
| `internal/kernel/errors/code.go` | 新增 `ErrChildEnvelopeInvalid` |
| `internal/handler/handler.go` | `mediator.NewMediator()` 不再呼叫 `WithQuery` |

**測試**

- 配合簽章調整：
  - `runtime/engine/bfs_engine_mocks_test.go`：`mockStore` 實作 `Create`。
  - `runtime/engine/executor_test.go`：`ExecuteBatch` 多傳一個 Repo。
  - `handler/entry_cf_category_test.go`：`newCFTestEngine` 不再呼叫 `WithQuery`。
- 既有的 engine 測試（`bfs_engine_test.go`）預期不需修改：子事件沒有填 aggregate，依 4.1 第 3 點沿用父事件的 aggregate；mock store 一律回傳版本 0。實作時若有測試需要修改，在 review 時逐一說明原因。
- `runtime/engine/executor_scenarios_test.go` 新增轉傳情境（D3）：`Dispatch` 與 `ExecuteBatch` 把傳入的非 nil Repo 原樣交給每個 handler（比對指標）。
- 新增 `internal/runtime/engine/bfs_engine_tx_scenarios_test.go`、`bfs_engine_tx_test.go`：沿用既有的 Ginkgo suite（`bfs_engine_suite_test.go`），以 `test.NewTestDBGinkgo()`（單一連線）與 `repos.NewUnitOfWork(db)` 執行，情境見 4.3。
- 新增 `internal/kernel/event/envelop_test.go`（標準 table-driven）：`NewDerivedEvent` 的 aggregate ID 只由父事件 UUID 與 `purpose` 決定（相同輸入得到相同 ID、不同 `purpose` 得到不同 ID），`ChildAggregate` 為 `ChildAggregateDerived`。

### 4.3 測試情境（engine，接 DB）

**測試資料**：沿用 suite 的 `typeA`（`transaction.created`）、`typeB`（`transaction.corrected`）與 `testPayload`，aggregate 類型為 `TRANSACTION`，aggregate ID 為測試內固定的字串，商戶 1、使用者名稱固定。ctx 帶商戶 1（checkpoint 依 ctx 的商戶寫入）。mediator handler 與 projector 由各情境以函式設定。這些事件只在測試 DB 內寫入 `event_store`、`aggregate_versions`、`projection_checkpoints`，不涉及帳務資料。

「全部回復」指 `event_store`、`aggregate_versions`、`projection_checkpoints` 在 `Run` 之後與 `Run` 之前相同（情境 7 預先寫入的版本紀錄維持原值）。

| # | 情境 | 預期 |
|---|---|---|
| 1 | 單一根事件，handler 與 projector 成功 | 成功；1 筆事件（版本 1）、1 筆 checkpoint |
| 2 | 根事件產生一個同 aggregate 的子事件 | 成功；同一個 aggregate 有版本 1、2 的事件，`aggregate_versions` 為 2 |
| 3 | 根事件產生兩個同 aggregate 的子事件（兄弟） | 成功；版本 1、2、3 |
| 4 | 根事件 → 子事件 → 孫事件，都在同一個 aggregate | 成功；版本 1、2、3 |
| 5 | 根事件以 `NewDerivedEvent` 產生衍生 aggregate 的子事件 | 成功；子事件的 aggregate 為指定的類型與 `uuidx.NewFromEvent(根事件 UUID, purpose)`、版本 1；根事件的 aggregate 版本 1 |
| 6 | 同 5，衍生的子事件再產生同 aggregate 的孫事件 | 成功；衍生 aggregate 有版本 1、2 的事件，`aggregate_versions` 為 2；根事件的 aggregate 版本 1 |
| 7a | 同 5，但衍生的 aggregate 已有版本 0 的紀錄（測試先寫入） | `ErrVersionConflict`；全部回復，版本紀錄仍為 0 |
| 7b | 同 5，但衍生的 aggregate 已有版本 3 的紀錄 | `ErrVersionConflict`；全部回復，版本紀錄仍為 3 |
| 8 | 同一次 `Run` 兩個子事件衍生出相同的 aggregate | 第二個回傳 `ErrVersionConflict`；全部回復 |
| 9 | 子事件的 handler 回傳錯誤 | `Run` 回傳該錯誤；全部回復（根事件與它的 checkpoint 也不存在） |
| 10 | 子事件的 projector 回傳錯誤 | `Run` 失敗；全部回復 |
| 11 | 子事件被 middleware 擋下（`DepthGuard(0)`） | `ErrBFSDepthExceeded`；全部回復 |
| 12 | `BackpressureScheduler(1)`，根事件產生兩個子事件：根事件已寫入與套用 projector 後，middleware 才判定超量 | `ErrQueueOverflow`；全部回復 |
| 13 | 根事件的 handler 取消 ctx 後回傳子事件 | `Run` 回傳錯誤，子事件的 handler 沒有執行；全部回復。ctx 取消後，根事件的 append 或 checkpoint 可能先因 ctx 失敗，所以不限定錯誤的種類；`Run` 迴圈本身的檢查由 13a 確認 |
| 13a | 同 13，但不設定 UnitOfWork（不接 DB） | `Run` 回傳 `context.Canceled`，子事件的 handler 沒有執行 |
| 14 | 子事件的 handler 以收到的 Repo 查詢根事件（`Event.GetByAggregate`）；ctx 設 1 秒逾時 | 查得到尚未 commit 的根事件，`Run` 成功 |
| 15 | 根事件的 handler 產生子事件時，在子事件填入其他 `MerchantID`、`UpdatedBy`、`CorrelationID` | `event_store` 中子事件的 `merchant_id`、`updated_by` 與父事件相同；metadata 的 `correlation_id`、`request_id` 與父事件相同，`causation_id` 為父事件 UUID |
| 16 | 子事件直接填入與父事件不同的 aggregate，沒有使用 `NewDerivedEvent` | `ErrChildEnvelopeInvalid`；全部回復 |
| 17 | 根事件的 `ExpectedVersion` 過期（aggregate 已在版本 1，送出 0） | `Run` 失敗；`event_store` 不變（既有行為） |

`handler/entry_cf_category_test.go` 既有的 F1、F5 情境（含經 `NewMux` 的情境）必須全數通過，用來確認 4.5 列出「維持不變」的行為。

### 4.4 反向檢查

在容器內的副本分別還原下列修改，確認對應情境失敗：

| 還原方式 | 預期失敗 |
|---|---|
| `Run` 改回每個事件各自呼叫 `uow.Do` | 9～12：根事件已 commit，不會全部回復 |
| 移除 `Run` 迴圈的 `ctx.Err()` 檢查 | 13a：子事件的 handler 仍會執行 |
| terminal 改傳全域 DB 的 Repo（例如 `query.NewQueryRepository(db)`） | 14：等不到唯一的連線而逾時 |
| 移除版本紀錄，子事件一律以 0 寫入 | 2～4、6：子事件撞上主鍵 |
| 衍生 aggregate 改以 `Store.Append`（`ExpectedVersion` 0）寫入 | 7a：版本 0 的紀錄被更新成 1，`Run` 成功 |
| 移除封套欄位的覆寫 | 15 |

### 4.5 預期的行為變更

1. **事件圖整棵在一個交易內提交**（A1）：任一事件的 middleware、handler、append、projector 失敗，整次 `Run` 回復；原本是每個事件各自提交。正式環境唯一的 BFS 事件不產生子事件，結果不變。
2. **mediator handler 的查詢改在 `Run` 的交易內執行**：可以讀到同一次 `Run` 尚未提交的寫入；查詢 Repo 拒絕寫入。
3. **交易提前到 handler 之前開始**：唯一的連線被佔用的時間包含 handler 的查詢。A1 第 4 點已列出這個代價；深度與佇列上限不變。
4. **ctx 已取消時，`Run` 在處理下一個事件前中止**並回復。
5. **子事件的封套欄位與版本由 engine 配置**（A2）；aggregate 與父事件不同的子事件必須以 `NewDerivedEvent` 建立，否則失敗；衍生的 aggregate 只能新建，已存在（含版本 0）時失敗。

**維持不變**：

- `PUT /api/txn/{txn_uuid}/cf-category` 的請求、回應內容與狀態碼；寫入的事件（aggregate、版本、`merchant_id`、`updated_by`、payload）與 `entry_cf_categories` 的結果。
- 根事件的寫入方式（`Store.Append`）與目前相同。
- safety middleware 的語義與狀態隔離（F4）；每 50 個版本建立 snapshot；每個 projector 每個事件更新 checkpoint。
- 舊路徑完全不變。

---

## 5. 段 3：事件註冊表

### 5.1 設計

`internal/domain/event/registry.go`（D5）登錄每種遷移到 BFS 的事件類型（ADR-027 規則第 1 點）：

| 欄位 | 內容 |
|---|---|
| `EventType` | 事件類型，取自 payload 的 `EventType()`，不另外手寫 |
| `Decode` | 把 `event_store` 的 payload JSON 解碼成 payload。回傳**值型別**，與 mediator handler、projector 的型別斷言相同（`process/base.go:10`、`persistence/handle/utils.go:17`） |
| `Aggregate` | 由 payload 推導事件所屬的 aggregate（A5、A7） |
| `Version` | 版本策略：`新建`、`驗證`、`交易內`（ADR-027 規則第 4 點） |
| `Legacy` | 經 `POST /api/event/append` 時接受的舊 aggregate 映射（規則第 3 點）；nil 表示不接受經該 API 寫入 |

- 對外提供 `Lookup(eventType) (Spec, bool)` 與 `Decode(eventType, raw) (any, error)`。事件類型未登錄時回傳 `kernel/errors.ErrUnknownEventType`，JSON 無法解碼時回傳 `ErrDecodePayload`；兩者都已定義在 contract layer（`kernel/errors/code.go:7`、`:9`），但目前沒有使用。
- 登錄以泛型輔助函式完成，`Decode` 由 payload 型別產生，不逐一手寫。
- 同一種事件類型重複登錄時，由測試檢查。
- **依賴限制**（D5）：註冊表只負責宣告、解碼與身分規則，只 import `enums`、`enums/event_types`、`kernel/event`、`kernel/errors` 與標準函式庫，不依賴 engine、mediator、persistence 與 `repos`。R2 的派送與 R3 的重建依賴註冊表，反過來不行。

**`transaction.cf_category_updated` 的登錄內容**：

| 欄位 | 值 | 依據 |
|---|---|---|
| payload | `domain/event.TransactionCFCategoryUpdatedPayload` | mediator handler 與 projector 使用的型別 |
| aggregate | `('TRANSACTION', payload.txn_uuid)` | A5；與 `handler/entry_cf_category.go:77`、`:88` 相同 |
| 版本策略 | 驗證 | 專屬 API 送出 `expected_version`（A2 第 1 點） |
| 舊映射 | 無 | 前端只經專屬 API 寫入；舊入口的處理見 D2 |

R1 不修改 `entry_cf_category.go`：handler 仍自行設定 aggregate，兩者的一致性由 5.3 第 2 點的測試確認；R2 建立派送時改由註冊表推導。

### 5.2 修改內容

| 檔案 | 修改 |
|---|---|
| `internal/domain/event/registry.go`（新增） | 5.1 的型別、`Lookup`、`Decode`，以及 cf 的登錄 |
| `internal/domain/event/registry_test.go`（新增） | 標準 table-driven，情境 1～5 |
| `internal/persistence/handle/decode_apply_test.go`（新增） | package `handle_test`，標準 testing，情境 6 |

### 5.3 測試情境

| # | 情境 | 預期 |
|---|---|---|
| 1 | `Lookup(transaction.cf_category_updated)` | 找得到；版本策略為「驗證」，沒有舊映射 |
| 2 | 以 cf 的 payload 呼叫 `Aggregate` | `('TRANSACTION', txn_uuid)` |
| 3 | 把 cf 的 payload 序列化後以 `Decode` 解碼 | 回傳值型別的 payload，內容與原本相同 |
| 4 | `Decode` 未登錄的事件類型、無法解碼的 JSON | 分別回傳 `ErrUnknownEventType`、`ErrDecodePayload` |
| 5 | 檢查整份註冊表 | 沒有重複的事件類型；每一筆的 `EventType` 等於解碼後 payload 的 `EventType()` |
| 6 | 經 BFS 寫入一筆 cf 事件（測試 DB 與 F1 測試相同：一筆交易、兩筆分錄，資料在測試內明確寫入）→ 刪除 `entry_cf_categories` → 以 `query.Event.GetByAggregate` 讀出事件，依欄位組回 `event.Event`、以 `Decode` 解碼 payload → 在 `persistence` 的 UnitOfWork 內呼叫 `handle.Registry.ApplyAll` | `entry_cf_categories` 回到刪除前的內容（`cf_category`、`is_confirmed`、`updated_by`） |

### 5.4 反向檢查

| 還原方式 | 預期失敗 |
|---|---|
| `Decode` 改為回傳指標 | 3、6（projector 的型別斷言失敗） |
| 移除 cf 的登錄 | 1、6 |

### 5.5 預期的行為變更

無。註冊表在 R1 只由測試使用；R2 的派送與 R3 的重建才會在正式流程中使用。

---

## 6. 文件

| 時機 | 文件 | 內容 |
|---|---|---|
| 開始實作 | `doc/STEP.md` | R1 改為進行中，加上三段的子進度 |
| 段 2 | `backend/CLAUDE.md` 的 BFS 遷移過渡規則 | 補充：整棵事件圖一個交易；mediator handler 只能使用 `Run` 傳入的交易內查詢；子事件以 `NewNextEvent`／`NewDerivedEvent` 建立，版本由 engine 配置，衍生的 aggregate 以 `Store.Create` 新建 |
| 段 2 | `doc/ARCHITECTURE.md` ADR-023 | A2 第 3 點註明以 `Store.Create` 實作（含版本 0 的情況）；第 4 點註明延到第一個需要的 Phase（D1） |
| 段 3 | `doc/ARCHITECTURE.md` ADR-027 | 規則第 1 點註明註冊表的位置與依賴限制 |
| 完成後 | `doc/TODO.md` | D3（`ExecuteBatch` 是否移除） |
| 完成後 | 上層計畫書 5.3 的 R1、`doc/STEP.md` | 狀態與驗證結果 |

D2 的問題已登錄為 ISSUE-028，並在上層計畫書 5.2、`doc/STEP.md`、`doc/TODO.md` 加上 F7；F7 的實作計畫另行確認。

---

## 7. 風險與範圍外

1. **A2 第 4 點延後**（D1）：第一個需要寫入其他既有 aggregate 的 Phase，要先補上這條規則與測試。
2. **根事件以 `ExpectedVersion` 0 送出時**：沿用 `Store.Append`，與目前相同。已有版本 0 的紀錄時會被當成既有 aggregate 更新成版本 1；版本大於 0 時撞上主鍵，回傳 `ErrStoreWriteFailed` 而不是 `ErrVersionConflict`。R1 不改變根事件的寫入方式；R2 的「新建」策略改用 `Store.Create`，「驗證」策略的錯誤碼是否統一，在 R2 的計畫決定。
3. **projector 必須只使用交易**：`NewCashFlowCategoryProjector(db)` 收到全域 DB 但沒有使用。之後的 projector 若改用全域 DB，在單一連線下會卡住；R1 不刪除這個參數（無關清理），review 新 projector 時檢查。
4. **連線佔用**：事件圖越大，佔用唯一連線的時間越長（A1 第 4 點）；深度與佇列上限維持 `NewMux` 目前的設定。
5. **兩種 cf payload 型別並存**：舊 projection 使用 `model/payload.TransactionCFCategoryUpdatedPayload`，註冊表使用 `domain/event` 的型別；R10 移除舊路徑時一併移除前者。

---

## 8. 各段的 commit 範圍

| 段 | 程式與測試 | 文件 |
|---|---|---|
| 1 | 3.1 的 7 個檔案 | `STEP.md` |
| 2 | 4.2 的 9 個程式檔、配合調整的 3 個測試檔、`executor_scenarios_test.go` 的轉傳情境、3 個新增的測試檔 | `STEP.md`、`backend/CLAUDE.md`、`ARCHITECTURE.md`（ADR-023） |
| 3 | 5.2 的 3 個檔案 | `STEP.md`、`ARCHITECTURE.md`（ADR-027）、`TODO.md`、上層計畫書 |

本計畫書 review 通過後，先單獨 commit；ISSUE-028 與 F7 的登錄一併 commit。
