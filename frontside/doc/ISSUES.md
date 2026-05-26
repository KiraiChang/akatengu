# Issues

開發過程中遇到的問題、技術障礙與待解事項。問題解決後更新狀態，**不得刪除紀錄**（保留歷史供參考）。

---

## 狀態說明

| 狀態 | 說明 |
|------|------|
| 🔴 Open | 尚未解決 |
| 🟡 In Progress | 處理中 |
| 🟢 Resolved | 已解決 |
| ⚫ Wontfix | 確認不修復，附理由 |

---

## 問題清單

<!-- 新增問題時複製以下範本，依日期降冪排列 -->

### [ISSUE-005] AccountSelect 根目錄層無結果：傳入 descendants 導致空下拉選單
- **狀態**：🟢 Resolved
- **日期**：2026-05-26
- **嚴重程度**：High
- **位置**：`src/views/Ledger.svelte:48`（`filteredAccounts` derived）
- **描述**：
  `Ledger.svelte` 的帳本新增 Modal 依 LedgerTypeConfig 過濾可選科目時，將 `filteredAccounts` 設為 `activeLedgerTypeConfig.descendants`（配置科目下的所有子孫）。但 `AccountSelect` 元件在開啟時以 `currentParentId = null` 為初始狀態，只顯示 `parent_id === null` 的頂層科目。由於 `descendants` 內所有科目的 `parent_id` 均非 `null`（它們都指向其父科目），導致初始畫面一筆都不顯示，下拉選單完全空白。
- **影響範圍**：
  帳本管理→新增帳戶→選擇現有科目，在 LedgerTypeConfig 已配置（`account_id` 非空、`descendants` 非空）的情況下，科目下拉選單無任何選項，使用者無法新增帳戶。
- **解決方向**：
  在 `filteredAccounts` 中除了 `descendants` 外，還必須加入從根節點到配置科目的完整祖先鏈（ancestors）。這樣 AccountSelect 在根目錄層就能找到頂層祖先，使用者可逐層鑽入，最終到達配置科目，再看到所有 descendants。
- **解決紀錄**：
  將 `filteredAccounts` 與 `filteredParentAccounts` 的計算邏輯改為 IIFE：從 `activeLedgerTypeConfig.account_id` 出發，沿 `parent_id` 向上追溯所有祖先，最終回傳 `[...path.reverse(), ...descendants]`。`filteredParentAccounts` 設為與 `filteredAccounts` 相同（兩者本質上是同一份祖先 + 子孫集合）。`npm run check` 與 `npm run build` 均無錯誤。

---

### [ISSUE-004] Svelte 唯讀展示用 `<label>` 未關聯控制項導致 a11y 警告
- **狀態**：🟢 Resolved
- **日期**：2026-05-19
- **嚴重程度**：Low
- **位置**：`src/views/Accounts.svelte`、`src/views/Ledger.svelte`、`src/views/Investment.svelte`
- **描述**：
  在 modal 內新增唯讀審計欄位（更新者 / 更新時間）時，`<label>` 未加 `for` 屬性，`npm run check` 回報 `a11y_label_has_associated_control` 警告，共 6 筆。
- **影響範圍**：
  僅警告，不影響功能，但 `npm run check` 輸出有雜訊。
- **解決方向**：
  即使是唯讀 `<input>`，仍需 `for` + `id` pair；或改用 `<p>` 等非互動元素搭配 `<span>` 顯示，完全不需 label。
- **解決紀錄**：
  為每個唯讀欄位補上靜態 `id`（`f-acct-updated-by`、`f-ldgr-updated-by`、`f-inv-updated-by` 等），`<label for="...">` 對應。警告全部消除，`npm run check` 回報 0 errors 0 warnings。

---

### [ISSUE-003] `npm run lint` 執行失敗（eslint: command not found）
- **狀態**：🟢 Resolved
- **日期**：2026-05-16
- **嚴重程度**：Low
- **位置**：開發環境（frontside）
- **描述**：
  執行 `npm run lint` 時回傳 `sh: eslint: command not found`，即使在正確的 frontside 目錄下也無法執行。
- **影響範圍**：
  無法透過 lint 驗證程式碼品質。
- **解決方向**：
  Node.js / npm 執行環境異常，`node_modules/.bin/eslint` 未被正確解析（可能為 PATH 問題或 nvm/fnm 環境未載入）。向使用者說明狀況並請求支援，不嘗試自行繞過（見 CLAUDE.md 行為規範）。
- **解決紀錄**：
  由使用者修正執行環境後恢復正常。

### [ISSUE-002] Svelte 5 `$effect` 內同時讀寫同一 `$state` 導致無限更新迴圈
- **狀態**：🟢 Resolved
- **日期**：2026-05-15
- **嚴重程度**：High
- **位置**：`src/routes/Home.svelte`
- **描述**：
  為了讓導覽到報表子路由時自動展開 sidebar sub-menu，在 `$effect` 內透過展開運算子讀取 `expandedParents`（`[...expandedParents]`），再將新的 `Set` 賦值回去。每次賦值觸發 effect 重新執行，形成無限迴圈，Svelte 5 runtime 拋出 `effect_update_depth_exceeded`，導致報表子選單無法正常顯示與導覽。
- **影響範圍**：
  財務報表 sub-menu 展開狀態異常，無法正常導覽至資產負債表、損益表、現金流量表。
- **解決方向**：
  Svelte 5 規則：`$effect` 內不應對自己追蹤的 reactive state 進行寫入。改用 `$derived` 純衍生計算取代 `$effect` + 狀態寫入。
- **解決紀錄**：
  移除 `$effect`，改宣告 `const isReportsExpanded = $derived(expandedParents.has('reports') || currentPath.startsWith('/home/reports'))`。展開狀態由手動切換（`expandedParents`）與目前路徑共同決定，無副作用。

### [ISSUE-001] `replace_all` Edit 工具誤換 `$derived` 宣告內部的同名表達式
- **狀態**：🟢 Resolved
- **日期**：2026-05-15
- **嚴重程度**：Low
- **位置**：`src/routes/Home.svelte`
- **描述**：
  使用 `replace_all: true` 將所有 `expandedParents.has('reports')` 替換為 `isReportsExpanded` 時，連 `$derived` 宣告本身內部的 `expandedParents.has('reports')` 也被替換，造成 `isReportsExpanded = $derived(isReportsExpanded || ...)` 的循環參考，tsc 報 `implicitly has type 'any'` 與 `used before its declaration` 錯誤。
- **影響範圍**：
  型別檢查失敗，無法建置。
- **解決方向**：
  `replace_all` 只能用於確實要全域替換的場景。若宣告式本身含有相同子字串，應改用精確的 `old_string` 取代，不使用 `replace_all`。
- **解決紀錄**：
  手動將 `$derived` 宣告內的表達式還原為 `expandedParents.has('reports')`，僅保留 template 裡的替換結果。


<!--
### [ISSUE-XXX] 標題
- **狀態**：🔴 Open
- **日期**：YYYY-MM-DD
- **嚴重程度**：High / Medium / Low
- **位置**：`path/to/file.go:行號`
- **描述**：
  問題的具體現象與重現步驟。
- **影響範圍**：
  哪些功能或模組受影響。
- **解決方向**：
  目前已知的解法或待確認事項。
- **解決紀錄**：（Resolved 後填寫）
  如何解決、相關 commit hash。
-->
