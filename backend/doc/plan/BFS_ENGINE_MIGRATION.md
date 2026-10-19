# BFS Engine 遷移與事件重建計畫書

| 項目 | 內容 |
|---|---|
| 狀態 | 📝 草案：待完成 Phase 0 的架構決策與範圍決策（第 7 節） |
| 建立日期 | 2026-10-19 |
| 最後修訂 | 2026-10-19：依兩輪 review 補上事件圖交易邊界、子事件規則、銀行對帳資料處理與一致性比較規格；bug fix 與重構分開；新增 R0（交易內查詢），F2 拆為 F2a／F2b，F3 改為冪等止血；第三輪補上 F3 的交易查找流程，並把舊路徑移除拆為 R9～R11（含移除門檻） |
| 分支 | `refactor`（基準 commit `82e8990`） |
| 關聯 | ISSUE-020～023；ADR-001、ADR-014、ADR-017、ADR-018、ADR-020、ADR-021 |

> 本文件是計畫書。Phase 0 的決策確認後，依 `backend/CLAUDE.md` 的需求處理流程寫成 ADR，並在 `doc/STEP.md` 建立進度表。
> 文中行號以基準 commit 加上 2026-10-19 的工作區狀態為準。

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

### 2.3 事件重建機制的缺口

現有機制是 `POST /api/audit/replay`（`services/event.go:153`）：清空 projection 表後，對每個事件重跑 pipeline 與所有 projection。

| # | 缺口 | 影響 | 紀錄 |
|---|---|---|---|
| 1 | 清空清單漏表：`bank_csv_templates`、`bank_pdf_templates`、`bank_statement_imports`、`bank_statement_txns`（含各 `*_ledgers`）、`entry_cf_categories` 未清（`repos/unit_of_work/event_store/truncate.go:28`） | 銀行相關 UUID 由事件決定且 INSERT 未處理衝突，只要存在銀行範本或匯入紀錄，全量重建就會撞 UNIQUE 而中斷；`bank_statement_txns` 以整數 FK 引用分錄與交易，重建後關聯斷裂。但若直接清除銀行資料表，ADR-021 直接寫入的比對狀態與 `created_txn_id` 會遺失（見架構決策 A3） | ISSUE-021 |
| 2 | 非原子：清空先 commit，之後每個事件各自一個交易 | 中途失敗時 DB 停在半重建狀態，無法回復。若直接改成單一交易，舊 pipeline 透過全域 DB 查詢（`services/event.go:43`、`:192`）會等不到被交易占用的連線，也讀不到同一交易內尚未提交的重建結果（見架構決策 A4） | ISSUE-022 |
| 3 | 重跑 pipeline：10 個舊 projection 依賴 pipeline 當下計算的 state（分錄等），這些資料沒有存入事件 | 重建結果取決於當下的程式碼與 DB；驗證規則變更後舊事件可能重播失敗，未清空的資料（如 `exchange_rates`）可能讓數字不同 | 第 4 節 |
| 4 | BFS 事件不在重建範圍 | BFS 事件的 `merchant_id` 為 0，被 replay 查詢的商戶過濾排除；BFS projector 也未接入 replay | ISSUE-020 |

---

## 3. 遷移範圍

### A. 會計交易：必須遷移

| 功能 | 事件類型 | 備註 |
|---|---|---|
| 傳票 | `transaction.created`／`corrected`／`voided` | 核心，最優先 |
| 期間結帳 | `period.month_started`／`month_closed`／`month_reopened`、`period.annual_started`／`annual_closed`／`annual_reopened` | 年結會產生結帳分錄 |
| 銀行對帳核准 | `ApproveTxn` 產生 `transaction.created` | 目前分兩個交易（ISSUE-023）；修正位置見範圍決策 S4 |
| 預付費用 | `prepaid.created`、`created_with_installment`、`amortized`、`disposed` | 內嵌分期（ADR-020） |
| 固定資產 | `asset.purchased`、`purchased_with_installment`、`depreciated`、`disposed` | 內嵌分期（ADR-020） |
| 分期 | `installment.created`／`period_paid` | |
| 投資 | `investment.bought`／`sold`／`stock_split`／`dividend_received`／`unrealized_marked` | |

### B. 影響帳務計算的主檔與設定：建議一併遷移（範圍決策 S1）

`account.created`／`updated`／`bulk_imported`、`ledger_account.created`／`updated`、`sys_config.sys_account_updated`、`account_config.ledger_type_updated`／`asset_type_updated`、`prepaid_category.*`、`fixed_asset_category.*`、`investment.created`／`updated`／`fx_rate_updated`。

