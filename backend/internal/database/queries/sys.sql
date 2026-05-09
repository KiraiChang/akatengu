-- name: GetSysAccounts :many
SELECT sys_code, description, account_id
FROM sys_accounts;

-- name: UpdateSysAccount :exec
UPDATE sys_accounts
SET account_id = @account_id,
    description = @description
WHERE sys_code = @sys_code;