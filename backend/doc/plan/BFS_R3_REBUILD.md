# R3 實作計畫書：BFS 重建機制

| 項目 | 內容 |
|---|---|
| 狀態 | ✅ review 通過（2026-10-27），實作中；進度見 `doc/STEP.md` 的 R3 子階段 |
| 建立日期 | 2026-10-27 |
| 最後修訂 | 2026-10-27：依 review 修訂兩處。(1) D3：舊 projection 與 BFS 共用 `CreateLedgerAccount` 查詢與 mapper（`repos/unit_of_work/event_store/projection_repo/account.go:50`、`persistence/repos/projection/account.go:35`）。若直接修改這個查詢，舊 projection 沒有設定 `LedgerId`，重新產生 mapper 後會寫入 0，舊 replay 重播第二個明細帳戶時就會撞主鍵。改為 BFS 另外新增可以指定 ID 的查詢，原查詢不動，段 1 補上舊 replay 重播兩個明細帳戶的情境。(2) D9：`transaction.cf_category_updated` 的 handler 必須查到有效的交易與分錄（`process/cashflow.go:30-55`），這組前置資料只能以 SQL 建立。改為明列前置資料，以 BFS 寫入前後的差異找出清空範圍，並驗證前置的交易與分錄不變 |
| 分支 | `refactor`（基準 commit `9b18ba6`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md`：5.3 的 R3、第 8 節 |
| 前置 | R0、R1、R2、F9（皆已完成） |
| 關聯 | ADR-023（原則 1～5、8）、ADR-025（A4）、ADR-027 規則 1；ISSUE-021、ISSUE-022、ISSUE-031、ISSUE-033、ISSUE-034、ISSUE-035 |

> R3 是重構，依 `AGENTS.md`「變更幅度與重構」，只做本計畫列出的行為變更（第 4 節）；盤點時發現的 bug 另開 F10，不併入 R3。
> 文中行號以基準 commit 為準。

---

## 1. 範圍

### 1.1 重建的資料

R3 只重建 BFS projector 的讀模型，涵蓋註冊表（`domain/event/registry.go:57`）目前登錄的 17 種事件。重建單位是單一商戶，在一個交易內完成（ADR-025 A4）。

| projector（checkpoint 名稱） | 事件類型 | 清空範圍（D1） |
|---|---|---|
| `accountProjector` | `account.created`／`updated`／`bulk_imported`、`ledger_account.created`／`updated` | `accounts`、`ledger_accounts`、`account_closure`（由 `accounts` 的 INSERT trigger 維護，`database/migrations/20260514005_fix_closure_trigger.sql:7`） |
| `configProjector` | `sys_config.sys_account_updated`、`account_config.ledger_type_updated`／`asset_type_updated` | `sys_accounts`、`ledger_account_type_config`、`asset_type_account_config` |
| `prepaidCategoryProjector` | `prepaid_category.*` | `prepaid_categories` |
| `fixedAssetCategoryProjector` | `fixed_asset_category.*` | `fixed_asset_categories` |
| `periodProjector` | `period.month_started`、`period.annual_started` | `period_closings` |
| `cashFlowCategoryProjector` | `transaction.cf_category_updated` | `entry_cf_categories` |

事件核心的資料表：
- `aggregate_versions` 依事件重算（D6）。
- `projection_checkpoints` 只重建上表 6 個 BFS projector 的列（D6）。
- `snapshots` 不動（D7）。
- `event_store` 只讀取。

**範圍外**：重建前後必須不變。
- 其他商戶的所有資料。
- 本商戶的 `event_store`、`snapshots`。
- 舊 projection 的資料表與它們的 checkpoint。
- 直接寫入的資料：`transaction_templates`、`transaction_template_entries`、`merchants`、`users`、`user_merchants`。
- 全域的 `exchange_rates`。

### 1.2 盤點時發現的問題

1. **ISSUE-031 需要 engine 支援，而且 `ledger_id` 必須在重建後維持不變，只是事件查得到還不夠。**
   - mediator handler 無法補充根事件的 payload：`bfs_engine.go:102` 的 `Dispatch` 回傳 `result.EventResult`（`kernel/result/result.go:13`），裡面只有子事件與 policy；`:111` 寫入的是原本的事件。`ledger_account.created` 的 handler 只做 payload 驗證（`process/registry.go:15`）。
   - 以整數 `ledger_id` 引用明細帳戶的不只是事件：
     - 前端有 24 個檔案使用 `ledger_id`。
     - 交易範本 `transaction_template_entries.ledger_id`（`database/migrations/20260525001_transaction_template.sql:19`）是直接寫入的資料，不在重建範圍內，重建也不會改寫它。
   - 因此「事件改用 `ledger_uuid` 引用」仍會讓 `ledger_id` 在重建後改變，交易範本就會指到錯的明細帳戶（D3）。
2. **有兩張讀模型表與舊路徑共用。**
   - `period_closings`：BFS 只負責開帳。舊路徑的結帳會新增下一個期間（`services/projection/period.go:94`、`:177`），重開會修改狀態（`:110`、`:192`），年結會回寫交易 ID（`services/projection/transaction.go:144`）。
   - `entry_cf_categories`：舊的傳票 projection 會寫入預設分類（`services/projection/entry_cf_category.go:192`、`:233`、`:253`）。
   - 對有舊事件的商戶清空這兩張表，會遺失舊路徑寫入的狀態。由 D5 的前置檢查排除（原則 8）。
3. **舊路徑寫入過已登錄類型的事件。**
   - R2 之前，類別建立事件寫在隨機的 aggregate（ISSUE-029），期間開帳寫在 `('ACCOUNT', '')`。
   - migration `20261022003` 為這些類別補了沒有對應事件的版本紀錄。
   - 若依事件重算 `aggregate_versions`，這些紀錄會消失，之後的修改就會版本衝突。同樣由 D5 排除。
4. **ISSUE-033 影響重建範圍內的 3 張表。**
   - 以容器內的 Python sqlite3（3.46.1）套用全部 migration 確認：`accounts`、`ledger_accounts`、`sys_accounts` 的 `PRAGMA foreign_key_check` 會回傳 `foreign key mismatch`。
   - 範圍內只有 `ledger_account_type_config` 的外鍵定義正確，其餘資料表沒有宣告外鍵（D8）。
5. **唯一鍵缺少 `merchant_id` 的盤點（`TODO.md`，原訂 R3 之前完成）。**
   - 方法同上：列出最終 schema 中所有主鍵與唯一索引。
   - **有問題**：只有 `investments` 的 `idx_investments_symbol(symbol, currency)`（`database/migrations/20260310001_init_table.sql:238`）是缺少 `merchant_id` 的業務鍵。`CreateInvestment` 是一般的 INSERT（`database/queries/investment.sql:95`），因此一個商戶建立某個代號與幣別的投資後，其他商戶建立相同代號與幣別時會撞唯一鍵。這是依 schema 推導，沒有實際執行。登錄為 ISSUE-034，以 F10 修正（1.3）。
   - **沒有問題**：其餘缺少 `merchant_id` 的唯一鍵分為三類，都不會跨商戶衝突。
     - 全域唯一的 UUID：`event_uuid`、`category_uuid`、`template_uuid`、`import_uuid`、`bank_txn_uuid`。
     - 由全域整數 ID 組成：`account_balance_snapshots(closing_id, account_id)`、`ledger_account_balance_snapshots(closing_id, ledger_id)`、`investment_movements(investment_id, event_id)`、`investment_positions(investment_id)`。
     - 刻意設計的全域資料：`exchange_rates` 由 A6 在 R9 處理；`users.username` 本來就是全域唯一。
   - 各 `*_ledgers` 表沒有 `merchant_id` 欄位，它們的唯一鍵也都是 UUID。
6. **SQL 種子直接寫入讀模型（ISSUE-035）。**
   - 每次啟動都會執行 `accounts.sql`、`sys_accounts.sql`（`bootstrap/db.go:43-44`，UPSERT）。
   - `bootstrapAccounts`（`bootstrap/event_seed.go:45`）只在 aggregate 還沒有事件時，才以資料表的內容建立 `account.bulk_imported`。
   - 因此首次啟動之後才加入種子檔的科目或系統科目代碼，會不經事件直接寫入商戶 1；`accounts.sql` 的 UPSERT 也會直接更新 `cash_flow_category`。
   - 重建商戶 1 時，這些資料會消失，下次啟動才又直接寫回，仍然沒有事件。
   - 不阻擋 R3：測試的 fixture 只執行一次種子。必須在 R10 以新機制取代 replay 之前處理。
7. **`snapshots`**：
   - 寫入端有兩個：engine（`bfs_engine.go:124`）與舊 `Append`。
   - 讀取端只有稽核列表（`repos/query/audit.go:111`）；`query.Event.GetSnapshot` 沒有呼叫點。
   - 舊的清空清單會刪除 snapshots，之後卻不會重建（`truncate.go:51`）。
   - 處理方式見 D7。
8. **其他被 payload 引用的整數 ID**（後續 Phase 的清單）：
   - `closing_id`：期間結帳與重開（`model/payload/period.go:17` 等），R5 處理。
   - 交易 ID：`model/payload/transaction.go:40`、`:90-92`、`:123-124`，R4 處理。
   - `bank_txn_id`、`matched_entry_id`、`created_txn_id`、`template_id`：`model/payload/bank_statement.go:109-171`，R6 處理。
   - 傳票、投資與銀行對帳引用的 `ledger_id`：D3 之後已經穩定。

   R3 只處理 R2 事件引用的 `ledger_id`；其他 ID 在各自的 Phase 依 D3 的規則處理（`TODO.md`）。

### 1.3 新增的 bug fix（不屬於 R3 的實作範圍）

| 項目 | 內容 | 依賴 |
|---|---|---|
| F10 | ISSUE-034：以 migration 把 `investments` 的唯一索引改為 `(merchant_id, symbol, currency)`；補測試重現「兩個商戶建立同一代號與幣別的投資」 | 無；R9 之前必須完成。不阻擋 R3：D5 會拒絕有投資事件的商戶，投資從 R9 起才納入重建 |

F10 依 `ISSUES.md` 以小幅修改實作與 review，實作計畫另行確認。ISSUE-033、ISSUE-032 維持獨立的 bug fix，不是 R3 的前置（D8）。ISSUE-035 在 R10 之前另行規劃（`TODO.md`）。

---

## 2. 設計決策

| # | 問題 | 建議 | 理由 |
|---|---|---|---|
| D1 | 清空範圍如何宣告 | `Projector` 新增 `Reset(ctx, tx *repos.DbTransaction) error`，商戶取自 ctx，與 checkpoint 相同（`persistence/repos/db_transaction.go:29`）。每個 projector 以 sqlc 的 `DELETE ... WHERE merchant_id = ?` 清除自己的資料表，查詢與方法放在 `persistence/repos/projection`；`accountProjector` 一併清除 `account_closure`。`Registry` 新增 `ResetAll` | 原則 4。依 `backend/CLAUDE.md` 使用 sqlc。不採用「projector 宣告資料表名稱，再由共用程式以 sqlx 組出 DELETE」：表名是動態的，屬於需要另行許可的 sqlx 例外；而且 trigger 維護的表（`account_closure`）與 R6 沒有 `merchant_id` 的 `*_ledgers` 都需要個別的條件。範圍是否完整由 D9 的測試保證，不依賴人工清單 |
| D2 | 重建入口 | `(*BFSEngine).Rebuild(ctx, merchantID int64)`，以 `merchantID` 覆寫 ctx 的商戶。使用 engine 的 UnitOfWork 與 projector registry，由 `command.NewEngine` 組裝（`application/command/engine.go`），與寫入使用同一組 projector。以 `domain/event` 的註冊表解碼。R3 不對外，到 R10 之前只有測試呼叫 | 上層計畫要求以 engine 方法提供。engine 可以依賴註冊表，反方向不行（ADR-027 規則 1；`persistence/handle` 已經依賴 `domain/event`）。共用 `NewEngine` 可以避免重建漏掉某個 projector |
| D3 | ISSUE-031：整數 `ledger_id` | **建立事件記錄整數 ID**：(1) engine 允許 handler 回傳它所處理事件的補充後 payload（`EventResult.Payload`），型別必須與原本的 payload 相同，在寫入前取代；子事件不受影響。(2) `ledger_account.created` 的 handler 以交易內查詢配置 `ledger_id`，值為全部商戶的 `MAX(ledger_id) + 1`，與目前 `INTEGER PRIMARY KEY`（沒有 AUTOINCREMENT）的 rowid 配置方式相同；寫入 payload 的新欄位 `ledger_id`。只有 BFS 型別有這個欄位，前端不修改，送來的值一律覆寫。(3) projector 以明確的 `ledger_id` 新增，值必須大於 0。使用 BFS 專用的新查詢 `CreateLedgerAccountWithID`；舊 projection 共用的 `CreateLedgerAccount`（`database/queries/account.sql:105`）與它的 mapper 不修改，因為舊 projection 不會設定 ID（`services/projection/account.go:107`），共用查詢若加入 `ledger_id` 就會寫入 0。重建時以相同 ID 寫回；ID 是全域主鍵，不會與其他商戶衝突。**後續 Phase 沿用的規則**：被其他事件或直接寫入的資料引用的整數 ID，由 handler 在交易內配置，並記錄在建立事件中 | 原則 1 要求 handler 把 projector 需要的資料寫進事件，R4 的交易 ID 也需要同一個機制。配置在 `Run` 的交易內，單一連線會讓寫入依序執行，不會競爭。不採用的方案：(b) 清空前先讀出 `ledger_uuid` → `ledger_id` 的對照，重建時沿用：projector 在重建與寫入時的行為不同，而且 R10 匯入時沒有舊的讀模型可以對照。(c) 事件改以 `ledger_uuid` 引用：前端 24 個檔案與交易範本都要改，而且 `ledger_id` 仍會在重建後改變，交易範本會指到錯的帳戶（1.2 第 1 點） |
| D4 | 交易與事件順序 | 在一個 `uow.Do` 內完成。以交易內查詢 `tx.Query.Event.Replay(ctx, 0, nil)`（`repos/query/event.go:161`）讀取商戶的全部事件，依 `event_id` 排序（`database/queries/event_store.sql:59`）。開頭執行 `PRAGMA defer_foreign_keys = ON`（同 `truncate.go:63`）：外鍵目前沒有開啟（ISSUE-032），這行沒有效果；開啟後清空的順序也不受影響，外鍵在 commit 時才檢查。依 `event_id` 順序套用 projector | ADR-025 A4、ISSUE-022。同一次 `Run` 內，BFS 依序寫入根事件與子事件，`event_id` 的順序就是因果順序。事件一次全部讀進記憶體，以目前的資料量可以接受（第 6 節） |
| D5 | 前置檢查（清空之前） | 在交易內、刪除任何資料之前，逐一檢查商戶的事件：(1) 事件類型已登錄（`Lookup`）；(2) payload 可以解碼；(3) 以解碼後的事件呼叫 `Spec.Aggregate`，結果等於儲存的 `(aggregate_type, aggregate_id)`；(4) `ledger_account.created` 有 `ledger_id`，且大於 0（D3）。任一項不符就回傳錯誤，指出 `event_id` 與原因，不刪除任何資料 | 原則 8：新機制只處理 BFS 事件。(1) 排除有舊路徑事件的商戶，1.2 第 2 點的共用資料表因此安全。(2)、(3) 排除舊路徑寫入的已登錄類型事件（1.2 第 3 點），讓重算的 `aggregate_versions` 正確。(4) 排除 D3 實作之前寫入的事件。R10 之前，任何有傳票的商戶都會被拒絕，這符合「尚不對外」；R10 會清空開發 DB（S2） |
| D6 | `aggregate_versions` 與 checkpoint | 刪除商戶的 `aggregate_versions`，再依 `event_store` 寫入每個 aggregate 的最大 `aggregate_version`（以 sqlc 的 `INSERT ... SELECT` 實作）。checkpoint：套用事件前刪除商戶 6 個 BFS projector 名稱的列；`ApplyAll` 的 upsert（`persistence/handle/projector.go:46`）會把每個 projector 寫回它最後套用的 `event_id`，與寫入時的結果相同。舊 projection 的 checkpoint 名稱不同（`enums.ProjectionType*`），不受影響 | 上層計畫 R3 的要求；舊 replay 也有同樣的步驟，但分在不同交易（`services/event.go:242`）。D5 通過後，每個版本紀錄都有對應的事件，重算的結果與原本相同。R10 匯入時沒有版本紀錄，必須依事件重算 |
| D7 | `snapshots` | 不清除、不重建 | snapshots 在事件端，內容是每 50 個版本時的事件 JSON（`bfs_engine.go:124`），不是讀模型；重建不改動 `event_store`，除了稽核列表也沒有其他讀取端（1.2 第 7 點）。重新產生的內容與寫入時不一定相同，沒有好處。驗證時列入範圍外，確認重建前後不變 |
| D8 | 外鍵檢查與 ISSUE-033 | R3 不等 ISSUE-033 修正。重建測試的孤兒檢查分兩部分：(1) 範圍內外鍵定義正確的資料表（`ledger_account_type_config`），執行 `PRAGMA foreign_key_check`；(2) 定義錯誤的 3 個引用（`accounts.parent_id`、`ledger_accounts.account_id`、`sys_accounts.account_id`）與沒有宣告外鍵的引用（`account_closure` 引用 `accounts`），改以帶 `merchant_id` 條件的 `LEFT JOIN` 檢查（`doc/NOTES.md` 資料庫第 6 點）。範圍內其他以科目編號引用 `accounts`、沒有宣告外鍵的欄位，實作時依 schema 列出，一併檢查。ISSUE-033、ISSUE-032 仍是獨立的 bug fix（033 先於 032） | 修正 ISSUE-033 要重建 7 張表，其中 `journal_entries`、`investments` 屬於 R4、R9 的範圍，不宜在 R3 之前處理。帶 `merchant_id` 的 `LEFT JOIN` 比定義錯誤的外鍵更嚴格。上層計畫第 8 節的「執行 `PRAGMA foreign_key_check`」在實作時補上這個做法 |
| D9 | 如何驗證清空範圍完整 | 以 BFS 寫入前後的差異找出清空範圍。步驟：(1) 以 SQL 為商戶 2（沒有種子）寫入前置資料，也就是一筆 ACTIVE 的 `transactions` 與它的兩筆 `journal_entries`，做法同 R1 的 `persistence/handle/decode_apply_test.go`；不寫入 `entry_cf_categories`，那是 cf projector 的資料表。(2) 從 `sqlite_master` 列出所有有 `merchant_id` 欄位的表，逐表匯出商戶 2 的完整資料列，作為基準。(3) 經 BFS 寫入每種已登錄事件的 fixture，再匯出一次。除 `event_store`、`aggregate_versions`、`projection_checkpoints`、`snapshots` 之外，資料有差異的表就是清空範圍。(4) 驗證：清空範圍內的表在基準中必須沒有商戶 2 的資料，也就是前置資料沒有寫到 projector 的表；`ResetAll` 後，清空範圍內沒有商戶 2 的資料；其他資料表（包含前置的交易與分錄）與基準相同；商戶 1 的資料不變。fixture 若漏掉 `domainevent.Specs()` 中的任一事件類型，測試直接失敗 | ISSUE-021 的要求：新增 projector 或資料表卻沒有清空時，不必靠人工清單也能攔下。`transaction.cf_category_updated` 的 handler 必須查到有效的交易與分錄（`process/cashflow.go:30-55`）。這組前置資料若經舊路徑建立，會被 D5 拒絕；若以 SQL 建立，又不能當作清空範圍。因此以寫入前後的差異判斷，並驗證前置資料不變。限制：沒有 `merchant_id` 的 `*_ledgers` 找不到，R6 改依上層資料表擴充這項檢查 |
| D10 | canonical 比較的規格 | 逐表選取欄位並依業務鍵或 UUID 排序。**排除**：AUTOINCREMENT 的代理鍵（`prepaid_categories.id`、`fixed_asset_categories.id`、`period_closings.closing_id`、`ledger_account_type_config.id`、`asset_type_account_config.id`），以及寫入當下產生的時間欄位（`updated_at`、`created_at`）。**包含**：`ledger_id`（D3 之後是決定性的，也是 ISSUE-031 的驗收）、`updated_by`、`version`。另外比較 `aggregate_versions`，以及 BFS projector 的 checkpoint（排除 `updated_at`） | 上層計畫第 8 節 |

---

## 3. 分段

每段連同自己的測試完成，依 1 → 3 各自 review、commit；開始實作時在 `doc/STEP.md` 加上 R3 子階段。三段修改的檔案都超過 5 個，各自再分為三個 phase，每個 phase 完成後執行 `go build ./...`。

### 段 1：建立事件記錄 `ledger_id`（D3）

- **1a**：`kernel/result/result.go` 的 `EventResult` 新增 `Payload`；`runtime/engine/bfs_engine.go` 的 `terminal` 在寫入前檢查型別並取代 payload。
- **1b**：
  - `domain/event/account.go` 的 `LedgerAccountCreatePayload` 加上 `ledger_id`。
  - 新增唯讀查詢 `NextLedgerID`：`database/queries/`、`repos/query/`。只有讀取，`tx.go` 的寫入清單不需要修改。
  - `process` 新增 `ledger_account.created` 的 handler，取代 `registry.go:15` 的單純驗證。
  - 新增 BFS 專用查詢 `CreateLedgerAccountWithID`，寫入明確的 `ledger_id`：
    - `database/queries/account.sql`；
    - `model/db/projection/account.go` 的 `LedgerAccount` 加上 `//dbmap:sqlcdb=CreateLedgerAccountWithIDParams`；
    - `persistence/repos/projection/account.go` 改用新查詢。
  - 原本的 `CreateLedgerAccount`（`:105`）與舊路徑的 `projection_repo/account.go` 不修改。
  - `persistence/handle/account.go`。