理由：pipeline 依這些資料決定分錄科目，重建時它們必須和交易事件依同一個全域順序重播。

### C. 已定義但未實作的事件：若要實作，直接寫在 BFS 上

`account.deactivated`、`ledger_account.deactivated`、`installment.cancelled`／`completed`、`reconciliation.*`（`reconciliations` 等資料表已建立，但沒有任何寫入）。

### D. 暫不納入 BFS

| 資料 | 現況 | 說明 |
|---|---|---|
| 銀行範本（CSV／PDF）、`bank_statement.imported`／`completed` | 舊路徑事件 | 非帳務資料，最後隨舊路徑一起收斂 |
| 對帳比對狀態（`match_status`、`created_txn_id`） | 直接更新（ADR-021） | 重建時的處理方式見架構決策 A3；`bank_statement.txn_matched`／`txn_ignored`／`adjustment_approved` 已註冊 pipeline，但沒有任何呼叫點 |
| 交易範本 `transaction_templates` | 直接寫入，不產生事件 | 無法由事件重建；重建流程不得清除或覆寫 |
| 商戶、使用者 | 直接寫入 | 身分與租戶資料，不屬於帳務 |

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
8. **過渡期使用統一的 replay runner。**
   新舊路徑寫入同一個 `event_store`；runner 依 `event_id` 全域順序，把每個事件交給負責它的 projection 組（舊事件走舊 pipeline 加 projection，BFS 事件走 BFS projector）。

---

## 5. 執行計畫

依 `AGENTS.md`「變更幅度與重構」，修正類與重構類分開進行、分開 review。

### 5.1 Phase 0：決策

| Phase | 內容 | 驗收條件 | 狀態 |
|---|---|---|---|
| 0 | 完成第 7 節的架構決策與範圍決策；寫 ADR（含是否修訂 ADR-021）；建立 `doc/STEP.md` | 使用者確認 | ⬜ pending |

### 5.2 相關 bug fix：不屬於本計畫的實作範圍

依 `AGENTS.md`，bug fix 不併入重構的實作範圍；這裡只列出與本計畫的前後依賴。每一項都依 `ISSUES.md` 各自以小幅修改實作、各自 review。

| 項目 | 內容 | 依賴 | 驗收條件 | 狀態 |
|---|---|---|---|---|
| F1 | ISSUE-020：修正 payload 型別、註冊 projector、補齊根事件封套欄位（aggregate 用交易 UUID、`ExpectedVersion` 取自請求） | 無 | `PUT /api/txn/{txn_uuid}/cf-category` 端到端可用；補 handler 測試；完成後關閉 ISSUE-020 | ⬜ pending |
| F2a | ISSUE-021：舊 replay 的清空範圍，含 `bank_statement_txns` 改用 UUID 關聯 | A3；選 (a) 時，比對狀態的事件化與 backfill 屬於新功能，需另列項目先完成 | 依第 8 節的 canonical 比較通過 | ⬜ pending |
| F2b | ISSUE-022：舊 replay 原子化 | A4；選 (a) 時需先完成 R0 | 重建中途失敗時，資料回到重建前的狀態 | ⬜ pending |
| F3 | ISSUE-023 冪等止血，只改 `ApproveTxn`：(1) 交易 aggregate ID 改由銀行交易 UUID 衍生的固定值（目前是 `uuid.NewV7()` 隨機產生，`services/bank_statement_review.go:168`）。(2) 首次核准：用 `Append` 回傳事件的 `event_uuid`，依 projection 既有規則 `uuidx.NewFromEvent(eventUUID, "txn:")`（`services/projection/transaction.go:265`）推導交易 UUID 後查交易、寫入關聯；修正目前以 aggregate ID 查交易、必定查不到的問題（`bank_statement_review.go:186`）。(3) 重試或並行：先以 `query.Event.GetByAggregate` 找既有的 `transaction.created` 事件，存在就用它的 `event_uuid` 推導交易並補寫關聯，不再 append；並行時被樂觀鎖擋下的一方也走這條路徑。既有的未關聯交易需另行盤點，不在 F3 範圍 | S4 選 (a) | 三種情境都要有測試：首次核准成功並寫入關聯；回寫失敗後重試只補關聯；並行核准只產生一筆交易 | ⬜ pending |

### 5.3 重構類：BFS 遷移

依 `AGENTS.md`，每個 Phase 須列出預期的行為變更；未列出的行為一律維持不變。

