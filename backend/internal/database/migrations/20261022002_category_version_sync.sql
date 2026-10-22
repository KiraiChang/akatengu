-- +goose Up
-- Category updates and deletes never bumped the read-model version, so the
-- screen kept sending 1 while the category_uuid aggregate moved to 2, and the
-- next update or delete always failed (backend ISSUE-029). Align categories
-- that have a version row for their merchant; categories never updated have
-- none and keep version 1 (BFS R2 segment 4 adds their rows). Only version is
-- changed: this is a repair, not an edit.
UPDATE prepaid_categories
SET version = (
    SELECT av.current_version FROM aggregate_versions av
    WHERE av.aggregate_type = 'PREPAID_CATEGORY'
      AND av.aggregate_id = prepaid_categories.category_uuid
      AND av.merchant_id = prepaid_categories.merchant_id)
WHERE EXISTS (
    SELECT 1 FROM aggregate_versions av
    WHERE av.aggregate_type = 'PREPAID_CATEGORY'
      AND av.aggregate_id = prepaid_categories.category_uuid
      AND av.merchant_id = prepaid_categories.merchant_id);

UPDATE fixed_asset_categories
SET version = (
    SELECT av.current_version FROM aggregate_versions av
    WHERE av.aggregate_type = 'FIXED_ASSET_CATEGORY'
      AND av.aggregate_id = fixed_asset_categories.category_uuid
      AND av.merchant_id = fixed_asset_categories.merchant_id)
WHERE EXISTS (
    SELECT 1 FROM aggregate_versions av
    WHERE av.aggregate_type = 'FIXED_ASSET_CATEGORY'
      AND av.aggregate_id = fixed_asset_categories.category_uuid
      AND av.merchant_id = fixed_asset_categories.merchant_id);

-- +goose Down
-- No-op: the previous read-model versions were wrong, and restoring them would
-- bring ISSUE-029 back.
