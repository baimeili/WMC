-- name: UpsertInteraction :one
INSERT INTO interactions (actor_user_id, target_user_id, action, source)
VALUES ($1, $2, $3, $4)
ON CONFLICT (actor_user_id, target_user_id) DO UPDATE SET
    action = EXCLUDED.action,
    source = EXCLUDED.source
RETURNING *;

-- name: GetInteraction :one
SELECT * FROM interactions WHERE actor_user_id = $1 AND target_user_id = $2;

-- name: ListLikesBetween :one
SELECT COUNT(*) FROM interactions
WHERE actor_user_id = $1 AND target_user_id = $2 AND action = 1;

-- name: CreateMatch :one
INSERT INTO matches (user_a_id, user_b_id, status)
VALUES (LEAST($1::uuid, $2::uuid), GREATEST($1::uuid, $2::uuid), 1)
ON CONFLICT (user_a_id, user_b_id) DO NOTHING
RETURNING *;

-- name: GetMatchPair :one
SELECT * FROM matches WHERE user_a_id = LEAST($1::uuid, $2::uuid) AND user_b_id = GREATEST($1::uuid, $2::uuid);

-- name: GetMatchByID :one
SELECT * FROM matches WHERE id = $1;

-- name: CancelMatch :exec
UPDATE matches SET status = 2, ended_at = now() WHERE id = $1;

-- name: SetConfirmedMeeting :exec
UPDATE matches SET confirmed_meeting = TRUE WHERE id = $1;

-- name: ListMatchesByUser :many
SELECT m.id, m.user_a_id, m.user_b_id, m.status, m.matched_at, m.confirmed_meeting,
       u.id AS peer_id, COALESCE(p.nickname, '') AS peer_nickname,
       COALESCE(p.city_code, '') AS peer_city_code,
       p.birth_date AS peer_birth_date,
       COALESCE((SELECT ph.object_key FROM user_photos ph
         WHERE ph.user_id = u.id AND ph.is_primary AND ph.deleted_at IS NULL AND ph.moderation_status = 2
         ORDER BY ph.sort_order LIMIT 1), '')::text AS peer_photo_key
FROM matches m
JOIN users u ON u.id = CASE WHEN m.user_a_id = $1 THEN m.user_b_id ELSE m.user_a_id END
LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE (m.user_a_id = $1 OR m.user_b_id = $1) AND m.status = 1
  AND NOT EXISTS (
      SELECT 1 FROM user_blocks b
      WHERE (b.blocker_user_id = m.user_a_id AND b.blocked_user_id = m.user_b_id)
         OR (b.blocker_user_id = m.user_b_id AND b.blocked_user_id = m.user_a_id)
  )
ORDER BY m.matched_at DESC
LIMIT $2 OFFSET $3;

-- name: UpsertOutcome :one
INSERT INTO match_outcomes (match_id, reporter_user_id, outcome_type, occurred_at, note)
VALUES ($1, $2, $3, COALESCE($4::timestamptz, now()), $5)
ON CONFLICT (match_id, reporter_user_id, outcome_type) DO UPDATE SET
    note = EXCLUDED.note,
    occurred_at = EXCLUDED.occurred_at
RETURNING *;

-- name: ListOutcomesByMatch :many
SELECT * FROM match_outcomes WHERE match_id = $1 ORDER BY created_at;

-- name: CountOutcomeReporters :one
SELECT COUNT(DISTINCT reporter_user_id) FROM match_outcomes WHERE match_id = $1 AND outcome_type = $2;

-- name: UpsertBlock :exec
INSERT INTO user_blocks (blocker_user_id, blocked_user_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteBlock :exec
DELETE FROM user_blocks WHERE blocker_user_id = $1 AND blocked_user_id = $2;

-- name: IsBlocked :one
SELECT EXISTS (
    SELECT 1 FROM user_blocks
    WHERE (blocker_user_id = $1 AND blocked_user_id = $2) OR (blocker_user_id = $2 AND blocked_user_id = $1)
);

-- name: CreateReport :one
INSERT INTO user_reports (reporter_user_id, reported_user_id, reason, description, evidence_keys)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: AdminListReports :many
SELECT r.*, COALESCE(p.nickname, '') AS reported_nickname
FROM user_reports r
LEFT JOIN user_profiles p ON p.user_id = r.reported_user_id
WHERE ($1::int = 0 OR r.status = $1::int)
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: AdminResolveReport :exec
UPDATE user_reports
SET status = $2::smallint,
    reviewer_id = $3,
    reviewed_at = now(),
    resolution_note = $4
WHERE id = $1;

-- name: AdminListPendingPhotos :many
SELECT p.id, p.user_id, p.object_key, p.created_at, COALESCE(up.nickname, '') AS nickname
FROM user_photos p
LEFT JOIN user_profiles up ON up.user_id = p.user_id
WHERE p.moderation_status = 1 AND p.deleted_at IS NULL
ORDER BY p.created_at
LIMIT $1 OFFSET $2;

-- name: AdminListUserReports :many
SELECT * FROM user_reports WHERE reported_user_id = $1 ORDER BY created_at DESC LIMIT 20;
-- name: BlockMatchesBetween :exec
UPDATE matches SET status = 3, ended_at = now()
WHERE (user_a_id = LEAST($1::uuid, $2::uuid) AND user_b_id = GREATEST($1::uuid, $2::uuid)) AND status = 1;
