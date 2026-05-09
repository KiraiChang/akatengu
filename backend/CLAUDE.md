# CLAUDE.md

本文件提供 Claude Code（claude.ai/code）在此程式庫工作時的指引。
始終用繁體中文說明。

---

## 限制事項（Constraints）

- **禁止**修改或讀取任何 `*_gen.go` 檔案。
- **禁止**修改或讀取 `internal/database/sqlcdb` 目錄下由 sqlc 產生的程式碼。
- 未經明確許可，**禁止**自動執行任何 `git push` 指令。
- **禁止**使用 AI 生成的假資料來測試功能；請使用真實或手動建立的測試資料。
- 遇到問題時**直接反映**，不要盲目修改；未釐清原因前不得做推測性修改。
- 瀏覽程式碼時**不重複閱讀**同一檔案，已讀過的內容直接引用記憶。

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

---

### 快取（Cache）

`internal/pkg/cache/` 提供**兩層快取**（本地記憶體 + 可選的分散式快取）。`internal/cache_service/account.go` 封裝帳戶查詢的快取邏輯，供 Pipeline 使用。