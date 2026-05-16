-- name: GetSysAccounts :many
SELECT sys_code, description, account_id
FROM sys_accounts
WHERE merchant_id = @merchant_id;

-- name: UpdateSysAccount :exec
UPDATE sys_accounts
SET account_id  = @account_id,
    description = @description
WHERE sys_code = @sys_code AND merchant_id = @merchant_id;

-- name: BulkInsertSysAccounts :exec
INSERT INTO sys_accounts (sys_code, merchant_id, description, account_id)
VALUES (?, ?, ?, ?)
ON CONFLICT(sys_code, merchant_id) DO NOTHING;
