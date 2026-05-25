-- name: GetTemplates :many
SELECT * FROM transaction_templates
WHERE merchant_id = @merchant_id
  AND (@search = '' OR name LIKE '%' || @search || '%' OR description LIKE '%' || @search || '%')
ORDER BY id DESC;

-- name: GetTemplateByID :one
SELECT * FROM transaction_templates
WHERE id = @id AND merchant_id = @merchant_id;

-- name: InsertTemplate :one
INSERT INTO transaction_templates(merchant_id, name, description, tag, updated_by)
VALUES(@merchant_id, @name, @description, @tag, @updated_by)
RETURNING *;

-- name: UpdateTemplate :exec
UPDATE transaction_templates SET
    name        = @name,
    description = @description,
    tag         = @tag,
    updated_by  = @updated_by,
    updated_at  = datetime('now'),
    version     = version + 1
WHERE id = @id AND merchant_id = @merchant_id;

-- name: DeleteTemplate :exec
DELETE FROM transaction_templates WHERE id = @id AND merchant_id = @merchant_id;

-- name: GetTemplateEntriesByTemplateID :many
SELECT * FROM transaction_template_entries
WHERE template_id = @template_id AND merchant_id = @merchant_id
ORDER BY sort_order ASC;

-- name: InsertTemplateEntry :exec
INSERT INTO transaction_template_entries(
    merchant_id, template_id, sort_order, account_id, ledger_id,
    debit, credit, note, cash_flow_category
) VALUES(
    @merchant_id, @template_id, @sort_order, @account_id, @ledger_id,
    @debit, @credit, @note, @cash_flow_category
);

-- name: DeleteTemplateEntriesByTemplateID :exec
DELETE FROM transaction_template_entries WHERE template_id = @template_id AND merchant_id = @merchant_id;
