-- name: ListActiveQuestions :many
SELECT * FROM compatibility_questions WHERE active = TRUE ORDER BY sort_order, code;

-- name: ListAllQuestions :many
SELECT * FROM compatibility_questions ORDER BY sort_order, code;

-- name: GetQuestionByID :one
SELECT * FROM compatibility_questions WHERE id = $1;

-- name: CreateQuestion :one
INSERT INTO compatibility_questions (code, category, title, description, question_type, options, active, sort_order, version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateQuestion :one
UPDATE compatibility_questions SET
    category = $2,
    title = $3,
    description = $4,
    question_type = $5,
    options = $6,
    active = $7,
    sort_order = $8,
    version = version + 1
WHERE id = $1
RETURNING *;

-- name: ListRulesByUser :many
SELECT * FROM user_compatibility_rules WHERE user_id = $1;

-- name: ListRulesByUsers :many
SELECT * FROM user_compatibility_rules WHERE user_id = ANY($1::uuid[]);

-- name: DeleteRulesByUser :exec
DELETE FROM user_compatibility_rules WHERE user_id = $1;

-- name: UpsertRule :one
INSERT INTO user_compatibility_rules (user_id, question_id, own_answer, accepted_answers, importance, is_hard)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, question_id) DO UPDATE SET
    own_answer = EXCLUDED.own_answer,
    accepted_answers = EXCLUDED.accepted_answers,
    importance = EXCLUDED.importance,
    is_hard = EXCLUDED.is_hard
RETURNING *;