- **1c**：調整 R2 的契約測試 `TestAccountPayloads_LegacyJSON`（`domain/event/account_test.go:58`）；新增測試情境。

測試情境：

| # | 情境 | 預期 |
|---|---|---|
| 1 | handler 回傳補充後的 payload | `event_store` 與 projector 看到的都是補充後的 payload；子事件不受影響 |
| 2 | 補充後的 payload 型別與原本不同 | 錯誤，整次 `Run` 回復 |
| 3 | handler 沒有補充 payload | 既有的 engine 情境全數通過，行為不變 |
| 4 | 通用 API 送出 `ledger_account.created` | 事件 payload 的 `ledger_id` 等於讀模型的 `ledger_id`；`ledger_uuid` 為事件 UUID |
| 5 | 商戶 2 已有 `ledger_id` 2，商戶 1 建立明細帳戶 | 商戶 1 取得 3，與目前的 rowid 行為相同 |
| 6 | 送出的 payload 帶 `ledger_id: 99` | 被覆寫為配置的值 |
| 7 | 新舊 payload 的 JSON | BFS 型別只多出 `ledger_id`；舊型別可以解碼 BFS 寫入的 JSON |
| 8 | 商戶 2 經 BFS 建立科目與兩個明細帳戶後，以舊 `EventStoreService.Replay(ctx, 0, nil)` 全量重播 | 成功；兩個明細帳戶都存在，`ledger_uuid` 各自對應建立事件的 UUID。確認舊 replay 的行為不變（第 4 節） |

