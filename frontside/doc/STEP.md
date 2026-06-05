# 銀行對帳單 UI 實作計畫

根據 commit `bbf6959`（add upload csv bank transaction statement）所新增的後端 API，分三個 Phase 建立前端介面。

---

## Phase 1 — CSV 範本管理（本次對話）

**目標**：讓使用者能新增、編輯、停用 CSV 解析範本（對應不同銀行的對帳單格式）。

### 新增檔案

| 檔案 | 說明 |
|---|---|
| `src/types/bankStatement.ts` | 所有 bank statement 相關型別定義 |
| `src/api/bankStatement.ts` | 所有 bank statement API 呼叫封裝 |
| `src/views/BankStatementTemplate.svelte` | CSV 範本 CRUD 頁面 |
| `src/styles/views/bankStatement.css` | 樣式（只放 bank statement 相關） |

### 修改檔案

| 檔案 | 修改內容 |
|---|---|
| `src/routes/Home.svelte` | 加入「銀行對帳」選單群組與 `/home/bank-statement/template` 路由 |
| `src/main.ts` | import 新樣式檔 |

### 涉及 API

- `GET /bank-statement/template` — 取得所有範本
- `POST /bank-statement/template` — 建立範本
- `PUT /bank-statement/template/{template_id}` — 更新範本
- `DELETE /bank-statement/template/{template_id}` — 停用範本

### 範本欄位說明

| 欄位 | 型別 | 說明 |
|---|---|---|
| `template_name` | string | 範本名稱 |
| `encoding` | string | 檔案編碼（UTF-8、Big5 等） |
| `skip_rows` | number | 跳過前幾行（表頭） |
| `date_column` | number | 日期欄位索引（0-based） |
| `date_format` | string | 日期格式（e.g. `2006-01-02`） |
| `description_column` | number | 說明欄位索引 |
| `debit_column` | number? | 借方欄位索引（選填） |
| `credit_column` | number? | 貸方欄位索引（選填） |
| `amount_column` | number? | 金額欄位索引（選填，debit/credit 的替代方案） |
| `balance_column` | number? | 餘額欄位索引（選填） |
| `reference_column` | number? | 參考編號欄位索引（選填） |
| `note` | string? | 備註（選填） |

---

## Phase 2 — 對帳單匯入與比對

**目標**：讓使用者上傳 CSV 對帳單、查看解析結果、執行自動/手動比對。

### 新增檔案

| 檔案 | 說明 |
|---|---|
| `src/views/BankStatement.svelte` | 對帳單主頁（匯入清單） |
| `src/views/BankStatementImport.svelte` | 匯入表單（含上傳 CSV） |
| `src/views/BankStatementResult.svelte` | 匯入結果（交易明細 + 比對操作） |

### 修改檔案

| 檔案 | 修改內容 |
|---|---|
| `src/routes/Home.svelte` | 加入 `/home/bank-statement` 與 `/home/bank-statement/:id/result` 路由 |

### 涉及 API

- `POST /bank-statement/import` — 上傳 CSV（multipart: `ledger_id`, `statement_date`, `template_id`, `file`, `note`）
- `GET /bank-statement/paged` — 匯入清單（query: `ledger_id`, `page`, `size`）
- `GET /bank-statement/{import_id}/result` — 取得解析後交易明細
- `POST /bank-statement/{import_id}/auto-match` — 自動比對
- `POST /bank-statement/{import_id}/re-match` — 重新比對
- `PUT /bank-statement/{import_id}/txn/{bank_txn_id}/match` — 手動比對（body: `entry_id`）

---

## Phase 3 — 審查與完成匯入

**目標**：讓使用者逐筆審查比對結果，核准或忽略，最終完成匯入。

### 新增檔案

| 檔案 | 說明 |
|---|---|
| `src/views/BankStatementReview.svelte` | 審查頁（核准 / 忽略 / 完成） |

### 涉及 API

- `GET /bank-statement/{import_id}/review` — 取得審查項目
- `POST /bank-statement/{import_id}/txn/{bank_txn_id}/approve` — 核准（body: `account_id`, `counter_account_id`, `ledger_id` 等）
- `POST /bank-statement/{import_id}/txn/{bank_txn_id}/ignore` — 忽略
- `POST /bank-statement/{import_id}/complete` — 完成匯入（body: `import_uuid`, `expected_version`）
