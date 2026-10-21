# BFS Engine 遷移與事件重建計畫書

| 項目 | 內容 |
|---|---|
| 狀態 | ✅ 計畫書 review 通過（2026-10-19）；Phase 0 尚有架構決策 A1～A5 與範圍決策 S3 待決 |
| 建立日期 | 2026-10-19 |
| 最後修訂 | 2026-10-19：依「目前沒有需要保留的資料」的前提改版。舊 replay 不再修正（ISSUE-021、022 改為 Wontfix，原 F2a／F2b 取消）；不轉換歷史事件（S2 採 (c)，原 R10 歷史事件轉換取消）；主檔與設定提前遷移，取消新舊事件共用的統一 replay runner；重構類重新排序為 R0～R10。後續 review：期間開帳事件提前到 R2；R10 補上事件匯入的切換與格式檢查。review 通過後：新增 A5（交易的 aggregate 身分），S4 確定採 (a)；補上 A2／A5 的身分映射規則與 R4、R5 驗收條件，修正 2.4 的現況描述 |
| 分支 | `refactor`（基準 commit `82e8990`） |
| 關聯 | ISSUE-020～023；ADR-001、ADR-014、ADR-017、ADR-018、ADR-020、ADR-021 |

> 本文件是計畫書。Phase 0 的決策確認後，依 `backend/CLAUDE.md` 的需求處理流程寫成 ADR，並在 `doc/STEP.md` 建立進度表。
> 文中行號以基準 commit 加上 2026-10-19 的工作區狀態為準。

> **前提（2026-10-19 確認）**：目前系統沒有需要保留的資料。因此本計畫不轉換歷史事件、不盤點或 backfill 既有資料，舊 replay 也不再修正；遷移完成後經使用者確認，清空開發 DB 重新開始。

---

## 1. 目標

1. **重大交易一律走 BFS engine**：凡會產生分錄或影響餘額的寫入，都經 `internal/runtime/engine.BFSEngine` 處理。
2. **可由事件重建 DB**：所有由事件衍生的讀模型，都能從 `event_store` 全量重建，且重建結果與重建前一致（比較規格見第 8 節）。

---

## 2. 現況

### 2.1 寫入路徑

| 路徑 | 流程 | 涵蓋範圍 |
|---|---|---|
| 舊路徑 | `POST /api/event/append` 等 → `EventStoreService.Append` → Pipeline → UnitOfWork → Projection | 53 種事件類型（`services/pipelines/factory/base.go`），涵蓋所有會計交易 |
| BFS | `BFSEngine.Run` → mediator handler → `persistence` UnitOfWork → `handle.Registry` | 1 種：`transaction.cf_category_updated`（`PUT /api/txn/{txn_uuid}/cf-category`） |
| 直接寫入 | 不產生事件 | 對帳比對狀態（ADR-021）、交易範本、商戶、使用者 |

### 2.2 BFS 路徑的缺口

**現有功能的 bug（ISSUE-020）**：目前唯一接上 BFS 的功能無法運作。

1. **payload 型別不一致**：HTTP handler 建立 `domain/event.TransactionCFCategoryUpdatedPayload`（`handler/entry_cf_category.go:69`），mediator handler 卻斷言 `model/payload.TransactionCFCategoryUpdatedPayload`（`process/cashflow.go:22`），每次請求都回 400。
2. **projector 未註冊**：`handler/handler.go:33` 建立空的 `handle.Registry`，`NewCashFlowCategoryProjector` 沒有呼叫點；該 projector 也斷言舊型別（`persistence/handle/cash_flow_cateory.go:45`）。
3. **根事件封套欄位缺漏**：`AggregateType`、`MerchantID`、`UpdatedBy`、`ExpectedVersion` 皆未填；aggregate ID 每次隨機產生（`entry_cf_category.go:70`），樂觀鎖失效，事件以 `merchant_id = 0` 寫入。

