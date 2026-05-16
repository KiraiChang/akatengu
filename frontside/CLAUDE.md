# AKATENGU

使用 Svelte 5 開發的會計軟體前端。

## 技術棧

- 框架：Svelte 5.55.4
- 語言：TypeScript 6.0.2（strict mode）
- 建置工具：Vite 8.0.9
- 套件管理：npm
- Node.js：24.12.2

## 常用指令

- `npm run dev`：啟動開發伺服器（port 5174）
- `npm run build`：建置
- `npm run check`：svelte-check + tsc 型別檢查（唯一型別驗證指令）
- `npm run preview`：預覽建置結果

---

## 需求處理流程（Requirement Workflow）

接獲任何新需求時，**必須**依照以下順序執行，不得跳過任一步驟：

1. **重複需求、請求確認**
   以自己的話複述需求內容（包含範圍、前提條件、預期結果），明確詢問使用者是否與原意相符，**待使用者確認後才進入下一步**。

2. **撰寫計畫書、請求確認**
   確認需求後，提出包含影響範圍、修改步驟及潛在風險的計畫書，**待使用者確認計畫書無誤後才開始動手修改**。

3. **依計畫書實作**
   按已確認的計畫逐步實作，範圍不得超出計畫書所述。

4. **現有測試失敗時：優先完成需求，禁止立即修復**
   若修改過程中發現既有測試案例出現錯誤，**強烈禁止**立即中斷實作去修復測試；應先完成當前需求的全部修改，再統一處理測試問題。若情況複雜，**需向使用者回報並取得指示**，不得自行決定修復策略。

5. **撰寫測試案例驗證需求**
   需求實作完成後，補充對應的測試案例，執行 `go test ./...` 確認新需求正常運作，並在回報中附上測試輸出。

6. **開發問題與阻礙 → `doc/ISSUES.md`**
   開發過程中遇到尚未解決的問題、技術障礙、已知 Bug 或待釐清事項，**立即**記錄到 `doc/ISSUES.md`，並在解決後更新狀態，避免問題重複發生或遺忘。

7. **臨時想法與筆記 → `doc/NOTES.md`**
   開發中產生的臨時想法、Prompt 設計、Debug 筆記、架構草稿等非正式內容，記錄到 `doc/NOTES.md`，不需要完整，但要能讓未來的自己看懂。

8. **長期架構決策 → `doc/ARCHITECTURE.md`**
   凡是影響系統整體設計的決策（選型理由、設計取捨、不採用某方案的原因），記錄到 `doc/ARCHITECTURE.md`，格式採用 ADR（Architecture Decision Record）。

9. **計畫待辦事項 → `doc/TODO.md`**
   需求完成後、回顧時，或發現「現在不做但之後要做」的事項（功能擴充、技術債、效能優化、文件補充），記錄到 `doc/TODO.md`。每次開發前先確認 TODO 清單，避免遺漏已規劃的工作。

10. **需求完成後回顧 → 更新 `doc/`**
    需求實作並驗證全部通過後，主動回顧本次開發，依下列清單確認是否有內容需要記錄或補充。
    **此為必做步驟，不得省略。**

    | 回顧項目 | 目標檔案 |
    |---|---|
    | 開發中遇到的問題、技術障礙、已知 Bug（含已解決） | `doc/ISSUES.md` |
    | 架構決策、後端合作知識、設計取捨、不採用某方案的原因 | `doc/ARCHITECTURE.md` |
    | 值得保留的技術筆記、設計想法、Debug 心得 | `doc/NOTES.md` |
    | 本次未處理但值得做的項目（功能、技術債、UI 改善、文件補充） | `doc/TODO.md` |

    沒有要記錄的項目時，**明確向使用者說明「本次無需更新 doc/」**，不得靜默略過。

---

## 行為規範

- 需求不明確或資訊不足時，**先提問**，不自行假設或補充資料
- 修改前列出計畫與修改範圍（將讀取哪些檔案、將修改哪些檔案、修改原因），確認後才開始執行
- 遇到問題時**直接反映**，不要盲目修改；未釐清原因前不得做推測性修改。
- 執行指令時遇到**執行環境問題**（如 `command not found`、權限不足、路徑錯誤等），立即向使用者說明錯誤狀況並請求支援，**不得自行嘗試繞過**（如切換目錄、改用 npx、修改 PATH 等），等待使用者修正後再繼續。
- **需要瀏覽專案時，必須先向使用者說明瀏覽目的並取得確認，禁止未獲許可主動瀏覽任何檔案或目錄。**
- 每個檔案在同一任務中只讀取一次；只讀取與當前任務直接相關的檔案，不掃描無關目錄
- 修改完成後執行驗證，並**回報實際結果**（型別錯誤、lint 警告、測試失敗等），不要只回報「build 成功」
- 建置失敗時**回報錯誤狀況**，不要盲目嘗試修復
- 修改涉及 UI 排版或樣式的程式碼後，必須檢查「已知問題與處理慣例 > UI 設計慣例」中的預防檢查點，確認未觸發已知問題後才算完成

## 驗證流程

每次修改後依序執行：

1. `npm run lint`：ESLint 檢查
2. `npm run check`：svelte-check + tsc 型別檢查
3. `npm run build`：建置並檢查告警或錯誤

純樣式修改（CSS 異動、無 TS/Svelte 邏輯變更）可省略步驟 2。
回報格式：列出每個步驟的實際輸出，有錯誤則附上錯誤訊息與行號。

## 命名規範