反向檢查：
- `terminal` 忽略補充後的 payload → 情境 1、4 失敗。
- handler 不配置 ID，改由資料庫自動配置 → 情境 4 失敗。
- 改為修改共用的 `CreateLedgerAccount` 加入 `ledger_id`，並重新 `go generate` → 情境 8 失敗（第二個明細帳戶撞主鍵）。

### 段 2：`Reset`（D1、D9）

- **2a**：各資料表的 `DELETE ... WHERE merchant_id = ?` 查詢（`database/queries/`）與 `persistence/repos/projection` 的方法，以及刪除指定名稱 checkpoint 的查詢。這個 phase 不修改介面。
- **2b**：
  - `persistence/handle/projector.go`：`Projector` 新增 `Reset`，`Registry` 新增 `ResetAll`。
  - 6 個 projector 實作 `Reset`。
  - 實作 `Projector` 的測試替身一併補上 `Reset`。
- **2c**：新增測試情境。

測試情境：

| # | 情境 | 預期 |
|---|---|---|
| 1 | D9：以 SQL 寫入前置交易與分錄、取得基準後，經 BFS 為商戶 2 寫入涵蓋所有已登錄事件類型的 fixture，再執行 `ResetAll` | BFS 寫入前後有差異的資料表，在基準中沒有商戶 2 的資料；`ResetAll` 後，這些表中商戶 2 的資料全部清空 |
| 2 | 同上 | 前置的 `transactions`、`journal_entries` 與其他資料表都與基準相同；商戶 1（有種子）的資料不變 |
| 3 | 連續執行 `ResetAll` 兩次 | 不出錯 |