**engine 設計尚未支援的部分**（目前不影響現有功能，由重構類處理）：

4. **projector 依賴 handler 的即時 state**：`runtime/engine/bfs_engine.go:40` 把未持久化的 `r.State` 傳給 projector，replay 時無法取得（目前的 CF projector 沒有使用 state）。
5. **每個事件各自一個交易**：`BFSEngine.Run` 每取出一個 queue 事件就處理一次（`bfs_engine.go:91`），每次各自呼叫 `uow.Do`（`bfs_engine.go:32`）。根事件 commit 後才處理子事件，子事件失敗時根事件已經生效，留下半套結果。
6. **handler 的查詢不在交易內**：mediator 的 `query.Repo` 綁定全域 DB（`handler/handler.go:25`），而 SQLite 預設 `MaxOpenConns=1`。若直接把整棵事件圖包進同一個交易，handler 的查詢會等不到連線而卡住。
7. **子事件封套規則未定義**：`NewNextEvent`（`kernel/event/envelop.go:133`）沒有帶入 `AggregateType`、`MerchantID`、`UpdatedBy`、`ExpectedVersion`。子事件若沿用父事件的 aggregate 與 `ExpectedVersion`，會因父事件寫入後版本已遞增而被樂觀鎖拒絕。

### 2.3 舊 replay 的缺口（不再修正）

現有機制是 `POST /api/audit/replay`（`services/event.go:153`）：清空 projection 表後，對每個事件重跑 pipeline 與所有 projection。依前提，舊 replay 不再修正，將在 R10 由新的 BFS 重建機制（R3）取代；以下缺口改列為新機制的驗收條件。

| # | 缺口 | 影響 | 紀錄 |
|---|---|---|---|
| 1 | 清空清單漏表：`bank_csv_templates`、`bank_pdf_templates`、`bank_statement_imports`、`bank_statement_txns`（含各 `*_ledgers`）、`entry_cf_categories` 未清（`repos/unit_of_work/event_store/truncate.go:28`） | 銀行相關 UUID 由事件決定且 INSERT 未處理衝突，只要存在銀行範本或匯入紀錄，全量重建就會撞 UNIQUE 而中斷；`bank_statement_txns` 以整數 FK 引用分錄與交易，重建後關聯斷裂。但若直接清除銀行資料表，ADR-021 直接寫入的比對狀態與 `created_txn_id` 會遺失（見架構決策 A3） | ISSUE-021（Wontfix） |
| 2 | 非原子：清空先 commit，之後每個事件各自一個交易 | 中途失敗時 DB 停在半重建狀態，無法回復。若直接改成單一交易，舊 pipeline 透過全域 DB 查詢（`services/event.go:43`、`:192`）會等不到被交易占用的連線，也讀不到同一交易內尚未提交的重建結果 | ISSUE-022（Wontfix） |
| 3 | 重跑 pipeline：10 個舊 projection 依賴 pipeline 當下計算的 state（分錄等），這些資料沒有存入事件 | 重建結果取決於當下的程式碼與 DB；驗證規則變更後舊事件可能重播失敗，未清空的資料（如 `exchange_rates`）可能讓數字不同 | 第 4 節原則 1 |
| 4 | BFS 事件不在重建範圍 | BFS 事件的 `merchant_id` 為 0，被 replay 查詢的商戶過濾排除；BFS projector 也未接入 replay | ISSUE-020 |
| 5 | 事件匯入（`POST /api/audit/event/import`，`services/event.go:317`）只檢查匯出檔版本是否為 `"1"`，接著在一個交易內清除並寫入事件、commit，之後才呼叫舊 replay | replay 失敗時原有事件已被取代，無法回到匯入前的狀態；含銀行資料的匯出檔會因缺口 1 必定失敗 | ISSUE-021、022（Wontfix）；要求改列為 R10 驗收條件 |

### 2.4 交易的 aggregate 身分不一致

