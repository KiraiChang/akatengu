-- name: InsertFixedAsset :one
INSERT INTO fixed_assets (
    merchant_id, asset_uuid, name,
    asset_account_id, accum_depreciation_account_id, depreciation_expense_account_id,
    cost, residual_value, useful_life_months,
    depreciation_method, payment_type, purchase_date, updated_by
) VALUES (
    @merchant_id, @asset_uuid, @name,
    @asset_account_id, @accum_depreciation_account_id, @depreciation_expense_account_id,
    @cost, @residual_value, @useful_life_months,
    @depreciation_method, @payment_type, @purchase_date, @updated_by
) RETURNING *;

-- name: UpdateFixedAssetTxn :exec
UPDATE fixed_assets SET txn_id = @txn_id WHERE id = @id AND merchant_id = @merchant_id;

-- name: GetFixedAssetByID :one
SELECT * FROM fixed_assets WHERE id = @id AND merchant_id = @merchant_id;

-- name: GetFixedAssetByUUID :one
SELECT * FROM fixed_assets WHERE asset_uuid = @asset_uuid AND merchant_id = @merchant_id;

-- name: GetActiveFixedAssetsByMerchant :many
SELECT * FROM fixed_assets WHERE merchant_id = @merchant_id AND status = 'ACTIVE' ORDER BY id ASC;

-- name: GetAllFixedAssetsByMerchant :many
SELECT * FROM fixed_assets WHERE merchant_id = @merchant_id ORDER BY id ASC;

-- name: UpdateFixedAssetDepreciation :exec
UPDATE fixed_assets SET
    total_depreciated   = total_depreciated + @delta_amount,
    depreciated_periods = depreciated_periods + 1,
    updated_by          = @updated_by,
    updated_at          = datetime('now'),
    version             = version + 1
WHERE id = @id AND merchant_id = @merchant_id;

-- name: UpdateFixedAssetDisposed :exec
UPDATE fixed_assets SET
    status       = 'DISPOSED',
    disposal_date = @disposal_date,
    updated_by   = @updated_by,
    updated_at   = datetime('now'),
    version      = version + 1
WHERE id = @id AND merchant_id = @merchant_id;

-- name: InsertFixedAssetDepreciation :exec
INSERT INTO fixed_asset_depreciations (
    merchant_id, depreciation_uuid, asset_uuid, asset_id, txn_id, period_date, amount, updated_by
) VALUES (
    @merchant_id, @depreciation_uuid, @asset_uuid, @asset_id, @txn_id, @period_date, @amount, @updated_by
);

-- name: GetFixedAssetDepreciationsByAssetID :many
SELECT * FROM fixed_asset_depreciations
WHERE asset_id = @asset_id AND merchant_id = @merchant_id
ORDER BY period_date ASC;
