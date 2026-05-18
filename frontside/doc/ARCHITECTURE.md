# Architecture Decision Records (ADR)

系統長期架構決策紀錄。每筆 ADR 記錄「為什麼這樣設計」，包含被拒絕的替代方案與取捨理由。
**ADR 一旦建立不得刪除**，狀態變更時更新 `狀態` 欄位即可。

---

## 狀態說明

| 狀態 | 說明 |
|------|------|
| Proposed | 提案中，尚未決定 |
| Accepted | 已採用 |
| Deprecated | 曾採用，已廢棄（被新 ADR 取代） |
| Superseded by ADR-XXX | 被指定 ADR 取代 |

---

## 決策清單

| 編號 | 標題 | 狀態 | 日期 |
|------|------|------|------|
| ADR-001 | 變更操作透過 Event Sourcing 送出 | Accepted | — |
| ADR-002 | TypeScript 型別與後端 Go struct 對齊慣例 | Accepted | — |
| ADR-003 | CashFlowCategory 在帳目與分錄兩個層級的設計 | Accepted | — |
| ADR-004 | 財務報表導覽從 tab-nav 改為 sidebar sub-menu | Accepted | 2026-05-15 |
| ADR-005 | 財務報表視圖拆分為獨立元件並建立共用工具層 | Accepted | 2026-05-15 |
| ADR-006 | Svelte 5 衍生 UI 狀態一律使用 `$derived` 而非 `$effect` | Accepted | 2026-05-15 |
| ADR-007 | 後端 `decimal.Decimal` 對應前端 `string`，顯示層才轉數字 | Accepted | 2026-05-15 |

---

## ADR-001：變更操作透過 Event Sourcing 送出

- **狀態**：Accepted
- **日期**：—
- **背景**：
  後端採用 Event Sourcing 架構，所有狀態變更（新增交易、更正、作廢）不走傳統 REST PATCH/POST，
  而是透過統一的 Event Store 端點寫入事件。
- **決策**：
  前端所有「寫入」操作一律呼叫 `src/api/aggregate.ts` 的 `appendEvent()`，
  而非直接打對應資源的 REST endpoint。流程如下：

  1. 呼叫 `getAggregateVersion(aggregateType, aggregateId)` 取得目前版本號
  2. 呼叫 `appendEvent({ aggregate_type, aggregate_id, expected_version, event_type, payload, metadata })`
  3. 後端驗證版本號（樂觀鎖），事件合法則寫入並更新投影
- **替代方案**：
  - **直接使用 REST PATCH/POST**：傳統做法，但與後端架構不符，無法享有樂觀鎖與事件歷史。
- **後果**：
  - 正面：前端 payload 結構與後端 Event Payload struct 一一對應，型別對齊直覺。
  - 負面：新增任何寫入功能前，需先確認後端對應的 `event_type` 字串與 payload 結構，不可自行假設。
  - 注意：`aggregate_id` 可為空字串 `''`（交易類聚合不需要特定 ID）。

---

## ADR-002：TypeScript 型別與後端 Go struct 對齊慣例

- **狀態**：Accepted
- **日期**：—
- **背景**：
  後端使用 Go，前端使用 TypeScript strict mode。兩者的型別系統不同，需要建立明確的對齊規則，
  避免欄位遺漏或型別錯誤導致執行期錯誤。
- **決策**：
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
- **替代方案**：
  - **各元件自行轉換**：分散在各處，難以維護，欄位遺漏風險高。
- **後果**：
  - 正面：型別對齊有明確規則，新增欄位時有依據；後端 enum 新增值時，前端 `type` 定義與 label 對照表都需同步更新。
  - 負面：新增或修改後端欄位時，前端 `src/types/` 需同步更新；Payload 送出數字、查詢回傳字串，同一欄位在不同 interface 中型別不同，需注意區分。

---

## ADR-003：CashFlowCategory 在帳目與分錄兩個層級的設計

- **狀態**：Accepted
- **日期**：—
- **背景**：
  現金流量分類（`cash_flow_category`）同時存在於：
  - **科目（Account）層級**：作為該科目的預設分類，由帳目管理員設定
  - **交易分錄（Entry）層級**：實際記帳時每筆分錄可指定，覆蓋科目預設值
