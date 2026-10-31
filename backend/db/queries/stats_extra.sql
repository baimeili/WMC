-- name: StatsCountEvents :one
SELECT COUNT(*) FROM product_events WHERE event_name = $1;

-- name: StatsCountEventUsers :one
SELECT COUNT(DISTINCT COALESCE(user_id::text, anonymous_id)) FROM product_events WHERE event_name = $1;

-- name: ListSettingsRows :many
SELECT * FROM system_settings ORDER BY key;