反向檢查：
- 從 `Reset` 移除 `account_closure` → 情境 1 失敗。
- 移除某個 projector 的一條刪除 → 情境 1 失敗。
- 拿掉 `merchant_id` 條件 → 情境 2 失敗。
- 在某個 `Reset` 加入刪除商戶的 `journal_entries` → 情境 2 失敗。

### 段 3：`Rebuild`（D2、D4～D8、D10）

- **3a**：
  - `aggregate_versions` 的查詢：依商戶刪除；從 `event_store` 寫入最大版本。
- **3b**：
  - `runtime/engine` 的 `Rebuild`：前置檢查、`defer_foreign_keys`、`ResetAll`（段 2 實作時已包含清除 checkpoint）、依序套用、重算版本。
  - 把 `event_store` 的資料列轉回 kernel 事件，沿用 R1 測試的轉換方式（`persistence/handle/decode_apply_test.go`），改為正式的程式。
  - 依需要新增錯誤代碼。
- **3c**：
  - 測試共用的 canonical 匯出與孤兒檢查。
  - 新增測試情境。

測試情境（R3 的驗收）。fixture 都是手動撰寫、可追溯的 payload：
- 商戶 1 以真實的種子檔執行 `BootstrapEventSeeds`。
- 商戶 2 以 `account.created` 等事件建立自己的科目。
- 前置資料：兩個商戶各以 SQL 寫入一筆 ACTIVE 的 `transactions` 與兩筆 `journal_entries`（同 D9），供 `transaction.cf_category_updated` 的 handler 查詢。前置資料不是事件，D5 不會因此拒絕；實際環境中，交易與分錄只會來自舊路徑的事件，那時 D5 會拒絕。
- 兩個商戶都經 `command.Build` 與 `Writer.Append` 寫入全部 17 種事件，包含明細帳戶的修改。
- 商戶 2 的明細帳戶建立在商戶 1 之後，讓商戶 2 持有較大的 `ledger_id`（ISSUE-031 的回歸）。

