# Architecture Decision Records (ADR)

---

## ADR-001：現金流量表分類來源改為 journal_entries

**日期：** 2026-05-16
**狀態：** 已採用

### 背景

原始設計將 `cash_flow_category`（OPERATING / INVESTING / FINANCING）掛在 `accounts` 資料表，
導致同一科目下的所有分錄必須歸入同一個現金流量類別，無法針對個別交易彈性分類。

例如：股利支付在 IFRS 可歸入 OPERATING 或 FINANCING；
應收帳款依交易性質可能是 OPERATING 或 INVESTING。

### 決策

將 OPERATING / INVESTING / FINANCING 分類移至 `journal_entries.cash_flow_category`（nullable）。
每筆分錄可獨立設定，前端依所選科目的 `accounts.cash_flow_category` 作為預填預設值，使用者可自行修改。

`accounts.cash_flow_category` 保留以下用途：
- `CASH`：識別現金及約當現金科目，用於計算期初 / 期末現金餘額（`queryCashBefore` / `queryCashUpTo`）。
- `OPERATING / INVESTING / FINANCING`：提供前端預設值參考，不再參與計算。

### 取捨

| 面向 | 舊設計（account-based） | 新設計（entry-based） |
|------|------------------------|----------------------|
| 彈性 | 低，同科目只能一個分類 | 高，每筆分錄獨立設定 |
| 資料完整性 | account 設定後即生效 | 需前端 / pipeline 主動填入 |
| 系統自動分錄 | 靠科目自動帶入 | 需逐案評估並補上（見 TODO.md） |
| 查詢複雜度 | 較低 | 略高，需 GROUP BY (account_id, cash_flow_category) |

### 不採用方案

- **保留 account-based**：無法支援同一科目多分類需求，排除。
- **同時維護兩個來源**：增加維護成本與資料一致性風險，排除。

---

## ADR-002：直接法現金流量表（Direct Method）架構設計

**日期：** 2026-05-16
**狀態：** 已採用

### 背景

間接法（Indirect Method）以本期淨利為起點加減調整項，適合財務人員分析，但不直觀。
直接法（Direct Method）直接呈現實際現金收入（CashReceived）與現金支出（CashPaid），
對個人財務使用者更容易理解，且兩法 Operating Total 數學恆等，可作為交互驗證手段。

### 決策

新增獨立的 `GetDirectCashFlowStatement` 方法，與間接法並存，提供不同使用情境：
- 間接法（`/report/cash_flow_statement`）：財務分析、損益拆解
- 直接法（`/report/cash_flow_statement_direct`）：現金收支明細、兩法驗證

#### Operating 活動現金計算方式

1. 識別「營業活動交易（operating_txns）」—— 取兩個條件的聯集：
   - 含 `a.type IN ('INCOME','EXPENSE')` 分錄的交易（損益交易）
   - 含 `je.cash_flow_category = 'OPERATING'` 分錄的交易（非損益的OPERATING調整）
2. 對這些交易中所有 `accounts.cash_flow_category = 'CASH'` 的分錄加總借方與貸方：
   - `CashReceived = SUM(debit)` （現金流入）
   - `CashPaid = SUM(credit)` （現金流出，以正數呈現）
   - `Total = CashReceived - CashPaid`

#### Investing / Financing 重用間接法查詢

直接法的投資 / 籌資活動數字與間接法完全相同，因此：
- 直接重用 `queryCashFlowChanges`，只跳過 OPERATING rows 不累入 Total
- 無需另寫 SQL，維護成本最低

#### 期初 / 期末現金重用

`queryCashBefore` / `queryCashUpTo` 兩個 CASH 科目聚合查詢被兩法共用，無需複製。

### 取捨

| 面向 | 直接法 | 間接法 |
|------|--------|--------|
| 使用者易讀性 | 高（顯示實際現金流） | 中（需理解調整項意義） |
| 查詢複雜度 | 略高（operating_txns 兩層 CTE） | 低（直接聚合 OPERATING entries） |
| 兩法可互相驗證 | ✓ Operating Total 恆等 | ✓ |
| 混合分類交易精準度 | 邊界行為（整筆歸入 OPERATING） | 精確（按 entry 分類） |

### 已知限制與可接受邊界行為

一筆交易同時含不同 CF 分類的分錄（如：CR INCOME 600 + CR INVESTING 400 = DR CASH 1000），
`queryDirectOperatingCash` 會將整筆交易歸入 OPERATING（因偵測到 INCOME 分錄），
現金流入 1000 全算入 CashReceived 而非按比例拆分 600/400。

**判斷**：個人財務情境中幾乎不會出現此類混合交易，視為可接受的邊界行為，不增加查詢複雜度處理。

### 不採用方案

- **改寫間接法 SQL 同時輸出直接法欄位**：兩個呈現目的不同，合併會讓查詢難以維護，排除。
- **在 Service 層從間接法結果換算直接法**：跳過了「實際現金移動」的查詢，無法正確還原直接法，排除。
