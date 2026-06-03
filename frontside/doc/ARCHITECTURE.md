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
| ADR-008 | 所有 Projection 模型一律加入 `updated_by` / `updated_at` 審計欄位 | Accepted | 2026-05-19 |
| ADR-009 | Sidebar 可展開選單項以函式泛化展開邏輯 | Accepted | 2026-05-20 |
| ADR-010 | 帳本建立科目限制透過 `filteredAccounts` prop 傳遞而非在子元件內載入 | Accepted | 2026-05-20 |
| ADR-011 | 跨頁面科目導航使用 URL hash query param 傳遞預選科目 | Accepted | 2026-05-26 |
| ADR-012 | 登入後以 pendingToken 暫存 user token，選商戶後換 merchant token | Accepted | 2026-06-03 |

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

---

## ADR-008 所有 Projection 模型一律加入 `updated_by` / `updated_at` 審計欄位

- **狀態**：Accepted
- **日期**：2026-05-19
- **背景**：
  後端 `internal/model/db/projection/` 的所有查詢用 struct（Account、LedgerAccount、Transaction、Entry、Installment、InstallmentPayment、Investment、InvestmentLot、InvestmentMovement、PeriodClosing）補入了 `updated_at *string` 與 `updated_by *string`。前端需決定是否在 UI 顯示這兩個欄位，以及顯示的位置與層級。
- **決策**：
  - **TypeScript interface**：所有對應的 interface 補上 `updated_by: string | null` 與 `updated_at: string | null`（nullable，對應 Go `*string`）。
  - **列表**（HTML table）：加入 `更新者` / `更新時間` 欄（`hidden md:table-cell`，桌面才顯示）。
  - **子表格**（`grid-template-columns` 自訂）：在 CSS 加寬度，並在 HTML header + row 各補 `<span>`。
  - **Detail / 展開面板**：能對應到資料列的地方直接顯示；無獨立 detail modal 的子表格（Entry、Payment、Lot、Movement）只在 grid row 顯示。
  - **Edit modal**：唯讀展示（readonly input），放在備註欄下方，僅在 `mode === 'edit'` 時顯示。
- **替代方案**：
  - **只在 types 補欄位，不在 UI 顯示**：型別安全，但損失了審計可見性，使用者無法從 UI 確認最後修改者。
  - **獨立「歷史紀錄」展開面板**：更完整，但目前後端未提供歷史查詢 API，超出本次範圍。
- **後果**：
  - 正面：使用者可在列表與 modal 直接看到最後更新者與時間，提升可稽核性。
  - 負面：每個自訂 grid 子表格需同步修改 CSS 與 HTML 兩處（已記錄於 NOTES.md）；新增唯讀 label 需搭配 `for`/`id`，否則觸發 a11y 警告（見 ISSUE-004）。

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


## ADR-009 Sidebar 可展開選單項以函式泛化展開邏輯

- **狀態**：Accepted
- **日期**：2026-05-20
- **背景**：
  原 `Home.svelte` 的 sidebar 展開邏輯以 `isReportsExpanded` 這個 hardcoded `$derived` 實作，
  只處理「財務報表」的展開狀態。當「系統設定」也需要子選單時，若繼續各自宣告一個 `$derived`，
  每新增一個可展開選單項都要同時修改 script 區塊和 template，維護成本線性增長。
- **決策**：
  改為通用 `isExpanded(key: string): boolean` 函式搭配 `parentBasePaths` lookup table。
  `MenuItem` 介面加入可選的 `key` 欄位，template 中用 `{@const key = item.key ?? item.path}` 取得 key，
  展開邏輯一律走 `isExpanded(key)`。

  ```typescript
  const parentBasePaths: Record<string, string> = {
    reports:  '/home/reports',
    settings: '/home/settings',
  };

  function isExpanded(key: string): boolean {
    return expandedParents.has(key) || currentPath.startsWith(parentBasePaths[key] ?? '');
  }
  ```

  新增可展開選單項只需：① 在 `menuItems` 加 `key` 與 `children`；② 在 `parentChildPaths` 加對應路徑陣列。

  延伸 ADR-006 的原則：展開狀態仍為純計算（不寫入 state），讀取 `expandedParents` 與 `currentPath` 兩個來源。

  **後續延伸（2026-05-25）**：當「分期管理」子路徑（`/home/installment`、`/home/prepaid`、`/home/fixed-asset`）無公共前綴時，`parentBasePaths: Record<string, string>` 的單一字串無法涵蓋所有路徑。改為 `parentChildPaths: Record<string, string[]>` 搭配 `some(p => currentPath.startsWith(p))` 判斷，支援任意數量的子路徑。

