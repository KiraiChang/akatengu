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

### [ISSUE-007] SQL 查詢檔案使用 Unicode 注釋字元導致 sqlc 解析失敗

- **狀態**：🟢 Resolved
- **日期**：2026-05-20
- **嚴重程度**：Medium
- **位置**：`internal/database/queries/*.sql`
- **描述**：
  在 `config.sql` 初稿中，為了視覺分隔使用了 Unicode 裝飾注釋（如 `── ledger_account_type_config ──────────────────`）。
  執行 `go generate ./...`（sqlc generate）時，sqlc parser 對這些非 ASCII 字元報錯：
  ```
  mismatched input '─' expecting ...
  ```
  導致整個 sqlc 程式碼產生中斷，後續 `go build` 連鎖失敗。
- **影響範圍**：
  `go generate` 失敗 → sqlcdb 未更新 → `go build` 因引用不存在的 sqlc 型別而失敗。
- **解決紀錄**：
  將所有 Unicode 裝飾注釋改為純 ASCII `--` 注釋，問題解除。
  **結論：`*.sql` 檔案中只允許使用純 ASCII 字元，包含注釋內容。**

---

### [ISSUE-008] 呼叫不存在的套件函數或錯誤 import path 導致 build 失敗

- **狀態**：🟢 Resolved
- **日期**：2026-05-20
- **嚴重程度**：Medium
- **位置**：`internal/handler/setting.go`、`internal/repos/query/config.go`
- **描述**：
  實作過程中出現兩類因函數/套件名稱憑印象寫入而導致的 build 失敗：
  1. `ctxkey.GetUpdatedBy(ctx)` — 此函數不存在，正確名稱為 `ctxkey.GetUserName(ctx)`
  2. `akatengu/internal/model/db/dbmapconv` — 此路徑不存在，正確路徑為 `akatengu/internal/pkg/dbmapconv`
- **影響範圍**：
  `go build` 報 `undefined: ctxkey.GetUpdatedBy` / `cannot find package`，需回頭修正。
- **解決紀錄**：
  分別修正函數名稱與 import path。
  **結論：使用任何套件函數前，必須先以 Grep 確認函數確實存在及 import path 正確。**

---

### [ISSUE-006] sellEntries 稅金科目錯誤使用手續費 sys_code（已在重構中修正）

- **狀態**：🟢 Resolved
- **日期**：2026-05-20
- **嚴重程度**：High
- **位置**：`internal/services/pipelines/factory/investment.go`，原第 358 行
- **描述**：
  `sellEntries` 中稅金分錄的科目查詢呼叫了 `getAssetTypeFeeSysCode`（手續費），
  應為 `getAssetTypeTaxSysCode`（交易稅）。此 bug 導致出售時手續費科目被重複使用兩次，
  交易稅科目完全未被記錄。
- **影響範圍**：
  投資出售（SELL）交易分錄中，稅金（`p.Tax`）對應的科目錯誤；財務報表費用科目金額不正確。
- **解決紀錄**：
  在 asset_type_account_config 重構（Step 9）時一併修正：改為直接使用 `config.TaxAccountID`，bug 不復存在。

---

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

### [ISSUE-009] SELL 分錄借貸不平衡：RealizedGain 重複扣除手續費與交易稅

- **狀態**：🟢 Resolved
- **日期**：2026-05-22
- **嚴重程度**：High
- **位置**：`internal/services/pipelines/factory/investment.go`
- **描述**：
  `sellEntries` 中 `ct.State.RealizedGain` 計算公式為 `sellAmount - originalCostBasis - Fee - Tax`，
  同時銀行分錄使用 `NetProceeds = sellAmount - Fee - Tax`，且手續費/交易稅另以獨立 DR 分錄記錄。
  導致手續費與交易稅被重複扣除，分錄借方合計比貸方多出 `Fee + Tax`，違反借貸平衡原則。
  `applyTransaction` 的餘額驗證 `debit != credit` 在有手續費/稅的 SELL 交易時必然失敗。
- **影響範圍**：
  所有含手續費或交易稅的投資出售（SELL）交易均無法成功寫入。
- **根本原因**：
  `RealizedGain` 既從銀行淨收款中扣除，又從獲利計算中扣除，而手續費/稅已由獨立費用科目分錄覆蓋。
