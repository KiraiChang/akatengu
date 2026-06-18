# BFS Single-Event + Persistence 整合 — 實作進度

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 建立 STEP.md | ✅ done |
| 1 | safety: BatchProcessor → EventProcessor，更新 4 個 middleware | ✅ done |
| 2 | safety tests 更新（suite / scenarios / spec） | ✅ done |
| 3 | Queue 改寫：FIFO + depth（Push/Pop） | ✅ done |
| 4 | kernel/mediator 變更：event.AggregateType、EventResult.State、Executor.Dispatch | ✅ done |
| 5 | persistence/eventstore + persistence/projection 新增 | ✅ done |
| 6 | BFSEngine 改寫：Run(single evt) + WithStore + WithProjector + engine tests | ✅ done |
| 7 | go test ./... 全套驗證 | ✅ done |

## 最終驗證結果（Phase 7 完成後）

```
go build ./...  → 通過（無錯誤）
go test ./...   → 全部通過

ok  akatengu/internal/handler                                         16.046s
ok  akatengu/internal/handler/middleware                              (cached)
ok  akatengu/internal/pkg/enumx                                       (cached)
ok  akatengu/internal/repos/query                                     27.486s
ok  akatengu/internal/repos/unit_of_work/event_store/projection_repo  20.795s
ok  akatengu/internal/runtime/engine                                  (cached)
ok  akatengu/internal/runtime/safety                                  15.908s
ok  akatengu/internal/services                                        20.198s
ok  akatengu/internal/services/pipelines/factory                      14.432s
ok  akatengu/internal/services/projection                             16.061s
```