同一筆交易的事件分散在不同 aggregate，樂觀鎖保護的範圍也不一致（架構決策 A5）：

| 寫入來源 | 使用的 aggregate ID | 位置 |
|---|---|---|
| 前端建立傳票（`transaction.created`） | 空字串 `""`，所有傳票共用同一個 aggregate；版本號以 `getAggregateVersion('TRANSACTION', '')` 取得 | `frontside/src/api/transaction.ts:37-40` |
| 現金流分類修改（`transaction.cf_category_updated`） | 前端以交易的 `txn_uuid` 查詢版本號並送出；但後端目前仍以 `NewAggregateID` 每次隨機產生 aggregate ID，F1 會改為使用 `txn_uuid` | `frontside/src/api/cfCategory.ts`、`handler/entry_cf_category.go:70` |
| 銀行對帳核准 | 每次以 `uuid.NewV7()` 隨機產生 | `services/bank_statement_review.go:168` |

而交易本身的 `txn_uuid` 是由事件 UUID 推導的 `uuidx.NewFromEvent(eventUUID, "txn:")`（`services/projection/transaction.go:265`），和建立時的 aggregate ID 無關。影響：

- 同一筆交易的建立與後續修改記錄在不同的 aggregate，無法從單一 aggregate 看到完整歷史。
- 所有前端建立的傳票共用一個 aggregate，同時建立兩筆傳票時，其中一筆會被樂觀鎖拒絕。
- 前端目前沒有送出 `transaction.corrected`／`voided`，因此尚未出現更正、作廢的 aggregate 用法。

---

## 3. 遷移範圍

### A. 會計交易：必須遷移

| 功能 | 事件類型 | 遷移 Phase | 備註 |
|---|---|---|---|
| 傳票 | `transaction.created`／`corrected`／`voided` | R4 | 核心 |
| 期間開帳 | `period.month_started`、`period.annual_started` | R2 | 傳票、分期、預付、固定資產、投資的 pipeline 都要求交易日期所在月份為 OPEN（`validPeriodMonthlyStatus`，`services/pipelines/factory/` 共 14 處）；開帳事件本身只驗證 payload，不依賴其他資料，可以最先遷移 |
| 期間結帳 | `period.month_closed`／`month_reopened`、`period.annual_closed`／`annual_reopened` | R5 | 年結會產生結帳分錄，依賴傳票，因此留在傳票之後 |
| 銀行對帳核准 | `ApproveTxn` 產生 `transaction.created` | R7 | 目前必定失敗（ISSUE-023），先以 F3 止血 |
| 預付費用 | `prepaid.created`、`created_with_installment`、`amortized`、`disposed` | R8 | 內嵌分期（ADR-020） |
| 固定資產 | `asset.purchased`、`purchased_with_installment`、`depreciated`、`disposed` | R8 | 內嵌分期（ADR-020） |
| 分期 | `installment.created`／`period_paid` | R8 | |
| 投資 | `investment.created`／`updated`／`bought`／`sold`／`stock_split`／`dividend_received`／`unrealized_marked`、`investment.fx_rate_updated` | R9 | |

### B. 影響帳務計算的主檔與設定：必須遷移，且最先遷移（R2）

`account.created`／`updated`／`bulk_imported`、`ledger_account.created`／`updated`、`sys_config.sys_account_updated`、`account_config.ledger_type_updated`／`asset_type_updated`、`prepaid_category.*`、`fixed_asset_category.*`，以及啟動時寫入種子事件的 `BootstrapEventSeeds`（`internal/bootstrap/event_seed.go:27`）。

理由：handler 依這些資料決定分錄科目。先遷移它們，之後每個 Phase 的事件都來自 BFS，新的重建機制只需要處理 BFS 事件，不必同時支援新舊兩種事件。R2 也一併遷移 A 類中的期間開帳事件。

### C. 已定義但未實作的事件：若要實作，直接寫在 BFS 上

