# ISSUES

---

## ISSUE-001：summary_cf 父科目重複計入現金流量 Total

**狀態：** 已解決
**發生時機：** 實作 entry-based `queryCashFlowChanges` 後

### 問題描述

新的 `queryCashFlowChanges` 查詢同時回傳葉節點（leaf_cf）與父科目彙總行（summary_cf）。
`GetCashFlowStatement` 在計算各活動小計時，對所有回傳行一律累加，導致父科目金額被重複計算。

**範例**：1103-01（+800）和 1103-02（+500）都有 OPERATING entry，
查詢結果同時包含葉節點（各 800、500）與父科目 1103（彙總 1300）。
若三行都累加，OPERATING 調整項 = 800 + 500 + 1300 = 2600，應為 1300。

### 根本原因

舊設計（account-based）由使用者決定要設定在葉節點還是彙總節點，避免雙重設定。
新設計（entry-based）父科目彙總行是系統自動產生的，無法由使用者控制是否出現。

### 解決方式

`GetCashFlowStatement` 計算 `OperatingActivities.Total` / `InvestingActivities.Total` / `FinancingActivities.Total` 時，
加入 `if !row.IsSummary` 條件，只有葉節點行才累入小計；彙總行僅用於顯示階層，不計入總額。

---

## ISSUE-003：直接法 queryDirectOperatingCash 無法用 sqlc 管理

**狀態：** 已確認，設計決策
**發生時機：** 實作 `GetDirectCashFlowStatement` 時

### 問題描述

`queryDirectOperatingCash` 的 `operating_txns` CTE 中，`a.type IN ('INCOME', 'EXPENSE')` 包含動態 IN 清單，
即使此 IN 值固定，sqlc 在解析 CTE 子查詢時也無法靜態追蹤 aggregate 函數的欄位來源（`COALESCE(SUM(...))`），
無法自動對應覆寫型別（`decimal.Decimal`）。

此外，同一查詢中混用了以科目型別判斷的 CTE（`operating_txns`）與以帳戶屬性判斷的 CASH 過濾，
sqlc 無法對這種跨 CTE 的多條件 JOIN 產生型別安全的 Go 方法。

### 處理方式

按照 CLAUDE.md 例外條款：「動態性質確實無法以 sqlc 靜態產生」，改用 `sqlx.Named` 搭配 `db.GetContext` 直接執行。
在程式碼行內以注釋說明原因（`// sqlx.Named is required because...`）。

結果欄位 `debit_total` / `credit_total` 透過掃描至既有的 `cashSumRow` struct（`decimal.Decimal` 型別），
符合金額欄位使用 `decimal.Decimal` 的規定，無需建立 View。

---

## ISSUE-002：sqlc nullable TEXT + custom type override 不產生 pointer 型別

**狀態：** 已確認，設計決策
**發生時機：** 在 sqlc.yaml 新增 `journal_entries.cash_flow_category` override 後

### 問題描述

sqlc 在 `emit_pointers_for_null_types: true` 設定下，nullable TEXT 欄位搭配基本型別（如 `*string`）可自動產生指標型別。
但若使用 custom type override（如 `enums.CashFlowCategory`），sqlc 會直接使用指定型別，**不會**自動加上 `*`，
產生的 params struct 欄位為 `CashFlowCategory enums.CashFlowCategory`（非指標）。

### 處理方式

調查 `enumx.Enum[T]` 的 `Value()` 方法：
> **零值（empty string）的 `Value()` 回傳 `nil`，即 SQL NULL。**

因此可以直接用零值代表「未設定」，不需要指標型別：
- `projection.Entry.CashFlowCategory enums.CashFlowCategory`（零值 → INSERT NULL）
- `payload.TransactionEntryPayload.CashFlowCategory *enums.CashFlowCategory`（nil → 轉換為零值再寫入）

在 `applyTransaction` 做 nil check 後解參考，確保 payload nil 正確轉為 DB NULL。
