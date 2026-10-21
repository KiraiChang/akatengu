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

### [ISSUE-015] appendClosingEntry 迭代父科目導致年度結帳分錄借貸不平衡

- **狀態**：🟢 Resolved
- **日期**：2026-05-28
- **嚴重程度**：High
- **位置**：`internal/services/pipelines/factory/period.go`，`appendClosingEntry` 函式
- **描述**：
  `appendClosingEntry` 負責產生年度結帳時的損益歸零分錄。它從 `GetIncomeStatement` 取得損益表資料，
  遍歷 `is.Income` 與 `is.Expenses` 並對每一行產生沖銷分錄。
  但 `queryIncomeStatementNoSnap` 同時回傳葉節點（`is_summary=0`）**與**父科目彙總行（`is_summary=1`）。
  父科目的金額是其所有子孫葉節點的加總，因此若遍歷時不跳過父科目，
  同一金額會被計入兩次（葉節點一次 + 父科目彙總一次），造成借方合計 ≠ 貸方合計。

  **範例**：帳戶 "4101-01"（葉節點）有 50000 收入，其父 "4101"（`is_summary=TRUE`）彙總也顯示 50000。
  若兩行都產生 DR 分錄，DR 合計 = 100000，但 CR（本期淨利）= 50000，借貸不平衡。
  `applyTransaction` 的 `debit != credit` 驗證必然失敗，年度結帳事件寫入被拒絕。
- **影響範圍**：
  所有執行 `EventPeriodAnnualClosed` 的商戶。只要損益表科目有父子關係（幾乎必然），
  結帳分錄即無法成功寫入，年度結帳功能完全失效。
- **根本原因**：
  `GetIncomeStatement` 回傳完整的科目樹（含彙總節點）供前端渲染階層用，
  但 `appendClosingEntry` 設計時未考慮彙總節點的存在，對所有行一律沖銷。
- **解決紀錄**：
  在 `appendClosingEntry` 的收入與支出迴圈頭部分別加入 `row.HasChild ||` guard：
  ```go
  for _, row := range is.Income {
      if row.HasChild || row.Amount.IsZero() {
          continue
      }
      // ...
  }
  for _, row := range is.Expenses {
      if row.HasChild || row.Amount.IsZero() {
          continue
      }
      // ...
  }
  ```
  只對葉科目（`HasChild=false`）產生沖銷分錄，父科目彙總行僅供顯示，不重複沖銷。

---

### [ISSUE-016] enums.InitEnums() 必須在任何 enum 使用前呼叫，不可放在 handler 初始化中

- **狀態**：🟢 Resolved
- **日期**：2026-05-28
- **嚴重程度**：High
- **位置**：`cmd/main.go`（原位於 `internal/handler/handler.go` 的 `NewMux`）
- **描述**：
  `enums.InitEnums()` 負責將所有 enum 值註冊到全域 registry，使 `Parse*`、`.Enum()`、`.String()` 等方法正確運作。
  原本放在 `handler.NewMux()` 中執行，但 `bootstrap.BootstrapEventSeeds` 在 `InitDB` 階段（比 `NewMux` 更早）就呼叫了 enum 方法，
  導致 event seed bootstrap 執行時 registry 尚未初始化，產生 runtime 錯誤。
- **影響範圍**：
  任何在 `NewMux` 呼叫之前使用 enum 的程式碼（如 bootstrap、seeder、排程任務）都會受影響。
- **解決紀錄**：
  將 `enums.InitEnums()` 移至 `cmd/main.go` 的最頂端，確保早於 `InitDB`（及其內部的 `BootstrapEventSeeds`）執行。
  **結論：`enums.InitEnums()` 屬於全域初始化，必須是程式啟動的第一步，不得放在任何具體的服務或 handler 初始化函式內。**

---

### [ISSUE-017] 新增 AggregateType 需同步更新 DB schema 的 CHECK constraint