- **解決紀錄**：
  將 `ct.State.RealizedGain` 改為稅前毛利：`sellAmount.Sub(originalCostBasis)`（不扣費用）。
  手續費與交易稅由各自的費用科目分錄承擔，Movement 記錄已有 `Fee` / `Tax` 個別欄位可計算淨利。
  修正後借貸平衡：DR = NetProceeds + Fee + Tax + AccUnrealized = CR = InvestmentAsset + GrossGain。

---

### [ISSUE-010] 預付/資產 txn_id 雞生蛋問題：Projection 執行順序導致無法在 INSERT 時提供 txn_id

- **狀態**：🟢 Resolved
- **日期**：2026-05-22
- **嚴重程度**：Medium
- **位置**：`internal/services/projection/prepaid.go`、`internal/services/projection/asset.go`、`internal/repos/unit_of_work/event_store/projection_repo/`
- **描述**：
  `prepaids` / `fixed_assets` 表中 `txn_id` 用於關聯對應的會計交易。但在事件處理流程中，
  Projection 的執行順序為：PrepaidProjectionService（INSERT prepaid）→ TransactionProjectionService（INSERT transaction）。
  PrepaidProjectionService 執行時 transaction 尚未建立，無法取得 txn_id，
  若將 `txn_id NOT NULL`，則 INSERT 必然失敗。
- **影響範圍**：
  預付費用創建（PREPAID_CREATED）與固定資產購入（ASSET_PURCHASED）無法成功寫入。
- **解決紀錄**：
  將 `prepaids.txn_id` 與 `fixed_assets.txn_id` 定義為 nullable（`INTEGER` 無 `NOT NULL`）。
  Projection 的 INSERT SQL 不包含 `txn_id`，改為在 TransactionProjectionService 建立交易後，
  呼叫新增的 `UpdatePrepaidTxn` / `UpdateFixedAssetTxn` 將 txn_id 回寫。
  此設計參照 installment 表的 `UpdateInstallmentTxn` 既有模式。
  `prepaid_amortizations` / `fixed_asset_depreciations` 的 `txn_id` 則在 TransactionProjectionService
  建立交易後直接以正確 txn_id 一次完成 INSERT，不需二次更新。

---

### [ISSUE-011] dbmap_gen.go 型別不一致導致 build 失敗（TxnID *int64 變更後未重新 generate）

- **狀態**：🟢 Resolved
- **日期**：2026-05-22
- **嚴重程度**：Medium
- **位置**：`internal/model/db/projection/dbmap_gen.go`
- **描述**：
  將 projection model 的 `TxnID` 改為 `*int64`（nullable）後，未立即執行 `go generate ./...`。
  舊版 `dbmap_gen.go` 根據修改前的型別（`int64` non-nullable）生成了 `dbmapconv.Ptr()` / `dbmapconv.Deref()` 轉換，
  與新的 `*int64` 不相容，導致 `go build` 失敗：
  ```
  cannot use dbmapconv.Ptr(m.TxnID) (value of type **int64) as *int64
  ```
- **影響範圍**：
  `go build` 失敗，阻斷所有後續開發。
- **解決紀錄**：
  執行 `go generate ./...` 重新產生 `dbmap_gen.go`，使轉換函式與最新型別一致，問題解除。
  **結論：修改 projection struct 欄位型別後，必須立即執行 `go generate ./...`，
  否則 `dbmap_gen.go` 與 `sqlcdb` 型別不一致，導致 build 失敗。**

---

### [ISSUE-012] v_installments_active 引用不存在欄位，RENAME TABLE 時觸發驗證失敗

- **狀態**：🟢 Resolved
- **日期**：2026-05-26
- **嚴重程度**：Medium
- **位置**：`internal/database/migrations/20260310001_init_table.sql`（view 定義）、`20260526002_add_merchant_id_to_event_core.sql`（觸發點）
- **描述**：
  `v_installments_active` 視圖定義中參照 `i.paid_periods`，但 `installments` 表從未有此欄位。
  SQLite 不在 CREATE VIEW 時驗證欄位存在性（延遲驗證），因此視圖建立成功。
  但 SQLite 3.26+ 在執行 `ALTER TABLE ... RENAME TO ...` 時，強制驗證所有視圖（含不相關視圖），
  導致 migration 20260526002 的 RENAME 操作因此報錯失敗。
- **影響範圍**：
  所有使用完整 migration 序列的測試均因 migration 失敗而無法執行。
