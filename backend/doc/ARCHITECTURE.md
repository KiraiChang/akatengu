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
| [ADR-001](#adr-001-事件溯源作為核心寫入架構) | 事件溯源作為核心寫入架構 | Accepted | 2024-01-01 |
| [ADR-002](#adr-002-使用-sqlite-作為資料庫) | 使用 SQLite 作為資料庫 | Accepted | 2024-01-01 |
| [ADR-003](#adr-003-金額欄位使用-decimaldecimal) | 金額欄位使用 decimal.Decimal | Accepted | 2024-01-01 |
| [ADR-004](#adr-004-現金流量表採用間接法account-based) | 現金流量表採用間接法（account-based） | Superseded by ADR-007 | 2026-05-15 |
| [ADR-005](#adr-005-可為空的-enum-欄位型別映射策略) | 可為空的 enum 欄位型別映射策略 | Accepted | 2026-05-15 |
| [ADR-006](#adr-006-股東權益變動表的本期淨利補入策略) | 股東權益變動表的本期淨利補入策略 | Accepted | 2026-05-15 |
| [ADR-007](#adr-007-現金流量表分類來源改為-journal_entriesentry-based) | 現金流量表分類來源改為 journal_entries（entry-based） | Accepted | 2026-05-16 |
| [ADR-008](#adr-008-直接法現金流量表direct-method架構設計) | 直接法現金流量表（Direct Method）架構設計 | Accepted | 2026-05-16 |
| [ADR-009](#adr-009-審計欄位-updated_by--updated_at-架構設計) | 審計欄位（updated_by / updated_at）架構設計 | Accepted | 2026-05-19 |
| [ADR-010](#adr-010-投資-pipeline-科目來源改由-asset_type_account_config-查詢) | 投資 Pipeline 科目來源改由 asset_type_account_config 查詢 | Accepted | 2026-05-20 |

---

## ADR-001 事件溯源作為核心寫入架構

- **狀態**：Accepted
- **日期**：2024-01-01
- **背景**：
  會計系統需要完整的稽核軌跡，每一筆交易都必須可追溯，且支援重播（replay）以重建任意時間點的狀態。
- **決策**：
  採用事件溯源（Event Sourcing）模式，所有寫入操作透過 `EventStoreService.Append` 進入，事件不可變更，讀模型（Projection）由事件重播建立。
- **替代方案**：
  - CRUD 直寫：實作簡單，但無法保留完整歷史，稽核困難。
- **後果**：
  - 正面：完整稽核軌跡、可重播、讀寫分離清晰。
  - 負面：寫入鏈路較長，新增事件類型需同步新增 Pipeline 與 Projection。

---

## ADR-002 使用 SQLite 作為資料庫

- **狀態**：Accepted
- **日期**：2024-01-01
- **背景**：
  本系統為個人/小團隊財務管理工具，部署環境為單機，不需要分散式資料庫。
- **決策**：
  採用 SQLite（WAL 模式，單一連線），透過 `golang-migrate` 管理 schema 版本。
- **替代方案**：
  - PostgreSQL：功能強大但需額外部署，對單機場景過重。
- **後果**：
  - 正面：零部署成本、檔案即資料庫、WAL 模式下讀寫效能尚可。
  - 負面：不支援水平擴展，並發寫入受單一連線限制（`MaxOpenConns=1`）。

---

## ADR-003 金額欄位使用 decimal.Decimal

- **狀態**：Accepted
- **日期**：2024-01-01
- **背景**：
  浮點數（float64）在金融計算中會產生精度誤差，不適合儲存金額。
- **決策**：
  所有金額欄位一律使用 `github.com/shopspring/decimal` 的 `Decimal` 型別，並在 `sqlc.yaml` 的 `overrides` 中宣告型別映射。
- **替代方案**：
  - float64：有精度問題，拒絕。
  - int64（以分為單位）：可行但需處理幣別小數位數，後續擴充較複雜。
- **後果**：
  - 正面：精確的十進位運算，符合會計需求。
  - 負面：aggregate 函數結果（SUM、COALESCE）需額外建立 View 才能正確映射型別。

---

## ADR-004 現金流量表採用間接法（account-based）

- **狀態**：Superseded by ADR-007
- **日期**：2026-05-15
- **背景**：
  需要提供現金流量表報表，有間接法（Indirect Method）和直接法（Direct Method）兩種實作方式。間接法從淨利出發，調整非現金科目的期間變動；直接法直接列示現金收付明細。
- **決策**：
  採用間接法。計算公式統一以 `期間 credit - 期間 debit` 表達所有非 CASH 科目的現金流量影響。此公式對借方正常科目（資產）與貸方正常科目（負債、權益）均語義正確：資產增加（借方增加）→ 現金流出為負；負債增加（貸方增加）→ 融資流入為正。
- **資料設計**：
  在 `accounts` 資料表新增 `cash_flow_category TEXT CHECK (... IN ('CASH','OPERATING','INVESTING','FINANCING'))` 欄位。`CASH` 類別用於識別現金科目（期初/期末餘額），其餘三類用於分組調整項目。查詢端使用 `sqlx` 動態 SQL（非 sqlc），原因是篩選條件需在執行期依分類動態組裝。
- **替代方案**：
  - 直接法：需追蹤每筆交易的現金收付意圖，現有 journal_entries 資料結構無法直接支援，需大幅改動寫入模型。
- **後果**：
  - 正面：公式簡單、與現有 journal_entries 資料直接對應、不需改動寫入路徑。
  - 負面：前端收到的是調整項目清單，需自行判斷呈現格式（設計上有意為之，以保持後端彈性）。
- **取代原因**：
  account-based 設計導致同一科目下所有分錄必須歸入同一現金流量類別，無法針對個別交易彈性分類。由 ADR-007（entry-based）取代。

---

## ADR-005 可為空的 enum 欄位型別映射策略

- **狀態**：Accepted
- **日期**：2026-05-15
- **背景**：
  SQLite 的 nullable TEXT 欄位（如 `accounts.cash_flow_category`）在 sqlc 生成時會產生 `*string`，但系統的 Projection Model 應使用型別安全的 `*enums.CashFlowCategory`，否則需在 Projection Service 手動進行字串轉換，且 dbmap-gen 無法自動映射。
- **決策**：
  在 `sqlc.yaml` 的 `overrides` 區段宣告型別覆寫（`column: "accounts.cash_flow_category"` → `enums.CashFlowCategory`）。搭配已設定的 `emit_pointers_for_null_types: true`，sqlc 自動生成 `*enums.CashFlowCategory`。Projection Model 跟著使用相同指標型別，dbmap-gen 即可直接映射，Projection Service 不需任何轉換邏輯。
- **替代方案**：
  - Projection Model 保留 `*string`：需在 Service 層手動轉換（`s := p.CashFlowCategory.String(); cfCat = &s`），且 dbmap-gen 無法感知 enum 語義。
  - 使用 `NullEnum[T]`：適合明確需要區分「NULL」與「零值」的場景，但此處指標語義已足夠。
- **後果**：
  - 正面：型別安全貫穿 DB → sqlcdb → Projection Model → Payload 全鏈路，Service 層零轉換邏輯。
  - 負面：每新增一個 nullable enum 欄位都需在 `sqlc.yaml` 手動宣告 override，維護略繁。

---

## ADR-006 股東權益變動表的本期淨利補入策略

- **狀態**：Accepted
- **日期**：2026-05-15
- **背景**：
  股東權益變動表需要顯示期間權益的完整變動，包含「本期淨利」對權益的貢獻。問題在於：若當期帳未關帳，損益類科目（INCOME/EXPENSE）的餘額尚未結轉至保留盈餘（EQUITY 科目），單純查詢 EQUITY 科目的期間變動會低估實際權益增減。
- **決策**：
  採用後端自動補入策略（Direction B）：後端呼叫現有的 `GetIncomeStatement` 計算本期淨利，以 `is_virtual: true` 的虛擬行附加在權益科目清單尾端，同時將淨利加入 `total_period_change` 與 `total_end_balance`。各個真實 EQUITY 科目的 `end_balance` 保持原始帳面值不修改，讓前端能清楚區分已關帳的真實餘額與虛擬的本期損益貢獻。
- **替代方案**：
  - Direction A（前端自行合併）：後端只回傳 EQUITY 科目原始資料，前端另呼叫損益表 API 再合併。優點是後端邏輯單純；缺點是要求前端具備會計知識，且需多一次 API 呼叫。
- **後果**：
  - 正面：API 單次呼叫即可取得完整的股東權益變動表，前端零負擔。
  - 負面：`GetIncomeStatement` 被 `GetEquityStatement` 隱式依賴；若未來損益表計算邏輯變動，權益表的淨利數字也會連動改變（屬於正確行為，但需留意）。

---

## ADR-007 現金流量表分類來源改為 journal_entries（entry-based）

- **狀態**：Accepted
- **日期**：2026-05-16
- **背景**：
  原始設計（ADR-004）將 `cash_flow_category`（OPERATING / INVESTING / FINANCING）掛在 `accounts` 資料表，
  導致同一科目下的所有分錄必須歸入同一個現金流量類別，無法針對個別交易彈性分類。

  例如：股利支付在 IFRS 可歸入 OPERATING 或 FINANCING；
  應收帳款依交易性質可能是 OPERATING 或 INVESTING。
- **決策**：
  將 OPERATING / INVESTING / FINANCING 分類移至 `journal_entries.cash_flow_category`（nullable）。
  每筆分錄可獨立設定，前端依所選科目的 `accounts.cash_flow_category` 作為預填預設值，使用者可自行修改。

  `accounts.cash_flow_category` 保留以下用途：
  - `CASH`：識別現金及約當現金科目，用於計算期初 / 期末現金餘額（`queryCashBefore` / `queryCashUpTo`）。
  - `OPERATING / INVESTING / FINANCING`：提供前端預設值參考，不再參與計算。
- **替代方案**：
  - 保留 account-based：無法支援同一科目多分類需求，排除。
  - 同時維護兩個來源：增加維護成本與資料一致性風險，排除。
- **後果**：

  | 面向 | 舊設計（account-based） | 新設計（entry-based） |
  |------|------------------------|----------------------|
  | 彈性 | 低，同科目只能一個分類 | 高，每筆分錄獨立設定 |
  | 資料完整性 | account 設定後即生效 | 需前端 / pipeline 主動填入 |
  | 系統自動分錄 | 靠科目自動帶入 | 需逐案評估並補上（見 TODO.md） |
  | 查詢複雜度 | 較低 | 略高，需 GROUP BY (account_id, cash_flow_category) |

---

## ADR-008 直接法現金流量表（Direct Method）架構設計

- **狀態**：Accepted
- **日期**：2026-05-16
- **背景**：
  間接法（Indirect Method）以本期淨利為起點加減調整項，適合財務人員分析，但不直觀。
  直接法（Direct Method）直接呈現實際現金收入（CashReceived）與現金支出（CashPaid），
  對個人財務使用者更容易理解，且兩法 Operating Total 數學恆等，可作為交互驗證手段。
- **決策**：
  新增獨立的 `GetDirectCashFlowStatement` 方法，與間接法並存，提供不同使用情境：
  - 間接法（`/report/cash_flow_statement`）：財務分析、損益拆解
  - 直接法（`/report/cash_flow_statement_direct`）：現金收支明細、兩法驗證

  **Operating 活動現金計算方式**：
  1. 識別「營業活動交易（operating_txns）」—— 取兩個條件的聯集：
     - 含 `a.type IN ('INCOME','EXPENSE')` 分錄的交易（損益交易）
     - 含 `je.cash_flow_category = 'OPERATING'` 分錄的交易（非損益的OPERATING調整）
  2. 對這些交易中所有 `accounts.cash_flow_category = 'CASH'` 的分錄加總借方與貸方：
     - `CashReceived = SUM(debit)` （現金流入）
     - `CashPaid = SUM(credit)` （現金流出，以正數呈現）
     - `Total = CashReceived - CashPaid`

  **Investing / Financing 重用間接法查詢**：直接重用 `queryCashFlowChanges`，只跳過 OPERATING rows 不累入 Total。

  **期初 / 期末現金重用**：`queryCashBefore` / `queryCashUpTo` 被兩法共用。
- **替代方案**：
  - 改寫間接法 SQL 同時輸出直接法欄位：兩個呈現目的不同，合併會讓查詢難以維護，排除。
  - 在 Service 層從間接法結果換算直接法：跳過了「實際現金移動」的查詢，無法正確還原直接法，排除。
- **後果**：

  | 面向 | 直接法 | 間接法 |
  |------|--------|--------|
  | 使用者易讀性 | 高（顯示實際現金流） | 中（需理解調整項意義） |
  | 查詢複雜度 | 略高（operating_txns 兩層 CTE） | 低（直接聚合 OPERATING entries） |
  | 兩法可互相驗證 | ✓ Operating Total 恆等 | ✓ |
  | 混合分類交易精準度 | 邊界行為（整筆歸入 OPERATING） | 精確（按 entry 分類） |

  **已知限制**：一筆交易同時含不同 CF 分類的分錄（如：CR INCOME 600 + CR INVESTING 400 = DR CASH 1000），`queryDirectOperatingCash` 會將整筆交易歸入 OPERATING，現金流入 1000 全算入 CashReceived 而非按比例拆分。個人財務情境中幾乎不會出現此類混合交易，視為可接受的邊界行為。

---

## ADR-009 審計欄位（updated_by / updated_at）架構設計

- **狀態**：Accepted
- **日期**：2026-05-19
- **背景**：
  所有 projection 表與 event_store 需記錄「誰觸發」（`updated_by`）與「何時寫入」（`updated_at`）。
  JWT Claims 已帶有 `UserName` 欄位，Middleware 將 `*jwt.Claims` 注入 context，
  需要一個統一的取用方式，並在整個寫入鏈路中傳遞。
- **決策**：

  **1. 來源**：`internal/pkg/ctxkey.GetUserName(ctx)` 統一讀取 context 中的 Claims。
  回傳空字串代表無使用者（Replay 場景），呼叫端轉為 `nil *string` 寫入 DB NULL。

  **2. 傳遞路徑**：
  ```
  JWT Claims（ctx）
    └─► ctxkey.GetUserName(ctx)
          └─► EventStoreService.Append → ct.UpdatedBy = username
                └─► pipelines.Result.UpdatedBy（string，空字串表示無使用者）
                      └─► 各 Projection Service Apply 方法
                            └─► toUpdatedBy(ct.UpdatedBy) → *string（nil 或 &username）
                                  └─► projection model struct 的 UpdatedBy 欄位
                                        └─► projection_repo 方法 → sqlcdb INSERT/UPDATE
  ```

  **3. updated_at 策略**：應用層**不傳遞** `updated_at`，改由 SQLite 自動填入：
  - INSERT：`DEFAULT (datetime('now'))` 自動填入
  - UPDATE：SQL SET 子句顯式加入 `updated_at = datetime('now')`

  **4. Replay 場景**：context 無 user → `GetUserName` 回傳 `""` → `toUpdatedBy("")` 回傳 `nil` → DB 存 NULL，此為預期行為。

- **替代方案**：
  - 直接在每個 Projection Service 讀取 context：需在所有 Apply 方法中重複讀取 context，且 Projection 介面的 context 參數需明確支援（目前已有）。此方案耦合度較高，被 `pipelines.Result` 集中傳遞的方式取代。
  - 在 Pipeline 層計算後寫入事件 payload：侵入 payload 結構，且 Replay 時 payload 中的 username 是歷史值而非當下執行者，語義不正確。
- **後果**：
  - 正面：username 在 `Append` 入口點統一取得，各 Projection 只需讀取 `ct.UpdatedBy`，無需重複讀 context。Replay 場景安全，存 NULL 不影響功能。
  - 負面：`pipelines.Result` 新增了一個與事件業務無關的欄位，輕度職責擴散。

---

## ADR-010 投資 Pipeline 科目來源改由 asset_type_account_config 查詢

- **狀態**：Accepted
- **日期**：2026-05-20
- **背景**：
  投資 BUY / SELL / MARK pipeline 中，手續費、交易稅、已實現損益、未實現評價損益與 OCI 的科目 ID
  原本透過 `sys_accounts`（一張系統設定表）以 sys_code 字串查詢取得。
  此設計的問題是科目無法由前端調整：同一 sys_code 對所有 AssetType 共用同一筆 sys_accounts 記錄，
  無法針對 STOCK / FUND / GOLD / FX 分別設定不同科目。
- **決策**：
  新增 `asset_type_account_config` 表，每個 (merchant_id, asset_type) 一筆記錄，
  存放 7 個科目 ID（`realized_gain`、`realized_loss`、`unrealized_gain`、`unrealized_loss`、
  `oci`（nullable）、`fee`、`tax`），提供前端 CRUD API（`/api/setting/asset-type`）供動態設定。

  Pipeline 中 `buyEntries`、`sellEntries`、`fvEntries` 均改為：
  ```go
  config, err := e.query.Config.GetAssetTypeAccountConfig(ctx, assetType)
  // 直接使用 config.FeeAccountID、config.TaxAccountID 等欄位
  ```
  若 `oci_account_id` 為 nil 而嘗試執行 FVOCI 評價，回傳明確錯誤，不使用 fallback。

  舊的 helper 函數（`getAccountIdByFunc`、`getAssetTypeFeeSysCode`、`getAssetTypeTaxSysCode`、
  `getAssetTypeGainSysCode`、`getAssetTypeLossSysCode`、`getFVTPLUnrealizedGainSysCode`）一律移除。
- **替代方案**：
  - 保留 sys_accounts，改在 sys_accounts 多一筆 per-AssetType 紀錄：修改量大，且 sys_accounts 的語義是「系統預設，不可前端修改」，混入投資設定會破壞此語義。
  - 在 Pipeline 初始化時批次載入所有 config：可減少 DB 查詢次數，但目前 Pipeline 每次都是新 goroutine，無需快取，過早優化。
- **後果**：
  - 正面：每個 AssetType 可獨立設定科目，前端可動態調整；同時修正了 sellEntries 中稅金科目長期使用手續費 sys_code 的 bug（ISSUE-006）。
  - 負面：merchant 若未設定 `asset_type_account_config`（Seeder 未執行或未覆蓋），投資操作會報錯。Seeder 需確保所有 AssetType 均有預設值。

---

## 寫入路徑修改指引（Modification Guide）

> 本節記錄常見修改場景的**最小影響範圍**，避免每次需求都需全面瀏覽程式碼。

---

### 場景 A：新增 DB 欄位（含審計欄位）

需同時修改以下檔案（順序重要）：

| 步驟 | 檔案 | 說明 |
|------|------|------|
| 1 | `internal/database/schema.sql` | 在對應 `CREATE TABLE` 加入新欄位（**sqlc 的唯一源泉**） |
| 2 | `internal/database/migrations/YYYYMMDDNNN_xxx.sql` | `ALTER TABLE` 加入欄位（goose 格式：`-- +goose Up` / `-- +goose Down`） |
| 3 | `internal/database/queries/xxx.sql` | INSERT 加欄位與 `?`；UPDATE SET 加 `col = ?`；SELECT 加欄位 |
| 4 | `go generate ./internal/database/...` | 重新產生 `internal/database/sqlcdb/` |
| 5 | `internal/model/db/projection/xxx.go` | Projection Model struct 加欄位 + `//dbmap:sqlcdb=TypeName` 確認 |
| 6 | `go generate ./...` | 重新產生 `dbmap_gen.go` |
| 7 | `internal/repos/unit_of_work/event_store/projection_repo/interface.go` | 若方法簽名需新增參數，在此更新介面 |
| 8 | `internal/repos/unit_of_work/event_store/projection_repo/xxx.go` | 實作更新，傳入新欄位至 sqlcdb Params |
| 9 | `internal/services/projection/xxx.go` | Apply 方法中設定新欄位值後呼叫 repo |

> ⚠️ **常見錯誤**：只改 migration 忘記改 `schema.sql`，導致 sqlc 產生的 Params struct 缺少欄位。

---

### 場景 B：新增 Event 寫入欄位（event_store 表）

| 步驟 | 檔案 |
|------|------|
| 1 | `internal/database/schema.sql`（event_store 表） |
| 2 | `internal/database/migrations/xxx.sql` |
| 3 | `internal/database/queries/event_store.sql`（InsertEvent + 所有 SELECT） |
| 4 | `go generate ./internal/database/...` |
| 5 | `internal/model/db/event.go`（EventStore struct） |
| 6 | `internal/repos/unit_of_work/event_store/interface.go`（InsertEventParams） |
| 7 | `internal/repos/unit_of_work/event_store/event.go`（Insert 實作） |
| 8 | `internal/services/event.go`（Append 方法中注入值） |

---

### 場景 C：新增 Projection Repo 方法（全新 UPDATE/INSERT）

| 步驟 | 檔案 |
|------|------|
| 1 | `internal/database/queries/xxx.sql`（新增 named query） |
| 2 | `go generate ./internal/database/...` |
| 3 | `internal/repos/unit_of_work/event_store/projection_repo/interface.go`（新增方法簽名） |
| 4 | `internal/repos/unit_of_work/event_store/projection_repo/xxx.go`（實作） |
| 5 | `internal/services/projection/xxx.go`（在 Apply 中呼叫） |

---

### 場景 D：新增 EventType（全新事件）

| 步驟 | 檔案 |
|------|------|
| 1 | `internal/enums/event_types/`（宣告新 EventType 常數） |
| 2 | `go generate ./internal/enums/...` |
| 3 | `internal/model/payload/`（新增 Payload struct） |
| 4 | `internal/services/pipelines/`（新增 TypedPipeline） |
| 5 | `internal/services/pipelines/factory/base.go`（註冊 Pipeline） |
| 6 | 各相關 `internal/services/projection/xxx.go`（Apply switch case） |
| 7 | `internal/handler/`（新增 HTTP handler，組裝 AppendCmd） |

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