- **狀態**：🟢 Resolved
- **日期**：2026-05-28
- **嚴重程度**：High
- **位置**：`internal/database/schema.sql`、對應 migration 檔
- **描述**：
  `event_store` 與 `aggregate_versions` 兩張表均有如下 CHECK constraint：
  ```sql
  CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION'))
  ```
  新增 `AggregateSysConfig`（'SYS_CONFIG'）與 `AggregateAccountConfig`（'ACCOUNT_CONFIG'）後，
  嘗試寫入這兩類 aggregate 的事件時，SQLite 因 CHECK constraint 不符而拒絕 INSERT，
  導致 `BootstrapEventSeeds` 與後續所有 SYS_CONFIG / ACCOUNT_CONFIG 相關事件寫入失敗。
- **影響範圍**：
  - `BootstrapEventSeeds`：`bootstrapSysAccounts`、`bootstrapAccountConfigs` 呼叫 `es.Append` 全部失敗
  - 任何透過 `/api/event/append` 寫入新 aggregate type 的請求均被 DB constraint 拒絕
- **解決紀錄**：
  新增 `internal/database/migrations/20260528001_update_aggregate_type_constraint.sql`，
  以 SQLite 慣用模式（CREATE new → INSERT SELECT → DROP old → RENAME）重建兩張表，
  同時搭配 `PRAGMA legacy_alter_table = ON` 跳過 `v_installments_active` 既有視圖的欄位驗證（見 ISSUE-012）。
  更新後的 CHECK constraint：
  ```sql
  CONSTRAINT chk_aggregate_type CHECK (aggregate_type IN ('ACCOUNT', 'TRANSACTION', 'SYS_CONFIG', 'ACCOUNT_CONFIG'))
  ```
  **結論：往後每次新增 `aggregateTypeVal` 常數，必須同步新增 migration 擴充此 constraint；
  schema.sql 無需修改（sqlc 不需要知道 CHECK constraint）。**

---

### [ISSUE-019] EventStoreService.Append 回傳 *db.EventStore，非 *pipelines.Result

- **狀態**：🟢 Resolved
- **日期**：2026-06-03
- **嚴重程度**：Low
- **位置**：`internal/services/bank_statement_review.go`
- **描述**：
  `ApproveTxn` 呼叫 `es.Append(...)` 後，嘗試以 `txnResult.Event.EventUuid` 取得新建交易的 UUID。
  `pipelines.Result` struct 確實有 `.Event db.EventStore` 欄位，但 `EventStoreService.Append` 的回傳型別為 `*db.EventStore`，
  非 `*pipelines.Result`，因此 `.Event` 欄位不存在，`go build` 報：
  ```
  txnResult.Event undefined (type *db.EventStore has no field or method Event)
  ```
- **影響範圍**：
  `bank_statement_review.go` build 失敗，Phase 5 無法完成。
- **解決紀錄**：
  在 `ApproveTxn` 組裝 `AppendCmd` 前，已用 `uuid.NewV7()` 生成 `txnAggID` 作為 `AggregateID`。
  Transaction projection 將 `aggregate_id` 存為 `txn_uuid`，因此直接以 `txnAggID.String()` 呼叫 `GetByUUID` 即可，
  不需讀取 `Append` 的回傳值中的 event UUID。
  修正：`_, err = s.es.Append(...)` + `s.q.Transaction.GetByUUID(ctx, txnAggID.String())`。

---

### [ISSUE-018] 全量 Replay 後整數 FK 斷裂問題

- **狀態**：🟢 Resolved
- **日期**：2026-05-29
- **嚴重程度**：High
- **位置**：所有 event sourcing projection 資料表
- **描述**：
  全量重播（TruncateProjections + 逐事件 Apply）後，SQLite AUTOINCREMENT 計數器未重置，
  projection 記錄取得比原始更大的整數 ID。跨事件引用的整數 FK（如 `investment_lots.movement_id`、
  `journal_entries.txn_id`、`prepaid_amortizations.prepaid_id` 等）因 ID 改變而無法對應到正確主檔，
  讀模型關聯斷裂，財務報表資料不正確。
- **影響範圍**：
  所有依賴 replay 重建讀模型的商戶；以及多租戶情境下測試 replay 的場景。
- **解決紀錄**：
  為所有 projection 資料表新增程式端生成的 UUID 欄位（migration `20260529002_add_uuid_to_projections.sql`），
  跨表引用改用 UUID FK（與整數 FK 並存）。UUID 從事件的 `event_uuid`（UUIDv7）以 `uuidx.NewFromEvent` 衍生，
  確保相同事件 replay 產生相同 UUID，FK 關聯正確重建。詳見 ADR-018。