| # | 情境 | 預期 |
|---|---|---|
| 1 | 匯出 canonical → `Rebuild(1)` → 再次匯出；商戶 2 同樣執行一次 | 兩次匯出相同（D10），包含 `ledger_id`、`aggregate_versions`、checkpoint |
| 2 | 重建商戶 1 | 商戶 2 的 canonical 匯出與原始資料列不變；範圍外資料不變：`event_store`、`snapshots`、`exchange_rates`、`merchants`、`users`、前置的 `transactions` 與 `journal_entries`，以及引用商戶 1 `ledger_id` 的交易範本（仍對應同一個 `ledger_uuid`） |
| 3 | 測試組裝的 engine 多一個在第 N 個事件失敗的 projector | `Rebuild` 回傳錯誤；範圍內資料表、`aggregate_versions`、checkpoint 都與重建前相同（ISSUE-022） |
| 4a | 商戶有未登錄類型的事件（經 `EventStoreService.Append` 寫入的 `investment.created`） | 拒絕，資料不變 |
| 4b | 已登錄類型的事件，但 aggregate 與推導結果不符（以 SQL 寫入的舊式類別建立事件） | 拒絕，資料不變 |
| 4c | `ledger_account.created` 沒有 `ledger_id`（以 SQL 寫入） | 拒絕，資料不變 |
| 5 | 重建前刪除或竄改商戶的 `aggregate_versions` | 重建後與事件一致；之後以讀模型的版本經 API 修改科目，成功 |
| 6 | 重建後的孤兒檢查（D8） | 沒有孤兒引用 |

