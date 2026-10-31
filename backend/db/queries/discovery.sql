-- name: ListPreferencesByUsers :many
SELECT * FROM partner_preferences WHERE user_id = ANY($1::uuid[]);

-- name: ListInteractionsByActorTargets :many
SELECT * FROM interactions WHERE actor_user_id = $1 AND target_user_id = ANY($2::uuid[]);
