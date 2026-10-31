-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByPhoneHash :one
SELECT * FROM users WHERE phone_hash = $1;

-- name: CreateUser :one
INSERT INTO users (phone_hash, phone_encrypted, status, role, is_seed)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = now() WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $2, deleted_at = $3 WHERE id = $1;

-- name: SetUserDiscoveryEnabled :exec
UPDATE users SET discovery_enabled = $2 WHERE id = $1;

-- name: RequestUserDeletion :exec
UPDATE users
SET status = 5, discovery_enabled = FALSE, deletion_requested_at = now()
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE status <> 4;

-- name: CountUsersByRole :one
SELECT COUNT(*) FROM users WHERE role = $1 AND status <> 4;
