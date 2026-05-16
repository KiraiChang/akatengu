# CLAUDE.md

本文件提供 Claude Code（claude.ai/code）在此程式庫工作時的指引。
始終用繁體中文說明。

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

---

## 限制事項（Constraints）

- **禁止**修改或讀取任何 `*_gen.go` 檔案。
- **禁止**修改或讀取 `internal/database/sqlcdb` 目錄下由 sqlc 產生的程式碼。
- 未經明確許可，**禁止**自動執行任何 `git push` 指令。
- **禁止**使用 AI 生成的假資料來測試功能；請使用真實或手動建立的測試資料。
- 遇到問題時**直接反映**，不要盲目修改；未釐清原因前不得做推測性修改。
- **需要瀏覽專案時，必須先向使用者說明瀏覽目的並取得確認，禁止未獲許可主動瀏覽任何檔案或目錄。**
- 瀏覽程式碼時**不重複閱讀**同一檔案，已讀過的內容直接引用記憶。
- **Repository 層一律使用 sqlc 產生的型別安全查詢**；嚴禁直接撰寫 `sqlx` 查詢，除非該 SQL 動態性質（如欄位清單或 `WHERE` 條件在執行期才確定）確實無法以 sqlc 靜態產生，且**須在行內以註解說明為何不得不改用 `sqlx`**。如遇此情況，應先提出說明並取得明確許可，再動手實作。
- **所有金額相關欄位必須使用 `decimal.Decimal` 型別**，禁止用 `float64` 儲存金額資料。在 `sqlc.yaml` 的 `overrides` 區段以 `column: "table.column"` 格式宣告型別覆寫（`github.com/shopspring/decimal` → `Decimal`）。若欄位來自計算式（如 `SUM`、`COALESCE` 等 aggregate 函數），sqlc 無法直接追蹤欄位來源，必須先建立 Database View（參考 `v_account_balances`、`v_parent_balance_agg`），再對 View 欄位加覆寫，不可在 Go 程式碼中使用 `decimal.NewFromFloat()` 轉型來規避。

---

## 驗證流程

每次修改後依序執行：

1. `go generate ./...`：生成 sqlc, enums, dbmap，檢查以上生成遇到的錯誤
2. `go build`：專案建置並檢查告警或錯誤
3. `go test ./...`：執行全部測試，確認無測試失敗

回報格式：列出每個步驟的實際輸出，有錯誤則附上錯誤訊息與行號，測試失敗時須提出修正建議。

---

## 常用指令（Commands）

```bash
# 啟動開發伺服器
go run ./cmd/main.go

# 建置專案
go build ./...

# 執行全部測試
go test ./...

# 執行單一套件的測試（範例：services 層）
go test ./internal/services/...

# 執行列舉型別的程式碼產生器（修改 enum 定義後必須執行）
go generate ./internal/enums/...

# 或直接執行產生器二進位
go run ./cmd/enumx-gen
```

---

## 術語對照（Terms）

| 中文說法 | 程式碼名稱 | 資料表 | 說明 |
|---------|-----------|--------|------|
| 科目 | Account | `accounts` | 會計科目樹（如 1100 現金、5100 食費） |
| 帳戶 / 實體帳戶 | LedgerAccount | `ledger_accounts` | 具體的銀行帳戶、信用卡等 |
| 科目餘額 | AccountBalance | `v_account_summary` | 科目層級的借貸加總 |
| 帳戶餘額 | LedgerAccountBalance | `v_account_balances` | 帳戶層級的借貸加總 |
| 月結 / 關帳 | PeriodClose | `period_closings` | 月度或年度結帳紀錄 |

---

## 系統架構（Architecture）

本系統是**基於 Go 語言的事件溯源（Event Sourcing）會計系統**，資料層使用 `sqlite`、`sqlx`、`sqlc` 與 `golang-migrate`。伺服器監聽 `:8080`，採用標準 `net/http` 函式庫。

---

### 核心寫入流程（Event Flow）

每一筆寫入操作都必須經過以下完整鏈路：

