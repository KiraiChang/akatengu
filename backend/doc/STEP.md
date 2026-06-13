# PDF / CSV / XLSX 銀行對帳單多帳本匯入 — 實作進度

| Phase | 說明 | 狀態 |
|-------|------|------|
| 0 | 建立 STEP.md | ✅ done |
| 1 | 新增 BankType 列舉 + pdfparser interface/factory 骨架 + 新增 BANK_PDF_TEMPLATE aggregate | ✅ done |
| 2 | 資料庫 Migration（新增 4 張表 + 修改 bank_statement_imports + bank_statement_txns） | ✅ done |
| 3 | SQL queries 更新（bank_statement.sql、bank_pdf_template.sql）+ go generate | ✅ done |
| 4 | bank_pdf_template 全層（Payload / Projection Model / Event Types / Projection Service / Pipeline / Query Repo / Projection Repo / Service） | ✅ done |
| 5 | 修改 bank_csv_template 加入 ledger 清單（Payload / Projection Service / Projection Repo / Query Repo） | ✅ done |
| 6 | 修改 BankStatementImported 多帳本（Payload / Projection Service / Projection Repo / Pipeline） | ✅ done |
| 7 | Handler 新增 PDF template CRUD + ImportPDF；更新 CSV/XLSX import | ✅ done |
| 8 | go build + go test ./... 全套驗證 | ✅ done |
