-- name: InsertFixedAssetCategory :exec
INSERT INTO fixed_asset_categories (
    merchant_id, category_uuid, name,
    asset_account_id, accum_depreciation_account_id, depreciation_expense_account_id,
    updated_by
) VALUES (
    @merchant_id, @category_uuid, @name,
    @asset_account_id, @accum_depreciation_account_id, @depreciation_expense_account_id,
    @updated_by
);

-- name: UpdateFixedAssetCategory :exec
UPDATE fixed_asset_categories
SET name                            = @name,
    asset_account_id                = @asset_account_id,
    accum_depreciation_account_id   = @accum_depreciation_account_id,
    depreciation_expense_account_id = @depreciation_expense_account_id,
    updated_by                      = @updated_by,
    updated_at                      = datetime('now')
WHERE category_uuid = @category_uuid AND merchant_id = @merchant_id;

-- name: SoftDeleteFixedAssetCategory :exec
UPDATE fixed_asset_categories
SET is_active  = 0,
    updated_by = @updated_by,
    updated_at = datetime('now')
WHERE category_uuid = @category_uuid AND merchant_id = @merchant_id;

-- name: GetFixedAssetCategoryByUUID :one
SELECT * FROM fixed_asset_categories
WHERE category_uuid = @category_uuid AND merchant_id = @merchant_id;

-- name: GetAllFixedAssetCategoriesByMerchant :many
SELECT * FROM fixed_asset_categories
WHERE merchant_id = @merchant_id AND is_active = 1
ORDER BY name;