反向檢查：
- 不重算 `aggregate_versions` → 情境 5 失敗。
- 拿掉前置檢查 → 情境 4 失敗。
- 清空後先 commit 再套用 → 情境 3 失敗。
- `ledger_id` 改回由資料庫配置 → 情境 1、2 失敗。
- 拿掉 `merchant_id` 條件 → 情境 2 失敗。

---

## 4. 預期的行為變更

1. `ledger_account.created` 的 payload 新增 `ledger_id`，由 handler 在交易內配置，配置的值與目前的 rowid 規則相同。通用 API 回應中的事件 payload 也會多出這個欄位；前端送出的請求不變（前端 ADR-013），不需要修改。舊 Replay（Wontfix）會忽略這個欄位。
2. engine：handler 可以回傳它所處理事件的補充後 payload（內部 API）。
3. `Projector` 新增 `Reset`，`Registry` 新增 `ResetAll`（內部 API）。
4. 新增 `BFSEngine.Rebuild`，不對外。`POST /api/audit/replay`、`POST /api/audit/event/import` 到 R10 之前都維持舊機制。

與上層計畫的差異：上層計畫 5.3 的 R3 把「預期的行為變更」列為「無」。第 1 點來自 D3，經 review 確認後，實作時更新上層計畫。

維持不變：
- 其他 16 種事件的 payload。
- 所有 API 的請求與回應，只有第 1 點的事件 payload 例外。
- 版本策略與 aggregate 身分。
- 舊路徑與舊 replay，包含舊 projection 使用的 `CreateLedgerAccount` 查詢與 mapper（D3、段 1 情境 8）。