`account.deactivated`、`ledger_account.deactivated`、`installment.cancelled`／`completed`、`reconciliation.*`（`reconciliations` 等資料表已建立，但沒有任何寫入）。

### D. 非帳務資料

| 資料 | 現況 | 處理方式 |
|---|---|---|
| 銀行範本（CSV／PDF）、`bank_statement.imported`／`completed` | 舊路徑事件 | R6 遷移到 BFS；銀行核准（R7）依賴這些資料，必須先遷移 |
| 對帳比對狀態（`match_status`、`created_txn_id`） | 直接更新（ADR-021） | 依架構決策 A3，在 R6 處理；`bank_statement.txn_matched`／`txn_ignored`／`adjustment_approved` 已註冊 pipeline，但沒有任何呼叫點 |
| 交易範本 `transaction_templates` | 直接寫入，不產生事件 | 維持現狀；無法由事件重建，重建流程不得清除或覆寫 |
| 商戶、使用者 | 直接寫入 | 維持現狀；身分與租戶資料，不屬於帳務 |

---

## 4. 設計原則

1. **事件是已決定的事實，重建只跑 projector，不重跑 handler。**
   handler（`internal/process/`）負責驗證與計算，並把 projector 需要的資料全部寫進事件；分錄以子事件表達（例如 `journal.posted`，對應 `TODO.md`「預計專案結構」的 `domain/event/journal_posted.go`）。
2. **projector 只依賴事件內容與自身讀模型。**
   移除 `Projector.Apply` 的 `state` 參數。
3. **每個事件都要持久化。**
   BFS 已把根事件與每個子事件都寫入 `event_store`，重建時依 `event_id` 順序套用 projector 即可，不會重新產生子事件。
4. **清空範圍由 projector 自行宣告。**
   每個 projector 提供 `Reset()`（或宣告所屬資料表），取代手動維護的清空清單，避免再次漏表。
5. **重建必須可回復。**
   策略見架構決策 A4；同時重建 `aggregate_versions` 與 checkpoint。
6. **根事件封套必填欄位。**
   `AggregateType`、以業務 ID 作為 aggregate ID（例如交易 UUID）、`MerchantID` 與 `UpdatedBy`（取自 ctx）、`ExpectedVersion`（取自請求）；replay 時 `Metadata.Execution.Source` 設為 `"replay"`。
7. **子事件的交易邊界、aggregate 與版本由 engine 統一處理。**
   規則見架構決策 A1、A2，handler 不自行處理版本。
8. **依賴先行，不做新舊事件共用的 replay。**
   Phase 順序保證每個功能依賴的資料都已先遷移到 BFS；新的重建機制只處理 BFS 事件。過渡期由舊路徑寫入的開發資料不保證能重建，遷移完成後清空。
9. **遷移一種事件類型時，一併接手它在舊路徑上的所有 projection。**
   每個 Phase 開始前，列出該事件類型在 `services/projection/` 的所有處理者並逐一對應，避免讀模型漏更新。

---

## 5. 執行計畫

依 `AGENTS.md`「變更幅度與重構」，修正類與重構類分開進行、分開 review。

### 5.1 Phase 0：決策

| Phase | 內容 | 驗收條件 | 狀態 |
|---|---|---|---|
| 0 | 完成第 7 節的待決事項；寫 ADR（含是否修訂 ADR-021）；建立 `doc/STEP.md` | 使用者確認 | ⬜ pending |

### 5.2 相關 bug fix：不屬於本計畫的實作範圍

依 `AGENTS.md`，bug fix 不併入重構的實作範圍；這裡只列出與本計畫的前後依賴。每一項都依 `ISSUES.md` 各自以小幅修改實作、各自 review。原 F2a、F2b（修正舊 replay）已隨 ISSUE-021、022 改為 Wontfix 而取消。

