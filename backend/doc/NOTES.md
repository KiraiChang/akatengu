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
