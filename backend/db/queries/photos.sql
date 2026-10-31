-- name: CreatePhoto :one
INSERT INTO user_photos (user_id, object_key, sort_order, is_primary, moderation_status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPhotoByID :one
SELECT * FROM user_photos WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPhotosByUser :many
SELECT * FROM user_photos WHERE user_id = $1 AND deleted_at IS NULL ORDER BY sort_order, created_at;

-- name: CountPhotosByUser :one
SELECT COUNT(*) FROM user_photos WHERE user_id = $1 AND deleted_at IS NULL;

-- name: ClearPrimaryPhotos :exec
UPDATE user_photos SET is_primary = FALSE WHERE user_id = $1 AND is_primary AND deleted_at IS NULL;

-- name: SetPrimaryPhoto :exec
UPDATE user_photos SET is_primary = TRUE WHERE id = $1;

-- name: SoftDeletePhoto :exec
UPDATE user_photos SET deleted_at = now(), is_primary = FALSE WHERE id = $1;

-- name: ListPendingPhotos :many
SELECT p.*, up.nickname
FROM user_photos p
JOIN user_profiles up ON up.user_id = p.user_id
WHERE p.moderation_status = 1 AND p.deleted_at IS NULL
ORDER BY p.created_at
LIMIT $1 OFFSET $2;

-- name: ModeratePhoto :exec
UPDATE user_photos SET moderation_status = $2 WHERE id = $1;