| 項目 | 內容 | 依賴 | 驗收條件 | 狀態 |
|---|---|---|---|---|
| F1 | ISSUE-020：修正 payload 型別、註冊 projector、補齊根事件封套欄位（aggregate 用交易 UUID、`ExpectedVersion` 取自請求） | 無 | `PUT /api/txn/{txn_uuid}/cf-category` 端到端可用；補 handler 測試；完成後關閉 ISSUE-020 | ✅ done（2026-10-19；另修正 BFS Store 的初始版本檢查） |
| F3 | ISSUE-023 冪等止血，只改 `ApproveTxn`：(1) 交易 aggregate ID 改由銀行交易 UUID 衍生的固定值（目前是 `uuid.NewV7()` 隨機產生，`services/bank_statement_review.go:168`）。(2) 首次核准：用 `Append` 回傳事件的 `event_uuid`，依 projection 既有規則 `uuidx.NewFromEvent(eventUUID, "txn:")`（`services/projection/transaction.go:265`）推導交易 UUID 後查交易、寫入關聯；修正目前以 aggregate ID 查交易、必定查不到的問題（`bank_statement_review.go:186`）。(3) 重試或並行：先以 `query.Event.GetByAggregate` 找既有的 `transaction.created` 事件，存在就用它的 `event_uuid` 推導交易並補寫關聯，不再 append。(4) 並行：舊路徑 `Append` 遇到版本衝突時，回傳的是寫入版本紀錄的唯一鍵錯誤，沒有統一的版本衝突錯誤（`services/event.go:75`），因此 `Append` 失敗後一律重查固定 aggregate，確認已有 `transaction.created` 才走補關聯；查不到就保留原本的失敗，不得視為成功。目前沒有需要保留的資料，既有的未關聯交易不另行處理 | S4 (a)，已確定 | 測試涵蓋：首次核准成功並寫入關聯；已核准後再按回錯誤；回寫失敗後重試只補關聯；並行：強制兩個請求都先查到沒有事件（以替換事件查詢的 stub，讓對手在本次建立前完成核准），本次 `Append` 確實衝突後走恢復路徑成功，銀行明細為 `APPROVED` 且關聯到唯一的交易；`Append` 失敗且重查不到事件時回傳錯誤、不建立交易 | ✅ done（2026-10-19；ISSUE-023 維持 In Progress，原子化待 R7） |

### 5.3 重構類：BFS 遷移

依 `AGENTS.md`，每個 Phase 須列出預期的行為變更；未列出的行為一律維持不變。R2 之後的每個 Phase 都要依原則 9 接手舊 projection。