- **解決紀錄**：
  在 migration 20260526002 的所有 RENAME 操作前後加入：
  ```sql
  PRAGMA legacy_alter_table = ON;
  PRAGMA legacy_alter_table = OFF;
  ```
  `PRAGMA legacy_alter_table = ON` 告知 SQLite 跳過視圖完整性檢查，migration 成功執行。
  **注意**：`v_installments_active` 本身仍為無效視圖（`paid_periods` 欄位不存在），查詢時會報錯。

---

### [ISSUE-013] TruncateProjections 未隔離商戶，全量重播清除所有商戶資料

- **狀態**：🟢 Resolved
- **日期**：2026-05-26
- **嚴重程度**：High（多租戶環境）
- **位置**：`internal/repos/unit_of_work/event_store/truncate.go`
- **描述**：
  `TruncateProjections` 在全量重播（`fromEventID == 0`）時清除所有 projection 資料，
  但清除操作未加入 `merchant_id` 過濾條件。在多商戶環境下，
  重播單一商戶的事件會清除所有商戶的 projection 資料，造成其他商戶資料遺失。
- **影響範圍**：
  多商戶環境中，任一商戶執行全量重播均會破壞其他商戶的讀模型。
- **根本原因**：
  `TruncateProjections` 在引入商戶隔離（`merchant_id`）之前設計，未考慮多租戶場景。
- **解決紀錄**：
  在 `TruncateProjections` 開頭以 `ctxkey.GetMerchantID(ctx)` 取得 merchantID，
  所有 DELETE 語句加入 `WHERE merchant_id = ?`。
  同時補入後加功能遺漏的表（`prepaids`、`prepaid_amortizations`、`fixed_assets`、`fixed_asset_depreciations`、`account_closure`）、
  移除無 merchant_id 的全域資料 `exchange_rates`（replay 時 UPSERT 補回），
  以及移除多租戶下不適用的 `sqlite_sequence` 重置。
  `deleteUserAccounts` CTE 的起點（sys_accounts）、遞迴 JOIN 與外層 DELETE 均加入 merchant_id 過濾。

---

### [ISSUE-014] AccountBalanceRealtimeProjection 未處理投資事件，running balance 未更新

- **狀態**：🟢 Resolved（投資三事件；其餘事件見 TODO）
- **日期**：2026-05-27
- **嚴重程度**：High
- **位置**：`internal/services/projection/account_balance_realtime.go`
- **描述**：
  `AccountBalanceRealtimeProjection.Apply` 的 switch 只處理三個基本交易事件：
  `EventTransactionCreated`、`EventTransactionVoided`、`EventTransactionCorrected`。
  其餘所有透過 `TransactionProjectionService.applyTransaction` 產生分錄的事件，
  雖然 journal_entries 被正確寫入，但 `account_running_balance` 與 `ledger_running_balance`
  的 Upsert 從未被呼叫，導致即時餘額永遠不更新。
  受影響的事件類型（完整清單）：
  - `EventInvestmentBought` / `EventInvestmentSold` / `EventDividendReceived`（**本次修正**）
  - `EventInvestmentMarked`（`UnrealizedMarkedState.Transaction != nil` 但 TransactionProjectionService 亦未處理，屬複合 bug）
  - `EventInstallmentCreated` / `EventInstallmentPeriodPaid`
  - `EventPrepaidCreated` / `EventPrepaidAmortized` / `EventPrepaidDisposed`
  - `EventAssetPurchased` / `EventAssetDepreciated` / `EventAssetDisposed`
  - `EventPeriodAnnualClosed` / `EventPeriodAnnualReopened`
- **根本原因**：
  `AccountBalanceRealtimeProjection` 在設計之初只對應手動交易（EventTransactionCreated）。
  後續新增投資、分期付款、預付費用、固定資產等功能時，忘記同步補上對應的 case，
  形成「TransactionProjectionService 寫分錄 → AccountBalanceRealtimeProjection 靜默跳過」的系統性缺口。
- **解決紀錄**：
  在 `account_balance_realtime.go` 補上三個投資事件的 case，
  透過 `checkAndGetState` 從 pipeline Result 取得已計算好的 `st.Transaction.Entries`，
  呼叫 `applyFromTxnPayload` → `applyDeltas` 完成 running balance 的 Upsert。
  此模式不需額外 DB 查詢，與現有 `applyCreated` 讀 payload.Entries 的模式一致。
  其餘未修正的事件類型已記錄於 TODO（P1）。

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
