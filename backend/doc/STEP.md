# F1：BFS 現金流分類修改修正（ISSUE-020）— 實作進度

> 關聯：ISSUE-020；計畫書 `doc/plan/BFS_ENGINE_MIGRATION.md` 5.2 節 F1。
> 依 `AGENTS.md`「變更幅度與重構」，本項是 bug fix，與 BFS 遷移的重構分開實作、分開 review。
> 共 6 個檔案，達到 `backend/CLAUDE.md` 的大型標準，因此拆成兩段。

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 建立 STEP.md | ✅ done |
| 1 | 程式修正（4 個檔案）：`internal/process/cashflow.go`、`internal/persistence/handle/cash_flow_cateory.go` 改用 `domain/event` 的 payload 型別；`internal/handler/handler.go` 註冊現金流分類 projector；`internal/handler/entry_cf_category.go` 補齊事件封套（`AggregateType=TRANSACTION`、aggregate ID 改用路徑的 `txn_uuid`、`ExpectedVersion` 取自請求、`MerchantID` 與 `UpdatedBy` 取自登入資訊）。驗證：`go build ./...` | ⬜ pending |
| 2 | handler 測試（2 個檔案）：`internal/handler/entry_cf_category_scenarios_test.go`、`internal/handler/entry_cf_category_test.go`，共用既有的 Ginkgo suite；測試資料以 SQL 直接寫入一筆交易與分錄；情境：儲存成功（事件欄位正確、`entry_cf_categories` 更新為已確認）、版本號不符、分錄不屬於該交易、交易不存在。驗證：`go generate ./...`、`go build ./...`、`go test ./...` | ⬜ pending |

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

（Phase 2 完成後填寫）
