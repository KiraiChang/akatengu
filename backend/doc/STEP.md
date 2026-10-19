# F1：BFS 現金流分類修改修正（ISSUE-020）— 實作進度

> 關聯：ISSUE-020；計畫書 `doc/plan/BFS_ENGINE_MIGRATION.md` 5.2 節 F1。
> 依 `AGENTS.md`「變更幅度與重構」，本項是 bug fix，與 BFS 遷移的重構分開實作、分開 review。
> 共 7 個檔案，達到 `backend/CLAUDE.md` 的大型標準，因此拆成兩段。

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 建立 STEP.md | ✅ done |
| 1 | 程式修正（5 個檔案）：`internal/process/cashflow.go`、`internal/persistence/handle/cash_flow_cateory.go` 改用 `domain/event` 的 payload 型別；`internal/handler/handler.go` 註冊現金流分類 projector；`internal/handler/entry_cf_category.go` 補齊事件封套（`AggregateType=TRANSACTION`、aggregate ID 改用路徑的 `txn_uuid`、`ExpectedVersion` 取自請求、`MerchantID` 與 `UpdatedBy` 取自登入資訊）；`internal/persistence/repos/store.go` 找不到版本紀錄時，只允許 `expected_version` 為 0 建立新 aggregate（review 補充，否則新 aggregate 可繞過版本檢查）。驗證：`go build ./...` | ✅ done（2026-10-19：加入 `store.go` 修正後重新驗證，Docker 內 `go generate ./...`、`go build ./...` 皆 exit 0；`gofmt -l` 無輸出） |
| 2 | handler 測試（2 個檔案）：`internal/handler/entry_cf_category_scenarios_test.go`、`internal/handler/entry_cf_category_test.go`，共用既有的 Ginkgo suite；測試資料以 SQL 直接寫入一筆交易與分錄；情境（共 8 個）：儲存成功（事件欄位正確、`entry_cf_categories` 更新為已確認）、既有版本 1 正常遞增到 2、過期版本號、首次寫入時 `expected_version` 不為 0（7、-1）、分錄不屬於該交易、交易不存在；以及 1 個經正式 `NewMux`（含 JWT 與商戶驗證）的成功情境，用來攔截正式組裝漏註冊 projector 的回歸（review 補充）。驗證：`go generate ./...`、`go build ./...`、`go test ./...` | ✅ done（2026-10-19，見最終驗證結果） |

## 範圍外

- 被註解掉的舊程式碼（無關清理，依規範不順手刪除）。
- 舊路徑使用的 `model/payload.TransactionCFCategoryUpdatedPayload`（`POST /api/event/append` 仍在使用）。
- API、前端，以及 engine 的交易邊界與 projector 的 `state` 參數（計畫書 R1）。

## 驗證環境

- 主機無法直接執行 `go`，以 Docker（`golang:1.26`）驗證。
- 主機只有 2 GB 記憶體：使用預建的 sqlc binary，並設定 `GOFLAGS=-p=1`、`GOGC=20`。
- `go test` 附帶的 vet 若被 OOM kill，另以 `go test -vet=off ./...` 確認測試結果，並在此照實記錄。
- in-memory SQLite 每條新連線都是空 DB，測試需將連線數設為 1。

## 最終驗證結果

2026-10-19，Docker（`golang:1.26`，Go 1.26.9）：

```
go generate ./...        → exit 0
go build ./...           → exit 0
gofmt -l（新增的測試檔） → 無輸出
go test ./...            → exit 1：所有套件皆 ok，但 go test 附帶的 vet 分析
                           modernc.org/sqlite/lib 時被 OOM kill（主機 2 GB 記憶體）
go test -vet=off ./...   → exit 0

ok  akatengu/internal/handler                                        2.397s
ok  akatengu/internal/handler/middleware
ok  akatengu/internal/pkg/enumx
ok  akatengu/internal/pkg/enumx_tools
ok  akatengu/internal/pkg/jwt
ok  akatengu/internal/repos/query
ok  akatengu/internal/repos/unit_of_work/event_store/projection_repo
ok  akatengu/internal/runtime/engine
ok  akatengu/internal/runtime/safety
ok  akatengu/internal/services
ok  akatengu/internal/services/pipelines/factory
ok  akatengu/internal/services/projection

go test -vet=off -count=1 -v ./internal/handler/ -ginkgo.v
Ran 18 of 18 Specs — 18 Passed（新增 8 個：7 個 UpdateCFCategory、1 個 NewMux；既有 10 個 merchant）

回歸檢查：在容器內的副本刪除 NewMux 的 projector 註冊後重跑
Ran 18 of 18 Specs — 17 Passed | 1 Failed
失敗的是 NewMux 情境（entry_cf_categories 查無資料），確認可攔截此回歸
```

vet 因記憶體不足未能在本機完整執行；在記憶體足夠的環境執行 `go test ./...` 可補上 vet 檢查。
