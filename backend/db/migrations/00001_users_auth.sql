-- +goose Up
-- 用户、认证、验证、同意记录

-- 通用更新时间触发器函数
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash            BYTEA       NOT NULL,
    phone_encrypted       BYTEA       NOT NULL,
    status                SMALLINT    NOT NULL DEFAULT 1,
    role                  SMALLINT    NOT NULL DEFAULT 1,
    discovery_enabled     BOOLEAN     NOT NULL DEFAULT FALSE,
    is_seed               BOOLEAN     NOT NULL DEFAULT FALSE,
    deletion_requested_at TIMESTAMPTZ,
    last_login_at         TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ,
    CONSTRAINT users_status_check CHECK (status BETWEEN 1 AND 5),
    CONSTRAINT users_role_check CHECK (role BETWEEN 1 AND 4)
);
COMMENT ON TABLE users IS '用户账号。手机号只保存 HMAC 摘要与 AES-GCM 密文';
COMMENT ON COLUMN users.status IS '1 pending/2 active/3 suspended/4 deleted/5 deletion_pending';
COMMENT ON COLUMN users.role IS '1 user/2 moderator/3 admin/4 super_admin';

CREATE UNIQUE INDEX users_phone_hash_uidx ON users (phone_hash);
CREATE INDEX users_status_discovery_idx ON users (status, discovery_enabled);

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Refresh Token 只保存 hash，支持轮换、吊销与重用检测
CREATE TABLE auth_refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash  BYTEA       NOT NULL,
    family_id   UUID        NOT NULL,
    user_agent  TEXT        NOT NULL DEFAULT '',
    ip_hash     BYTEA,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    replaced_by UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX auth_refresh_tokens_hash_uidx ON auth_refresh_tokens (token_hash);
CREATE INDEX auth_refresh_tokens_user_idx ON auth_refresh_tokens (user_id, expires_at DESC);
CREATE INDEX auth_refresh_tokens_family_idx ON auth_refresh_tokens (family_id);

-- 验证记录：第一版只实现 phone
CREATE TABLE verification_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type        SMALLINT    NOT NULL,
    status      SMALLINT    NOT NULL DEFAULT 1,
    evidence    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    verified_at TIMESTAMPTZ,
    expire_at   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT verification_records_type_check CHECK (type BETWEEN 1 AND 5),
    CONSTRAINT verification_records_status_check CHECK (status BETWEEN 1 AND 5)
);
CREATE UNIQUE INDEX verification_records_user_type_uidx ON verification_records (user_id, type);
CREATE INDEX verification_records_status_idx ON verification_records (type, status);

CREATE TRIGGER verification_records_set_updated_at BEFORE UPDATE ON verification_records
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 同意记录：必须保存版本号
CREATE TABLE consent_records (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    consent_type   SMALLINT    NOT NULL,
    policy_version TEXT        NOT NULL,
    consented_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at     TIMESTAMPTZ,
    CONSTRAINT consent_records_type_check CHECK (consent_type BETWEEN 1 AND 3)
);
CREATE UNIQUE INDEX consent_records_user_type_uidx ON consent_records (user_id, consent_type);
CREATE INDEX consent_records_user_idx ON consent_records (user_id, consented_at DESC);

-- +goose Down
DROP TABLE IF EXISTS consent_records;
DROP TABLE IF EXISTS verification_records;
DROP TABLE IF EXISTS auth_refresh_tokens;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS set_updated_at();
