# NOTES

臨時想法、Debug 筆記、架構草稿等非正式內容。

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
個人財務場景中此情況極少，視為可接受的邊界行為（見 ARCHITECTURE.md ADR-002）。
