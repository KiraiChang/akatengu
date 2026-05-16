-- name: CreateMerchant :execlastid
INSERT INTO merchants (name, display_name, currency)
VALUES (?, ?, ?);

-- name: GetMerchant :one
SELECT merchant_id, name, display_name, currency, status, created_at
FROM merchants
WHERE merchant_id = ?;

-- name: AddUserMerchant :exec
INSERT INTO user_merchants (user_id, merchant_id, role)
VALUES (?, ?, ?);

-- name: GetUserMerchants :many
SELECT m.merchant_id, m.name, m.display_name, m.currency, m.status, m.created_at,
       um.role, um.joined_at
FROM user_merchants um
JOIN merchants m ON um.merchant_id = m.merchant_id
WHERE um.user_id = ?
  AND m.status = 'ACTIVE';

-- name: GetUserMerchant :one
SELECT um.role
FROM user_merchants um
WHERE um.user_id = ? AND um.merchant_id = ?;