- **替代方案**：
  - **每個可展開項各宣告一個 `$derived`**：最直觀，但每次新增子選單都要改 script 與 template 兩處，容易遺漏。
  - **將展開邏輯移入 `MenuItem` struct，作為物件方法**：Svelte 5 reactive system 無法追蹤外部物件方法的依賴，難以正確更新。
- **後果**：
  - 正面：新增任意數量的可展開選單項，template 零改動；`parentBasePaths` 作為集中設定，一目瞭然。
  - 負面：新增選單項若忘記在 `parentBasePaths` 登記，直接以 URL 導覽到子路由時不會自動展開父項（手動點擊仍正常）。

---

## ADR-010 帳本建立科目限制透過 `filteredAccounts` prop 傳遞而非在子元件內載入

- **狀態**：Accepted
- **日期**：2026-05-20
- **背景**：
  後端 `GET /api/setting/ledger-account-type` 回傳 `LedgerAccountTypeConfigResult`，其中 `descendants: Account[]`
  是配置科目下所有子孫科目。帳本建立時需依帳戶類型（`BANK_ACCOUNT` / `CREDIT_CARD` / `LOAN`）
  限制可選的關聯科目。
- **決策**：
  限制邏輯集中在頁面元件 `Ledger.svelte`，以 `$derived` 計算出 `filteredAccounts` 與 `filteredParentAccounts`，
  再透過 prop 傳入 `NewLedgerFormSection`，最終傳給 `AccountSelect` 的 `accounts` prop。
  子元件完全不感知「限制」這件事，只負責顯示傳入的 accounts 清單。

  ```
  Ledger.svelte ($derived filteredAccounts)
    → NewLedgerFormSection (accounts / accountsForParent props)
      → AccountSelect (accounts prop, 只顯示傳入清單)
      → NewAccountFormSection (parentAccounts prop)
  ```

  帳戶類型切換時，透過 `$effect` 監測 type 變化，清空已選的 `newLedgerAccountId` 與 `newAccountForm.parent_id`，
  避免殘留不符合限制的選擇值。

- **替代方案**：
  - **在 `AccountSelect` 加 `allowedIds?: string[]` prop 在元件內過濾**：讓 AccountSelect 感知限制邏輯，破壞其無狀態、可重用的特性；且需要在每個使用 AccountSelect 的地方判斷是否傳入 allowedIds。
  - **在 `NewLedgerFormSection` 內部呼叫 API 載入 config**：元件自帶資料來源，但導致 API 每次開 modal 就重新呼叫；且不符合「子元件接收資料、頁面元件管理狀態」的分層原則。
- **後果**：
  - 正面：`AccountSelect` 保持無狀態，可在任何地方重用；限制邏輯集中在 Ledger.svelte，易於測試與修改。
  - 負面：config 需要在 Ledger.svelte 的 `openModal()` 時額外呼叫 API；若日後其他頁面（如 Installment）也需要同樣限制，需要在各頁面重複相同的 `$derived` 邏輯。

---

## ADR-011 跨頁面科目導航使用 URL hash query param 傳遞預選科目

- **狀態**：Accepted
- **日期**：2026-05-26
- **背景**：
  多個頁面（JournalEntry、BalanceSheet、IncomeStatement、Accounts）顯示科目名稱，
  使用者希望點擊科目名稱直接跳轉至科目分析（AccountAnalysis）並自動載入該科目的資料，
  無需在目標頁面重新手動選擇。需要決定「跳轉時傳遞預選科目 ID」的機制。