| Phase | 內容 | 預期的行為變更 | 驗收條件 | 狀態 |
|---|---|---|---|---|
| R0 | 交易內查詢：讓 `query.Repo` 可綁定 `sqlx.Tx`。目前 `query.NewQueryRepository` 只接受 `*sqlx.DB`，account、investment、report 等 repo 也直接使用它（`repos/query/base.go:37`） | 無 | 既有測試全數通過；新增在交易內查詢的測試 | ⬜ pending |
| R1 | engine 基礎：事件圖交易邊界、子事件封套與版本規則、handler 改用 R0 的交易內查詢、移除 projector 的 `state` 參數 | 事件圖的提交方式依 A1 改變 | engine 測試涵蓋子事件失敗回復、同 aggregate 連續版本、跨 aggregate 子事件 | ⬜ pending |
| R2 | 主檔與設定（B 類），含 `BootstrapEventSeeds` 改走 BFS；以及期間開帳 `period.month_started`、`period.annual_started` | 無（事件內容不變） | 啟動時的種子事件與開帳事件經 BFS 寫入且封套欄位完整；主檔、設定與期間 API 回應不變；既有測試全數通過 | ⬜ pending |
| R3 | BFS 重建機制：只跑 projector；各 projector 以 `Reset()` 宣告清空範圍；依 A4 可回復；重建 `aggregate_versions` 與 checkpoint。承接 ISSUE-021、022 的要求。`POST /api/audit/replay` 維持舊機制，到 R10 才切換 | 無（新機制先以 engine 方法與測試提供，尚不對外） | 以 R2 的事件通過第 8 節的 canonical 比較；重建中途失敗時資料回到重建前的狀態；清空範圍涵蓋所有 BFS projector 的資料表 | ⬜ pending |
| R4 | 傳票 `transaction.*`，產生 `journal.posted` 子事件；依 A5 統一交易的 aggregate 身分；接手處理傳票事件的舊 projection（以 `transaction.created` 為例：`transaction`、`entry_cf_category`、`account_balance_realtime`） | `event_store` 新增 `journal.posted`；傳票改為每筆交易各自一個 aggregate（A5） | API 回應不變；canonical 比較通過；並行建立不同交易都成功；以過期版本修改同一筆交易必定失敗 | ⬜ pending |
| R5 | 期間結帳與重開（月結、年結） | 年結分錄改由子事件產生 | canonical 比較通過；年結產生的結帳、開帳兩筆交易 UUID 不同，且重建後保持相同 | ⬜ pending |
| R6 | 銀行範本（CSV／PDF）、`bank_statement.imported`／`completed` 改走 BFS；依 A3 處理對帳比對狀態 | 依 A3（選 a 時比對狀態改為事件） | 同上；銀行資料可由事件重建，或依 A3 明確排除 | ⬜ pending |
| R7 | 銀行對帳核准原子化：核准事件產生 `transaction.created` 子事件，與關聯回寫在同一交易完成（ISSUE-023 的根本修正，取代 F3 的止血做法） | 核准改為事件產生子事件 | 同上；中途失敗時不留下沒有關聯的交易 | ⬜ pending |
| R8 | 預付費用、固定資產、分期 | 分錄改由子事件產生 | 同上 | ⬜ pending |
| R9 | 投資（含 `investment.created`／`updated`、`investment.fx_rate_updated`） | 分錄改由子事件產生 | 同上 | ⬜ pending |
| R10 | 移除舊 pipeline、projection 與舊 replay；`POST /api/audit/replay` 切換到 R3 的新機制；`POST /api/audit/event/import` 改用新機制：提升匯出檔版本（目前 `Import` 只接受 `"1"`，`services/event.go:318`），不支援的版本或事件格式在清除資料前拒絕，清除、寫入與重建在同一交易完成（依 A4）；經使用者確認後清空開發 DB 重新開始 | replay 與匯入改用新機制；匯入只接受新版匯出檔；不再支援舊路徑寫入的事件 | 門檻：R0～R9 全部完成。舊寫入入口無呼叫點；全套測試與 canonical 比較通過；新版匯出 → 匯入 → 重建的 canonical 比較一致；不相容的匯入失敗後原資料不變；清空 DB 前先以 `VACUUM INTO` 備份 | ⬜ pending |

---

## 6. 風險

| 風險 | 對策 |
|---|---|
| R4 之後，舊 replay 無法處理 BFS 新增的事件類型（例如 `journal.posted`）；新機制也無法處理舊路徑寫入的事件 | 依前提，R10 之前不以 `POST /api/audit/replay` 重建開發資料；重建一致性只在測試中以新機制驗證；開發 DB 可隨時清空重來 |
| 遷移某事件類型時漏接舊路徑上的 projection | 依原則 9，每個 Phase 開始前列出處理者清單並在 review 時核對 |
| SQLite 單一連線下，交易內再用全域 DB 查詢會卡住，且讀不到交易內尚未提交的資料 | R0 讓查詢可綁定交易；R1 起 handler 一律使用交易內查詢，並以測試覆蓋 |
| 清空開發 DB 是破壞性操作 | R10 執行前取得使用者確認，並先以 `VACUUM INTO` 備份 |
| 匯入舊版或不相容的事件檔，先覆蓋現有資料才因重建失敗而報錯 | R10 提升匯出檔版本，並在清除資料前檢查版本與事件格式；清除、寫入與重建在同一交易完成 |
| 事件 payload 結構變更 | 使用 `Metadata.Schema.SchemaVersion` 區分版本 |
| 前端 API 契約變動 | 依範圍決策 S3；若改用專屬 endpoint，需同步修改 `frontside/src/api/` 與 `src/types/` |
| 重建一致性測試需要真實資料 | 依 `backend/CLAUDE.md`，使用手動建立、可追溯的測試事件，不使用 AI 生成的假資料 |

