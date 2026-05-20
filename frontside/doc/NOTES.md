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

### 2026-05-20｜Setting Config 串接模式

#### 唯讀顯示欄用 `<span>` 完全避免 a11y 警告

`<label>` 元素需對應一個可互動控制項（`for` + `id`），否則 svelte-check 報 `a11y_label_has_associated_control`。
若該欄只用來顯示文字（如 modal 中唯讀的「帳戶類型」），直接改用 `<span class="form-label">` 即可，不需要 `for`，也不產生警告。

```svelte
<!-- ✅ 唯讀顯示：改用 <span>，不需要 for/id -->
<span class="form-label">帳戶類型</span>
<p style="font-size:13px;color:#dedad3;margin:0;padding:8px 0;">{LABELS[type]}</p>

<!-- ✅ 對應 AccountSelect 的 label：加 for 屬性（任意值即可） -->
<label class="form-label" for="some-id">配置科目 *</label>
<AccountSelect ... />

<!-- ❌ 唯讀顯示用 <label> 但無 for → 觸發 a11y 警告 -->
<label class="form-label">帳戶類型</label>
<p>...</p>
```

> 見 ISSUE-004 的解決紀錄（補 `for` + `id` 的做法）與本次更乾淨的替代：唯讀欄改 `<span>`。

#### Setting API 為非 Event Sourcing 的直接 REST 端點

Setting 相關 API（`/api/setting/*`）是直接的 REST PUT/GET，不走 Event Sourcing 的 `appendEvent` 流程。
與 Account、LedgerAccount 等必須用 `appendEvent` 的不同，這類「系統設定」通常只維護最新狀態，無版本號樂觀鎖。

```typescript
// ✅ Setting API：直接 PUT
await apiFetch(`/api/setting/ledger-account-type/${type}`, {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ account_id: accountId }),
});

// ❌ 不走 appendEvent
await appendEvent({ event_type: 'setting.updated', ... });
```

#### 以 Map 補全後端「可能不回傳所有 type」的顯示清單

後端 setting API 只回傳已設定的 type（未設定的 type 根本不存在於資料庫），
但前端需要顯示所有 type 的 row（未設定者顯示「—」）。
標準做法：以固定陣列為基準，用 `configMap.get(type)` 取值，取不到就顯示「—」。

```typescript
const ALL_TYPES = ['BANK_ACCOUNT', 'CREDIT_CARD', 'LOAN'] as const;
const configMap = $derived(new Map(configs.map(c => [c.type, c])));

// template 中
{#each ALL_TYPES as type}
  {@const cfg = configMap.get(type)}
  <td>{cfg?.account_id || '—'}</td>
{/each}
```

---

### 2026-05-19｜Projection audit 欄位補充模式

#### 唯讀審計欄（更新者 / 更新時間）顯示模式

Modal 內顯示唯讀審計資料的標準寫法：

```svelte
{#if mode === 'edit'}
  {@const item = items.find(i => i.id === editId)}
  {#if item}
    <div class="form-row">
      <div class="form-group">
        <label class="form-label" for="f-xxx-updated-by">更新者</label>
        <input id="f-xxx-updated-by" class="form-input" type="text"
               value={item.updated_by ?? '—'} readonly
               style="background:#f5f0e8;cursor:default;" />
      </div>
      <div class="form-group">
        <label class="form-label" for="f-xxx-updated-at">更新時間</label>
        <input id="f-xxx-updated-at" class="form-input" type="text"
               value={item.updated_at ?? '—'} readonly
               style="background:#f5f0e8;cursor:default;" />
      </div>
    </div>
  {/if}
{/if}
```

注意：`<label>` 一定要有 `for`，對應 `<input id="..."`，否則 `npm run check` 報 `a11y_label_has_associated_control`（見 ISSUE-004）。

#### 自訂 grid 子表格新增欄時的同步清單

凡是使用 `grid-template-columns` 的子表格（`inv-lot-row`、`inv-movement-row`、`inst-payment-item` 等），新增欄位需同步修改兩處：

1. **CSS** — `src/styles/views/<檔案>.css`：在 header + row 的 `grid-template-columns` 加上對應寬度
2. **HTML** — view 檔案的 header `<span>` 與 row `<span>` 各補一個

若只改其中一處，欄位會錯位或多出空白欄，視覺上難以察覺。

---

### 後端合作慣例

#### 新增後端欄位時的前端 Checklist

後端 Go struct 新增欄位，前端需逐一確認：

1. `src/types/` 對應 interface 補欄位（注意 nullable 轉換，見 ADR-002）
2. API payload interface 若為送出用途，也需同步更新（`TransactionEntryPayload` 等）
3. 若欄位為 enum：在 `src/types/` 加型別定義，並加顯示用 label 對照表（放在同一檔案）
4. 若欄位有預設值來源（如從關聯資料自動帶入），在 UI 補自動填入邏輯
5. 表單 UI 補對應輸入元件

#### 確認後端 Event Type 字串

呼叫 `appendEvent()` 時的 `event_type` 必須與後端 Event Handler 的 case 字串完全一致。
不確定時，查後端 `internal/` 的 event handler 或 `doc/` 文件，不可自行猜測。

已知 event type：
- `transaction.created`
- `transaction.corrected`
- `transaction.voided`

---

### 2026-05-16｜新增交易分錄 cash_flow_category

#### 設計決策

**共用常數放置位置**

`CASH_FLOW_CATEGORIES` 與 `CASH_FLOW_CATEGORY_LABELS` 集中於 `src/types/account.ts`，
與 `CashFlowCategory` 型別定義並排，理由是「型別與對應的顯示文字應該住在一起」。
若未來需要在更多地方顯示現金流量類別，直接 import 即可，不會出現多處定義不一致的問題。