- 變數、函式：camelCase
- 元件、型別、介面：PascalCase
- 常數：UPPER_SNAKE_CASE
- 元件檔案：PascalCase（`UserProfile.svelte`）
- 非元件檔案：camelCase（`userService.ts`）
- 布林值：`is`/`has`/`should` 前綴

## TypeScript 規範

- 禁止 `any`，必要時用 `unknown` 搭配型別守衛
- 物件型別用 `interface`，聯合型別或工具型別用 `type`
- 所有匯出函式必須標註回傳型別

## Svelte 5 元件規範

- 使用 runes 語法：`$props()`、`$state()`、`$derived()`、`$effect()`
- Props 型別以 `interface` 定義後傳入 `$props<T>()`
- 共用邏輯抽成 `src/lib/` 下的 `.svelte.ts` 檔
- 超過 200 行的元件拆分為子元件

## 錯誤處理

- API 呼叫用 try-catch 包裹
- 統一使用 `AppError` 類別

## 專案結構慣例

- 新頁面：`src/routes/` 或 `src/views/`（依現有結構）
- API 串接：`src/services/`，型別定義於 `src/types/`
- 使用 `apiClient` 封裝，不直接使用 fetch 或 axios
- 樣式一律集中於 `src/styles/`，依類型分子目錄（`views/`、`components/` 等），並在 `src/main.ts` 統一 import；**禁止**在 `.svelte` 元件內使用 `<style>` 區塊

## 回應規範

- 一律使用繁體中文
- 程式碼中的變數名稱與註解使用英文

## 已知問題與處理慣例

### 建置告警

#### TypeScript
<!-- 格式：錯誤碼 / 描述 → 根本原因 → 解決方式 -->

#### Vite / 建置工具
<!-- 格式：告警描述 → 根本原因 → 解決方式 -->

#### ESLint
<!-- 格式：規則名稱 → 觸發情境 → 解決方式 -->

**`a11y_click_events_have_key_events`**：`<div onclick={...}>` 缺少鍵盤事件處理器
→ 在同一元素加上 `onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && handler()}`

**`a11y_no_static_element_interactions`**：非互動元素（`<div>`）有 click handler 但缺 ARIA role
→ 加上 `role="button"` 與 `tabindex="0"`（兩個告警通常同時出現，需同時處理）

**純阻止事件冒泡的 `<div onclick={(e) => e.stopPropagation()}>` 包裹層**（同時觸發上述兩個告警）
→ 移除包裹 div，改將 `e.stopPropagation()` 移入各子元素的 onclick 內：`onclick={(e) => { e.stopPropagation(); handler(); }}`

**`eslint.config.js` 中 `parserOptions.project` 指向錯誤 tsconfig（所有 `.ts` 檔案 parsing error）**
→ 根本原因：根目錄 `tsconfig.json` 有 `"files": []`，不包含任何原始碼；須改指向 `./tsconfig.app.json`
→ 正確設定：`parserOptions: { project: './tsconfig.app.json' }`

**`.svelte.ts` 檔案中 Svelte rune（`$state` 等）被 `no-undef` 標記為未定義**
→ 根本原因：ESLint TypeScript 設定區塊未宣告 Svelte 5 compiler 巨集為全域
→ 在 `eslint.config.js` TypeScript 區塊的 `globals` 加入：`$state`, `$derived`, `$effect`, `$props`, `$bindable`, `$inspect`, `$host`（均設為 `'readonly'`）

**`_` 前綴解構變數（如 `_v`）仍被 `@typescript-eslint/no-unused-vars` 警告**
→ 規則需加入 `{ varsIgnorePattern: '^_', argsIgnorePattern: '^_' }` 選項
→ 需同時套用於 TypeScript 區塊與 Svelte 區塊，兩處都要設定


### UI 設計慣例

#### RWD 排版
<!-- 修改元件或改版後，若發生排版跑版問題，記錄於此 -->
<!-- 格式：問題描述 → 根本原因 → 修正方式 → 預防檢查點 -->

**Header `.header-brand-sub`（已修）**：「會計管理系統」在 mobile 下變直書
**根本原因**：header 為 `justify-content: space-between` flex 容器，小螢幕空間壓縮使無 `nowrap` 保護的中文觸發直書
**修正方式**：`white-space: nowrap; writing-mode: horizontal-tb;`
**預防檢查點**：修改 header 寬度、padding、或左右子元素尺寸後，需在 375px 確認品牌文字仍為橫書

#### 文字排版
- **問題**：改版後 flex 容器寬度縮小，導致橫書文字在 768px 以下變成直書並溢出至相鄰區塊
  **根本原因**：父層 `overflow: hidden` 配合 `writing-mode` 未明確設定，瀏覽器自動折行
  **修正方式**：在該元件明確設定 `writing-mode: horizontal-tb`，並給容器最小寬度
  **預防檢查點**：凡修改此元件的寬度、flex 方向、或父層 overflow 屬性後，
  必須在 375px / 768px / 1280px 三個斷點確認文字方向與容器邊界正常

- **flex 子元素文字溢出**：flex 容器內的文字元素若未設 `min-width:0`，在空間壓縮時會撐破容器或觸發直書
  **預防檢查點**：凡新增 flex 子元素含文字時，確認已加 `min-width:0` 與適當 overflow 保護

#### 元件層疊與定位
<!-- z-index、overflow、position 相關問題 -->
<!-- 格式：問題描述 → 根本原因 → 修正方式 → 預防檢查點 -->