---

### [ISSUE-020] BFS 路徑的現金流分類修改無法運作

- **狀態**：🟢 Resolved
- **日期**：2026-10-19
- **嚴重程度**：High
- **位置**：`internal/handler/entry_cf_category.go:69-70`、`internal/process/cashflow.go:22`、`internal/persistence/handle/cash_flow_cateory.go:45`、`internal/handler/handler.go:33`
- **描述**：
  `PUT /api/txn/{txn_uuid}/cf-category` 是目前唯一接上 BFS engine 的功能，但有三個問題：
  1. HTTP handler 建立 `domain/event.TransactionCFCategoryUpdatedPayload`，mediator handler 與 projector 卻斷言 `model/payload.TransactionCFCategoryUpdatedPayload`。兩者是不同型別，斷言永遠失敗，每次請求都回 400。
  2. `handler.go` 建立空的 `handle.Registry`，`NewCashFlowCategoryProjector` 沒有呼叫點，讀模型不會更新。
  3. 事件未填 `AggregateType`、`MerchantID`、`UpdatedBy`、`ExpectedVersion`，aggregate ID 每次隨機產生，樂觀鎖失效，事件以 `merchant_id = 0` 寫入，因此也不在 replay 的商戶範圍內。

  另外，`bfs_engine.go:40` 會把 handler 回傳、未持久化的 `r.State` 傳給 projector，replay 時無法取得。目前的 CF projector 沒有使用 state，不影響此功能，因此不列入本 issue，改由計畫書重構類 R1 處理。
- **影響範圍**：
  前端 CF 分類審視頁（CFCategoryReview）的儲存功能；後續所有遷移到 BFS 的功能都會沿用同一套地基。
- **解決方向**：
  見 `doc/plan/BFS_ENGINE_MIGRATION.md` F1；F1 完成後即可關閉本 issue。
- **解決紀錄**：
  2026-10-19 依計畫書 F1 分兩段修正（進度見 `doc/STEP.md`）：
  1. Phase 1（commit `5f6b731`）：mediator handler 與 projector 改用 `domain/event` 的 payload 型別；在 `handler.go` 註冊現金流分類 projector；事件的 aggregate ID 改用 `txn_uuid`，並補上 `AggregateType`、`MerchantID`、`UpdatedBy`、`ExpectedVersion`。review 另外發現 BFS Store 找不到版本紀錄時未要求 `expected_version` 為 0，一併修正（`internal/persistence/repos/store.go`；舊路徑的相同問題見 ISSUE-024）。
  2. Phase 2：新增 `internal/handler/entry_cf_category_scenarios_test.go`、`entry_cf_category_test.go`，以 in-memory DB 執行整條 BFS 路徑，8 個情境全數通過。其中 1 個經正式 `NewMux`（含 JWT 與商戶驗證）；刪除 `NewMux` 的 projector 註冊時，該情境會失敗，可攔截本 issue 第 2 點的回歸。

---

### [ISSUE-021] 全量 Replay 的清空清單漏表，存在銀行資料時重建失敗

- **狀態**：⚫ Wontfix（舊 replay 將被取代，且目前沒有需要保留的資料）
- **日期**：2026-10-19
- **嚴重程度**：High
- **位置**：`internal/repos/unit_of_work/event_store/truncate.go:28`
- **描述**：
  `TruncateProjections` 沒有清除 `bank_csv_templates`、`bank_csv_template_ledgers`、`bank_pdf_templates`、`bank_pdf_template_ledgers`、`bank_statement_imports`、`bank_statement_import_ledgers`、`bank_statement_txns`、`entry_cf_categories`。
  銀行相關資料表的 UUID 由事件決定（`template_uuid`、`import_uuid` 取自 payload 或 `event_uuid`），且 INSERT 未處理衝突，因此只要存在任何銀行範本或匯入紀錄，全量重播時就會撞 UNIQUE 而中斷。
  另外 `bank_statement_txns.matched_entry_id`、`created_txn_id` 只有整數 FK，重建後分錄與交易的 ID 改變，關聯會斷裂（同 ISSUE-018 的問題類型）。
