# Notes

臨時想法、Debug 筆記、Prompt 設計、架構草稿等非正式內容。不需完整，能讓未來的自己看懂即可。
正式化後的架構決策請移至 `ARCHITECTURE.md`，已解決的問題請移至 `ISSUES.md`。

---

## 分類

- `[IDEA]` 功能想法或改善方向
- `[DEBUG]` 除錯過程與發現
- `[PROMPT]` Prompt 設計與調整紀錄
- `[DRAFT]` 架構草稿（待正式化）
- `[REF]` 參考資料或外部連結

---

## 筆記

<!-- 新增時在最上方插入，格式如下 -->

### [DRAFT] 2026-05-29 Event Sourcing UUID 設計要點

**問題根因**：SQLite AUTOINCREMENT 計數器（`sqlite_sequence`）不在 DELETE 後重置，
全量重播（TruncateProjections + 逐事件 Apply）後 projection 記錄的整數 ID 與原始不同，
跨事件引用的 FK 斷裂（如 InvestmentSold 引用先前 InvestmentBought 建立的 lot_id）。

**UUID 生成策略**：
- 1:1 事件：`ct.Event.EventUuid` 直接作為主體 UUID（e.g., `txn_uuid`、`movement_uuid`、`ledger_uuid`、`prepaid_uuid`、`asset_uuid`）
- 1:N 事件：`uuidx.NewFromEvent(ct.Event.EventUuid, qualifier)` 衍生，qualifier 帶索引（e.g., `"entry:0"`, `"lot"`, `"payment:0"`）
- UUID v5（SHA1 deterministic）確保相同輸入永遠輸出相同值，replay 完全冪等

**`journal_entries.ledger_uuid` 特殊處理**：
TransactionCreatedPayload 不帶 LedgerAccount 狀態，改在 `sqlxTransactionRepo.UpsertJournalEntries` 內
自動 `SELECT ledger_uuid FROM ledger_accounts WHERE ledger_id = ?` 補入，不需修改 payload 或 pipeline。
lookup 失敗時寫 warn log 並繼續（ledger_uuid 留空字串），不中斷寫入。

**`investment_lot_disposals.lot_uuid` 特殊處理**：
`calcFIFOCostBasis` pipeline 從 DB 查詢 `GetNotCloseLots` 取得 `lot.LotUuid`，直接填入 `InvestmentLotDisposals.LotUuid`。
schema 與 SQL query 已更新讓 `GetNotStatusLotsRow` 包含 `lot_uuid` 欄位，dbmap 自動映射。

**整數 ID 保留**：UUID 欄位與現有整數 PK/FK 並存，不移除整數欄位，保持 SQLite btree 效能。

### [DEBUG] 2026-05-28 事件會計整合測試的設計模式

**測試入口**：`internal/services/journal_entry_test.go`（`package services_test`）

**核心原則**：所有寫入一律透過 `services.EventStoreService.Append`，不繞過 pipeline 直接操作 DB。
此模式與正式 API 路徑完全一致，能驗證 pipeline 驗證邏輯、分錄產生及 running balance 更新。

**測試前置資料（DB helpers）**：
- `insertLedger(t, db, id, typ, acctID)` — INSERT INTO ledger_accounts
- `insertPeriodOpen(t, db, id, startDate)` — INSERT 月結紀錄（MONTHLY, OPEN）
- `insertAllMonthsClosed(t, db, year, baseID)` — 12 筆 CLOSED 月結（年度結帳前置條件）
- `insertAnnualPeriodOpen(t, db, id, year)` — INSERT 年度結算期（ANNUAL, OPEN）
- `insertSysAccounts(t, db)` — 5 筆 sys_accounts（PREPAID_INTEREST / LOAN:INTEREST 等）
- `insertAssetTypeConfig(t, db)` — STOCK 投資科目設定（asset_type_account_config）
- `insertInvestment(t, db, id, acctID, assetType, costMethod, ifrsCategory)` — 投資主檔
- `insertDividendAccounts(t, db)` — 補插 "4210"（INCOME）/ "5920"（EXPENSE）（seeds 中缺少）

**sqlx 掃描陷阱**：`db.GetContext(ctx, &anonymousStruct{}, ...)` 無法正確掃描（field 沒有 `db:` tag）。
一律改用具名 struct 搭配 `db:` tag，例如 `rbScanRow{DebitTotal decimal.Decimal \`db:"debit_total"\`}`。

**年度結帳測試的雞生蛋問題**：年度結帳 pipeline 需讀取當年損益表，但建立收入交易需要開放的月結期。
解法：用原始 SQL 直接插入 `transactions` + `journal_entries` 繞過 pipeline，
`GetIncomeStatement` 直接查 journal_entries，不依賴 EventStoreService，故此繞過合法。

