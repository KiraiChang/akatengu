# NOTES

開發中的臨時想法、設計草稿、Debug 紀錄。

---

## 後端合作慣例

### 新增後端欄位時的前端 Checklist

後端 Go struct 新增欄位，前端需逐一確認：

1. `src/types/` 對應 interface 補欄位（注意 nullable 轉換，見 ADR-002）
2. API payload interface 若為送出用途，也需同步更新（`TransactionEntryPayload` 等）
3. 若欄位為 enum：在 `src/types/` 加型別定義，並加顯示用 label 對照表（放在同一檔案）
4. 若欄位有預設值來源（如從關聯資料自動帶入），在 UI 補自動填入邏輯
5. 表單 UI 補對應輸入元件

### 確認後端 Event Type 字串

呼叫 `appendEvent()` 時的 `event_type` 必須與後端 Event Handler 的 case 字串完全一致。
不確定時，查後端 `internal/` 的 event handler 或 `doc/` 文件，不可自行猜測。

已知 event type：
- `transaction.created`
- `transaction.corrected`
- `transaction.voided`

---

## 2026-05-16｜新增交易分錄 cash_flow_category

### 設計決策

**共用常數放置位置**

`CASH_FLOW_CATEGORIES` 與 `CASH_FLOW_CATEGORY_LABELS` 集中於 `src/types/account.ts`，
與 `CashFlowCategory` 型別定義並排，理由是「型別與對應的顯示文字應該住在一起」。
若未來需要在更多地方顯示現金流量類別，直接 import 即可，不會出現多處定義不一致的問題。

### `StringFormLineField` 型別技巧

`FormLine` 加入 `cash_flow_category: CashFlowCategory | null` 後，
原本的 `updateLine(id, field: keyof FormLine, value: string)` 會有型別問題——
TypeScript 無法保證 `string` 值可以合法指派給所有 `keyof FormLine` 對應的屬性。

解法：定義 `type StringFormLineField = 'account_id' | 'ledger_id' | 'debit' | 'credit'`，
讓 `updateLine` 只接受字串欄位；`cash_flow_category` 由專屬的 `updateLineCashFlow()` 處理。

### 自動帶入邏輯

- **選擇科目（AccountSelect）**：`onselect` 中查 `activeAccounts`，帶入 `acct.cash_flow_category ?? null`
- **選擇金融帳戶（LedgerSelect）**：`selectLedger()` 中，先取帳戶連結的 `account_id`，再查 `allAccounts` 取得 category
  - 注意：此處查的是 `allAccounts`（含停用），而非 `activeAccounts`，確保帳戶關聯的科目即使停用也能正確帶入
- 使用者可在下拉選單手動覆蓋自動帶入的值
