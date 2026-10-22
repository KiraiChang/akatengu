-- +goose Up
-- BFS R2 segment 4 writes category updates and deletes to the category_uuid
-- aggregate at the version the screen sends, so every category's read-model
-- version must equal its aggregate version (BFS R2 D6).
--
-- 1. Between F8 (20261022002) and segment 4 the legacy path still created the
--    version row at "sent version + 1" on a category's first update, while the
--    read model went to 2: a first update sent with version 0 leaves read
--    model 2 and aggregate 1. Align the read model with the row again, as F8
--    does; the row is kept so the aggregate versions of the stored events never
--    go back. Only version is changed: this is a repair, not an edit.
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
      AND av.merchant_id = prepaid_categories.merchant_id
      AND av.current_version <> prepaid_categories.version);

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
      AND av.merchant_id = fixed_asset_categories.merchant_id
      AND av.current_version <> fixed_asset_categories.version);

-- 2. Categories never updated have no version row for their merchant, and BFS
--    rejects a version other than 0 without one; add the row at the read-model
--    version.
INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
SELECT 'PREPAID_CATEGORY', c.category_uuid, c.merchant_id, c.version
FROM prepaid_categories c
WHERE NOT EXISTS (
    SELECT 1 FROM aggregate_versions av
    WHERE av.aggregate_type = 'PREPAID_CATEGORY'
      AND av.aggregate_id = c.category_uuid
      AND av.merchant_id = c.merchant_id);

INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
SELECT 'FIXED_ASSET_CATEGORY', c.category_uuid, c.merchant_id, c.version
FROM fixed_asset_categories c
WHERE NOT EXISTS (
    SELECT 1 FROM aggregate_versions av
    WHERE av.aggregate_type = 'FIXED_ASSET_CATEGORY'
      AND av.aggregate_id = c.category_uuid
      AND av.merchant_id = c.merchant_id);

-- +goose Down
-- No-op: the previous read-model versions did not match their aggregates, and
-- the added rows cannot be told from rows BFS wrote afterwards; undoing either
-- would make the categories impossible to update again.