```
HTTP 請求
  └─► handler/               解析請求，組成 cmd.AppendCmd
        └─► EventStoreService.Append   唯一的寫入入口點
              └─► pipelines/           驗證與豐富事件 payload（每個 EventType 在 factory/base.go 中對應一個 PipelineRunner）
                    └─► UnitOfWork.Do  單一 SQLite 交易包覆所有寫入：
                          ├─ 樂觀鎖版本檢查（VersionRepository.UpdateIfVersionMatch）
                          ├─ 事件寫入
                          ├─ 所有 Projection.Apply 呼叫（讀模型更新）
                          └─ 每 50 個版本建立一次快照（Snapshot）
```

---

### 核心抽象層（Key Abstractions）

#### `enumx` 套件（`internal/pkg/enumx/`）

型別安全的列舉系統，由全域 registry 支撐。列舉型別以 `//enumx:enum` 注釋宣告，並透過 `cmd/enumx-gen` 產生程式碼。**修改列舉常數後，務必執行 `go generate ./internal/enums/...`。**

---

#### Projections（`internal/services/projection/`）

讀模型建構器（Read Model Builder）。每個 Projection 實作 `Projection` 介面（`Name()`、`Apply()`）。

- `base.go` 中的 `NewProjection()` 是**唯一的 Projection 註冊清單**，新增時一律在此登記。
- Projection 與事件寫入在**同一個資料庫交易**內執行。
- 各 Projection 只查詢自身需要的資料，**不依賴 Pipeline 傳入預組裝資料**（避免違反開放封閉原則）。

---

#### Pipelines（`internal/services/pipelines/`）

寫入前的業務邏輯層（Pre-write Business Logic）。`TypedPipeline[S, P]` 結合：

- `S`（State）：從資料庫取出的目前狀態
- `P`（Payload）：來自 JSON 的命令資料

Pipeline 負責驗證領域規則，確認後事件才落地。每個 EventType 在 `factory/base.go` 中完成註冊。

---

#### UnitOfWork（`internal/repos/unit_of_work/event_store/`）

`EventStoreRepositories` 打包所有交易性 Repository（Event、Version、Snap、Check、Projection），**僅在 `UnitOfWork.Do` 回呼內有效**。Service 層**不持有直接的資料庫參考**，一律透過 UoW 回呼取得 Repository。

---

#### 查詢層 `query.Repo`（`internal/repos/query/`）

**唯讀**查詢層，用於交易外的狀態查詢，例如 Pipeline 的狀態讀取與報表查詢。

---

### 設定（Configuration）

所有設定均由環境變數驅動，詳見 `internal/bootstrap/config.go`。主要預設值：

| 變數 | 預設值 |
|---|---|
| `DB_DRIVER` | `sqlite` |
| `DATA_SOURCE` | `../../output/akatengu/db.db` |
| `JWT_SECRET` | `this is a secret JWT` |
| `APP_ENV` | `local` |

---

### 資料庫（Database）

- **Migration**：位於 `internal/database/migrations/`，透過 `//go:embed` 嵌入，伺服器啟動時自動執行。
- **Seeds**：位於 `internal/database/seeds/`，`accounts.sql` 每次啟動都會執行（冪等設計）。
- SQLite 設定為 **WAL 模式**，單一連線（`MaxOpenConns=1`）。
- **查詢撰寫原則**：新增或修改 SQL 時，先在 `internal/database/sqlcdb/` 對應的 `.sql` 檔中定義查詢，執行 `sqlc generate` 產生型別安全程式碼，再於 Repository 層呼叫產生的方法。僅在 sqlc 靜態產生確實不可行時（例如動態欄位、動態 `IN` 清單），才可退而使用 `sqlx`，並須附上理由說明。

---

### 快取（Cache）

`internal/pkg/cache/` 提供**兩層快取**（本地記憶體 + 可選的分散式快取）。`internal/cache_service/account.go` 封裝帳戶查詢的快取邏輯，供 Pipeline 使用。