# ARCHITECTURE

長期架構決策，採用 ADR（Architecture Decision Record）格式。

格式：
- **狀態**：Proposed / Accepted / Deprecated / Superseded
- **背景**：為何需要做這個決策
- **決策**：採用什麼方案
- **後果**：帶來的影響（正面與負面）

---

## ADR-001：變更操作透過 Event Sourcing 送出

**狀態**：Accepted

**背景**

後端採用 Event Sourcing 架構，所有狀態變更（新增交易、更正、作廢）不走傳統 REST PATCH/POST，
而是透過統一的 Event Store 端點寫入事件。

**決策**

前端所有「寫入」操作一律呼叫 `src/api/aggregate.ts` 的 `appendEvent()`，
而非直接打對應資源的 REST endpoint。流程如下：

1. 呼叫 `getAggregateVersion(aggregateType, aggregateId)` 取得目前版本號
2. 呼叫 `appendEvent({ aggregate_type, aggregate_id, expected_version, event_type, payload, metadata })`
3. 後端驗證版本號（樂觀鎖），事件合法則寫入並更新投影

**後果**

- 正面：前端 payload 結構與後端 Event Payload struct 一一對應，型別對齊直覺
- 負面：新增任何寫入功能前，需先確認後端對應的 `event_type` 字串與 payload 結構，不可自行假設
- 注意：`aggregate_id` 可為空字串 `''`（交易類聚合不需要特定 ID）

---

## ADR-002：TypeScript 型別與後端 Go struct 對齊慣例

**狀態**：Accepted

**背景**

後端使用 Go，前端使用 TypeScript strict mode。兩者的型別系統不同，需要建立明確的對齊規則，
避免欄位遺漏或型別錯誤導致執行期錯誤。

**決策**

前端 `src/types/` 內的 interface 欄位必須與後端對應 Go struct 的 JSON tag 一致，型別轉換規則如下：

| Go 型別 | TypeScript 型別 |
|---|---|
| `string` | `string` |
| `int64` / `int32` | `number` |
| `*string`（nullable） | `string \| null` |
| `*int64`（nullable） | `number \| null` |
| `decimal.Decimal` | `string`（API 回傳字串）或 `number`（payload 送出數字） |
| `*enums.XxxEnum`（nullable enum） | `XxxEnum \| null` |
| Go enum（`type XxxEnum string`） | TypeScript `type XxxEnum = 'A' \| 'B' \| ...` |

**後果**

- 新增或修改後端欄位時，前端 `src/types/` 需同步更新
- 後端 enum 新增值時，前端 `type` 定義與對應的 label 對照表都需更新
- Payload 送出數字（`debit: number`），但 API 查詢回傳字串（`debit: string`），同一欄位在不同 interface 中型別不同，需注意區分

---

## ADR-003：CashFlowCategory 在帳目與分錄兩個層級的設計

**狀態**：Accepted

**背景**

現金流量分類（`cash_flow_category`）同時存在於：
- **科目（Account）層級**：作為該科目的預設分類，由帳目管理員設定
- **交易分錄（Entry）層級**：實際記帳時每筆分錄可指定，覆蓋科目預設值

**決策**

前端新增交易時的自動帶入邏輯：
1. 使用者選擇科目 → 自動帶入該科目的 `cash_flow_category`（可為 null）
2. 使用者選擇金融帳戶（LedgerAccount）→ 透過帳戶連結的 `account_id` 查出科目，再帶入 category
3. 使用者可在下拉選單手動覆蓋，最終以分錄層級的值為準送出

查詢科目時使用 `allAccounts`（含停用科目），而非 `activeAccounts`，
確保即使科目停用，關聯資料仍能正確帶入。

**後果**

- 減少使用者手動填寫的負擔，提升輸入速度
- 若科目未設定 category，自動帶入 null，使用者需自行選擇
- 共用常數 `CASH_FLOW_CATEGORIES` 與 `CASH_FLOW_CATEGORY_LABELS` 集中於 `src/types/account.ts`，避免多處定義不一致