- **影響範圍**：
  `POST /api/audit/replay` 的全量重建（`from_event_id = 0`）；以及 `POST /api/audit/event/import`，它在清除並寫入事件後會呼叫全量 replay，含銀行資料的匯出檔因此必定匯入失敗。
- **解決紀錄**：
  2026-10-19 決定不修正舊 replay：
  1. 目前系統沒有需要保留的資料，不需要以舊 replay 重建。
  2. 舊 replay 將由計畫書中只跑 BFS projector 的新重建機制（R3）取代，單獨修正舊機制屬於重工。

  本 issue 的要求改列為新重建機制的驗收條件：清空範圍由各 projector 自行宣告，且涵蓋所有由事件衍生的資料表；`bank_statement_txns` 改用 UUID 關聯；銀行對帳狀態依計畫書架構決策 A3 處理（直接清除銀行資料表會遺失 ADR-021 直接寫入的比對狀態與 `created_txn_id`）。
  在新機制完成前，只要存在銀行範本或匯入紀錄，`POST /api/audit/replay` 的全量重建仍會失敗。

---

### [ISSUE-022] 全量 Replay 非原子，中途失敗時 DB 停在半重建狀態

- **狀態**：⚫ Wontfix（舊 replay 將被取代，且目前沒有需要保留的資料）
- **日期**：2026-10-19
- **嚴重程度**：High
- **位置**：`internal/services/event.go:153`
- **描述**：
  `EventStoreService.Replay` 先在一個交易內清空 projection 並 commit，之後每個事件各自開一個交易套用 projection。任一事件失敗時，清空已經生效，後續事件不再套用，且沒有回復機制。
- **影響範圍**：
  全量重建失敗時，該商戶的讀模型與報表資料不完整。`POST /api/audit/event/import`（`services/event.go:317`）在一個交易內清除並寫入事件、commit 之後才呼叫 replay，重建失敗時原有事件已被取代，無法回到匯入前的狀態。
- **解決紀錄**：
  2026-10-19 決定不修正舊 replay，理由同 ISSUE-021。

  本 issue 的要求改列為新重建機制的驗收條件：重建中途失敗時，資料必須回到重建前的狀態，策略依計畫書架構決策 A4。
  匯入的對應要求改列為計畫書 R10 的驗收條件：不相容的匯入在清除資料前拒絕；清除、寫入與重建在同一交易完成，失敗後原資料不變。
  新機制只跑 projector、不重跑 pipeline，因此不會遇到舊 replay 的問題：改成單一交易時，舊 pipeline 透過全域 DB 查詢，在 `MaxOpenConns=1` 下等不到被交易占用的連線，也讀不到交易內尚未提交的重建結果。但新機制的 projector 必須只透過交易內的 repository 讀寫。
  在新機制完成前，舊 replay 中途失敗仍會留下半重建的資料。

---

### [ISSUE-023] 銀行對帳核准必定失敗並留下未關聯的交易，重按會產生重複交易

- **狀態**：🟡 In Progress（F3 止血完成；原子化待計畫書 R7）
- **日期**：2026-10-19
- **嚴重程度**：High
- **位置**：`internal/services/bank_statement_review.go:168`、`:174`、`:186`、`:195`；`internal/services/projection/transaction.go:265`
- **描述**：
  `ApproveTxn` 先以 `EventStoreService.Append` 寫入 `transaction.created`（自己的交易），接著以 aggregate ID 查詢剛建立的交易（`:186`），再以另一個 `UnitOfWork.Do` 回寫 `bank_statement_txns.created_txn_id`。
  但 projection 寫入的交易 UUID 是 `uuidx.NewFromEvent(eventUUID, "txn:")`（`projection/transaction.go:265`，自 `1f6fc25` 起），不等於 aggregate ID，因此查詢永遠回傳 nil，核准一律以「created transaction not found after event」失敗：交易已經建立，銀行明細卻沒有標成已核准。使用者重按時，因 aggregate ID 每次以 `uuid.NewV7()` 隨機產生（`:168`），會再建立一筆交易。
  即使修正查詢，寫入交易與回寫關聯仍分屬兩個交易，回寫失敗時同樣會留下未關聯的交易。
  以上依程式碼推導；`ApproveTxn` 目前沒有測試覆蓋。
