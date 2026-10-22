# F7 實作計畫書：舊寫入路徑拒絕已遷移到 BFS 的事件類型

| 項目 | 內容 |
|---|---|
| 狀態 | ⏳ 待 review |
| 建立日期 | 2026-10-22 |
| 最後修訂 | 2026-10-22：依 review 修訂。上層計畫的 F7 依賴改為「依賴 R1 段 3」；service 測試補上「已遷移事件搭配無效 payload」情境，確認檢查位於 pipeline 之前 |
| 分支 | `refactor`（基準 commit `be2c756`） |
| 上層計畫 | `doc/plan/BFS_ENGINE_MIGRATION.md`：5.2 的 F7 |
| 依賴 | R1 段 3 的事件註冊表（D2，已完成）；需在 R2 之前完成 |
| 關聯 | ISSUE-028；ADR-023、ADR-027；`doc/plan/BFS_R1_ENGINE_FOUNDATION.md` 的 D2（撰寫時 F7 尚未決定使用註冊表，該處「F7 與 R1 互不依賴」以本計畫為準） |

> F7 是 bug fix，依 `AGENTS.md`「變更幅度與重構」以小幅修改另行實作、另行 review，不併入 R1 或 R2 的重構。
> 文中行號以基準 commit 為準。

---

## 1. 問題

`transaction.cf_category_updated` 在 F1 改走 BFS（`PUT /api/txn/{txn_uuid}/cf-category`），但舊路徑仍能寫入它：

- 舊 pipeline 仍登錄這個事件類型（`internal/services/pipelines/factory/base.go:57`），舊 projection 也會處理它（`internal/services/projection/entry_cf_category.go:147`）。
- `POST /api/event/append`（`internal/handler/event.go:100`）把請求原樣交給 `EventStoreService.Append`（`internal/services/event.go:54`），因此這個事件類型可以經舊路徑寫入，違反過渡規則「同一種事件類型只能有一條寫入路徑」。
- 通用 API 接受呼叫端指定的 `aggregate_type`、`aggregate_id`，同一筆交易的分類修改可以寫到交易 UUID 以外的 aggregate，避開專屬 API 以交易 UUID 為 aggregate 的版本檢查。

前端只經專屬 API 寫入現金流分類：`frontside/src/` 沒有送出 `transaction.cf_category_updated` 的程式碼，`POST /api/event/append` 只由 `frontside/src/api/aggregate.ts:14` 呼叫。問題依程式碼推導，目前沒有實際的呼叫。

## 2. 設計

### 2.1 決策

| # | 問題 | 建議 | 理由 |
|---|---|---|---|
| D1 | 在哪一層拒絕 | **`EventStoreService.Append` 的開頭**，在 `es.factory.Dispatch` 之前 | `Append` 是舊寫入路徑唯一的入口：通用 API、各專屬 handler（設定、科目類別、銀行範本、銀行對帳）、`bootstrap/event_seed.go` 都經過它，在這裡拒絕可以涵蓋所有呼叫端，不只通用 API。舊 replay（`services/event.go:192`）與事件匯入（`services/event.go:317`，寫入後呼叫 replay）直接使用 `factory.Dispatch`，不經 `Append`，不受影響 |
| D2 | 「已遷移」以什麼為準 | **R1 的事件註冊表**：`domainevent.Lookup(cmd.EventType)` 找得到就拒絕 | `backend/CLAUDE.md` 的過渡規則已規定以 ADR-027 的註冊表為準；之後每遷移一種事件類型，舊路徑自動拒絕它，不需另外維護清單。`services` 依賴 `domain/event` 不會形成循環：註冊表只依賴 `enums`、`kernel` 與標準函式庫（ADR-027 規則 1） |
| D3 | 錯誤與 HTTP 狀態 | 新增 `services.ErrEventTypeMigrated`，以 `fmt.Errorf("%w: %v", ...)` 附上事件類型；**`EventHandler.Append` 不修改**，維持 500 與 problem 格式 | ADR-027 規則 7 要求遷移期間 `POST /api/event/append` 的錯誤格式維持不變（目前一律 500）；前端只判斷 `response.ok` 與 `detail`／`title`。改成 4xx 會讓這個 API 的狀態碼規則不一致，若要調整，應在 R2 建立派送時與舊映射的拒絕一起決定 |

### 2.2 修改內容

| 檔案 | 修改 |
|---|---|
| `internal/services/event.go` | 新增 `ErrEventTypeMigrated`；`Append` 開頭以 `domainevent.Lookup` 檢查，已登錄的事件類型直接回傳錯誤 |
| `internal/services/event_append_test.go`（新增） | 標準 table-driven 測試，見 3.1 |
| `internal/handler/event_test.go`（新增） | Ginkgo spec runner 與 helper，見 3.2 |
| `internal/handler/event_scenarios_test.go`（新增） | 3.2 的情境清單 |