---

## 7. 決策

### 7.1 已確定

| # | 問題 | 決定 | 理由 |
|---|---|---|---|
| S1 | 「重大交易」範圍 | (b) A＋B 類 | 主檔與設定影響分錄科目；調整後的 Phase 順序需要它們最先遷移 |
| S2 | 舊事件如何重建 | (c) 不保留歷史事件，遷移完成後清空開發 DB 重新開始（另兩個選項：(a) 過渡期沿用舊 pipeline 重播；(b) 一次性 upcast 補入分錄） | 目前沒有需要保留的資料；不需要歷史事件轉換與歷史重播相容層 |
| S4 | ISSUE-023 的處理方式 | (a) 先做冪等止血（F3），原子化留到 R7（另一個選項：(b) 先完成 R0 與 R1 的最小必要部分，再直接做原子修正） | 改動只在 `ApproveTxn` 一個函式，能立即讓核准功能可用並止住重複的會計交易；回寫失敗時仍會暫時留下沒有關聯的交易，但重試會補上，不會重複建立 |

### 7.2 架構決策：Phase 0 必須完成

| # | 問題 | 選項 | 建議 |
|---|---|---|---|
| A1 | BFS 事件圖的交易邊界 | (a) 整棵事件圖一個交易，任一事件失敗全部回復；handler 改用交易內的查詢　(b) 維持每個事件一個交易，失敗時以補償事件處理 | (a)：符合帳務原子性，也是在 BFS 上解決 ISSUE-023 的前提 |
| A2 | 子事件的 aggregate、版本與欄位繼承 | aggregate：(a) 沿用父事件　(b) 建立自己的 aggregate，ID 由父事件 UUID 衍生。版本：(c) engine 在事件圖內追蹤各 aggregate 的最新版本，依每次 append 的回傳值配置下一個　(d) engine 預先配置整批版本 | 同一 aggregate 的後續變化用 (a)，跨 aggregate 的衍生事實（如 `journal.posted`）用 (b)；版本用 (c)；`MerchantID`、`UpdatedBy`、`CorrelationID` 一律繼承父事件，`CausationID` 指向父事件；與 A5 的身分映射見本表下方 |
| A3 | ADR-021 直接寫入的銀行對帳狀態在重建中如何處理 | (a) 將比對、忽略、核准關聯事件化，並修訂 ADR-021　(b) 重建前保存直接寫入的欄位，重建後依 UUID 恢復　(c) 明確排除銀行對帳資料，不宣稱可由事件完整重建；重建時不清除也不重播銀行資料，但 `bank_statement_txns` 仍需改用 UUID 關聯 | (a)：相關事件類型已定義且已註冊 pipeline，成本較低，也最符合「可由事件重建」的目標；目前沒有需要保留的資料，不需 backfill |
| A5 | 交易的 aggregate 身分（見 2.4） | (a) 每筆交易各自一個 aggregate，aggregate ID 就是 `txn_uuid`：建立時先決定交易 UUID（由 handler 產生，或依 A2 由父事件 UUID 推導），寫入事件的 aggregate ID，projection 直接使用它，不再由事件 UUID 另行推導；之後的更正、作廢、現金流分類修改都使用同一個 aggregate　(b) 維持現狀：建立時共用一個 aggregate，後續修改以 `txn_uuid` 為 aggregate | (a)：同一筆交易的歷史集中在一個 aggregate，樂觀鎖保護的是單筆交易，同時建立多筆傳票也不會互相衝突。R4 實作時需處理前端建立傳票仍送 aggregate `""` 與其版本號的問題：依 S3 (a)，由後端在建立時產生交易 UUID 並以版本 0 寫入，前端不必修改；與 A2 的身分映射見本表下方 |
| A4 | 新重建機制的原子化策略 | (a) 在單一交易內完成全量重建，projector 只透過交易內的 repository 讀寫　(b) 重建到影子表後再切換　(c) 重建前以 `VACUUM INTO` 備份，失敗時還原；重建期間必須停止所有寫入，還原會影響整個 DB 而非單一商戶 | (a)：新機制只跑 projector，不會遇到舊 pipeline 在交易外查詢的問題，實作最簡單；代價是重建期間 DB 被鎖住 |

