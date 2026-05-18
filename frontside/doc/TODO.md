# TODO

計畫中的待辦事項。這裡記錄「知道要做但現在不做」的工作，包含功能擴充、技術債、效能優化與文件補充。
已確認本次要做的需求請直接進入需求處理流程，不需先記錄於此。

---

## 狀態說明

| 狀態 | 說明 |
|------|------|
| ⬜ Backlog | 計畫中，尚未排程 |
| 🔵 Planned | 已排入近期開發計畫 |
| 🟡 In Progress | 進行中 |
| ✅ Done | 已完成 |
| ❌ Cancelled | 取消，附理由 |

## 分類說明

| 分類 | 說明 |
|------|------|
| `[Feature]` | 新功能或功能擴充 |
| `[Refactor]` | 程式碼重構、架構調整 |
| `[Perf]` | 效能優化 |
| `[Test]` | 測試補充或改善 |
| `[Docs]` | 文件補充 |
| `[DX]` | 開發體驗改善（工具、腳本、設定） |
| `[Security]` | 安全性強化 |

---

## 待辦清單

### 功能

- [ ] ⬜ [Feature] 財務報表匯出 — 各報表頁加入「匯出 CSV / PDF」功能
- [ ] ⬜ [Feature] 財務報表列印版面 — 加入 `@media print` CSS，隱藏 sidebar/header，讓報表可直接從瀏覽器列印
- [ ] 🟡 [Feature] 分錄展開列顯示 cash_flow_category — `JournalEntry.svelte` 的 `entry-header` 與 `entry-row` 加入現金流量欄，顯示中文 label（或 `—`）；**前提**：需確認後端 `GET /api/txn/:id` 是否已回傳 `cash_flow_category`，若有則同步更新 `src/types/transaction.ts` 的 `Entry` interface

### 重構

- [ ] ⬜ [Refactor] 報表查詢表單元件化 — 四份報表的查詢表單結構相似（日期輸入 + 查詢按鈕），可抽成 `ReportQueryBar.svelte` 共用元件

### 技術債

- [ ] 🟡 [Refactor] `Entry` interface 補上 cash_flow_category — `src/types/transaction.ts` 的 `Entry`（查詢用途）目前缺少 `cash_flow_category: CashFlowCategory | null`；確認方式：查後端 `GET /api/txn/:id` 回應 schema 或實際呼叫 API

### UI

- [ ] 🟢 [Feature] Modal 寬度在小螢幕確認 — `JournalEntry.svelte` 新增分錄 Modal 改為 6 欄後，尚未在 375px / 768px 確認排版；檢查現金流量下拉選單是否正常顯示、分錄列是否有水平溢出

### 文件

- [ ] ⬜ [Docs] 後端 API 文件 — 整理現有 API 端點、請求/回應格式，供前後端對齊使用

---

## 完成紀錄

- [x] [Feature] 財務報表 sidebar sub-menu 結構重構 — 完成日期：2026-05-15，關聯：ADR-004、ADR-005
- [x] [Feature] 現金流量表（CashFlowStatement）頁面 — 完成日期：2026-05-15
- [x] [Feature] 權益變動表（EquityStatement）頁面 — 完成日期：2026-05-15
- [x] [Feature] 會計科目列表新增「現金流量分類」欄位 — 完成日期：2026-05-15
