-- name: InsertPrepaidCategory :exec
INSERT INTO prepaid_categories (
    merchant_id, category_uuid, name,
    account_id, expense_account_id,
    updated_by
) VALUES (
    @merchant_id, @category_uuid, @name,
    @account_id, @expense_account_id,
    @updated_by
);

-- name: UpdatePrepaidCategory :exec
UPDATE prepaid_categories
SET name               = @name,
    account_id         = @account_id,
    expense_account_id = @expense_account_id,
    updated_by         = @updated_by,
    updated_at         = datetime('now')
WHERE category_uuid = @category_uuid AND merchant_id = @merchant_id;

-- name: SoftDeletePrepaidCategory :exec
UPDATE prepaid_categories
SET is_active  = 0,
    updated_by = @updated_by,
    updated_at = datetime('now')
WHERE category_uuid = @category_uuid AND merchant_id = @merchant_id;

-- name: GetPrepaidCategoryByUUID :one
SELECT * FROM prepaid_categories
WHERE category_uuid = @category_uuid AND merchant_id = @merchant_id;

-- name: GetAllPrepaidCategoriesByMerchant :many
SELECT * FROM prepaid_categories
WHERE merchant_id = @merchant_id AND is_active = 1
ORDER BY name;
