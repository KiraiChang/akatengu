# F8 實作計畫書：類別的讀模型版本與 aggregate 版本一致

| 項目 | 內容 |
|---|---|
| 狀態 | ⏳ 待 review |
| 建立日期 | 2026-10-22 |
| 分支 | `refactor`（基準 commit `cec8f2c`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md`：5.2 的 F8 |
| 依賴 | 無；需在 R2 段 4（類別遷移）之前完成 |
| 關聯 | ISSUE-029；ADR-027 規則 6；`doc/plan/BFS_R2_MASTER_DATA.md` 的 1.2、D6、段 4 |

> F8 是 bug fix，依 `AGENTS.md`「變更幅度與重構」以小幅修改另行實作、另行 review。
> 文中行號以基準 commit 為準。

---

## 1. 問題

預付與固定資產類別建立後，只能修改或停用一次，之後一律失敗（ISSUE-029）。兩個原因疊加：

1. 修改與停用的 SQL 從不更新讀模型的 `version`，畫面一律送出 1。
2. 舊路徑在 aggregate 沒有版本紀錄時，以「送出版本 + 1」新建紀錄。第一次修改後，`category_uuid` 的 aggregate 是版本 2；第二次仍送 1，比對不符，撞上 `aggregate_versions` 的唯一鍵。

建立事件寫到另一個 aggregate（handler 隨機產生的 UUID），屬於 A7 的身分問題，由 R2 段 4 處理（R2 計畫書 D6），不在 F8 的範圍。

## 2. 現況盤點

1. **寫入流程**：兩種類別的專屬 handler 結構相同（`handler/prepaid_category.go`、`handler/fixed_asset_category.go`）。
   - 建立：aggregate 是 handler 產生的 UUID v7（`:52`），送出版本 0。
   - 修改、停用：aggregate 是路徑上的 `category_uuid`，送出 body 中的 `expected_version`（`UpdateCategory` `:81`、`DeleteCategory` `:123`）。
2. **讀模型**：兩張表的 `version` 是 `INTEGER NOT NULL DEFAULT 1`（`migrations/20260602001_asset_prepaid_categories.sql:15`、`:29`）。修改與停用的 SQL 沒有更新 `version`（`queries/prepaid_category.sql:12-26`、`queries/fixed_asset_category.sql:12-27`）。projection 只呼叫這 4 個查詢，Go 程式不需修改。
3. **舊路徑的版本規則**（`services/event.go:82-93`）：先以 `UpdateVersionIfMatch`（`queries/event_store.sql:102-109`）比對並遞增版本。沒有更新到任何列時，一律以「送出版本 + 1」新建紀錄，結果分兩種：
   - 沒有紀錄：新建成功。
   - 有紀錄但版本不符：撞上主鍵，整個交易回復。

   因此只要讀模型版本與 aggregate 版本一致，畫面送出過期版本時就會被拒絕。
4. **pipeline**：修改與停用前，會確認類別存在且啟用（`factory/prepaid_category.go:35`、`:66`，固定資產類別相同），不受 F8 影響。
5. **前端**：
   - 修改與停用時，送出讀模型的 `version`（`PrepaidCategoryConfig.svelte:117`、`:143`；`FixedAssetCategoryConfig.svelte:122`、`:149`）。
   - 每次寫入後重新載入列表（同檔 `:123`、`:145`；`:129`、`:151`）。

   所以修正後，畫面會自動送出遞增後的版本。API 的 JSON 已有 `version` 欄位，前端不需修改。
6. **其他讀模型的慣例**：修改時一律 `version = version + 1`（例如 `queries/account.sql:102`、`asset.sql:35`、`config.sql:15`）。
7. **sqlc**：只在 SET 子句加上 `version`，沒有新增參數；`schema.sql` 不需修改。產生的程式由 `go generate` 更新，且 `internal/database/sqlcdb` 不在版控中。
8. **既有測試**：`services/category_test.go`（Ginkgo）修改、停用時送出版本 0，也沒有檢查 `version`，不受影響。

撰寫本計畫時，以容器內的一次性探測測試，經 `EventStoreService.Append` 比較現行程式與草稿修正（兩種類別結果相同）：

| 操作（送出讀模型的版本） | 現行程式 | 草稿修正（D1、D2） |
|---|---|---|
| 第一次修改 | 成功；讀模型 1、aggregate 2 | 成功；兩者都是 2 |
| 第二次修改 | 唯一鍵錯誤 | 成功；兩者都是 3 |
| 以過期版本 1 修改 | 唯一鍵錯誤 | 唯一鍵錯誤，資料不變 |
| 停用 | 唯一鍵錯誤（修改過一次後） | 成功；兩者都是 4 |
| 舊資料（讀模型 1、aggregate 2）經 migration 後修改 | — | migration 後讀模型 2；修改成功，兩者都是 3 |

ISSUE-029 原本是依程式碼推導固定資產類別的情況，這次探測已實際確認兩種類別結果相同。

## 3. 設計

| # | 問題 | 建議 | 理由 |
|---|---|---|---|
| D1 | 讀模型版本如何更新 | 4 個查詢（兩種類別的修改與停用）在 SET 加上 `version = version + 1` | 沿用其他讀模型的慣例。每個修改或停用事件都會讓 aggregate 版本加 1，畫面也送出讀模型的版本，所以正常流程下兩者一致：第一次修改時 aggregate 沒有紀錄，送出 1 而新建為 2；讀模型同時由 1 變成 2 |
| D2 | 已不同步的既有資料 | 新增 migration：`category_uuid` 的 aggregate 在同一商戶下有版本紀錄時，讀模型的 `version` 改為該紀錄的版本。停用的類別也一起對齊；不更新 `updated_by`、`updated_at`。沒有版本紀錄的類別（從未修改）不變，由 R2 段 4 補上紀錄 | ISSUE-029 留下「讀模型 1、aggregate 2」的類別；只修改程式時，畫面仍送 1，這些類別依然無法修改。以商戶比對是因為 `aggregate_versions` 的主鍵包含 `merchant_id`。可以重複執行，結果相同 |
| D3 | Down | 不還原，只留註解（同 `20261021001_account_initial_version.sql`） | 原本的讀模型版本是錯的，還原會重新引入 ISSUE-029 |
| D4 | 過期版本的錯誤 | 不修改：仍是舊路徑的唯一鍵錯誤，專屬 API 回傳 400 | R2 段 4 把類別事件遷移到 BFS 後，修改與停用改以 `RootVersionExpected` 帶畫面版本，由 BFS Store 的樂觀鎖檢查（R2 段 4 情境 3），錯誤隨之改變，F8 不先調整。R2 D9 的業務版本只涵蓋 `account.updated`、`ledger_account.updated` 的 `payload.version`；類別的 payload 沒有版本欄位，不適用 |
| D5 | 測試方式 | 新增一個 table-driven 測試檔（`doc/TESTING.md` 的 services 規範）。每個情境都對兩種類別各執行一次；migration 情境以 F9 建立的 `database.MigrateUpTo` 遷移到前一版，再寫入 fixture。既有的 Ginkgo 測試不修改 | 兩種類別的程式相同，以同一張表驅動，避免兩份測試日後不一致 |

替代方案：修改時把讀模型版本設為事件的 aggregate 版本（`ct.Event.AggregateVersion`），而不是加 1。這樣在第一次修改送出任意版本時也能保持一致（見第 9 節），但要新增 sqlc 參數並修改兩個 projection，也和其他讀模型的慣例不同；而第 9 節的情況會在 R2 段 4 處理。不採用。

## 4. 修改內容

| 檔案 | 修改 |
|---|---|
| `internal/database/queries/prepaid_category.sql` | `UpdatePrepaidCategory`、`SoftDeletePrepaidCategory` 遞增 `version`（D1） |
| `internal/database/queries/fixed_asset_category.sql` | `UpdateFixedAssetCategory`、`SoftDeleteFixedAssetCategory` 遞增 `version`（D1） |
| `internal/database/migrations/20261022002_category_version_sync.sql`（新增） | D2、D3 |
| `internal/services/category_version_test.go`（新增） | 第 5 節的情境 |

程式 4 個檔案（sqlc 產生的程式由 `go generate` 更新，不在版控中），未達 `backend/CLAUDE.md` 的分段標準，一段完成。

## 5. 測試情境

fixture 以程式或 SQL 明確寫入，不使用產生的假資料：

- 科目沿用 `category_test.go` 的 `seeds/accounts.sql` 科目（預付 `1104-01`、`5201-01`；固定資產 `1601-01`、`1601-02`、`5601-01`）。
- migration 情境的 `category_uuid` 使用固定常數，商戶為 1，另以商戶 2 驗證比對條件。

每個情境都對預付與固定資產類別各執行一次。

### 5.1 寫入（`TestCategoryVersion_Write`，經 `EventStoreService.Append`）

封套同專屬 handler：

- 建立：aggregate 為新的 UUID v7，送出版本 0。
- 修改、停用：aggregate 為 `category_uuid`，送出讀模型目前的 `version`（同畫面）。

| # | 情境 | 預期 |
|---|---|---|
| 1 | 建立後連續修改兩次 | 都成功；每次修改後讀模型版本等於 aggregate 版本（2、3）；欄位為第二次修改的值 |
| 2 | 修改一次後停用 | 成功；讀模型版本與 aggregate 版本都是 3，類別停用 |
| 3 | 修改一次後，以修改前的版本 1 再修改 | 回傳 `aggregate_versions` 的唯一鍵錯誤；讀模型的所有欄位、aggregate 版本與 `event_store` 筆數都維持第一次修改後的狀態 |

### 5.2 migration（`TestCategoryVersion_Migration`）

fixture 的建立順序：

1. 以 `database.MigrateUpTo` 遷移到前一版（`20261022001`）。
2. 寫入商戶 1 的類別（讀模型版本 1，`updated_by`、`updated_at` 為固定值）與下表的版本紀錄。
3. 以 `database.RunMigrations` 遷移到最新版本，再寫入科目 seed（同 `test.NewTestDB`）。

| # | 情境（遷移前） | 預期 |
|---|---|---|
| 4 | 舊版修改過一次：讀模型 1、aggregate 2 | migration 後讀模型 2；以 2 修改成功，讀模型與 aggregate 都是 3 |
| 5 | 舊版停用過一次：讀模型 1、停用、aggregate 2 | migration 後讀模型 2，仍停用 |
| 6 | 從未修改：讀模型 1、沒有版本紀錄 | migration 後讀模型 1；以 1 修改成功，讀模型與 aggregate 都是 2 |
| 7 | 同一 `category_uuid` 只在商戶 2 下有版本紀錄（版本 5） | migration 後讀模型 1 不變 |

所有情境中，migration 都不改變 `version` 以外的欄位。情境 7 無法經由 API 產生（pipeline 以商戶查詢類別，找不到就拒絕），用來確認 migration 依商戶比對。

## 6. 預期的行為變更

1. 修改、停用類別後，讀模型的 `version` 遞增；查詢 API 回傳遞增後的版本。
2. 已修改或停用過一次的既有類別，讀模型版本對齊 aggregate 版本（1 → 2）。

結果：同一類別可以連續修改，也可以修改後停用；畫面送出過期的版本時回傳錯誤、資料不變（ADR-027 規則 6 開始有作用）。

維持不變：

- API 路徑與 JSON 欄位；
- 建立事件的 aggregate 身分（R2 段 4）；
- 過期版本的錯誤訊息與狀態碼（D4）；
- pipeline 的驗證；
- 前端。

## 7. 驗證

依 `doc/NOTES.md` 的 Docker 流程（`golang:1.26.9`、900 MB 記憶體上限）：

1. `go generate ./...`、`go build ./...`；`internal/services` 執行 `go vet`。
2. `go test -vet=off ./...` 全數通過。
3. 新增的測試檔 `gofmt -l` 無輸出；`git diff --check` 無輸出。
4. 反向確認（在容器副本中修改，修改查詢後重新執行 `go generate`，不留在 repo）：
   - 修改的查詢不遞增 `version`：情境 1、4 失敗。
   - 停用的查詢不遞增 `version`：情境 2 失敗。
   - migration 不更新：情境 4、5 失敗。
   - migration 不比對 `merchant_id`：情境 7 失敗。

## 8. 文件

- **本次（計畫書）**：
  - 上層計畫 5.2 的 F8 加上本計畫的路徑。
  - R2 計畫書段 4 的測試說明改為沿用 F9 建立的 `database.MigrateUpTo`（原本寫在 `shared/utils/test/db.go` 新增 helper）。
  - ISSUE-029 的固定資產類別改為已探測確認。
- **完成後**：ISSUE-029 改為 Resolved；上層計畫 5.2 與 `doc/STEP.md` 的 F8 改為 done；`doc/TODO.md` 勾選 F8。

## 9. 風險與範圍外

- **第一次修改送出任意版本**：aggregate 沒有版本紀錄時，舊路徑接受任何送出的版本。
  - 經通用 API 或自行組的請求送出 1 以外的版本時，aggregate 版本會是「送出版本 + 1」，讀模型只加 1，兩者再度不一致。
  - 畫面一律送出讀模型的版本，正常流程不會發生。
  - R2 段 4 會關閉這個缺口，靠以下三點：
    - 段 4 的 migration 補上從未修改類別的版本紀錄（D6）；
    - 修改與停用改以 `RootVersionExpected` 帶畫面版本；
    - BFS Store 的樂觀鎖在沒有紀錄、送出版本又不是 0 時視為衝突（`persistence/repos/store.go:47`）。

    F8 不在類別 projector 加入版本條件。
- **既有開發 DB**：migration 在啟動時自動執行，只改 `version`，重複執行結果相同。
- **舊 replay**（Wontfix）：重播既有的類別事件時，讀模型版本等於修改與停用事件數加 1；正常流程留下的事件與 aggregate 版本一致。R2 已不承諾舊 replay 的結果（R2 計畫書 D12）。
- **範圍外**：
  - 建立事件的 aggregate 身分與從未修改類別的版本紀錄（R2 段 4，D6）；
  - 類別事件遷移到 BFS，包含過期版本改由 BFS Store 的樂觀鎖回傳衝突（R2 段 4）。