| Phase | 內容 | 預期的行為變更 | 驗收條件 | 狀態 |
|---|---|---|---|---|
| R0 | 交易內查詢：讓 `query.Repo` 可綁定 `sqlx.Tx`。目前 `query.NewQueryRepository` 只接受 `*sqlx.DB`，account、investment、report 等 repo 也直接使用它（`repos/query/base.go:37`）。這是 F2b（A4 選 a 時）與 R1 的共同前置 | 無 | 既有測試全數通過；新增在交易內查詢的測試 | ⬜ pending |
| R1 | engine 基礎：事件圖交易邊界、子事件封套與版本規則、handler 改用 R0 的交易內查詢、移除 projector 的 `state` 參數 | 事件圖的提交方式依 A1 改變 | engine 測試涵蓋子事件失敗回復、同 aggregate 連續版本、跨 aggregate 子事件 | ⬜ pending |
| R2 | 統一 replay runner：舊事件走舊 pipeline，BFS 事件只跑 projector | replay 開始涵蓋 BFS 事件 | 混合新舊事件的 canonical 比較通過 | ⬜ pending |
| R3 | 傳票 `transaction.*`，產生 `journal.posted` 子事件 | `event_store` 新增 `journal.posted` | API 回應不變；canonical 比較通過 | ⬜ pending |
| R4 | 期間結帳（月結、年結） | 年結分錄改由子事件產生 | 同上 | ⬜ pending |
| R5 | 銀行對帳核准原子化：核准事件產生 `transaction.created` 子事件，與關聯回寫在同一交易完成（ISSUE-023 的根本修正） | 核准改為事件產生子事件 | 同上；中途失敗時不留下沒有關聯的交易 | ⬜ pending |
| R6 | 預付費用、固定資產、分期 | 分錄改由子事件產生 | 同上 | ⬜ pending |
| R7 | 投資，以及 `investment.fx_rate_updated` | 分錄改由子事件產生 | 同上 | ⬜ pending |
| R8 | 主檔與設定（依 S1） | 無 | 同上 | ⬜ pending |
| R9 | D 類舊路徑事件收斂：銀行範本（CSV／PDF）、`bank_statement.imported`／`completed`，以及依 A3 事件化的比對事件改走 BFS | 無（事件內容不變） | 同上 | ⬜ pending |
| R10 | 歷史事件轉換：依 S2 為歷史事件補上 projector 所需的事實（例如分錄），使其不必再靠舊 pipeline 重建；做法（讀取時 upcast 或補寫事件）在本 Phase 開始前另行決定 | 歷史事件的重建方式改變 | 只用 BFS projector 全量重建，canonical 比較通過 | ⬜ pending |
| R11 | 移除舊 pipeline、projection 與舊 replay | 無 | 門檻：R1～R10 全部完成。若 R10 尚未完成，只能移除舊的寫入入口，舊 pipeline 與 projection 保留為「歷史重播相容層」。舊寫入入口無呼叫點；全套測試與 canonical 比較通過 | ⬜ pending |

---

## 6. 風險

| 風險 | 對策 |
|---|---|
| 過渡期新舊路徑同時寫入，replay 分派錯誤 | R2 先完成統一 runner，並以混合事件的測試資料驗證 |
| SQLite 單一連線下，交易內再用全域 DB 查詢會卡住，且讀不到交易內尚未提交的資料 | R0 讓查詢可綁定交易；F2b 與 R1 都改用交易內查詢，並以測試覆蓋 |
| 舊事件沒有分錄資料，無法只靠 projector 重建 | 依範圍決策 S2：過渡期沿用舊 pipeline 重播；R11 完整移除舊 pipeline 的門檻是 R10 完成，在那之前保留歷史重播相容層 |
| 銀行對帳狀態事件化後（A3 選 a），既有紀錄沒有對應事件 | 撰寫一次性 backfill，為既有狀態補寫事件 |
| 事件 payload 結構變更 | 使用 `Metadata.Schema.SchemaVersion` 區分版本，必要時撰寫 upcast |
| 前端 API 契約變動 | 依範圍決策 S3；若改用專屬 endpoint，需同步修改 `frontside/src/api/` 與 `src/types/` |
| 重建一致性測試需要真實資料 | 依 `backend/CLAUDE.md`，使用手動建立、可追溯的測試事件，不使用 AI 生成的假資料 |

---

## 7. 待決事項

### 7.1 架構決策：Phase 0 必須完成