- **影響範圍**：
  銀行對帳審查頁的核准功能無法完成；每次核准都會留下未關聯的會計交易，重按還會重複建立，直接影響帳務正確性。目前系統沒有需要保留的資料，既有資料不另行盤點。
- **解決方向**：
  計畫書範圍決策 S4 已確定採 (a)：先做冪等止血（F3），只改 `ApproveTxn`：
  1. 交易 aggregate ID 改由銀行交易 UUID 衍生的固定值。
  2. 以 `Append` 回傳事件的 `event_uuid` 依 `"txn:"` 規則推導交易 UUID 後查交易，修正必定查不到的問題。
  3. 重試或並行時，先以 `query.Event.GetByAggregate` 找既有的 `transaction.created` 事件，存在就只補寫關聯，不再 append。
  原子化在重構類 R7 完成。
  目前的 BFS engine 每個事件各自一個交易，`EventStoreService.Append` 也無法在外部交易內執行，兩者都不能以小幅修改直接達成原子化。見 `doc/plan/BFS_ENGINE_MIGRATION.md`。
- **解決紀錄**：
  2026-10-19 完成 F3 冪等止血（`internal/services/bank_statement_review.go`）：
  1. 交易 aggregate ID 改為 `uuidx.NewFromEvent(銀行明細 UUID, "bank_approve")`。
  2. 建立前以 `query.Event.GetByAggregate` 查既有的 `transaction.created`，有就沿用；沒有才 `Append`。`Append` 失敗後一律重查，查到才走補關聯，查不到就回傳原本的錯誤。
  3. 交易 UUID 改以事件 UUID 依 `"txn:"` 規則推導，修正必定查不到交易的問題。

  測試 `internal/services/bank_statement_review_test.go` 5 個情境全數通過：首次核准、已核准再按、回寫失敗後重試、並行（以 stub 強制雙方都先查到沒有事件，本次 `Append` 衝突後走恢復路徑）、`Append` 失敗且查不到事件。在容器副本停用恢復路徑時，並行情境會失敗，確認測試確實經過恢復路徑。
  仍待處理：寫入交易與回寫關聯仍分屬兩個交易，回寫失敗時會暫時留下未關聯的交易（重試可補上）；原子化在計畫書 R7 完成後才關閉本 issue。

---

### [ISSUE-024] 舊路徑建立新 aggregate 時未要求 expected_version 為 0

- **狀態**：🔴 Open
- **日期**：2026-10-19
- **嚴重程度**：Medium
- **位置**：`internal/services/event.go:78`
- **描述**：
  `EventStoreService.Append` 在 `UpdateIfVersionMatch` 找不到符合的版本紀錄時，直接以 `expected_version + 1` 建立版本紀錄，沒有要求 `expected_version` 為 0。新的 aggregate 送 7 會以版本 8 寫入，送 -1 會以版本 0 寫入，繞過版本一致性檢查。
  BFS 的 Store（`internal/persistence/repos/store.go`）原本有相同問題，已在 F1（ISSUE-020）中修正。
- **影響範圍**：
  所有經 `POST /api/event/append` 等舊路徑寫入的事件。前端已確認的呼叫（建立傳票、現金流分類修改）都先以 `getAggregateVersion` 取得版本再送出，正常操作下不會觸發。
- **解決方向**：
  加入與 BFS Store 相同的檢查。修正前需先確認前端所有建立新 aggregate 的呼叫都送 0，避免既有功能被拒絕。舊路徑將在計畫書 R10 移除，也可評估改為 Wontfix。

---

### [ISSUE-025] BFS safety middleware 的狀態由所有請求共用，並行請求可能讓後端程式終止