程式 4 個檔案，未達 `backend/CLAUDE.md` 的分段標準，一段完成。舊 pipeline 與舊 projection 的登錄不移除，舊 replay 仍以它們重播既有的事件（Wontfix，R10 移除）。

## 3. 測試情境

fixture 沿用既有測試的寫法，以 SQL 明確寫入：一筆交易與兩筆分錄（科目來自 `seeds/accounts.sql`）。拒絕的情境 1、2 使用「沒有 F7 時舊路徑會寫入成功」的完整 payload，讓移除檢查時測試會失敗；情境 4 使用舊 pipeline 會拒絕的 payload，讓檢查被移到 pipeline 之後時測試會失敗。

### 3.1 Service（`TestAppend_MigratedEventType`）

| # | 情境 | 預期 |
|---|---|---|
| 1 | `transaction.cf_category_updated`，aggregate 為 `('TRANSACTION', 交易 UUID)` | 回傳 `ErrEventTypeMigrated`；`event_store`、`aggregate_versions`、`entry_cf_categories` 都不變 |
| 2 | 同上，aggregate 改為交易 UUID 以外的值（ISSUE-028 的繞過方式） | 同情境 1 |
| 3 | 對照：未遷移的 `account.created` | 寫入成功，`event_store` 多 1 筆 |
| 4 | `transaction.cf_category_updated` 搭配無效 payload `{}`（沒有 `txn_uuid`，舊 pipeline 的 `Validate` 會失敗） | 仍回傳 `ErrEventTypeMigrated`，不是 pipeline 的驗證錯誤；`event_store` 不變 |

### 3.2 Handler（`NewMux POST /api/event/append`）

| # | 情境 | 預期 |
|---|---|---|
| 1 | 經正式路由送出 `transaction.cf_category_updated` | 500、problem 格式，`detail` 含拒絕原因；`event_store` 沒有這筆交易的事件，`entry_cf_categories` 沒有資料 |
| 2 | 對照：經正式路由送出 `account.created` | 200，事件寫入 |

專屬 API 不受影響，由既有的 `entry_cf_category_test.go` 情境確認（不修改）。

## 4. 預期的行為變更

只有一項：`EventStoreService.Append` 收到事件註冊表中的事件類型（目前只有 `transaction.cf_category_updated`）時，回傳 `ErrEventTypeMigrated`、不執行 pipeline、不寫入事件。經 `POST /api/event/append` 時回 500 與 problem 格式。

維持不變：

- 未登錄的事件類型的寫入、版本處理與錯誤回應。
- 專屬 API `PUT /api/txn/{txn_uuid}/cf-category`（經 BFS）。
- 舊 replay（`POST /api/audit/replay`）與事件匯入：仍以舊 pipeline 重播既有的 `transaction.cf_category_updated` 事件。

## 5. 驗證

依 `doc/NOTES.md` 第 3 點的 Docker 流程（`golang:1.26.9`、900 MB 記憶體上限）：

1. `go generate ./...`、`go build ./...`；`internal/services`、`internal/handler` 執行 `go vet`。
2. `go test -vet=off ./...` 全數通過。
3. 新增與修改的檔案執行 `gofmt -l` 無輸出；`git diff --check` 無輸出。
4. 反向確認（在工作區副本中修改，不留在 repo）：
   - 移除 `Append` 的檢查：3.1 情境 1、2 與 3.2 情境 1 失敗。
   - 改成拒絕所有事件類型：3.1 情境 3 與 3.2 情境 2 失敗。
   - 把檢查移到 `es.factory.Dispatch` 之後：3.1 情境 4 失敗（回傳 pipeline 的驗證錯誤）。

## 6. 文件

完成後更新：

- `doc/ISSUES.md`：ISSUE-028 改為 Resolved，補解決紀錄。
- `doc/plan/BFS_ENGINE_MIGRATION.md` 5.2 的 F7、`doc/STEP.md` 的 F7：改為 done。
- `doc/TODO.md`：移除 F7 的待辦。
- `backend/CLAUDE.md` 的 BFS 遷移過渡規則：補一句「`EventStoreService.Append` 拒絕註冊表中的事件類型（F7）」。
- `doc/ARCHITECTURE.md` ADR-027 規則 2：補實作註記，R2 建立派送前，舊路徑以 F7 的檢查拒絕已登錄的事件類型。

## 7. 風險與範圍外

- **R2 的連帶要求**：R2 每登錄一種事件類型，舊路徑就立即拒絕它；同一個 Phase 必須把該類型的所有呼叫端（`BootstrapEventSeeds`、專屬 API、通用 API 的派送）一起改走 BFS。漏改的呼叫端會在測試中失敗，而不是悄悄寫入舊路徑，這是預期的保護。R2 計畫書需列出每種事件類型的呼叫端。
- **範圍外**：`POST /api/event/append` 依註冊表派送到 BFS、舊 aggregate 映射的檢查（R2）；錯誤狀態碼的調整（見 D3）；移除舊 pipeline 與舊 projection 的登錄（R10）。