| # | 問題 | 選項 | 建議 |
|---|---|---|---|
| A1 | BFS 事件圖的交易邊界 | (a) 整棵事件圖一個交易，任一事件失敗全部回復；handler 改用交易內的查詢　(b) 維持每個事件一個交易，失敗時以補償事件處理 | (a)：符合帳務原子性，也是在 BFS 上解決 ISSUE-023 的前提 |
| A2 | 子事件的 aggregate、版本與欄位繼承 | aggregate：(a) 沿用父事件　(b) 建立自己的 aggregate，ID 由父事件 UUID 衍生。版本：(c) engine 在事件圖內追蹤各 aggregate 的最新版本，依每次 append 的回傳值配置下一個　(d) engine 預先配置整批版本 | 同一 aggregate 的後續變化用 (a)，跨 aggregate 的衍生事實（如 `journal.posted`）用 (b)；版本用 (c)；`MerchantID`、`UpdatedBy`、`CorrelationID` 一律繼承父事件，`CausationID` 指向父事件 |
| A3 | ADR-021 直接寫入的銀行對帳狀態在重建中如何處理 | (a) 將比對、忽略、核准關聯事件化，並修訂 ADR-021　(b) 重建前保存直接寫入的欄位，重建後依 UUID 恢復　(c) 明確排除銀行對帳資料，不宣稱可由事件完整重建；重建時不清除也不重播銀行資料，但 `bank_statement_txns` 仍需改用 UUID 關聯 | (a)：相關事件類型已定義且已註冊 pipeline，成本較低，也最符合「可由事件重建」的目標；既有紀錄需 backfill |
| A4 | 重建的原子化策略 | (a) 在單一交易內完成全量重建；舊 pipeline 也必須在交易內查詢，前置為 R0　(b) 重建到影子表後再切換；pipeline 的查詢需改指向影子表　(c) 重建前以 `VACUUM INTO` 備份，失敗時還原；重建期間必須停止所有寫入，還原會影響整個 DB 而非單一商戶 | (a)：R0 本來就是 R1 的前置，不算額外工作；代價是重建期間 DB 被鎖住 |

### 7.2 範圍決策

| # | 問題 | 選項 | 建議 |
|---|---|---|---|
| S1 | 「重大交易」範圍 | (a) 只含 A 類　(b) A＋B 類 | (b)：B 類影響分錄科目與重建順序 |
| S2 | 舊事件如何重建 | (a) 過渡期沿用舊 pipeline 重播　(b) 一次性 upcast 補入分錄後全部只跑 projector | (a) 起步風險較低。但要完整移除舊 pipeline（R11），必須先完成 (b)（R10）；在那之前舊 pipeline 保留為歷史重播相容層 |
| S3 | API 形式 | (a) 維持 `POST /api/event/append`，後端依事件類型改派至 BFS　(b) 各功能改為專屬 endpoint | (a)：前端不需修改，跨端契約不變；需要事件類型對應 payload 的解碼註冊表 |
| S4 | ISSUE-023 的處理方式 | (a) 先做冪等止血（F3），原子化留到 R5；回寫失敗時仍會暫時留下沒有關聯的交易，但重試會補上，不會重複建立　(b) 先完成 R0 與 R1 的最小必要部分，再直接做原子修正 | (a)：改動只在 `ApproveTxn` 一個函式，能立即止住重複的會計交易。讓 `EventStoreService.Append` 在外部交易內執行並不是小幅修改，需要拆開 pipeline、寫入與 projection，並處理交易前計算造成的競態與巢狀交易 |

---

## 8. 驗證方式

1. 每個項目依 `backend/CLAUDE.md` 執行 `go generate ./...`、`go build ./...`、`go test ./...`。
2. **重建一致性測試**（F2a 建立，之後每個項目沿用），採 canonical 比較，不直接比對資料表原始內容：
   - 流程：寫入一組手動建立的測試事件 → 依下列規格匯出 projection 資料 → 全量重建 → 再次匯出並比對。
   - 以業務 UUID 排序與建立關聯，不比對 surrogate integer ID；多租戶下不重置 `sqlite_sequence`（`truncate.go:82`），重建後整數 ID 會改變。
   - 排除由重建當下產生的非決定性欄位（例如重建時間）。
   - 金額、狀態、UUID 關聯與筆數必須一致。
   - 執行 `PRAGMA foreign_key_check`，不得有 orphan FK。
   - 不在重建範圍內的資料（D 類，以及依 A3 排除的資料）另列清單，驗證重建前後沒有被改動。
3. 若 API 契約變動，另依 `frontside/CLAUDE.md` 執行 `npm run lint`、`npm run check`、`npm run build`。
