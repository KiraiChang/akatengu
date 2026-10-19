# Akatengu Frontend

Akatengu 會計管理系統的前端 SPA，使用 Svelte 5、TypeScript 與 Vite。正式建置產物由 Go 後端嵌入並提供。

## 本機開發

```bash
npm install
npm run dev
```

開發伺服器監聽 `http://localhost:5174`，並將 `/api` 代理到 `http://localhost:8080`；需先啟動後端，方式見[根目錄 README](../README.md)。

> `npm run build` 會清空並重建 `../backend/internal/web/static`，提交前請確認靜態資源差異是否符合預期。

## 相關文件

- [根目錄 README](../README.md)：專案結構、啟動方式與架構摘要
- [AGENTS.md](../AGENTS.md)：AI agent 共用規範與跨端契約
- [CLAUDE.md](CLAUDE.md)：前端開發規範與驗證流程
- [doc/](doc/)：前端 ADR、已知問題、技術筆記與待辦
