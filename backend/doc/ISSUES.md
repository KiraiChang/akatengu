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

### [ISSUE-001] Edit tool 字串重複導致無法定位修改位置

- **狀態**：🟢 Resolved
- **日期**：2026-05-15
- **嚴重程度**：Low
- **位置**：`internal/handler/report.go`
- **描述**：
  在 `report.go` 中新增 `GetCashFlowStatement` handler 時，檔案內存在兩段相同的結尾片段（`response.OK(w, result)\n}`），Edit tool 無法唯一定位要插入的位置，回傳 ambiguity 錯誤。
- **影響範圍**：
  編輯作業被阻斷，需重試。
- **解決方向**：
  提供更大範圍的上下文（包含前一個 handler 函式的函式簽名），確保匹配字串在檔案內唯一。
- **解決紀錄**：
  改用包含前函式 body 的較長字串作為 `old_string`，成功定位並插入。無需修改程式碼本身。

---

### [ISSUE-002] summary_cf 父科目重複計入現金流量 Total

- **狀態**：🟢 Resolved
- **發生時機**：實作 entry-based `queryCashFlowChanges` 後
- **嚴重程度**：High
- **位置**：`internal/repos/query/`
- **描述**：
  新的 `queryCashFlowChanges` 查詢同時回傳葉節點（leaf_cf）與父科目彙總行（summary_cf）。
  `GetCashFlowStatement` 在計算各活動小計時，對所有回傳行一律累加，導致父科目金額被重複計算。

  **範例**：1103-01（+800）和 1103-02（+500）都有 OPERATING entry，
  查詢結果同時包含葉節點（各 800、500）與父科目 1103（彙總 1300）。
  若三行都累加，OPERATING 調整項 = 800 + 500 + 1300 = 2600，應為 1300。
- **影響範圍**：
  現金流量表三大活動小計金額錯誤（重複計算）。
- **根本原因**：
  舊設計（account-based）由使用者決定要設定在葉節點還是彙總節點，避免雙重設定。
  新設計（entry-based）父科目彙總行是系統自動產生的，無法由使用者控制是否出現。
- **解決紀錄**：
  `GetCashFlowStatement` 計算 `OperatingActivities.Total` / `InvestingActivities.Total` / `FinancingActivities.Total` 時，
  加入 `if !row.IsSummary` 條件，只有葉節點行才累入小計；彙總行僅用於顯示階層，不計入總額。

---

### [ISSUE-003] 直接法 queryDirectOperatingCash 無法用 sqlc 管理

- **狀態**：⚫ Wontfix（確認，設計決策）
- **發生時機**：實作 `GetDirectCashFlowStatement` 時
- **嚴重程度**：Low
- **位置**：`internal/repos/query/`
- **描述**：
  `queryDirectOperatingCash` 的 `operating_txns` CTE 中，`a.type IN ('INCOME', 'EXPENSE')` 包含動態 IN 清單，
  即使此 IN 值固定，sqlc 在解析 CTE 子查詢時也無法靜態追蹤 aggregate 函數的欄位來源（`COALESCE(SUM(...))`），
  無法自動對應覆寫型別（`decimal.Decimal`）。

  此外，同一查詢中混用了以科目型別判斷的 CTE（`operating_txns`）與以帳戶屬性判斷的 CASH 過濾，
  sqlc 無法對這種跨 CTE 的多條件 JOIN 產生型別安全的 Go 方法。
- **影響範圍**：
  此查詢無法納入 sqlc 管理，需以 sqlx 手動維護。
- **解決紀錄**：
  按照 CLAUDE.md 例外條款：「動態性質確實無法以 sqlc 靜態產生」，改用 `sqlx.Named` 搭配 `db.GetContext` 直接執行。
  在程式碼行內以注釋說明原因（`// sqlx.Named is required because...`）。
  結果欄位 `debit_total` / `credit_total` 透過掃描至既有的 `cashSumRow` struct（`decimal.Decimal` 型別），
  符合金額欄位使用 `decimal.Decimal` 的規定，無需建立 View。

---

### [ISSUE-004] sqlc nullable TEXT + custom type override 不產生 pointer 型別

- **狀態**：⚫ Wontfix（確認，設計決策）
- **發生時機**：在 sqlc.yaml 新增 `journal_entries.cash_flow_category` override 後
- **嚴重程度**：Low
- **位置**：`sqlc.yaml`、`internal/model/db/projection/`
- **描述**：
  sqlc 在 `emit_pointers_for_null_types: true` 設定下，nullable TEXT 欄位搭配基本型別（如 `*string`）可自動產生指標型別。
  但若使用 custom type override（如 `enums.CashFlowCategory`），sqlc 會直接使用指定型別，**不會**自動加上 `*`，
  產生的 params struct 欄位為 `CashFlowCategory enums.CashFlowCategory`（非指標）。
- **影響範圍**：
  `journal_entries.cash_flow_category` 欄位在 sqlcdb 生成的 params struct 中為非指標型別，
  需在 Projection 層特別處理 NULL 語義。
- **解決紀錄**：
  調查 `enumx.Enum[T]` 的 `Value()` 方法：零值（empty string）的 `Value()` 回傳 `nil`，即 SQL NULL。
  因此可以直接用零值代表「未設定」，不需要指標型別：
  - `projection.Entry.CashFlowCategory enums.CashFlowCategory`（零值 → INSERT NULL）
  - `payload.TransactionEntryPayload.CashFlowCategory *enums.CashFlowCategory`（nil → 轉換為零值再寫入）
  在 `applyTransaction` 做 nil check 後解參考，確保 payload nil 正確轉為 DB NULL。

### [ISSUE-005] sqlc 不讀取 migration 檔，新增欄位必須同步更新 schema.sql

- **狀態**：🟢 Resolved
- **日期**：2026-05-19
- **嚴重程度**：Medium
- **位置**：`internal/database/schema.sql`
- **描述**：
  新增 `updated_by` / `updated_at` 欄位時，只在 `migrations/` 的 `ALTER TABLE` 語句中加入欄位，
  並更新了 `.sql` 查詢檔（INSERT / UPDATE / SELECT）。
  但執行 `go generate ./...`（sqlc generate）後，生成的 Params struct 仍不含 `UpdatedBy` 欄位，
  導致 `go build` 報 `unknown field UpdatedBy in struct literal` 錯誤。
- **根本原因**：
  sqlc 以 `schema.sql` 作為唯一的資料庫結構定義來源，**不讀取** `migrations/` 目錄下的 `ALTER TABLE`。
  `ALTER TABLE` 僅供 goose 在 runtime 執行，sqlc 看不到這些變更。
- **影響範圍**：
  所有透過 sqlc 生成的 Params / Row struct，若欄位未在 `schema.sql` 中定義，都不會出現在生成的 Go 程式碼中。
- **解決紀錄**：
  在 `schema.sql` 的對應 `CREATE TABLE` 語句中，同步加入 `updated_by TEXT` 與 `updated_at TEXT`，
  重新執行 `go generate ./...` 後問題解除。
  **結論：新增欄位時，`schema.sql` 與 migration 檔必須同時更新。**

---

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