**DividendReceived 帳戶 hardcode 問題**：
factory/investment.go 的股息接收 pipeline 直接以字串 `"4210"`（INCOME）與 `"5920"`（EXPENSE）查詢帳戶，
這兩個科目在 seeds/accounts.sql 中不存在，測試中需 `insertDividendAccounts` 手動插入。
這是已知的設計問題，詳見 TODO P2。

### [DEBUG] 2026-05-26 SQLite RENAME TABLE 觸發全視圖驗證

SQLite 3.26+ 在執行 `ALTER TABLE X RENAME TO Y` 時，會驗證所有視圖的欄位有效性，
不只是引用被重命名表格的視圖。`v_installments_active` 因引用不存在的 `i.paid_periods`，
在任何 RENAME 操作時都會觸發 `no such column` 錯誤。

解法：`PRAGMA legacy_alter_table = ON` 告知 SQLite 使用舊版行為，跳過視圖驗證。
記得在 RENAME 後恢復 `PRAGMA legacy_alter_table = OFF`。

### [DRAFT] 2026-05-26 多商戶事件核心隔離模式

`event_store` / `aggregate_versions` / `snapshots` 三表加入 `merchant_id` 後，
所有 UoW repo 改為從 `ctxkey.GetMerchantID(ctx)` 取值（與 checkpoint.go 同一模式），
不在介面參數中傳遞。這樣 Service 層呼叫時不需感知 merchantID，由 HTTP middleware 注入到 ctx。

### 2026-05-26｜儀表板 API 設計草案

#### 三支新 API 的職責分工

| API | 說明 | 實作複雜度 |
|-----|------|-----------|
| `GET /api/dashboard/summary` | Stat Card 四個數字（本月損益 + 淨資產 + 現金） | Medium — 並行呼叫現有 income_statement / balance_sheet 內部 query |
| `GET /api/dashboard/monthly-trend?months=12` | 過去 N 個月每月收入 / 支出 / 淨額 | Medium — journal_entries + accounts 依月分組加總 |
| `GET /api/ledger/balances` | 各 LedgerAccount 目前餘額 | Low — ledger_accounts LEFT JOIN journal_entries 依 ledger_id 加總 |

#### `DashboardSummary` struct 草案

```go
type DashboardSummary struct {
    AsOfDate          string          `json:"as_of_date"`
    Month             string          `json:"month"`           // "YYYY-MM"
    MonthIncome       decimal.Decimal `json:"month_income"`
    MonthExpense      decimal.Decimal `json:"month_expense"`
    TotalAssets       decimal.Decimal `json:"total_assets"`
    TotalLiabilities  decimal.Decimal `json:"total_liabilities"`
    TotalEquity       decimal.Decimal `json:"total_equity"`
    CashBalance       decimal.Decimal `json:"cash_balance"`
}
```

實作策略：handler 用 `errgroup` 並行執行三個 query：
1. `queryIncomeStatementAccounts`（本月日期範圍）→ sum INCOME / EXPENSE
2. `queryBalanceSheetAccounts`（today）→ sum ASSET / LIABILITY
3. `queryCashBefore` + `queryCashUpTo`（今天）→ 現金餘額

#### `MonthlyTrendItem` struct 草案

```go
type MonthlyTrendItem struct {
    Month   string          `json:"month"`   // "YYYY-MM"
    Income  decimal.Decimal `json:"income"`
    Expense decimal.Decimal `json:"expense"`
    Net     decimal.Decimal `json:"net"`
}
```

SQL 要點：
- 以 `strftime('%Y-%m', txn_date)` 分組
- JOIN `accounts` 過濾 `type IN ('INCOME','EXPENSE')`
- 注意與 `queryIncomeStatement` 相同：只取葉節點 (`NOT is_summary`)，避免彙總科目重複計算
- 起始月份：`date('now', '-N months')` 的月份第一天

#### `LedgerBalance` struct 草案

```go
type LedgerBalance struct {
    LedgerID    int64           `json:"ledger_id"`
    Name        string          `json:"name"`
    Institution string          `json:"institution"`
    Type        string          `json:"type"`        // BANK_ACCOUNT / CREDIT_CARD / LOAN
    Balance     decimal.Decimal `json:"balance"`     // 依 normal_balance 調整正負
}
```

SQL 要點：
- `FROM ledger_accounts la LEFT JOIN journal_entries je ON la.ledger_id = je.ledger_id`
- `SUM(je.debit) - SUM(je.credit)` 得到借方淨額
- 依 `la.account.normal_balance` 決定餘額符號（DEBIT normal → debit-credit；CREDIT normal → credit-debit）
- 排除已關閉的帳戶（若有 `is_active` 欄位）

