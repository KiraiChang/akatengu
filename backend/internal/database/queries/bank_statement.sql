-- ============================================================
-- bank_csv_templates
-- ============================================================

-- name: GetBankCsvTemplates :many
SELECT * FROM bank_csv_templates
WHERE merchant_id = @merchant_id AND is_active = 1
ORDER BY template_id DESC;

-- name: GetBankCsvTemplateByID :one
SELECT * FROM bank_csv_templates
WHERE template_id = @template_id AND merchant_id = @merchant_id;

-- name: GetBankCsvTemplateByUUID :one
SELECT * FROM bank_csv_templates
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- name: InsertBankCsvTemplate :one
INSERT INTO bank_csv_templates (
    template_uuid, merchant_id, template_name, encoding, skip_rows,
    date_column, date_format, description_column,
    debit_column, credit_column, amount_column,
    balance_column, reference_column, note, updated_by
) VALUES (
    @template_uuid, @merchant_id, @template_name, @encoding, @skip_rows,
    @date_column, @date_format, @description_column,
    @debit_column, @credit_column, @amount_column,
    @balance_column, @reference_column, @note, @updated_by
) RETURNING *;

-- name: UpdateBankCsvTemplate :exec
UPDATE bank_csv_templates SET
    template_name      = @template_name,
    encoding           = @encoding,
    skip_rows          = @skip_rows,
    date_column        = @date_column,
    date_format        = @date_format,
    description_column = @description_column,
    debit_column       = @debit_column,
    credit_column      = @credit_column,
    amount_column      = @amount_column,
    balance_column     = @balance_column,
    reference_column   = @reference_column,
    note               = @note,
    version            = version + 1,
    updated_by         = @updated_by,
    updated_at         = datetime('now')
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- name: DeactivateBankCsvTemplate :exec
UPDATE bank_csv_templates SET
    is_active  = 0,
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = datetime('now')
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- name: GetBankCsvTemplateIDByUUID :one
SELECT template_id FROM bank_csv_templates
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- ============================================================
-- bank_statement_imports
-- ============================================================

-- name: GetBankStatementImportByID :one
SELECT * FROM bank_statement_imports
WHERE import_id = @import_id AND merchant_id = @merchant_id;

-- name: GetBankStatementImportByUUID :one
SELECT * FROM bank_statement_imports
WHERE import_uuid = @import_uuid AND merchant_id = @merchant_id;

-- name: GetBankStatementImportsPaged :many
SELECT * FROM bank_statement_imports
WHERE merchant_id = @merchant_id
  AND (@ledger_id = 0 OR ledger_id = @ledger_id)
ORDER BY import_id DESC
LIMIT @page_size OFFSET @offset;

-- name: CountBankStatementImports :one
SELECT COUNT(*) FROM bank_statement_imports
WHERE merchant_id = @merchant_id
  AND (@ledger_id = 0 OR ledger_id = @ledger_id);

-- name: InsertBankStatementImport :one
INSERT INTO bank_statement_imports (
    import_uuid, merchant_id, ledger_id, template_id, template_uuid,
    pdf_template_id, pdf_template_uuid, bank_type,
    statement_date, import_source, filename, note, updated_by
) VALUES (
    @import_uuid, @merchant_id, @ledger_id, @template_id, @template_uuid,
    @pdf_template_id, @pdf_template_uuid, @bank_type,
    @statement_date, @import_source, @filename, @note, @updated_by
) RETURNING *;

-- name: UpdateBankStatementImportStatus :exec
UPDATE bank_statement_imports SET
    status     = @status,
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = datetime('now')
WHERE import_id = @import_id AND merchant_id = @merchant_id;

-- ============================================================
-- bank_statement_txns
-- ============================================================

-- name: InsertBankStatementTxn :one
INSERT INTO bank_statement_txns (
    bank_txn_uuid, import_id, merchant_id, txn_date,
    description, debit, credit, balance, reference_no, ledger_uuid, ledger_id
) VALUES (
    @bank_txn_uuid, @import_id, @merchant_id, @txn_date,
    @description, @debit, @credit, @balance, @reference_no, @ledger_uuid, @ledger_id
) RETURNING *;

-- name: GetBankStatementTxnsByImport :many
SELECT * FROM bank_statement_txns
WHERE import_id = @import_id AND merchant_id = @merchant_id
ORDER BY txn_date ASC, bank_txn_id ASC;

