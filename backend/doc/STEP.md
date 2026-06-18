# EventProcessor 回傳 result.EventResult + Policy 整合 — 實作進度

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 建立 STEP.md | ✅ done |
| 1 | result.Policy 新增；EventProcessor 回傳改 (result.EventResult, error) | ✅ done |
| 2 | 4 個 safety middleware 更新：回傳 EventResult + 設定 Policy | ✅ done |
| 3 | safety tests 更新（noop / withChildren / scenarios / spec） | ✅ done |
| 4 | EventError 移除 Retryable/Fatal；NewRuntimeError 簡化；更新所有呼叫點 | ✅ done |
| 5 | engine 層更新：bfs_engine / executor / tests | ✅ done |
| 6 | go test ./... 全套驗證 | ✅ done |

## 最終驗證結果（Phase 6 完成後）

```
go build ./...  → 通過（無錯誤）
go test ./...   → 全部通過

ok  akatengu/internal/handler                                         (cached)
ok  akatengu/internal/handler/middleware                              (cached)
ok  akatengu/internal/repos/query                                     (cached)
ok  akatengu/internal/repos/unit_of_work/event_store/projection_repo  (cached)
ok  akatengu/internal/runtime/engine                                  (cached)
ok  akatengu/internal/runtime/safety                                  (cached)
ok  akatengu/internal/services                                        (cached)
ok  akatengu/internal/services/pipelines/factory                      (cached)
ok  akatengu/internal/services/projection                             (cached)
```