---

## 5. 驗證與文件

**驗證**：每段依 `doc/NOTES.md` 測試第 3 點的 Docker 流程（900 MB 記憶體上限）執行；段 1、段 2 使用 `golang:1.26.9`，2026-10-27 升級後使用 `golang:1.27.3`：
- `go generate ./...`、`go build ./...`
- 修改到的套件執行 `go vet`
- `go test -vet=off ./...`
- 修改與新增的檔案執行 `gofmt -l`
- `git diff --check`

反向檢查在工作區副本中修改，不留在 repo。

**文件**：
- **本次（計畫書）**：
  - `doc/ISSUES.md`：登錄 ISSUE-034、ISSUE-035；ISSUE-031 的解決方向指向 D3；ISSUE-033 補上 1.2 第 4 點的影響。
  - 上層計畫 5.2、`doc/STEP.md`：加上 F10；R3 指向本計畫書。
  - `doc/TODO.md`：
    - 唯一鍵盤點改為完成，並記錄結果。
    - 新增 F10。
    - 新增 ISSUE-035。
    - 新增「後續 Phase 依 D3 處理整數 ID」。
- **實作時**：
  - ADR：D3 的規則建議另立 ADR-028；D1、D5、D7 補進 ADR-025。
  - 上層計畫第 8 節補上 D8 的孤兒檢查做法。
  - `doc/STEP.md` 加上 R3 子階段。
  - R3 完成後，關閉 ISSUE-031。

---

## 6. 風險與範圍外

- **重建期間 DB 被鎖住**：ADR-025 已知的後果。另外 D4 會一次讀入商戶的全部事件；資料量變大時，改為分批讀取。
- **R10 之前，多數商戶無法重建**：D5 會拒絕有舊事件的商戶，這是設計上的結果（原則 8）。R10 前不以新機制重建開發資料。
- **整數 ID**：`closing_id`、交易 ID 與銀行對帳的 ID（1.2 第 8 點）由 R4、R5、R6 依 D3 的規則處理；在那之前，重建後這些 ID 可能改變。
- **範圍外**：
  - 舊 `Replay` 與事件匯入：R10 處理。
  - `*_ledgers` 的清空條件與 D9 的擴充：R6 處理。
  - `exchange_rates`：R9 處理（A6）。
  - ISSUE-035：R10 之前處理。
  - ISSUE-033、ISSUE-032：獨立的 bug fix。
  - F10：R9 之前完成。
