-- name: CreateRefreshToken :one
INSERT INTO auth_refresh_tokens (user_id, token_hash, family_id, user_agent, ip_hash, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM auth_refresh_tokens WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE auth_refresh_tokens SET revoked_at = now() WHERE id = $1;

-- name: RevokeRefreshFamily :exec
UPDATE auth_refresh_tokens SET revoked_at = now() WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeAllUserTokens :exec
UPDATE auth_refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL;

-- name: MarkRefreshTokenReplaced :exec
UPDATE auth_refresh_tokens SET revoked_at = now(), replaced_by = $2 WHERE id = $1;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE FROM auth_refresh_tokens WHERE expires_at < now() - interval '7 days';

-- name: UpsertConsent :one
INSERT INTO consent_records (user_id, consent_type, policy_version)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, consent_type) DO UPDATE SET
    policy_version = EXCLUDED.policy_version,
    consented_at = now(),
    revoked_at = NULL
RETURNING *;

-- name: ListConsents :many
SELECT * FROM consent_records WHERE user_id = $1;

-- name: UpsertVerification :one
INSERT INTO verification_records (user_id, type, status, evidence, verified_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, type) DO UPDATE SET
    status = EXCLUDED.status,
    evidence = EXCLUDED.evidence,
    verified_at = EXCLUDED.verified_at
RETURNING *;

-- name: ListVerifications :many
SELECT * FROM verification_records WHERE user_id = $1;
