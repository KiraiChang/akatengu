-- name: GetUserByName :one
SELECT user_id, username, password, status
FROM users
WHERE username = ? AND status = ?;

-- name: CreateUser :execlastid
INSERT INTO users (username, password)
VALUES (?, ?);