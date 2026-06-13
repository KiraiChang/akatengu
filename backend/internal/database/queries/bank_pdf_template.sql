-- ============================================================
-- bank_pdf_templates
-- ============================================================

-- name: GetBankPdfTemplates :many
SELECT * FROM bank_pdf_templates
WHERE merchant_id = @merchant_id AND is_active = 1
ORDER BY template_id DESC;

-- name: GetBankPdfTemplateByID :one
SELECT * FROM bank_pdf_templates
WHERE template_id = @template_id AND merchant_id = @merchant_id;

-- name: GetBankPdfTemplateByUUID :one
SELECT * FROM bank_pdf_templates
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- name: InsertBankPdfTemplate :one
INSERT INTO bank_pdf_templates (
    template_uuid, merchant_id, template_name, bank_type, updated_by
) VALUES (
    @template_uuid, @merchant_id, @template_name, @bank_type, @updated_by
) RETURNING *;

-- name: UpdateBankPdfTemplate :exec
UPDATE bank_pdf_templates SET
    template_name = @template_name,
    bank_type     = @bank_type,
    version       = version + 1,
    updated_by    = @updated_by,
    updated_at    = datetime('now')
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- name: DeactivateBankPdfTemplate :exec
UPDATE bank_pdf_templates SET
    is_active  = 0,
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = datetime('now')
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- name: GetBankPdfTemplateIDByUUID :one
SELECT template_id FROM bank_pdf_templates
WHERE template_uuid = @template_uuid AND merchant_id = @merchant_id;

-- ============================================================
-- bank_pdf_template_ledgers
-- ============================================================

-- name: InsertBankPdfTemplateLedger :exec
INSERT INTO bank_pdf_template_ledgers (
    tpl_ledger_uuid, template_id, ledger_uuid, ledger_id, account_type, sort_order
) VALUES (
    @tpl_ledger_uuid, @template_id, @ledger_uuid, @ledger_id, @account_type, @sort_order
);

-- name: GetBankPdfTemplateLedgers :many
SELECT * FROM bank_pdf_template_ledgers
WHERE template_id = @template_id
ORDER BY sort_order ASC, tpl_ledger_id ASC;

-- name: DeleteBankPdfTemplateLedgers :exec
DELETE FROM bank_pdf_template_ledgers
WHERE template_id = @template_id;
