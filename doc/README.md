# Akatengu 文件索引

此目錄是全專案文件入口。後端與前端文件保留在各自目錄，以維持責任邊界；本頁提供跨端索引、閱讀順序與更新規則，避免複製後產生內容分歧。

## 建議閱讀順序

1. 根目錄 `README.md`：專案結構、啟動方式與架構摘要。
2. 根目錄 `AGENTS.md`：AI agent 共用工作規範與跨端契約。
3. 依任務選擇 backend 或 frontside 的 `CLAUDE.md`。
4. 只在需要時閱讀對應的 ADR、Issue、Note、TODO 或進度文件。

## 文件總覽

| 類別 | Backend | Frontend | 用途 |
|---|---|---|---|
| Agent 指引 | [`backend/CLAUDE.md`](../backend/CLAUDE.md) | [`frontside/CLAUDE.md`](../frontside/CLAUDE.md) | 子專案限制、慣例與驗證流程 |
| 架構決策 | [`backend/doc/ARCHITECTURE.md`](../backend/doc/ARCHITECTURE.md) | [`frontside/doc/ARCHITECTURE.md`](../frontside/doc/ARCHITECTURE.md) | 已採納或提議中的 ADR |
| 已知問題 | [`backend/doc/ISSUES.md`](../backend/doc/ISSUES.md) | [`frontside/doc/ISSUES.md`](../frontside/doc/ISSUES.md) | 問題、根因、狀態與解法 |
| 技術筆記 | [`backend/doc/NOTES.md`](../backend/doc/NOTES.md) | [`frontside/doc/NOTES.md`](../frontside/doc/NOTES.md) | 可重用的設計、除錯與實作知識 |
| 待辦事項 | [`backend/doc/TODO.md`](../backend/doc/TODO.md) | [`frontside/doc/TODO.md`](../frontside/doc/TODO.md) | 尚未納入當前需求的工作 |
| 執行進度 | [`backend/doc/STEP.md`](../backend/doc/STEP.md) | [`frontside/doc/STEP.md`](../frontside/doc/STEP.md) | 大型需求的 phase 與驗證紀錄 |
| 測試策略 | [`backend/doc/TESTING.md`](../backend/doc/TESTING.md) | `frontside/CLAUDE.md` 的驗證流程 | 測試風格、結構與指令 |
| 會計領域 | [`backend/doc/account/CASH_FLOW_GUIDE.md`](../backend/doc/account/CASH_FLOW_GUIDE.md) | 前端 ADR-003 | 現金流分類與前端呈現契約 |

## 跨端架構

```text
Browser
  -> Svelte SPA (`frontside/`)
  -> `/api` JSON contract
  -> Go HTTP handlers (`backend/internal/handler/`)
      -> query repository（讀取）
      -> EventStoreService + Pipeline + UnitOfWork + Projection（會計寫入）
  -> SQLite
```

開發時由 Vite 在 `:5174` 提供 SPA，並將 `/api` 代理至 Go server 的 `:8080`。正式建置由 `frontside/vite.config.ts` 將資源輸出至 `backend/internal/web/static/`，再由 Go 服務嵌入。

## API 契約快速索引

| Backend | Frontend | 規則 |
|---|---|---|
| `json:"snake_case"` | interface 的 `snake_case` 欄位 | 以實際 JSON tag 為準 |
| `decimal.Decimal` | `string` | 保留精度，顯示層才格式化 |
| `*T` | `T \| null` | `omitempty` 才使用 optional 欄位 |
| `time.Time`／日期字串 | `string` | API/service 層不轉 `Date` |
| problem JSON `{ title, detail }` | API error | 優先顯示 `detail`，其次 `title` |
| event type registry | `event_type` | 字串必須完全一致 |

## 文件歸屬

- 只影響單一端的內容，寫入該端 `doc/`。
- 跨端契約若形成長期決策，backend 與 frontside 各自建立或更新 ADR，並互相連結。
- 本頁只維護索引、全端資料流與文件歸屬，不在此複製完整 ADR 或歷史問題。
- `STEP.md` 是當前大型工作的進度檔；開始新計畫前應先確認舊內容是否已完成，再依子專案規範更新。
- 修改文件路徑或新增文件類別時，同步更新本索引與根目錄 `README.md`。
