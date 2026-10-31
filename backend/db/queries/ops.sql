-- name: ListSettings :many
SELECT * FROM system_settings ORDER BY key;

-- name: GetSetting :one
SELECT * FROM system_settings WHERE key = $1;

-- name: UpsertSetting :one
INSERT INTO system_settings (key, value, description, updated_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    description = COALESCE(NULLIF(EXCLUDED.description, ''), system_settings.description),
    updated_by = EXCLUDED.updated_by,
    updated_at = now()
RETURNING *;

-- name: InsertEvent :exec
INSERT INTO product_events (user_id, anonymous_id, event_name, properties)
VALUES ($1, $2, $3, $4);

-- name: CountEventsByName :one
SELECT COUNT(DISTINCT COALESCE(user_id::text, anonymous_id)) FROM product_events WHERE event_name = $1;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (actor_user_id, action, target_type, target_id, detail, ip_hash)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListAuditLogs :many
SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListRegions :many
SELECT * FROM regions ORDER BY level, sort_order, code;
