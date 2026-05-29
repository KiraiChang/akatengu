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
| [ADR-011](#adr-011-投資現金流量表investingfinancing-科目從-ni-重分類) | 投資現金流量表：INVESTING/FINANCING 科目從 NI 重分類 | Accepted | 2026-05-22 |
| [ADR-012](#adr-012-投資-pipeline-cashflowcategory-分錄標記策略) | 投資 Pipeline CashFlowCategory 分錄標記策略 | Accepted | 2026-05-22 |
| [ADR-013](#adr-013-直接法現金流量表固定利率分期還款已知限制) | 直接法現金流量表：固定利率分期還款已知限制 | Accepted | 2026-05-22 |
| [ADR-014](#adr-014-所有會計帳務異動必須透過事件溯源寫入查詢透過-queryrepo) | 所有會計帳務異動必須透過事件溯源寫入，查詢透過 query.Repo | Accepted | 2026-05-22 |
| [ADR-016](#adr-016-分頁-api-實作規範) | 分頁 API 實作規範 | Accepted | 2026-05-25 |
| [ADR-017](#adr-017-分錄組裝移至-pipeline-factory) | 分錄組裝移至 Pipeline Factory | Accepted | 2026-05-27 |
| [ADR-018](#adr-018-event-sourcing-projection-全面加入-uuid-以確保-replay-正確性) | Event Sourcing Projection 全面加入 UUID 以確保 Replay 正確性 | Accepted | 2026-05-29 |

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

## ADR-011 投資現金流量表：INVESTING/FINANCING 科目從 NI 重分類

- **狀態**：Accepted
- **日期**：2026-05-22
- **背景**：
  間接法現金流量表以淨利（NI）為起點，再調整非現金項目與營業外項目。
  若投資活動相關損益（已實現損益、手續費、交易稅）標記為 INVESTING，
  且這些科目已計入 NI，直接使用 `is.NetIncome` 會造成 INVESTING 與 OPERATING 雙重計算。
- **決策**：
  1. 在 `queryCashFlowChanges` 加入 `a.type AS account_type`，讓 Go 層識別科目類型。
  2. 在 `GetCashFlowStatement` 計算 `niReclassification`：
     加總所有標記 INVESTING 或 FINANCING 的 INCOME/EXPENSE 科目淨額（Credit - Debit）。
  3. CF 報表使用 `is.NetIncome - niReclassification` 作為 Operating NI，確保已重分類項目不重複計入。
  4. Direct Method（`GetDirectCashFlowStatement`）不受影響，Operating 以實際現金流計算，無需調整。
- **替代方案**：
  - **另建 queryOperatingNetIncome**：只查 `cash_flow_category IS NULL` 的損益科目計算 NI。
    缺點：需新增查詢，且與現有 `GetIncomeStatement` 邏輯分離，難以維護一致性。
  - **前端顯示時過濾**：前端自行減去 INVESTING 損益，不修改後端。
    缺點：CF 報表數字在後端即不正確，違反單一責任。
- **後果**：
  - 正面：CF 報表在投資損益重分類後仍能正確平衡（Beginning + Operating + Investing + Financing = Ending）。
  - 負面：`cfRawRow` 新增 `AccountType` 欄位，`queryCashFlowChanges` SQL 略微增長。

---

## ADR-012 投資 Pipeline CashFlowCategory 分錄標記策略

- **狀態**：Accepted
- **日期**：2026-05-22
- **背景**：
  投資買賣、公允價值評估的分錄預設無 CF 標記，現金流量表無法區分投資活動現金流。
- **決策**：

  | 事件 | 分錄 | CF 標籤 | 理由 |
  |------|------|---------|------|
  | BUY | DR 投資資產 | INVESTING | 非現金側：資產增加對應現金流出 |
  | BUY | CR 銀行 | NULL | 現金帳，始末餘額已捕捉 |
  | BUY | DR 手續費/交易稅費用 | INVESTING | 投資成本，重分類出 Operating NI |
  | SELL | CR 投資資產 | INVESTING | 資產減少對應現金流入（成本部分） |
  | SELL | CR/DR 已實現損益 | INVESTING | 損益重分類至投資活動 |
  | SELL | DR 手續費/交易稅費用 | INVESTING | 投資成本，重分類出 Operating NI |
  | SELL | DR/CR 累積未實現沖回（FVTPL/FVOCI） | INVESTING | 出售時一併沖回，屬投資活動 |
  | MARK FVTPL | DR/CR 投資資產 | OPERATING | 非現金調整，沖回 NI 膨脹 |
  | MARK FVTPL | CR/DR 未實現損益 | NULL | 留在 NI，由上方 OPERATING 抵銷 |
  | MARK FVOCI | 所有分錄 | NULL | 過 OCI（權益），不影響 NI，非現金交易於附註揭露 |
  | DIVIDEND | 所有分錄 | NULL | 股利收入屬於營業現金流入（IAS 7 允許） |

- **後果**：
  - 正面：CF 報表投資活動正確顯示完整收款（NetProceeds），MARK 非現金項目正確從 NI 沖銷。
  - 負面：FVOCI 非現金增值不在 CF 報表呈現，需於財報附註揭露（符合 IAS 7 規定）。

---

## ADR-013 直接法現金流量表：固定利率分期還款已知限制

- **狀態**：Accepted
- **日期**：2026-05-22
- **背景**：
  分期還款（InterestTypeFixedRate）每期銀行付款包含本金（融資活動）與利息（留在 NI，屬營業活動）。
  直接法透過識別「含 INCOME/EXPENSE 或 OPERATING 標記」的交易，取出其現金帳變動作為營業現金流。
  間接法則以 NI 為起點，加減調整項。

- **決策**：
  接受直接法對 FixedRate 分期的已知限制，不做額外修正。原因如下：

  1. **間接法正確**：利息費用已含於 NI，本金還款以 FINANCING 標籤正確歸類，CF 三大活動數字均正確。
  2. **直接法侷限**：銀行付款為單一分錄（`totalAmount = 本金 + 利息`），而 `operating_txns` 以交易為單位識別，
     利息費用為 EXPENSE 且未標 INVESTING/FINANCING，故整筆交易被納入 operating_txns。
     直接法將全部銀行付款視為營業現金流出，同時 FINANCING 也顯示本金還款 → 本金被重複計算。

- **替代方案**：
  - 拆分交易：將本金與利息拆成兩筆獨立交易 → 破壞「一次銀行轉帳對應一筆交易」的業務語義，不採用。
  - 以分錄層級識別 operating_txns：只取 INCOME/EXPENSE 的非投融資分錄金額，而非整筆交易銀行流動 → 需重構直接法查詢架構，範圍過大，非當前需求。

- **後果**：
  - 正面：間接法 CF 報表完整正確，適合作為主要財務報告工具。
  - 負面：直接法 CF 報表在含固定利率分期還款的期間，營業活動現金流出與融資活動現金流出合計會高於實際現金減少額（本金被重複計算）。使用者應優先參考間接法報表。

---

## ADR-014 所有會計帳務異動必須透過事件溯源寫入，查詢透過 query.Repo

- **狀態**：Accepted
- **日期**：2026-05-22
- **背景**：
  系統新增預付費用（Prepaid）與固定資產（Fixed Asset）模組，涉及分錄建立、攤提、折舊、處分等多個寫入操作。
  這些操作直接影響會計科目餘額與財務報表，若各模組各自處理寫入，將面臨以下風險：
  - 並發衝突導致帳務不一致（多個請求同時修改同一筆資產狀態）
  - Projection 讀模型與事件流不同步（若跳過事件系統直寫）
  - 欠缺完整稽核軌跡（無法重播事件以重建任意時間點狀態）

- **決策**：
  強制要求所有新模組的帳務寫入一律走事件溯源鏈路，不得例外：

  ```
  Handler → EventStoreService.Append
    → Pipeline（驗證領域規則，建立分錄 entries）
      → UnitOfWork.Do（單一 SQLite 交易）
          ├─ VersionRepository.UpdateIfVersionMatch（樂觀鎖）
          ├─ 事件持久化
          └─ Projection.Apply（讀模型同步）
  ```

  查詢操作（GET 系列 API）透過 `query.Repo`（唯讀層）執行，不接觸 UnitOfWork，不受樂觀鎖約束。
  具體對應關係：

  | 操作 | 路徑 |
  |------|------|
  | PREPAID_CREATED / AMORTIZED / DISPOSED | EventStoreService.Append |
  | ASSET_PURCHASED / DEPRECIATED / DISPOSED | EventStoreService.Append |
  | GET /api/prepaid/ , GET /api/asset/ | query.Repo（無鎖） |

- **替代方案**：
  - 直接在 Service 層 INSERT/UPDATE：實作簡單，但無樂觀鎖保護、無稽核軌跡、Projection 需手動維護，不採用。
  - 為查詢也加鎖：過度設計，SELECT 不修改狀態，加鎖只會降低吞吐量且無實際保護價值，不採用。

- **後果**：
  - 正面：帳務一致性由 UnitOfWork 樂觀鎖統一保護；完整事件軌跡供稽核與重播；Projection 自動同步。
  - 負面：新模組必須實作完整的 Pipeline + Projection 鏈路，初始開發量較直寫多；每次新增事件類型需在 `factory/base.go` 與 `projection/base.go` 同步登記。

---

## ADR-015 預付費用與固定資產的現金流量分類設計

- **狀態**：Accepted
- **日期**：2026-05-22
- **背景**：
  預付費用攤提與固定資產折舊均為非現金認列，若分錄不帶 `cash_flow_category` 標籤，間接法 CF 報表的 OperatingTotal 會因 NI 下降而出現負數，但實際並無現金流出。資產處分的利得/損失若留在 NI，會使 OperatingTotal 包含投資活動金額，破壞分類一致性。

- **決策**：
  依以下原則設定 `journal_entries.cash_flow_category`：

  | 事件 | 分錄 | CF 標籤 | 理由 |
  |------|------|---------|------|
  | 預付創建 | DR 預付科目 | `OPERATING` | 現金用於取得預付費用，屬營業活動現金流出；間接法：預付資產增加 → 負調整 |
  | 預付攤提 | CR 預付科目 | `OPERATING` | 非現金認列，加回 NI 中已扣除的費用，使 OperatingTotal=0 |
  | 預付提前終止 | CR 預付科目 | `OPERATING` | 同攤提，非現金一次認列 |
  | 資產購入（現金） | DR 資產帳戶 | `INVESTING` | 現金用於購置資產，屬投資活動現金流出 |
  | 資產購入（租賃） | 無標籤 | — | 非現金交易，NI=0 且無調整，NetChange=0 |
  | 資產折舊 | CR 累計折舊 | `OPERATING` | 非現金認列，加回 NI 中已扣除的折舊費用，使 OperatingTotal=0 |
  | 資產處分（利得/損失） | DR/CR 損益科目 | `INVESTING` | 將損益重分類至投資活動，從 NI 移出；InvestingTotal = 現金收款 |

- **替代方案**：
  - 依 `accounts.cash_flow_category` 自動推導：CF 報表 SQL 無法直接使用科目層級分類推導 `journal_entries` 標籤，需在查詢層做 JOIN 並附加條件，大幅增加查詢複雜度，不採用。
  - 不標 OPERATING，讓 OperatingTotal 自然反映：攤提/折舊期間 OperatingTotal 為負，但無現金流出，間接法與直接法數字會不一致，不採用。

- **後果**：
  - 正面：間接法與直接法 OperatingTotal 保持一致；6 個 CF 測試驗證各情境均通過。
  - 負面：每個 `applyXxx` 函式需要明確設定 CF 標籤，新增事件時不得遺漏。

---

## ADR-016 分頁 API 實作規範

- **狀態**：Accepted
- **日期**：2026-05-25
- **背景**：
  系統多個模組（投資、帳戶、分錄分析等）均需要分頁查詢 API。
  早期部分端點使用自訂 `Page`/`Size` 結構與獨立 COUNT 查詢，與投資模組的慣例不一致，造成前端串接負擔與開發標準分歧。

- **決策**：
  所有分頁 API 一律遵循以下四層規範：

  **1. SQL 查詢（`internal/database/queries/*.sql`）**
  - 分頁查詢函式名稱以 `Paged` 為後綴（如 `GetAccountJournalEntriesPaged`）
  - 使用 `WITH total AS (SELECT COUNT(*) AS cnt ...)` 將 total 嵌入主查詢，一次往返取得資料與總筆數，不另外發送獨立 COUNT 查詢
  - 範例結構：
    ```sql
    -- name: GetXxxPaged :many
    WITH total AS (
        SELECT COUNT(*) AS cnt FROM ... WHERE ...
    )
    SELECT ..., total.cnt AS total
    FROM ..., total
    WHERE ...
    ORDER BY ...
    LIMIT @limit OFFSET @offset;
    ```

  **2. Repository 層（`internal/repos/query/*.go`）**
  - 接受 `model.PaginationParams`（含 `Offset`、`Limit`，由 Handler 呼叫 `SetDefaults()` 後傳入）
  - 從 `rows[0].Total` 取得總筆數（空結果時回傳 0）
  - 函式簽章慣例：`GetXxxPaged(ctx, ...filterParams, req model.PaginationParams) ([]T, int64, error)`

  **3. Service 層**
  - 直接傳遞 `model.PaginationParams` 與篩選條件至 Repo，不做額外包裝

  **4. Handler 層（`internal/handler/*.go`）**
  - Query params 固定使用 `page`（頁碼，從 1 起）與 `page_size`（每頁筆數）
  - 呼叫 `req.SetDefaults()`：預設 `PageSize=20`，最大 `PageSize=100`
  - 回傳格式：`response.OK(w, model.PaginateWithTotal(result, req, total))`
  - 範例：
    ```go
    page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
    pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
    req := model.PaginationParams{Page: page, PageSize: pageSize}
    req.SetDefaults()
    result, total, err := h.s.GetXxxPaged(ctx, ..., req)
    response.OK(w, model.PaginateWithTotal(result, req, total))
    ```

- **替代方案**：
  - 自訂 `Page`/`Size` struct + 獨立 COUNT 查詢：多一次 DB 往返，且各端點行為不一致，不採用。
  - 全量載入後在 Go 中切頁（`PaginateWithoutTotal`）：適用於小資料集，但分析類查詢資料量不定，不應依賴此方式，不採用。

- **後果**：
  - 正面：所有分頁 API 對外格式統一（`data` + `meta` 含 `total_count`、`total_pages`）；DB 只需一次查詢；前端接入規則固定。
  - 負面：`SetDefaults()` 強制 `PageSize` 上限為 100，若特定場景需要更大批次，需另行評估替代方案（如游標分頁）。

---

## ADR-017 分錄組裝移至 Pipeline Factory

- **狀態**：Accepted
- **日期**：2026-05-27
- **背景**：
  `AccountBalanceRealtimeProjection` 負責維護即時餘額，需要讀取每個事件的 journal entry 分錄。
  對於 Installment / Prepaid / FixedAsset 等事件，分錄 entries 不存在 payload 中，而是由業務邏輯計算組裝。
  初始方案是由 `TransactionProjectionService` 在 Projection 階段組裝，再透過 state pointer mutation 傳給 `AccountBalanceRealtimeProjection`，
  但這造成兩個 Projection 之間的執行順序強依賴，容易在調整 `NewProjection()` 順序時靜默失效。

- **決策**：
  將分錄組裝邏輯從 `TransactionProjectionService` 移至 **Pipeline Factory**（`factory/installment.go`、`factory/prepaid.go`、`factory/asset.go`）：

  1. 在 `payload/` 套件新增 builder 函數（`BuildXxxTransaction`），接受 DB model 物件，返回 `TransactionCreatedPayload`。
  2. Factory 的 `Project()` 方法在完成 state 查詢後，呼叫對應 builder 並將結果存入 `c.Transaction`（Pipeline 階段，DB 交易開始前）。
  3. `TransactionProjectionService` 讀取 `st.Transaction` 直接呼叫 `applyTransaction` 寫入分錄，不再自行組裝 entries。
  4. `AccountBalanceRealtimeProjection` 透過 `applyStateTransaction[S txnHolder]` 讀取 `st.Transaction` 更新 running balance。
  5. 計算輔助函數 `AmortizationAmount`、`DepreciationAmount` 提升為 `payload/` 套件公開函數，`PrepaidProjectionService` 與 `FixedAssetProjectionService` 也改用這些公開函數（單一來源）。

  兩個 Projection 現在各自**獨立**讀取 factory 預建的 `st.Transaction`，執行順序不再有強依賴。

- **替代方案**：
  - **State Pointer Mutation**（舊方案）：`TransactionProjectionService` 組裝後注入 `st.Transaction`，後續 Projection 讀取。Projection 執行順序必須固定，新增事件時兩處都需更新，靜默失效風險高，已廢棄。
  - **DB 查詢回查**：`AccountBalanceRealtimeProjection` 以 txnId 查 `journal_entries`。txnId 在同一 TX 內不易取得，且增加 DB 往返，不採用。

- **後果**：
  - **正面**：分錄組裝邏輯集中在 `payload/` builder 函數，可獨立單元測試；Projection 互不依賴，執行順序可自由調整；新增事件只需在 factory + 一個 Projection（各加一行）。
  - **負面**：`AmortizationAmount`（攤提金額）和 `DepreciationAmount`（折舊金額）在 factory 和 `TransactionProjectionService`（`InsertPrepaidAmortization` / `InsertFixedAssetDepreciation` 的 Amount 欄位）各計算一次，但函數為純函數且結果一致，可接受。

---

## ADR-018 Event Sourcing Projection 全面加入 UUID 以確保 Replay 正確性

- **狀態**：Accepted
- **日期**：2026-05-29
- **背景**：
  Projection 資料表（如 `transactions`、`journal_entries`、`investment_movements`、`investment_lots` 等）
  的主鍵均為 SQLite AUTOINCREMENT 整數。全量重播（TruncateProjections + 逐事件 Apply）時，
  所有 projection 資料被清空後重新寫入，但 SQLite AUTOINCREMENT 的計數器不隨 DELETE 重置，
  因此重播後的自增 ID 會比原始值更大。此時跨表的整數 FK（如 `journal_entries.txn_id` → `transactions.txn_id`、
  `investment_lots.movement_id` → `investment_movements.movement_id`）雖然在「同一筆 Apply 呼叫內」仍可正確取得剛插入的 ID，
  但一旦需要「後序事件引用前序事件建立的記錄」（如 InvestmentSold 的 lot_uuid 引用先前 InvestmentBought 建立的 lot），
  整數 ID 在重播後與原始不同，FK 關聯斷裂。

- **決策**：
  為每個 projection 資料表新增程式端生成的 UUID 欄位，跨表引用改為 UUID 版 FK，確保 replay 後關聯正確重建：

  1. **UUID 生成策略**：所有 UUID 由 Go 程式碼在 Projection Apply 時生成，不依賴 SQLite DEFAULT：
     - 主體 UUID（如 `txn_uuid`、`movement_uuid`）：直接使用 `ct.Event.EventUuid`（事件自身的 UUIDv7），1:1 關係時語義最直觀。
     - 衍生 UUID（如 `entry_uuid`、`lot_uuid`、`payment_uuid`）：由 `uuidx.NewFromEvent(eventUUID, qualifier)` 衍生，
       使用 UUID v5（SHA1 deterministic），給定相同 event UUID + qualifier 永遠產生相同 UUID，確保 replay 冪等。
  2. **UUID 工具**：`internal/pkg/uuidx/NewFromEvent(eventUUID, qualifier)` 統一入口。
  3. **資料表變更範圍**：17 張 projection 資料表新增 UUID 欄位（詳見 migration `20260529002_add_uuid_to_projections.sql`）：
     - `transactions.txn_uuid`、`journal_entries.entry_uuid/txn_uuid/ledger_uuid`
     - `ledger_accounts.ledger_uuid`
     - `investment_movements.movement_uuid/investment_uuid/event_uuid`
     - `investment_lots.lot_uuid/investment_uuid/movement_uuid`
     - `investment_lot_disposals.disposal_uuid/lot_uuid/movement_uuid`
     - `investment_positions.position_uuid/investment_uuid`
     - `installments.installment_uuid`、`installment_payments.payment_uuid/installment_uuid`
     - `prepaids.prepaid_uuid`、`prepaid_amortizations.amortization_uuid/prepaid_uuid`
     - `fixed_assets.asset_uuid`、`fixed_asset_depreciations.depreciation_uuid/asset_uuid`
     - `period_closings.closing_uuid`、`snapshots.snapshot_uuid`
  4. **整數 ID 保留**：不移除現有整數 PK/FK，UUID 欄位與整數欄位並存，利用整數 PK 保持 SQLite 效能。
  5. **`journal_entries.ledger_uuid` 特殊處理**：因 `TransactionCreatedPayload` 不帶 LedgerAccount 狀態，
     `UpsertJournalEntries` 在 repo 層自動 SELECT `ledger_accounts.ledger_uuid` 補入，不需更動 payload 或 pipeline。

- **替代方案**：
  - **重置 sqlite_sequence**：TruncateProjections 時重置自增計數器。問題：多租戶環境下其他商戶資料仍存在，
    重置會導致 ID 與現有 row 衝突，且 AUTOINCREMENT 設計上就是保證不重用，此方案從根本上違反設計意圖。
  - **將 ID 存入 payload**：在 Handler 端生成 UUID 寫入事件 payload，Projection Apply 從 payload 讀取後以 INSERT ... WITH EXPLICIT ID 插入。
    優點是保留整數 PK 語義；缺點是所有 payload struct 必須大幅修改，且每個 BuildXxx 函式都需注入多個 UUID，維護負擔高。
  - **UUID 替換整數 PK**：完全以 UUID 取代整數 PK。優點是最純粹；缺點是 SQLite 效能下降，且現有大量程式碼依賴整數 PK，
    遷移成本極高。
- **後果**：
  - 正面：Replay 後所有跨表 UUID FK 關聯正確重建，不受 AUTOINCREMENT 計數器影響；UUID 從 event_uuid deterministic 衍生，同一事件重複 replay 結果完全一致（冪等）。
  - 負面：每次 INSERT 需設定多個 UUID 欄位，projection Apply 方法複雜度略增；`journal_entries.ledger_uuid` 需在 repo 層額外一次 SELECT lookup，有 N+1 查詢風險（實際影響很小，因 ledger 數量有限）。

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
