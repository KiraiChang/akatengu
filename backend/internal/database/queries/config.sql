-- name: GetLedgerAccountTypeConfigs :many
SELECT * FROM ledger_account_type_config WHERE merchant_id = @merchant_id;

-- name: GetLedgerAccountTypeConfig :one
SELECT * FROM ledger_account_type_config
WHERE merchant_id = @merchant_id AND type = @type;

-- name: UpsertLedgerAccountTypeConfig :exec
INSERT INTO ledger_account_type_config(merchant_id, type, account_id, updated_by)
VALUES(@merchant_id, @type, @account_id, @updated_by)
ON CONFLICT(merchant_id, type) DO UPDATE SET
    account_id = excluded.account_id,
    updated_by = excluded.updated_by,
    updated_at = datetime('now'),
    version    = version + 1;

-- name: GetAssetTypeAccountConfigs :many
SELECT * FROM asset_type_account_config WHERE merchant_id = @merchant_id;

-- name: GetAssetTypeAccountConfig :one
SELECT * FROM asset_type_account_config
WHERE merchant_id = @merchant_id AND asset_type = @asset_type;

-- name: UpsertAssetTypeAccountConfig :exec
INSERT INTO asset_type_account_config(
    merchant_id, asset_type,
    realized_gain_account_id, realized_loss_account_id,
    unrealized_gain_account_id, unrealized_loss_account_id,
    oci_account_id, fee_account_id, tax_account_id,
    account_id, updated_by
) VALUES(
    @merchant_id, @asset_type,
    @realized_gain_account_id, @realized_loss_account_id,
    @unrealized_gain_account_id, @unrealized_loss_account_id,
    @oci_account_id, @fee_account_id, @tax_account_id,
    @account_id, @updated_by
)
ON CONFLICT(merchant_id, asset_type) DO UPDATE SET
    realized_gain_account_id   = excluded.realized_gain_account_id,
    realized_loss_account_id   = excluded.realized_loss_account_id,
    unrealized_gain_account_id = excluded.unrealized_gain_account_id,
    unrealized_loss_account_id = excluded.unrealized_loss_account_id,
    oci_account_id             = excluded.oci_account_id,
    fee_account_id             = excluded.fee_account_id,
    tax_account_id             = excluded.tax_account_id,
    account_id                 = excluded.account_id,
    updated_by                 = excluded.updated_by,
    updated_at                 = datetime('now'),
    version                    = version + 1;

-- name: GetAccountDescendants :many
SELECT a.account_id, a.parent_id, a.name, a.type, a.normal_balance,
       a.currency, a.is_summary, a.is_active, a.note, a.version,
       a.cash_flow_category, a.updated_by, a.updated_at,
       EXISTS(
           SELECT 1 FROM accounts c
           WHERE c.parent_id = a.account_id AND c.merchant_id = a.merchant_id
       ) AS has_child
FROM accounts a
JOIN account_closure ac
     ON a.account_id = ac.descendant_id AND a.merchant_id = ac.merchant_id
WHERE ac.ancestor_id = @ancestor_id AND ac.merchant_id = @merchant_id AND ac.depth > 0
ORDER BY a.account_id ASC;