- **決策**：
  前端新增交易時的自動帶入邏輯：
  1. 使用者選擇科目 → 自動帶入該科目的 `cash_flow_category`（可為 null）
  2. 使用者選擇金融帳戶（LedgerAccount）→ 透過帳戶連結的 `account_id` 查出科目，再帶入 category
  3. 使用者可在下拉選單手動覆蓋，最終以分錄層級的值為準送出

  查詢科目時使用 `allAccounts`（含停用科目），而非 `activeAccounts`，
  確保即使科目停用，關聯資料仍能正確帶入。

  共用常數 `CASH_FLOW_CATEGORIES` 與 `CASH_FLOW_CATEGORY_LABELS` 集中於 `src/types/account.ts`，避免多處定義不一致。
- **替代方案**：
  - **只在分錄層級設定，不繼承科目預設**：使用者每次都需手動選擇，輸入負擔較高。
- **後果**：
  - 正面：減少使用者手動填寫的負擔，提升輸入速度。
  - 負面：若科目未設定 category，自動帶入 null，使用者需自行選擇。

---

## ADR-004 財務報表導覽從 tab-nav 改為 sidebar sub-menu

- **狀態**：Accepted
- **日期**：2026-05-15
- **背景**：
  原始 `Reports.svelte` 在頁面內用 tab-nav（資產負債表 / 損益表）切換報表種類。新增現金流量表後，tab 數量增加，且報表種類已是明確的獨立功能，不適合繼續用 tab 混在同一頁面。導覽邏輯也與其他模組（儀表板、投資管理等）不一致。
- **決策**：
  將三份報表（資產負債表、損益表、現金流量表）各自對應獨立路由（`/home/reports/balance-sheet`、`/home/reports/income-statement`、`/home/reports/cash-flow`），並在 sidebar「財務報表」項目下增設可展開的 sub-menu，與其他頂層選單項目的導覽方式保持一致。
- **替代方案**：
  - **保留 tab-nav、加入第三個 tab**：改動最小，但 tab 數量增加使版面更擠，且每份報表無法直接以 URL 書籤存取，不利分享與回上一頁。
  - **獨立頂層選單項目（不巢狀）**：與現有「投資管理」、「分期管理」平行列出三份報表，選單項目過多，語意上三份報表屬同一財務報表群組，不應拆散。
- **後果**：
  - 正面：每份報表有獨立 URL，可書籤；導覽一致性提升；擴充第四份報表只需新增 sub-item 與路由。
  - 負面：sidebar 新增巢狀層級，需處理展開/收合狀態；路由數量增加。

---

## ADR-005 財務報表視圖拆分為獨立元件並建立共用工具層

- **狀態**：Accepted
- **日期**：2026-05-15
- **背景**：
  原 `Reports.svelte` 約 430 行，同時包含資產負債表與損益表的所有狀態、圖表邏輯與 HTML。加入現金流量表後估計超過 600 行，遠超 CLAUDE.md 規範的 200 行上限，可讀性與維護性下降。
- **決策**：
  - 每份報表獨立為一個 Svelte 元件（`BalanceSheet.svelte`、`IncomeStatement.svelte`、`CashFlowStatement.svelte`）。
  - 抽出共用邏輯到 `src/lib/reportUtils.svelte.ts`（`Row` 介面、`visibleRows()` 樹狀可見列過濾、`fmt()` 數字格式化），符合 CLAUDE.md「共用邏輯抽成 `src/lib/` 下的 `.svelte.ts` 檔」規範。
  - 刪除原 `Reports.svelte`，`/home/reports` 路由直接指向 `BalanceSheet`。
- **替代方案**：
  - **保留單一 `Reports.svelte`、以路由參數切換**：所有三份報表仍在同一檔案，僅讀取 `window.location.hash` 決定顯示哪份。改動較少，但檔案持續膨脹，且三份報表的獨立狀態（loading、error、result）混雜在同一作用域。
- **後果**：
  - 正面：各元件職責單一；`reportUtils.svelte.ts` 集中共用邏輯，避免複製貼上；元件大小均在 200 行以內。
  - 負面：檔案數量增加；若日後需要跨報表共用某種視圖元件（例如通用表格），需再次抽取。