#### 新增路由規劃

```
GET /api/dashboard/summary          → handler.GetDashboardSummary
GET /api/dashboard/monthly-trend    → handler.GetMonthlyTrend  (query: months int, default 12)
GET /api/ledger/balances            → handler.GetLedgerBalances
```

`/api/ledger/balances` 掛在現有 `ledger.go` handler 下，與 `GET /api/ledger`（取得帳戶列表）並列。`/api/dashboard/*` 新建 `dashboard.go` handler。

#### 實作偏差（2026-05-26 完成後更新）

原草案計畫並行呼叫 income_statement / balance_sheet query，實際實作改用以下方式：

- **Summary**：`account_running_balances`（已含累計借貸總額）+ `v_monthly_income_expense`（當月），在 Go 合併計算，效能更佳，不需要掃描全量 journal_entries。
- **Monthly Trend**：新增 `v_monthly_income_expense` View（groupby merchant+month），由 sqlc 追蹤 SUM 欄位型別以產生 `decimal.Decimal`（符合 CLAUDE.md 約束，避免使用 `decimal.NewFromFloat()`）。
- **Ledger Balances**：`ledger_running_balances` + `ledger_accounts` 在 Go 合併（而非 LEFT JOIN journal_entries）。

---

---

## 現金流量表重構（entry-based category）

### enumx.Enum 零值行為

`enumx.Enum[T]` 實作了 `driver.Valuer`，零值（underlying string 為 `""`）的 `Value()` 回傳 `nil`，
對應 SQL NULL。可直接以非指標型別表達 nullable enum 欄位，不需要 wrapper struct（如 `sql.NullString`）。

- 查詢讀取：`Scan(nil)` 會保留零值，可用 `e.IsZero()` 判斷是否為 NULL。
- 插入 NULL：傳入零值即可，不需要額外處理。

這個行為在 `enumx_test.go` 中有測試：`TestEnum_Value_ZeroValue_ReturnsNil`。

### summary_cf 的 Total 計算原則

entry-based 設計下，`queryCashFlowChanges` 同時回傳葉節點與父科目彙總行。
計算活動小計時，**只能以葉節點行累加**，彙總行僅供前端渲染階層使用。

規則：`if !row.IsSummary { total += amount }`

類似處理：資產負債表 / 損益表的 `if !bsRow.HasChild { totalAssets += balance }` 相同概念。

### 測試輔助函式設計

`sumInsertTxnWithCF(t, db, txnID, date, debitAcct, creditAcct, amount, debitCF, creditCF *string)`
- `debitCF` / `creditCF` 各自獨立設定，允許同一筆交易中借方 INVESTING、貸方不分類（或反之）。
- `sumInsertTxn` 直接呼叫 `sumInsertTxnWithCF(... nil, nil)` 保持向後相容。

### accounts.cash_flow_category 的角色分工

| 值 | 用途 |
|----|------|
| `CASH` | 識別現金科目（`queryCashBefore` / `queryCashUpTo` 使用），**仍由 account 決定** |
| `OPERATING / INVESTING / FINANCING` | 前端填入 `journal_entries` 時的預設建議值，**不再參與計算** |

這個分工讓「哪個科目是現金」保持靜態（帳戶性質），而「每筆分錄屬於哪個活動」保持彈性（使用者可覆寫）。

---

## 直接法現金流量表（Direct Method）

### queryDirectOperatingCash 的 operating_txns 識別邏輯

直接法只需知道「哪些交易屬於營業活動」，然後對其中的 CASH 科目分錄加總借貸。

判斷「營業活動交易」的規則（取聯集）：
1. 交易中包含 `a.type IN ('INCOME', 'EXPENSE')` 的分錄 → 本期損益交易，直接屬於 OPERATING
2. 交易中包含 `je.cash_flow_category = 'OPERATING'` 的分錄 → 明確標記為 OPERATING

這兩個條件對應間接法的兩個計算項目：
- 條件 1 對應「本期淨利（Net Income）」
- 條件 2 對應「調整項（Adjustments）」

### 直接法 vs 間接法的數學恆等性

`Cash Received - Cash Paid = Net Income + OPERATING Adjustments`

這是雙式簿記的必然結果。只要每筆交易都借貸平衡，兩法 Operating Total 恆等，因此可用 `TestGetDirectCashFlowStatement_MatchesIndirectOperatingTotal` 互相驗證。

### 投資 / 籌資活動重用 queryCashFlowChanges

