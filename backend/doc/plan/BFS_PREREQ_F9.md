# F9 實作計畫書：期間的唯一鍵加入商戶

| 項目 | 內容 |
|---|---|
| 狀態 | ⏳ 待 review |
| 建立日期 | 2026-10-22 |
| 分支 | `refactor`（基準 commit `24a94bf`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md`：5.2 的 F9 |
| 依賴 | 無；建議在 R2 段 5（期間開帳遷移）之前完成，R3 之前必須完成 |
| 關聯 | ISSUE-030、ISSUE-032、ISSUE-033；ADR-025 A4；`doc/plan/BFS_R2_MASTER_DATA.md` 1.2 |

> F9 是 bug fix，依 `AGENTS.md`「變更幅度與重構」以小幅修改另行實作、另行 review。
> 文中行號以基準 commit 為準。

---

## 1. 問題

`period_closings` 的唯一鍵是 `UNIQUE(period_type, period_start)`（`internal/database/migrations/20260310001_init_table.sql:215`），之後加入 `merchant_id`（`20260514002_add_merchant_id.sql:16`）時沒有更新。一個商戶開立某期間後，其他商戶就無法開立同一期間；R2 計畫書盤點時已以探測測試重現（ISSUE-030）。

## 2. 現況盤點

1. **資料表定義**：以 migration 依序建立的欄位為 `closing_id`（`INTEGER PRIMARY KEY AUTOINCREMENT`）、`period_type`、`period_start`、`period_end`、`status`、`opening_txn_id`、`closing_txn_id`、`snapshot`、`closed_at`、`note`、`reopen_at`、`reopen_reason`（`20260310001`），`merchant_id`（`20260514002`），`updated_by`、`updated_at`（`20260519001_add_audit_fields.sql:47-48`），`closing_uuid`（`20260529002_add_uuid_to_projections.sql:79`）；4 個 CHECK；索引 `idx_period_closings_status`、`idx_period_closings_end`（`20260310001:218-219`）、`idx_period_closings_merchant`（`20260514002:26`）。實作時以遷移到前一版的 DB 中 `sqlite_master` 的定義為準。
2. **引用它的資料表**：`account_balance_snapshots`、`ledger_account_balance_snapshots` 的 `closing_id REFERENCES period_closings(closing_id)`（`20260508001:4`、`20260511001:4`），沒有 `ON DELETE`。沒有 view 或 trigger 引用 `period_closings`。另外，`account_balance_snapshots` 的另一個外鍵 `account_id REFERENCES accounts(account_id)`（`20260508001:5`）定義錯誤：`accounts` 的主鍵是 `(account_id, merchant_id)`，單獨的 `account_id` 並不唯一。因此不論外鍵是否開啟，對這張表執行 `PRAGMA foreign_key_check` 都會回傳 `foreign key mismatch`。同類外鍵共 7 個，登錄為 ISSUE-033，不在 F9 修正。`ledger_account_balance_snapshots` 的 `ledger_id` 外鍵沒有這個問題。
3. **查詢**：`database/queries/period.sql`、`account_analysis.sql`、`account_balance_snapshot.sql`、`ledger_account_balance_snapshot.sql` 與 `repos/unit_of_work/event_store/truncate.go:44`，每個查詢都以 `merchant_id` 或主鍵 `closing_id` 篩選，沒有 `ON CONFLICT`；沒有查詢依賴「跨商戶唯一」。
4. **sqlc**：schema 來自另行維護的 `internal/database/schema.sql`（`sqlc.yaml:5`），其中的 `period_closings` 沒有列出唯一鍵，不需修改；sqlc 產生的程式不受影響。
5. **外鍵實際上沒有開啟**：撰寫本計畫時發現，`bootstrap/config.go:45` 的預設 `DB_PARAM` 使用 go-sqlite3 的參數格式（`_foreign_keys=on` 等），modernc 驅動會忽略；探測結果為 `foreign_keys=0`、`busy_timeout=0`、`journal_mode=delete`，只有 `_pragma=foreign_keys(1)` 的格式會生效。登錄為 ISSUE-032，不在 F9 修正；但 F9 的 migration 必須在外鍵開啟或關閉時都正確，避免 ISSUE-032 修正後失效。

## 3. 設計

| # | 問題 | 建議 | 理由 |
|---|---|---|---|
| D1 | 如何更換唯一鍵 | 新增 migration 重建資料表，順序為：備份到暫存表 → `DROP TABLE period_closings` → 以相同名稱重建 → 寫回原資料 → 刪除暫存表 → 重建 3 個索引。開頭執行 `PRAGMA defer_foreign_keys = ON` | SQLite 無法移除資料表內宣告的 UNIQUE（自動索引不能 DROP），只能重建。不改名，子資料表外鍵寫的名稱 `period_closings` 始終不變，不受 `foreign_keys`、`legacy_alter_table` 的影響。外鍵開啟時，`DROP` 造成的違反會延後到 commit，寫回原本的 `closing_id` 後就消除（延後檢查只計算本交易造成的違反，不檢查既有資料，見第 9 節）；`defer_foreign_keys` 可以在 goose 的交易內設定，commit 後自動還原；外鍵關閉時這行沒有作用 |
| D2 | 新的定義 | 欄位、順序、預設值與 4 個 CHECK 與現況相同，唯一鍵改為 `UNIQUE(merchant_id, period_type, period_start)` | 只改唯一鍵，其他行為不變 |
| D3 | `closing_id` 與 AUTOINCREMENT 計數 | 寫回時保留原本的 `closing_id`；`sqlite_sequence` 中 `period_closings` 的計數先備份，寫回資料後還原為兩者中較大的值 | 快照表以 `closing_id` 引用期間；`DROP TABLE` 會刪除計數，若只依現有最大 ID 重算，曾被刪除的 ID 可能被重新使用（ISSUE-018 的同類問題） |
| D4 | Down | 不還原，只留註解（同 `20261021001_account_initial_version.sql` 的做法） | 還原舊唯一鍵會重新引入 bug，且兩個商戶已開立同一期間時會失敗 |
| D5 | migration 的測試方式 | `internal/database/migrate.go` 新增 `MigrateUpTo(ctx, db, version)`，與 `RunMigrations` 使用相同的內嵌檔案與 goose 設定；測試先遷移到 F9 的前一版、寫入 fixture，再遷移到最新版本 | 目前的測試 DB 一次執行所有 migration，無法驗證 migration 對既有資料的效果。F8 與 R2 段 4 的資料 migration 沿用同一個函式 |

撰寫本計畫時，以容器內的一次性探測測試比較兩種重建順序（父表含 AUTOINCREMENT 且刪除過一筆、子表有外鍵引用，各在外鍵開啟與關閉時執行）：

| 做法 | 外鍵關閉 | 外鍵開啟 |
|---|---|---|
| D1：備份 → DROP → 同名重建 → 寫回，並還原計數 | commit 成功；計數保留，下一筆 ID 跳過已刪除的 ID | commit 成功；`foreign_key_check` 沒有結果 |
| 建立 `_new` → 複製 → DROP → 改名 | commit 成功，但計數退回現有最大 ID，下一筆**重新使用已刪除的 ID** | commit 失敗（`FOREIGN KEY constraint failed`），整個交易回復 |

review 後，以實際 migration 再次探測 D1 的草稿 SQL。做法是遷移到 `20261021001`，寫入兩個商戶的期間與兩張快照表的資料，再經 goose 執行草稿，外鍵開啟與關閉各一次。結果如下：

- **兩種情況都 commit 成功**，結果相同：`closing_id` 與計數保留，下一筆 ID 大於原計數；唯一鍵為 `(merchant_id, period_type, period_start)`，3 個索引存在；兩張快照表的 `closing_id → period_closings(closing_id)` 外鍵定義保留；另一個商戶可以開立同一期間，同一商戶重複開立回傳唯一鍵錯誤。
- **預先寫入的孤兒快照**（引用不存在的 `closing_id`）在兩種情況下都保留不變。

替代方案：

- 沿用既有 migration 的「建立 `_new` → 複製 → DROP 舊表 → 改名」（例如 `20260614001_bank_pdf_multiledger.sql`）：如上表，外鍵開啟時 commit 失敗；外鍵關閉時也會失去 AUTOINCREMENT 計數。不採用。
- goose 的 `NO TRANSACTION` 搭配 `PRAGMA foreign_keys = OFF`：失去交易保護，中途失敗時資料表可能只剩一半。不採用。

## 4. 修改內容

| 檔案 | 修改 |
|---|---|
| `internal/database/migrations/20261022001_period_closing_merchant_unique.sql`（新增） | D1～D4 |
| `internal/database/migrate.go` | `MigrateUpTo`（D5） |
| `internal/database/period_closing_migration_test.go`（新增） | 5.2 的 migration 情境 |
| `internal/services/period_unique_test.go`（新增） | 5.1 的寫入情境 |

程式 4 個檔案，未達 `backend/CLAUDE.md` 的分段標準，一段完成。

## 5. 測試情境

fixture 以 SQL 明確寫入，科目來自 `seeds/accounts.sql`，商戶 1、2；不使用產生的假資料。

### 5.1 寫入（`TestPeriodStarted_PerMerchant`，經 `EventStoreService.Append`，封套同前端 `('ACCOUNT', '')`）

| # | 情境 | 預期 |
|---|---|---|
| 1 | 商戶 1、商戶 2 各自開立 2026-10 | 都成功，各有一筆 OPEN 的月期間 |
| 2 | 商戶 1、商戶 2 各自開立 2026 年度 | 都成功，各有一筆 OPEN 的年度期間 |
| 3 | 同一商戶重複開立 2026-10 | 第二次回傳唯一鍵錯誤；`event_store`、`aggregate_versions`、`period_closings` 都維持第一次後的狀態 |

### 5.2 migration（`TestPeriodClosingMerchantUnique_Migration`）

fixture：遷移到前一版（`20261021001`）後，在外鍵關閉時寫入以下資料。外鍵開啟時，寫入 `account_balance_snapshots` 會因 ISSUE-033 失敗，所以 fixture 一律在關閉時寫入。

- 兩個商戶的期間，狀態包含 OPEN、CLOSED（含 `closed_at`、`snapshot`）、REOPENED。
- 兩張快照表各一筆，引用 CLOSED 的期間；沒有孤兒引用。
- `sqlite_sequence` 的計數設為大於最大 `closing_id` 的值。

快照表的引用完整性以 `LEFT JOIN period_closings` 查詢找不到期間的快照，不使用 `PRAGMA foreign_key_check`：`account_balance_snapshots` 會因 ISSUE-033 直接報錯，`ledger_account_balance_snapshots` 則會連同 `ledger_id` 的引用一起檢查，超出 F9 的範圍。

| # | 情境 | 預期 |
|---|---|---|
| 4 | 遷移到最新版本（外鍵關閉，與目前的執行環境相同） | 每筆期間的所有欄位與遷移前相同（含 `closing_id`）；兩張快照表的資料不變，每筆快照的 `closing_id` 都找得到期間；`pragma_foreign_key_list` 仍有 `closing_id → period_closings(closing_id)`；3 個索引存在；唯一鍵的欄位為 `(merchant_id, period_type, period_start)`；下一筆新增的 `closing_id` 大於原本的計數 |
| 5 | 同情境 4，但在執行 F9 前開啟外鍵（`PRAGMA foreign_keys = ON`，同一條連線、交易外設定） | migration 成功，結果同情境 4 |

## 6. 預期的行為變更

只有一項：不同商戶可以開立同一期間。

維持不變：同一商戶重複開立同一期間仍回傳錯誤；期間的資料、`closing_id` 與快照的引用；所有查詢與 API；sqlc 產生的程式。

## 7. 驗證

依 `doc/NOTES.md` 第 3 點的 Docker 流程（`golang:1.26.9`、900 MB 記憶體上限）：

1. `go generate ./...`、`go build ./...`；`internal/database`、`internal/services` 執行 `go vet`。
2. `go test -vet=off ./...` 全數通過。
3. 新增與修改的檔案 `gofmt -l` 無輸出；`git diff --check` 無輸出。
4. 反向確認（在工作區副本中修改，不留在 repo）：
   - 唯一鍵改回 `(period_type, period_start)`：情境 1、2 失敗。
   - 移除唯一鍵：情境 3 失敗。
   - 移除 `defer_foreign_keys`：情境 5 失敗。
   - 不還原 `sqlite_sequence`：情境 4 的計數斷言失敗。

## 8. 文件

- 本次（計畫書）：ISSUE-032、ISSUE-033 登錄於 `doc/ISSUES.md`，`doc/TODO.md` 的「伺服器設定」加上對應待辦。
- 完成後：ISSUE-030 改為 Resolved；上層計畫 5.2 與 `doc/STEP.md` 的 F9 改為 done；`doc/TODO.md` 勾選 F9；`doc/NOTES.md` 補上「SQLite 重建有外鍵引用的資料表」的做法（D1、D3）與 modernc 的 DSN 參數格式。

## 9. 風險與範圍外

- **既有開發 DB**：migration 在啟動時自動執行。
  - **唯一鍵**：既有資料不會違反新唯一鍵，因為新鍵比舊鍵寬鬆。
  - **孤兒引用**：外鍵一直沒有開啟，快照表可能已有引用不存在期間的資料。F9 不檢查、不修復這類資料：不論外鍵開啟或關閉，migration 都會 commit 成功，孤兒引用保留不變（第 3 節的探測已確認）。原因是延後的外鍵檢查只計算本交易的語句造成的違反，開啟外鍵也不會追溯檢查既有資料。所以 migration 成功不能視為既有資料完整的證明。既有孤兒資料由 ISSUE-032 開啟外鍵前的檢查處理。
- **範圍外**：
  - ISSUE-032 的修正：改用 `_pragma` 格式；開啟外鍵前需先完成 ISSUE-033，並檢查既有資料與測試。
  - ISSUE-033：引用 `accounts(account_id)` 的 7 個外鍵定義錯誤。
  - 其他資料表的唯一鍵是否缺少 `merchant_id`（`doc/TODO.md` 的盤點項目）。
  - R2 段 5 的期間開帳遷移。
