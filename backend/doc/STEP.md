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
| R0 | 交易內查詢：讓 `query.Repo` 可綁定 `sqlx.Tx` | 🟡 in progress |
| R1 | engine 基礎：事件圖一個交易（A1）、子事件封套與版本規則（A2）、handler 改用 `Run` 建立的交易內查詢、移除 projector 的 `state` 參數、建立事件註冊表 | ⬜ pending |
| R2 | 主檔與設定（含 `BootstrapEventSeeds`）、期間開帳改走 BFS；`POST /api/event/append` 依註冊表派送 | ⬜ pending |
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
| 2 | 測試：`internal/repos/query/tx_test.go` | ⬜ pending |

## 最終驗證結果

R10 完成後記錄全套驗證結果。