直接法的 InvestingActivities / FinancingActivities 與間接法完全相同：
- 都是依 `journal_entries.cash_flow_category IN ('INVESTING','FINANCING')` 篩選分錄
- 金額公式相同（`credit - debit`）
- `GetDirectCashFlowStatement` 直接重用 `queryCashFlowChanges`，只過濾掉 OPERATING rows

### 已知邊界情況

一筆交易同時含不同 CF 分類（例：CR INCOME 600 + CR INVESTING 400 = DR CASH 1000），
`queryDirectOperatingCash` 因偵測到 INCOME 分錄而將整筆歸入 OPERATING，
現金流入 1000 全部算入 Operating，未按比例拆分。
個人財務場景中此情況極少，視為可接受的邊界行為（見 ARCHITECTURE.md ADR-008）。

---

### 2026-05-15 [REF] 報表整合測試的設計模式

**基礎建設**

所有報表整合測試位於 `internal/repos/query/report_test.go`，使用共用 helpers：
- `testutil.NewTestDB(t)` — in-memory SQLite，自動執行 migration + seed accounts，測試結束自動關閉
- `sumInsertTxn(t, db, txnID, date, debitAcct, creditAcct, amount)` — 插入一筆借貸對稱分錄
- `sumInsertPeriod(t, db, id, type, start, end, closed)` — 建立月結紀錄（closed=true 代表已關帳）
- `snapRepo.BulkInsert(ctx, merchantID, closingID)` — 建立餘額快照（用於快照路徑測試）
- `testCtx()` — 帶 `merchant_id=1` 的 context，與 seed 資料一致

**金額公式對照（避免測試寫錯方向）**

| 科目類型 | 正常方向 | balance / amount | period_change（權益表）|
|---------|---------|-----------------|----------------------|
| ASSET   | DEBIT   | debit − credit  | credit − debit（CF用）|
| LIABILITY | CREDIT | credit − debit  | —                    |
| EQUITY  | CREDIT  | credit − debit  | credit − debit        |
| INCOME  | CREDIT  | credit − debit  | —（損益表用）          |
| EXPENSE | DEBIT   | debit − credit  | —（損益表用）          |

**測試場景分類**

每個查詢函式應涵蓋兩條路徑：
1. **無快照路徑**（`ErrNoRows`）：直接全掃 journal_entries，用 `sumInsertTxn` 建立資料即可
2. **有快照路徑**：先建立月結 `sumInsertPeriod` + `snapRepo.BulkInsert`，再加 delta 交易

目前完成的測試覆蓋：
- `GetBalanceSheet`：無快照（葉/彙總/has_child/depth）、有快照（含/不含 delta）→ 完整
- `GetIncomeStatement`：無快照（葉/彙總）、有快照（月對齊/三月結 delta_pre 非空）→ 完整
- `GetCashFlowStatement`：無快照（期初期末現金/三分類/淨利/NetChange）→ **快照路徑待補**
- `GetEquityStatement`：無快照（期初期間區分/彙總/淨利虛擬行/total_* 合計）→ **快照路徑待補**

---

### 2026-05-15 [DRAFT] 股東權益變動表 SQL 設計要點

**期初／期間金額用單次 CTE 分離**

不用兩段查詢分別取期初和期間，而是在 `period_entries` CTE 中用 `CASE WHEN txn_date < start_date` 和 `CASE WHEN txn_date >= start_date AND txn_date <= end_date` 在同一次掃描分離兩段金額。這樣只需一次 journal_entries 掃描，leaf 和 summary 的計算都能直接 JOIN 這個 CTE。

**彙總節點透過 account_closure 聚合葉節點 period_entries**

summary 科目不直接 JOIN journal_entries（會重複計算），而是 JOIN `account_closure`（depth > 0）取得所有後裔葉節點，再 JOIN `period_entries` CTE 彙總。這與 `v_account_balances` view 的設計邏輯一致。

**total_* 只累計 parent_id IS NULL 的頂層科目**

若對所有科目（含彙總節點）累加，頂層彙總科目已包含子科目的值，再加子科目就會重複計算。只取 `parent_id IS NULL` 的科目代表真正的頂層彙總，等同於對整個權益類別取一次合計。

**本期淨利複用 GetIncomeStatement**

直接呼叫已有的 `GetIncomeStatement` 取得 `NetIncome`，而非重新寫一段 SQL。好處是：快照優化邏輯（`queryIncomeStatementWithSnaps`）自動繼承，不需在 equity 查詢中重新實作。

---

### 2026-05-15 [REF] 資料庫 Schema 新增欄位的完整步驟