-- name: GetBankStatementTxnByUUID :one
SELECT * FROM bank_statement_txns
WHERE bank_txn_uuid = @bank_txn_uuid AND merchant_id = @merchant_id;

-- name: GetBankStatementTxnByID :one
SELECT * FROM bank_statement_txns
WHERE bank_txn_id = @bank_txn_id AND merchant_id = @merchant_id;

-- name: GetUnmatchedBankTxns :many
SELECT * FROM bank_statement_txns
WHERE import_id = @import_id
  AND merchant_id = @merchant_id
  AND match_status = 'UNMATCHED'
ORDER BY txn_date ASC, bank_txn_id ASC;

-- name: GetBankTxnsForReview :many
SELECT * FROM bank_statement_txns
WHERE import_id = @import_id
  AND merchant_id = @merchant_id
  AND match_status NOT IN ('APPROVED', 'IGNORED')
ORDER BY match_status ASC, txn_date ASC, bank_txn_id ASC;

-- name: UpdateBankTxnMatch :exec
UPDATE bank_statement_txns SET
    match_status     = @match_status,
    matched_entry_id = @matched_entry_id,
    match_confidence = @match_confidence
WHERE bank_txn_id = @bank_txn_id AND merchant_id = @merchant_id;

-- name: UpdateBankTxnStatus :exec
UPDATE bank_statement_txns SET
    match_status = @match_status
WHERE bank_txn_id = @bank_txn_id AND merchant_id = @merchant_id;

-- name: UpdateBankTxnCreatedTxn :exec
UPDATE bank_statement_txns SET
    match_status   = 'APPROVED',
    created_txn_id = @created_txn_id
WHERE bank_txn_id = @bank_txn_id AND merchant_id = @merchant_id;

-- name: ResetNonConfirmedMatches :exec
UPDATE bank_statement_txns SET
    match_status     = 'UNMATCHED',
    matched_entry_id = NULL,
    match_confidence = NULL
WHERE import_id = @import_id
  AND merchant_id = @merchant_id
  AND match_status = 'MATCHED'
  AND match_confidence IN ('EXACT', 'FUZZY');

-- name: CountBankTxnsByStatus :one
SELECT COUNT(*) FROM bank_statement_txns
WHERE import_id = @import_id
  AND merchant_id = @merchant_id
  AND match_status = @match_status;

-- ============================================================
-- bank_statement_import_ledgers
-- ============================================================

-- name: InsertBankStatementImportLedger :exec
INSERT INTO bank_statement_import_ledgers (
    import_ledger_uuid, import_id, ledger_uuid, ledger_id, account_type
) VALUES (
    @import_ledger_uuid, @import_id, @ledger_uuid, @ledger_id, @account_type
);

-- name: GetBankStatementImportLedgers :many
SELECT * FROM bank_statement_import_ledgers
WHERE import_id = @import_id
ORDER BY import_ledger_id ASC;

-- ============================================================
-- bank_csv_template_ledgers
-- ============================================================

-- name: InsertBankCsvTemplateLedger :exec
INSERT INTO bank_csv_template_ledgers (
    tpl_ledger_uuid, template_id, ledger_uuid, ledger_id, account_type, sort_order
) VALUES (
    @tpl_ledger_uuid, @template_id, @ledger_uuid, @ledger_id, @account_type, @sort_order
);

-- name: GetBankCsvTemplateLedgers :many
SELECT * FROM bank_csv_template_ledgers
WHERE template_id = @template_id
ORDER BY sort_order ASC, tpl_ledger_id ASC;

-- name: DeleteBankCsvTemplateLedgers :exec
DELETE FROM bank_csv_template_ledgers
WHERE template_id = @template_id;

-- ============================================================
-- journal_entries for matching
-- ============================================================

-- name: GetLedgerEntriesForMatching :many
SELECT
    je.entry_id,
    je.entry_uuid,
    je.txn_id,
    je.ledger_id,
    je.account_id,
    je.debit,
    je.credit,
    t.txn_date,
    t.description
FROM journal_entries je
JOIN transactions t ON je.txn_id = t.txn_id AND t.merchant_id = je.merchant_id
WHERE je.ledger_id = sqlc.arg(ledger_id)
  AND je.merchant_id = sqlc.arg(merchant_id)
  AND t.status = 'ACTIVE'
  AND t.ref_txn_id IS NULL
  AND t.txn_date >= sqlc.arg(date_from)
  AND t.txn_date <= sqlc.arg(date_to)
ORDER BY t.txn_date ASC, je.entry_id ASC;
