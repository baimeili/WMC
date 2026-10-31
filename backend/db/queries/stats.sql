-- name: StatsCountInteractions :one
SELECT COUNT(*) FROM interactions WHERE action = $1;

-- name: StatsCountMatches :one
SELECT COUNT(*) FROM matches;

-- name: StatsCountOutcomes :one
SELECT COUNT(*) FROM match_outcomes WHERE outcome_type = $1;

-- name: StatsConfirmedMeetings :one
SELECT COUNT(*) FROM matches WHERE confirmed_meeting;

-- name: StatsActiveUsers :one
SELECT COUNT(DISTINCT user_id) FROM product_events
WHERE user_id IS NOT NULL AND occurred_at >= $1;

-- name: StatsCountCalculatorRuns :one
SELECT COUNT(*) FROM calculator_runs WHERE created_at >= $1;

-- name: StatsCountRegisteredSince :one
SELECT COUNT(*) FROM users WHERE created_at >= $1;

-- name: StatsCountMarketUsers :one
SELECT COUNT(*) FROM users u JOIN user_profiles p ON p.user_id = u.id
WHERE u.status = 2 AND u.discovery_enabled AND p.profile_completion >= $1;

-- name: StatsCountCompletedProfiles :one
SELECT COUNT(*) FROM user_profiles WHERE profile_completion >= $1;