每次在既有資料表新增欄位時，**必須同步修改以下五個位置**，缺一不可：

**1. Migration 檔案（`internal/database/migrations/`）**
- 建立新的 migration 檔，命名格式：`YYYYMMDDNNN_描述.sql`
- 使用 `+goose Up` / `+goose Down` 包覆 DDL（ALTER TABLE ADD COLUMN / DROP COLUMN）
- SQLite 的 `ALTER TABLE DROP COLUMN` 需 SQLite 3.35+，確認環境版本

**2. Schema 定義（`internal/database/schema.sql`）**
- 在對應 CREATE TABLE 中加入欄位定義
- 此檔案供 sqlc 靜態分析使用，不實際執行，必須與 migration 的最終狀態一致
- 若有 CHECK constraint，在此一併宣告（sqlc 會讀取 constraint 資訊）

**3. SQL 查詢（`internal/database/queries/*.sql`）**
- SELECT 查詢：將新欄位加入 SELECT 清單（所有讀取該資料表的查詢）
- INSERT 查詢：加入欄位名稱與對應的 `?` 佔位符
- UPDATE 查詢：加入 `欄位名 = ?` 到 SET 子句

**4. sqlc.yaml（若欄位需要型別映射）**
- 若欄位型別需覆寫（如 enum、decimal），在 `overrides` 區段加入：
  ```yaml
  - column: "table_name.column_name"
    go_type:
      import: "套件路徑"
      type: "型別名稱"
  ```
- nullable 欄位搭配 `emit_pointers_for_null_types: true` 會自動生成 `*型別`

**5. Projection Model（`internal/model/db/projection/*.go`）**
- 在對應的 struct 加入欄位定義（型別需與 sqlc.yaml override 一致）
- dbmap-gen 會依 `//dbmap:sqlcdb=...` 注釋自動重新生成轉換函式

**執行順序**
```
修改 migration → 修改 schema.sql → 修改 queries → 修改 sqlc.yaml（視需要）
→ 修改 projection model → go generate ./... → go build ./... → go test ./...
```

**常見錯誤**
- 只改了 schema.sql 沒改 migration：本機重建資料庫正常，但線上升級時欄位不存在
- 只改了 migration 沒改 schema.sql：sqlc 看不到欄位，生成的程式碼缺少對應欄位
- SELECT 查詢漏加新欄位：dbmap-gen 生成的 `FromRow` 函式取不到值，欄位永遠是零值

---

### 2026-05-15 [DEBUG] sqlc nullable enum 欄位的指標生成機制

問題：想讓 `accounts.cash_flow_category`（nullable TEXT）在 sqlcdb 中生成 `*enums.CashFlowCategory` 而非 `*string`。

關鍵發現：
- `sqlc.yaml` 已設定 `emit_pointers_for_null_types: true`
- 在 `overrides` 中宣告 `column: "accounts.cash_flow_category"` → `enums.CashFlowCategory`（不帶指標）
- sqlc 會自動合併兩者，最終生成 `*enums.CashFlowCategory`
- 不需要在 Go 端額外包裝 `EnumPtr` helper，也不需要 `NullEnum[T]`

結論：只要 `emit_pointers_for_null_types: true` 且 override 宣告基底型別，sqlc 自動處理指標化。

---

### 2026-05-15 [DRAFT] 現金流量表間接法公式推導

核心公式：非 CASH 科目的現金流量影響 = `period_credit − period_debit`

推導：
- 資產科目（借方正常）：資產增加 → 借方增加 → debit > credit → 結果為負 → 現金流出 ✓
- 負債科目（貸方正常）：負債增加 → 貸方增加 → credit > debit → 結果為正 → 現金流入 ✓
- 費用科目（借方正常）：費用發生 → debit 增加 → 結果為負 → 減少淨利，但本身代表非現金調整（如折舊）
- 收入科目（貸方正常）：已在淨利中反映，OPERATING 分類的收入調整項同理適用

此公式不需依科目類型分支處理，一條公式通吃所有分類，是 SQL 可以直接計算的形式。

---

### 2026-05-15 [IDEA] 現金流量表現有設計的已知限制

1. 非現金交易（如折舊費用）目前無法從科目資料自動識別是否為非現金項目，需人工標記或未來加欄位
2. `cash_flow_category` 目前只有 seed 預設值，若使用者新增科目沒有設定，該科目不會出現在報表中（靜默略過）
3. 期初現金計算用 `start_date`（exclusive），期末用 `end_date`（inclusive），與 IncomeStatement 的慣例一致，但需注意日期邊界

<!--
### YYYY-MM-DD [分類] 標題
內容（自由格式）
-->
