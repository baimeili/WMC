-- name: AdminListUsers :many
SELECT u.id, u.status, u.role, u.discovery_enabled, u.is_seed, u.created_at, u.last_login_at,
       COALESCE(p.nickname, '') AS nickname,
       COALESCE(p.profile_completion, 0) AS profile_completion,
       COALESCE(p.gender, 0) AS gender,
       COALESCE(p.city_code, '') AS city_code
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE ($1::text = '' OR COALESCE(p.nickname, '') ILIKE '%' || $1 || '%')
  AND ($2::int = 0 OR u.status = $2)
ORDER BY u.created_at DESC
LIMIT $3 OFFSET $4;

-- name: AdminCountUsers :one
SELECT COUNT(*)
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE ($1::text = '' OR COALESCE(p.nickname, '') ILIKE '%' || $1 || '%')
  AND ($2::int = 0 OR u.status = $2);

-- name: AdminGetUserDetail :one
SELECT u.id, u.status, u.role, u.discovery_enabled, u.is_seed, u.created_at, u.last_login_at,
       u.phone_encrypted,
       COALESCE(p.nickname, '') AS nickname,
       COALESCE(p.profile_completion, 0) AS profile_completion,
       COALESCE(p.gender, 0) AS gender,
       COALESCE(p.city_code, '') AS city_code,
       COALESCE(p.birth_date, '1900-01-01'::date) AS birth_date,
       COALESCE(p.education_level, 0) AS education_level,
       COALESCE(p.marital_status, 0) AS marital_status
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE u.id = $1;

-- name: AdminSetUserStatus :exec
UPDATE users
SET status = $2::smallint,
    discovery_enabled = CASE WHEN $2::smallint = 3 THEN FALSE ELSE discovery_enabled END
WHERE id = $1;

-- name: AdminCountByStatus :one
SELECT COUNT(*) FROM users WHERE status = $1;

-- name: AdminCountCompletedProfiles :one
SELECT COUNT(*) FROM user_profiles WHERE profile_completion >= $1;

-- name: AdminCountMarketUsers :one
SELECT COUNT(*) FROM users u
JOIN user_profiles p ON p.user_id = u.id
WHERE u.status = 2 AND u.discovery_enabled AND p.profile_completion >= $1;

-- name: AdminGenderDistribution :many
SELECT COALESCE(p.gender, 0) AS gender, COUNT(*) AS count
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE u.status <> 4
GROUP BY 1 ORDER BY 1;

-- name: AdminCityDistribution :many
SELECT COALESCE(p.city_code, '') AS city_code, COUNT(*) AS count
FROM users u
JOIN user_profiles p ON p.user_id = u.id
WHERE u.status <> 4 AND p.city_code <> ''
GROUP BY 1 ORDER BY count DESC LIMIT 20;

-- name: AdminAgeDistribution :many
SELECT width_bucket(EXTRACT(YEAR FROM age(p.birth_date)), 18, 60, 13) AS bucket, COUNT(*) AS count
FROM user_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.birth_date IS NOT NULL AND u.status <> 4
GROUP BY 1 ORDER BY 1;

-- name: AdminCountNewUsers :one
SELECT COUNT(*) FROM users WHERE created_at >= $1;
