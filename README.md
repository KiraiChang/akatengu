# Akatengu

Akatengu 是一套會計管理系統，採前後端分離開發、單一後端部署：

- `backend/`：Go 1.26、`net/http`、SQLite、sqlc，以及事件溯源（Event Sourcing）
- `frontside/`：Svelte 5、TypeScript、Vite；正式建置產物會寫入後端的嵌入式靜態資源目錄

## 專案結構

```text
akatengu/
├── backend/                 Go API、領域邏輯、資料庫與靜態檔案服務
│   ├── cmd/                 應用程式與 code generator 入口
│   ├── internal/            handler、service、repository、model 等內部套件
│   └── doc/                 後端 ADR、測試規範、問題、技術筆記與計畫書
├── frontside/               Svelte SPA
│   ├── src/api/             API client
│   ├── src/components/      共用元件
│   ├── src/routes/          路由元件
│   ├── src/views/           功能頁面
│   ├── src/styles/          全域、元件與頁面樣式
│   └── doc/                 前端 ADR、問題與技術筆記
├── doc/README.md            全專案文件索引與維護規則
├── AGENTS.md                Codex／AI agent 共用開發指引
└── CLAUDE.md                Claude Code 專案入口
```

## 本機開發

### 後端

```bash
cd backend
go run ./cmd/main.go
```

API 監聽 `http://localhost:8080`（目前寫死於 `backend/cmd/main.go`）。其餘設定由環境變數提供，主要預設值可在 `backend/internal/bootstrap/config.go` 查閱。

> `config.go` 雖定義了 `SERVER_ADDR`、`SERVER_READ_TIMEOUT`、`SERVER_WRITE_TIMEOUT`，但 `backend/cmd/main.go` 尚未使用：設定這三個變數目前不會改變監聽位址，也不會套用 timeout。

> 預設 `DATA_SOURCE` 為 `../../output/akatengu/db.db`，相對於執行目錄解析；從 `backend/` 啟動時會指向 repository 上一層的 `output/akatengu/`。程式不會自動建立該目錄，請先建立，或以 `DATA_SOURCE` 指定其他路徑。

### 前端

```bash
cd frontside
npm install
npm run dev
```

Vite 開發伺服器預設監聽 `http://localhost:5174`，並將 `/api` 代理到 `http://localhost:8080`。

> `npm run build` 的輸出目錄是 `backend/internal/web/static`，且建置時會清空該輸出目錄。提交前請一併確認靜態資源差異是否符合預期。

## 驗證

在 `backend/` 執行：

```bash
go generate ./...
go build ./...
go test ./...
```

在 `frontside/` 執行：

```bash
npm run lint
npm run check
npm run build
```

只修改單一端時執行該端的驗證；跨前後端契約或整合流程的修改，兩端都要驗證。

## 架構摘要

後端會計資料的寫入以事件溯源為核心：

```text
HTTP handler
  -> EventStoreService.Append
  -> Pipeline（領域驗證與 payload 補全）
  -> UnitOfWork（事件、樂觀鎖與 Projection 同一交易）
  -> Projection（更新查詢用 read model）
```

查詢走 `backend/internal/repos/query/` 的唯讀 repository。前端透過 `/api` 存取後端；正式前端資源由 Go binary 嵌入並提供 SPA fallback。

更完整的架構決策、已知問題與待辦，請從 [專案文件索引](doc/README.md) 開始閱讀。

## AI 協作

- Codex 與其他支援 `AGENTS.md` 的工具：從根目錄的 [AGENTS.md](AGENTS.md) 開始。
- Claude Code：從根目錄的 [CLAUDE.md](CLAUDE.md) 開始；它會載入共用指引，並保留子專案的專屬規範。