- **決策**：
  - `src/lib/navigate.ts` 封裝 `goToAccountAnalysis(accountId: string)` helper，
    設定 hash 為 `#/home/account-analysis?account=<encodeURIComponent(accountId)>`。
  - `AccountAnalysis.svelte` 在初始資料載入完成後（accounts 陣列已取得），解析 `window.location.hash`，
    讀取 `?account=` 參數，若對應科目存在則呼叫 `selectAccount()`。
  - 同時設定 `hasBack = true`，在頁面頂部顯示「← 返回」按鈕，呼叫 `history.back()`。

  ```typescript
  // src/lib/navigate.ts
  export function goToAccountAnalysis(accountId: string): void {
    window.location.hash = `#/home/account-analysis?account=${encodeURIComponent(accountId)}`;
  }
  ```

- **替代方案**：
  - **Svelte writable store 傳遞預選值**：需要在兩個路由元件之間共享 store 狀態，且 store 內容在重新整理後消失，無法從 URL 直接存取特定科目。
  - **僅導航不傳參數、讓使用者手動選擇**：降低摩擦感，但違背「點一下直達」的使用者期望。
  - **在 AccountAnalysis 元件接收 prop**：SPA hash routing 不支援直接對路由元件傳 prop，且需要修改 router 設定。
- **後果**：
  - 正面：URL 可書籤與分享；SPA hash 每次變更產生 history entry，`history.back()` 自然還原前一頁；`navigate.ts` 集中管理，各頁面只需 import 一個函式。
  - 負面：`AccountAnalysis` 的 query param 解析只在初始 mount 的 `$effect` 執行，若使用者在同一頁面多次點不同科目連結（hash 相同路徑但 param 不同），不會觸發重新選取，需要手動切換 AccountSelect（可接受，因這屬於邊緣情境）。

---

## ADR-012 登入後以 pendingToken 暫存 user token，選商戶後換 merchant token

- **狀態**：Accepted
- **日期**：2026-06-03
- **背景**：
  後端採用兩層 token 設計：登入取得的是 **user-scoped token**（可存取 merchant list / select API），
  呼叫 `POST /api/merchant/select` 後取得的才是 **merchant-scoped token**（可存取所有業務資料）。
  舊版 `auth.ts` 在 `login()` 內自動取第一個商戶並換 token，跳過了商戶選擇步驟。
  重構後需要讓使用者自行選擇商戶，因此必須在兩個路由之間安全地傳遞 user token。
- **決策**：
  - `authStore` 新增 `pendingToken`（`string | null`），存入 `sessionStorage`（而非 localStorage），
    使 token 在分頁關閉或重新整理後仍可還原，但不會跨 session 持留。
  - `login()` 只儲存 `pendingToken`，不再自動呼叫 `list` / `select`。
  - 登入成功後跳轉 `#/merchant-select`，此頁用 `pendingToken` 呼叫 `list` API，
    使用者選擇後呼叫 `select` 取得 merchant token，寫入 `authStore.token`（localStorage），
    同時清除 `pendingToken`，再跳轉 `#/home`。
  - 路由守衛：
    - 進入 `#/merchant-select` 時，若 `pendingToken` 不存在 → 跳回 `#/`；若 `token` 已存在 → 跳 `#/home`。

  ```
  login()
    → setPendingToken(userToken)  [sessionStorage]
    → navigate #/merchant-select
      → list(pendingToken) → filter ACTIVE
      → user clicks merchant
        → select(merchantId, pendingToken)
          → setToken(merchantToken)  [localStorage]
          → clearPendingToken()
          → navigate #/home
  ```

- **替代方案**：
  - **shared $state store 傳 user token（不存 sessionStorage）**：重新整理後 pendingToken 消失，使用者需重新登入；且若 SPA router 在 onMount 前已 re-render，store 可能尚未初始化。
  - **將 user token 與 merchant token 都存入 localStorage 各自 key**：user token 若未清除，日後可能被誤用於業務 API，有安全風險；sessionStorage 自然限制其存活範圍。
  - **在 Login 元件內嵌商戶選擇步驟（step 2 表單）**：不需要獨立路由，但 Login.svelte 會混入商戶選擇狀態與 UI，職責不清；無法透過 URL 直接進入商戶選擇頁（如重新整理後恢復）。
- **後果**：
  - 正面：職責分離清晰（Login 負責驗證、MerchantSelect 負責商戶選擇）；sessionStorage 自然限制 pendingToken 生命週期；重新整理停留在 `#/merchant-select` 仍可正常運作。
  - 負面：`authStore` 同時持有兩種語意的 token，需透過欄位命名（`token` vs `pendingToken`）區分；新增路由守衛需要維護。

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
