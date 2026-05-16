# TODO

「現在不做、但之後要做」的事項。每次開發前先確認，避免遺漏。

格式：
- **優先度**：🔴 高 / 🟡 中 / 🟢 低
- **分類**：功能 / 技術債 / UI / 文件
- **狀態**：待處理 / 進行中 / 完成

---

## 功能

### 🟡 分錄展開列顯示 cash_flow_category｜待處理

`JournalEntry.svelte` 的展開分錄列（`txn-entries`）目前欄位為：
會計科目 / 帳戶 / 借方 / 貸方 / 備註

未顯示 `cash_flow_category`，使用者新增後無法在列表確認設定是否正確。

**建議做法**：在 `entry-header` 與 `entry-row` 加入現金流量欄，顯示中文 label（或 `—`）。
**前提**：需確認後端 `GET /api/txn/:id` 回傳的 entry 資料是否已包含 `cash_flow_category` 欄位；
若有，同步更新 `src/types/transaction.ts` 的 `Entry` interface。

---

## 技術債

### 🟡 `Entry` interface 補上 cash_flow_category｜待處理

`src/types/transaction.ts` 的 `Entry`（查詢用途）目前缺少 `cash_flow_category` 欄位。
若後端 API 已回傳此欄位，前端型別需補上 `cash_flow_category: CashFlowCategory | null`。

**確認方式**：查後端 `GET /api/txn/:id` 的回應 schema，或呼叫 API 觀察實際回傳內容。

---

## UI

### 🟢 Modal 寬度在小螢幕確認｜待處理

`JournalEntry.svelte` 的新增分錄 Modal 改為 6 欄後，尚未在 375px / 768px 實際確認排版。
**檢查點**：現金流量下拉選單是否正常顯示、分錄列是否有水平溢出。