#### `StringFormLineField` 型別技巧

`FormLine` 加入 `cash_flow_category: CashFlowCategory | null` 後，
原本的 `updateLine(id, field: keyof FormLine, value: string)` 會有型別問題——
TypeScript 無法保證 `string` 值可以合法指派給所有 `keyof FormLine` 對應的屬性。

解法：定義 `type StringFormLineField = 'account_id' | 'ledger_id' | 'debit' | 'credit'`，
讓 `updateLine` 只接受字串欄位；`cash_flow_category` 由專屬的 `updateLineCashFlow()` 處理。

#### 自動帶入邏輯

- **選擇科目（AccountSelect）**：`onselect` 中查 `activeAccounts`，帶入 `acct.cash_flow_category ?? null`
- **選擇金融帳戶（LedgerSelect）**：`selectLedger()` 中，先取帳戶連結的 `account_id`，再查 `allAccounts` 取得 category
  - 注意：此處查的是 `allAccounts`（含停用），而非 `activeAccounts`，確保帳戶關聯的科目即使停用也能正確帶入
- 使用者可在下拉選單手動覆蓋自動帶入的值

---

### 2026-05-15 [DRAFT] 前後端界接準則

這份準則來自實際開發經驗，目的是讓新增 API 串接時少踩坑。

#### 1. 永遠先讀後端 model，再寫 TypeScript 型別

拿到後端 Go 的 struct，逐欄對應：

| Go 型別 | TypeScript 型別 | 備註 |
|---------|----------------|------|
| `string` | `string` | |
| `bool` | `boolean` | |
| `int`, `int64` | `number` | |
| `decimal.Decimal` | `string` | 精度保留，顯示時 `parseFloat` + `toLocaleString` |
| `*T`（pointer） | `T \| null` | |
| `time.Time` | `string` | 格式 `YYYY-MM-DD` 或 ISO8601 |
| 省略欄位（`omitempty`） | `T?`（optional） | 不一定出現，需 `?? default` |

欄位名稱以 json tag 為準（snake_case），不是 Go struct field 名稱。

#### 2. API 呼叫統一用 `apiFetch`，絕不直接 `fetch`

`apiFetch` 已封裝 Authorization header 與 base URL，直接 `fetch` 會漏帶 token。

```typescript
// ❌
const res = await fetch('/api/report/...');

// ✅
const res = await apiFetch('/api/report/...');
```

#### 3. 錯誤回應統一用 problem+JSON 格式解析

後端錯誤固定回傳 `{ title, detail }` 結構，統一用以下 fallback chain：

```typescript
const problem = await response.json();
throw new Error(problem.detail ?? problem.title ?? '預設錯誤訊息');
```

`detail` 優先（更具體），`title` 次之（通用），最後 fallback 到自訂字串。

#### 4. 日期一律用 string，不用 `Date` 物件

API 傳輸與本地狀態都用 `string`（`YYYY-MM-DD`），僅在 `<input type="date">` 的 `bind:value` 時自然對應。
禁止在 service 層或 state 中轉成 `Date` 物件再轉回，避免時區問題。

#### 5. decimal 金額顯示統一用 `fmt()`（`src/lib/reportUtils.svelte.ts`）

```typescript
// ❌ 各處自行 parseFloat + toLocaleString
parseFloat(item.amount).toLocaleString()

// ✅
fmt(item.amount)  // 已處理 NaN fallback
```

#### 6. 新增 API 端點的標準流程

1. 讀後端 model Go 檔案 → 確認回應結構
2. 在 `src/types/<模組>.ts` 新增 interface
3. 在 `src/api/<模組>.ts` 新增 async function
4. 建立或更新 view 元件
5. 在 `src/routes/Home.svelte` 加路由與選單項目（可展開子選單需同時更新 `parentBasePaths`）
6. `npm run check` → `npm run build` 驗證

**注意：若後端 API 有「可能不回傳所有項目」的特性**（如 setting 類 API，只回傳已設定的 type），
前端需以固定 ALL_TYPES 陣列為基準補全顯示，而非直接 `{#each configs}` 迭代（見 NOTES 2026-05-20）。

---

### 2026-05-15 [DEBUG] Svelte 5 `$effect` 無限迴圈模式

在 `$effect` 內不要對「effect 本身追蹤的 state」做賦值，常見觸發條件：

```svelte
// ❌ 錯誤：$effect 追蹤 mySet，賦值新 Set 又觸發 effect 重執行
$effect(() => {
  if (someCondition) {
    mySet = new Set([...mySet, 'value']); // 讀了 mySet 又寫 mySet
  }
});

// ✅ 正確：改用 $derived，純計算不產生副作用
const isExpanded = $derived(mySet.has('value') || someCondition);
```

適用場景：sidebar sub-menu 展開狀態要同時反映「手動切換」與「目前路由」時，用 `$derived` OR 兩個來源，比 `$effect` 更安全清晰。

---

### 2026-05-15 [DEBUG] `replace_all` Edit 工具的使用限制

`replace_all: true` 會替換檔案內所有符合的子字串，包含宣告式自身。若要替換的字串出現在新宣告的 `const` 初始值內，會造成循環參考。

**規則**：`replace_all` 只用於「確定整個檔案內所有出現都要換」的場景（如重新命名變數）。若新宣告包含舊字串，改用多次精確 `old_string` 替換。

<!--
### YYYY-MM-DD [分類] 標題
內容（自由格式）
-->
