# Akatengu Agent Guide

本文件是整個 repository 的共用 AI 開發指引，適用於 Codex、Claude 及其他 coding agents。回應使用繁體中文；程式碼識別字與程式碼註解使用英文。

## 指引層級

1. 本文件適用於整個 repository。
2. 修改 `backend/` 前，必須再閱讀 `backend/CLAUDE.md`；測試細節見 `backend/doc/TESTING.md`。
3. 修改 `frontside/` 前，必須再閱讀 `frontside/CLAUDE.md`。
4. 子目錄文件的領域規則比本文件更具體時，以子目錄文件為準。
5. 不要為了取得背景一次讀完所有歷史文件；依任務從 `doc/README.md` 的索引選讀相關 ADR、Issue、Note 或 TODO。

## Repository Map

| 範圍 | 技術與責任 | 主要入口 |
|---|---|---|
| `backend/` | Go 1.26 API、事件溯源、SQLite/sqlc、靜態檔案服務 | `backend/cmd/main.go`、`backend/internal/handler/handler.go` |
| `frontside/` | Svelte 5 SPA、TypeScript 6、Vite | `frontside/src/main.ts`、`frontside/src/App.svelte` |
| `backend/doc/` | 後端 ADR、測試策略、問題、筆記、TODO | `backend/doc/ARCHITECTURE.md` |
| `frontside/doc/` | 前端 ADR、問題、筆記、TODO | `frontside/doc/ARCHITECTURE.md` |
| `doc/` | 全專案文件入口與跨端維護規則 | `doc/README.md` |

## 工作方式

開始修改前先說明需求理解、預計範圍與驗證方式。需求不明、會造成不可逆影響或會改變既有架構時，先取得確認；其餘情況依已確認的目標執行，不擴張範圍。

- 先查看 `git status --short`，保留使用者既有變更，不覆寫、不回復無關差異。
- 優先搜尋再讀取，只開啟與任務直接相關的檔案；不要重複讀取同一內容。
- 大型工作要拆成 phase，並在受影響子專案的 `doc/STEP.md` 記錄進度；「大型」的判定標準與 phase 執行規則依受影響子專案的 `CLAUDE.md`。
- 遇到既有測試失敗時，先區分是否由本次變更造成；不要順手修改無關問題。
- 不自行執行 `git push`，也不提交未經要求的 commit。
- 禁止使用 AI 生成的假資料來測試功能；請使用真實或手動建立的測試資料，fixture 必須明確、可追溯且只服務測試案例。

## 變更幅度與重構

目的是讓每次修改都容易 review，避免簡單需求產生過大的 diff。

1. **新需求只做完成需求所需的最小修改。**
   - 不順手重構、搬移檔案、重新命名、調整格式或修改與需求無關的程式碼；發現值得改善的地方，記錄到對應端的 `doc/TODO.md`。
   - 優先沿用既有的結構、命名與寫法，不為單一需求另建新的抽象層。
   - 預估修改量達到子專案 `CLAUDE.md` 的「大型」標準時，先提出拆分方式，讓每一段都能獨立 review、獨立通過驗證。
2. **需要重構時，先寫重構計畫，經使用者確認後才動手。**
   - 計畫寫在對應端的 `doc/plan/`，至少說明重構原因、範圍、維持不變的行為、分階段步驟與驗證方式。
   - 重構與新需求分開進行，不混在同一次修改中；若新需求必須先重構，先完成並驗證重構，再實作需求。
   - 重構不得改變計畫未明載的行為；計畫中要列出每個預期的行為變更（例如事件結構、API）並說明理由。
   - bug fix 不併入重構的實作範圍，依第 1 點以小幅修改另行實作、另行 review；計畫書中可以列出相關 bug 與前後依賴。
   - 每個階段完成後都要通過該端的驗證流程。
3. **回報時說明本次修改的範圍**；若需求本身無法避免較大的 diff，說明原因。

## 跨端契約

修改 API 時，後端 model、JSON contract、前端 `src/types/` 與 `src/api/` 必須一起檢查。

- JSON 欄位以後端 `json` tag 為準，統一使用 `snake_case`。
- Go `decimal.Decimal` 在 JSON 與 TypeScript 中使用 `string` 保存精度，只在顯示層轉成數字格式。
- Go pointer 對應 TypeScript 的 `T | null`；有 `omitempty` 的欄位才視情況使用 optional `?`。
- 日期與時間在 API 邊界使用字串，不在 API service 層轉成 JavaScript `Date`。
- 前端 API 呼叫統一經 `frontside/src/api/http.ts` 的 `apiFetch`，避免遺漏授權資訊與 401 處理；唯一例外是尚未取得 token 的登入請求（`src/api/auth.ts`）。
- 前端 `event_type` 必須和後端已註冊的事件字串完全相同，禁止猜測。
- 後端對前端公開的回應不得直接暴露 `sqlcdb.*` 型別，必須轉為 `internal/model/db/projection/` 的 model。

## Backend Rules

後端的完整限制與流程以 `backend/CLAUDE.md` 為準（sqlc／sqlx、`decimal.Decimal`、生成碼、enumx、`panic()`、SQL 註解等），此處只保留跨端工作需要的架構摘要，不重複列出細則。

後端核心寫入路徑如下：

```text
Handler -> EventStoreService.Append -> Pipeline -> UnitOfWork -> Projection
```

- 所有會計帳務異動預設必須走上述事件溯源鏈路，並由 UnitOfWork 的樂觀鎖保護；只有既有 ADR 明載的例外可以直接更新。
- 查詢使用 `internal/repos/query/` 的唯讀 repository。

### Backend 驗證

在 `backend/` 執行：

```bash
go generate ./...
go build ./...
go test ./...
```

Handler／Middleware 測試使用 Ginkgo v2 + Gomega；Service、Repository 與 package 使用標準 Go table-driven tests。完整規範見 `backend/doc/TESTING.md`。

## Frontend Rules

前端的完整規範以 `frontside/CLAUDE.md` 為準（Svelte 5 runes、TypeScript、命名、目錄與樣式慣例、RWD 檢查點），此處只保留會影響其他目錄的規則：

- `npm run build` 會清空並重建 `backend/internal/web/static`，執行前後都要檢查該目錄差異。

### Frontend 驗證

在 `frontside/` 執行：

```bash
npm run lint
npm run check
npm run build
```

純 CSS 修改可省略 `npm run check`；其餘步驟仍需執行。專案目前沒有獨立的 frontend test script，不要以 `go test` 代替前端驗證。

## 文件維護

完成需求後，只更新真正受影響的文件：

| 內容 | 寫入位置 |
|---|---|
| 長期架構決策與取捨 | 對應端的 `doc/ARCHITECTURE.md`（ADR） |
| 已知問題、阻礙與解法 | 對應端的 `doc/ISSUES.md` |
| 可重用的實作或除錯知識 | 對應端的 `doc/NOTES.md` |
| 這次不做、未來要處理的工作 | 對應端的 `doc/TODO.md` |
| 尚待確認的大型計畫書與重構計畫 | 對應端的 `doc/plan/` |
| 大型需求的分階段進度 | 對應端的 `doc/STEP.md` |
| 全端文件入口、分類或路徑改動 | `doc/README.md` |

跨端決策要在 backend 與 frontside 的 ADR 互相連結，避免兩份文件各自描述不同契約。若本次沒有產生值得保存的決策、問題、筆記或待辦，回報時明確說明無需更新其他 `doc/` 文件。
