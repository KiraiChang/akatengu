# 現金流量分類抽離為獨立 Projection — 實作進度

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 建立 STEP.md | ✅ done |
| 1 | Migration + SQL queries + sqlc generate + Projection Model | ✅ done |
| 2 | 新 EventType + Payload + UoW Repo Interface + Repo 實作 | ✅ done |
| 3 | Pipeline Factory + 在 factory/base.go 完成註冊 | ✅ done |
| 4 | 新 CashFlowCategoryProjection + 更新 TransactionProjectionService + 更新 projection/base.go | ✅ done |
| 5 | Query Repo + Handler + 路由 | ✅ done |
| 6 | 重寫 CF 報表查詢（report.go） | ✅ done |
| 7 | 清理：移除 journal_entries.cash_flow_category（migration + struct + payload + payload builders） | ✅ done |

## 最終驗證結果（Phase 7 完成後）

```
go build ./...  → 通過（無錯誤）
go test ./...   → 全部通過（exit code 0）

ok  akatengu/internal/handler              (cached)
ok  akatengu/internal/handler/middleware   (cached)
ok  akatengu/internal/repos/query          (cached)
ok  akatengu/internal/repos/unit_of_work/event_store/projection_repo  (cached)
ok  akatengu/internal/services             13.057s
ok  akatengu/internal/services/pipelines/factory  (cached)
ok  akatengu/internal/services/projection  7.157s
```
