-- name: ListRulesWithQuestions :many
SELECT r.id, r.user_id, r.question_id, r.own_answer, r.accepted_answers, r.importance, r.is_hard,
       q.code, q.category, q.title, q.options
FROM user_compatibility_rules r
JOIN compatibility_questions q ON q.id = r.question_id
WHERE r.user_id = $1
ORDER BY q.sort_order;

-- name: CountRulesByUser :one
SELECT COUNT(*) FROM user_compatibility_rules WHERE user_id = $1;

-- name: UpsertProfileSettings :exec
UPDATE user_profiles SET profile_completion = $2 WHERE user_id = $1;

-- name: GetProfileWithUser :one
SELECT p.*, u.status, u.discovery_enabled, u.role
FROM user_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = $1;