- **狀態**：🟢 Resolved（2026-10-21，計畫書 F4）
- **日期**：2026-10-21
- **嚴重程度**：High
- **位置**：`internal/handler/handler.go:41`、`internal/runtime/safety/cycle_detector.go:16`、`internal/runtime/safety/backpressure.go:19`
- **描述**：
  `NewMux` 只建立一個 BFS engine，middleware 的狀態在建立時就產生，之後由所有請求共用：`CycleDetector` 的 `seen` map 與 `BackpressureScheduler` 的 `pending` 計數。`CycleDetector` 的註解說每次呼叫會產生獨立的實例，但實際上每個 engine 只呼叫一次。
  1. 兩個請求同時寫入 `seen` 時，Go runtime 會以 `fatal error: concurrent map read and map write` 終止整個程式，這種錯誤無法 recover。
  2. `pending` 有 data race，不同請求的計數互相干擾，可能誤判為超量（`ErrQueueOverflow`）或漏判。
  3. `seen` 只增不減，會記住所有處理過的事件 UUID，記憶體持續成長。

  重現（2026-10-21，暫時測試只放在容器副本，未加入 repo）：依 `NewMux` 的組態建立一條 middleware chain，由 8 個 goroutine 共用，各送 2000 個根事件。一般執行時程式以上述 fatal error 終止（`cycle_detector.go:19`）；`-race` 回報 `backpressure.go:24`、`:25`、`:37` 與 `cycle_detector.go:19`、`:23` 的 data race。
- **影響範圍**：
  目前唯一接上 BFS 的 `PUT /api/txn/{txn_uuid}/cf-category`：兩個請求同時進入 BFS 時，可能讓整個後端程式終止。計畫書 R1 起，遷移後的寫入都會經過 BFS，影響範圍會跟著擴大。`DeterministicLoopDetector` 也是相同寫法，但目前沒有使用。
- **解決方向**：
  middleware 的狀態改為每次 `Run` 各自獨立（計畫書 F4，需在 R1 之前完成）。注意事項：
  1. 只在每次 `Run` 重建 chain 還不夠：`CycleDetector()` 在被呼叫時就建立 map，重用同一個 middleware 值時狀態仍然共用。必須讓狀態在每次 `Run` 時重新建立，或明確傳入只屬於這次 `Run` 的狀態。
  2. 新的 chain 放在 `Run` 的區域變數，不能覆寫共用的 engine 欄位。
  3. 目前未使用的 `DeterministicLoopDetector` 依相同原則修正。
  4. 測試除了以 `-race` 確認，還要驗證各 middleware 的功能：
     - 並行測試共用同一個 engine，以同步屏障強制各次 `Run` 重疊。
     - 跨 `Run` 不互相干擾，同一次 `Run` 仍能擋下重複的 UUID。
     - `BackpressureScheduler`：兩個重疊的 `Run` 各自未超量時都成功；其中一個超量時，另一個不受影響。
     - `DeterministicLoopDetector`：同一次 `Run` 的事件類型循環仍被擋下，跨 `Run` 不殘留 graph。
     - `Run` 失敗後，下一次 `Run` 不殘留狀態。
  5. 狀態在 middleware 包裝 processor 時配置，每次 `Run` 一份，不是每個事件重建；middleware 的順序與錯誤 policy 不變。

  屬於 bug fix，依 `AGENTS.md` 以小幅修改另行實作與 review。
- **解決紀錄**：
  2026-10-21 完成計畫書 F4：
  1. `internal/runtime/safety/cycle_detector.go`、`backpressure.go`：`CycleDetector`、`DeterministicLoopDetector`、`BackpressureScheduler` 的狀態改在包裝 processor 時配置，重用同一個 middleware 值也不會共用狀態。
  2. `internal/runtime/engine/bfs_engine.go`：engine 保存最內層的 processor 與 middleware 清單，`Run` 開頭組一次 chain 放在區域變數；多次呼叫 `WithMiddleware` 時維持原本的順序（越晚加入的越外層）。`handler.go` 不需修改。
  3. `internal/runtime/engine/bfs_engine_test.go` 新增 8 個情境：同一事件再次 `Run`、同一次 `Run` 的重複 UUID、失敗後重試、8 個 goroutine 以屏障重疊執行、Backpressure 兩個 `Run` 都未超量、一個超量另一個不受影響、`DeterministicLoopDetector` 同一次 `Run` 的類型循環、跨 `Run` 不殘留因果圖。

  驗證（Docker）：engine 21 個情境全數通過，`go test -vet=off ./...` 全數 `ok`，`internal/runtime/...` 的 `-race` 沒有 data race。在容器副本換回修改前的檔案時，跨 `Run` 與 Backpressure 的 5 個情境都會失敗（只換回 engine、只換回 safety，結果相同）；並行情境在一般執行下仍會通過，加上 `-race` 才會回報 data race。

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
