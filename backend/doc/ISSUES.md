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
