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
- [ ] ⬜ [Feature] AssetTypeConfig 表格欄位優化 — 目前顯示 account_id 字串，可考慮改為顯示科目名稱（accountMap lookup）；但欄位多（7 欄），可能改用 tooltip 或點擊展開的方式顯示
- [ ] ⬜ [Feature] 帳本建立限制套用至 Installment — 目前 `NewLedgerFormSection` 的 `accountsForParent` 限制只在 `Ledger.svelte` 傳入，`Installment.svelte` 的 `LedgerSelectSection` 尚未套用同樣限制
### 重構

- [ ] ⬜ [Refactor] 報表查詢表單元件化 — 四份報表的查詢表單結構相似（日期輸入 + 查詢按鈕），可抽成 `ReportQueryBar.svelte` 共用元件

### 技術債

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
- [x] [Feature] 分錄展開列顯示 cash_flow_category — 完成日期：2026-05-19
- [x] [Refactor] 所有 projection 模型補上 `updated_by` / `updated_at`（後端 + 前端 types + 頁面） — 完成日期：2026-05-19
- [x] [Feature] Setting API 串接 — `LedgerTypeConfig` / `AssetTypeConfig` 頁面 + 帳本建立科目限制 — 完成日期：2026-05-20，關聯：ADR-009、ADR-010
- [x] [Feature] 預付費用（Prepaid）頁面 — 建立 / 攤提 / 提前終止，含攤提明細與分錄展開 — 完成日期：2026-05-25
- [x] [Feature] 固定資產（FixedAsset）頁面 — 購入 / 折舊 / 處分，含折舊明細與分錄展開 — 完成日期：2026-05-25
- [x] [Refactor] 分期管理 sidebar 改為父子選單（分期付款 / 預付費用 / 固定資產），`isExpanded` 改為多路徑陣列比對 — 完成日期：2026-05-25
- [x] [Feature] 現金流量表新增直接法（直接法）tab — 完成日期：2026-05-26
- [x] [Feature] 傳票輸入範本功能（開 modal 時載入範本套用 + 分錄展開儲存為範本） — 完成日期：2026-05-26
- [x] [Feature] 範本管理頁面（列表 / 編輯 / 刪除），加入「系統設定」子選單 — 完成日期：2026-05-26
- [x] [Feature] 科目分析（AccountAnalysis）頁面 — 子科目餘額（含鑽取）/ 分錄明細（含分頁）/ 月別趨勢（含長條圖） — 完成日期：2026-05-26
- [x] [Feature] 科目名稱點擊導航至科目分析，並支援「← 返回」回上一頁 — 完成日期：2026-05-26，關聯：ADR-011
- [x] [Feature] 儀表板串接真實資料 — 4 張 stat card + 近期傳票 + 帳戶餘額列表 + 月度損益趨勢（含內嵌長條圖）— 完成日期：2026-05-26
- [x] [Feature] 稽核查詢（Audit）頁面 — 4 tab（彙總版本、事件紀錄分頁、Checkpoint、快照）+ 重建 Projection 按鈕 — 完成日期：2026-05-26
- [x] [Feature] 匯率管理（ExchangeRate）頁面 — 列表含幣別篩選 + 手動輸入（透過 EventRateUpdated 事件）— 完成日期：2026-05-26