---

## ADR-006 Svelte 5 衍生 UI 狀態一律使用 `$derived` 而非 `$effect`

- **狀態**：Accepted
- **日期**：2026-05-15
- **背景**：
  實作 sidebar sub-menu 自動展開功能時，最初使用 `$effect` 監聽路由變化並寫入展開狀態 Set。由於 `$effect` 同時讀取（`[...expandedParents]`）與寫入（`expandedParents = new Set(...)`）同一個 `$state`，觸發 `effect_update_depth_exceeded` 無限迴圈，導致元件無法正常運作（見 ISSUE-002）。
- **決策**：
  凡是「可以從現有 state 直接計算出來」的 UI 狀態（例如「某個 sub-menu 是否展開」），一律使用 `$derived` 宣告，不透過 `$effect` 寫入中間狀態。`$effect` 僅用於真正的副作用（DOM 操作、事件監聽、外部 API 呼叫等）。

  ```svelte
  // ✅ 正確：$derived 純計算，不產生副作用
  const isReportsExpanded = $derived(
    expandedParents.has('reports') || currentPath.startsWith('/home/reports')
  );
  ```

- **替代方案**：
  - **`$effect` + `untrack`**：用 Svelte 提供的 `untrack()` 包裹寫入，使 effect 不追蹤該 state 的變化，可解除迴圈。但語意較不直觀，且 `untrack` 屬繞過響應式系統的逃生出口，能用 `$derived` 解決就不應動用。
- **後果**：
  - 正面：`$derived` 語意清晰（「這個值由那些 state 決定」）；不可能發生因寫入觸發自身的迴圈；程式碼量更少。
  - 負面：若衍生計算涉及非同步操作，`$derived` 無法直接使用，仍需 `$effect`。

---

## ADR-007 後端 `decimal.Decimal` 對應前端 `string`，顯示層才轉數字

- **狀態**：Accepted
- **日期**：2026-05-15
- **背景**：
  後端使用 `github.com/shopspring/decimal` 做高精度金融運算，序列化後為字串（例如 `"12345.67"`）。前端需要決定：在 TypeScript 型別、state 與顯示層分別用什麼型別存放金額。
- **決策**：
  - TypeScript 型別：金額欄位一律宣告為 `string`，與 API JSON 一對一對應，不在型別層做轉換。
  - State：維持 `string`，不在 `$state` 或 `$derived` 中提前轉成 `number`。
  - 顯示層：需要格式化時才呼叫 `fmt()`（`parseFloat` + `toLocaleString`），其餘需要比較大小才用 `parseFloat()`。
  - 共用工具：`fmt()` 集中在 `src/lib/reportUtils.svelte.ts`，所有報表元件統一呼叫。
- **替代方案**：
  - **型別宣告為 `number`，API 層轉換**：在 `apiFetch` 回傳前遞迴轉換所有金額欄位為 `number`。代價是：轉換邏輯難以維護（哪些欄位是金額？）；JavaScript `number` 有 IEEE 754 精度限制，大金額可能失真；轉換層與業務層耦合。
  - **型別宣告為 `Decimal` 物件（引入 decimal.js）**：保留精度，但增加套件依賴，且目前前端不做運算，無此需求。
- **後果**：
  - 正面：型別與 API 一對一，zero-overhead；不引入精度問題；`fmt()` 集中管理，修改格式只改一處。
  - 負面：每次需要顯示金額都要呼叫 `fmt()`，不能直接輸出；若未來需要前端加減運算，需明確 `parseFloat`，不夠直覺。

---


<!-- 新增 ADR 時複製以下範本 -->

<!--
## ADR-XXX 標題

- **狀態**：Proposed
- **日期**：YYYY-MM-DD
- **背景**：
  為什麼需要做這個決策？當時面臨什麼問題或限制？
- **決策**：
  最終採用的方案是什麼？
- **替代方案**：
  - 方案 A：為何不採用？
  - 方案 B：為何不採用？
- **後果**：
  - 正面：這個決策帶來哪些好處？
  - 負面：這個決策引入哪些代價或限制？
-->