**A2 與 A5 的身分映射規則**（建議，與 A2、A5 一併確認）

區分「事件所屬的 aggregate」與「事件影響的交易」：

1. **直接建立的交易**（例如前端建立傳票）：handler 產生交易 UUID，作為該事件的 aggregate ID，也就是交易的 `txn_uuid`。
2. **由其他事件衍生的交易**（例如年結、銀行核准、預付建立時產生的交易）：`txn_uuid` 由父事件 UUID 加上用途識別推導，沿用目前 `uuidx.NewFromEvent(eventUUID, "txn:"+用途)` 的做法（`services/projection/transaction.go:265`）。年結會從同一個事件產生結帳與開帳兩筆交易（`:134`、`:139`），用途識別 `closing`、`opening` 確保兩者的 UUID 不同且可重建。衍生交易的 aggregate ID 與它的 `txn_uuid` 相同。
3. **不屬於交易 aggregate 的事件**（例如依 A2 使用自己 aggregate 的 `journal.posted`）：payload 必須帶上目標交易的 `txn_uuid`；projector 依 payload 判斷分錄屬於哪筆交易，不得把事件自己的 aggregate ID 當成交易 ID。

### 7.3 範圍決策：待決

| # | 問題 | 選項 | 建議 |
|---|---|---|---|
| S3 | API 形式 | (a) 維持 `POST /api/event/append`，後端依事件類型改派至 BFS　(b) 各功能改為專屬 endpoint | (a)：前端不需修改，跨端契約不變；需要事件類型對應 payload 的解碼註冊表 |

---

## 8. 驗證方式

1. 每個項目依 `backend/CLAUDE.md` 執行 `go generate ./...`、`go build ./...`、`go test ./...`。
2. **重建一致性測試**（R3 建立，之後每個 Phase 沿用），採 canonical 比較，不直接比對資料表原始內容：
   - 流程：以 BFS 寫入一組手動建立的測試事件 → 依下列規格匯出 projection 資料 → 全量重建 → 再次匯出並比對。
   - 以業務 UUID 排序與建立關聯，不比對 surrogate integer ID；多租戶下不重置 `sqlite_sequence`（`truncate.go:82`），重建後整數 ID 會改變。
   - 排除由重建當下產生的非決定性欄位（例如重建時間）。
   - 金額、狀態、UUID 關聯與筆數必須一致。
   - 執行 `PRAGMA foreign_key_check`，不得有 orphan FK。
   - 不在重建範圍內的資料（D 類中維持直接寫入的資料，以及依 A3 排除的資料）另列清單，驗證重建前後沒有被改動。
3. **匯入驗證**（R10）：新版匯出 → 匯入 → 重建，canonical 比較一致；匯入不相容版本或事件格式時，在清除前拒絕，原資料不變。
4. 若 API 契約變動，另依 `frontside/CLAUDE.md` 執行 `npm run lint`、`npm run check`、`npm run build`。
