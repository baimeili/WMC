-- +goose Up
-- 用户资料、择偶偏好、照片

CREATE TABLE user_profiles (
    user_id             UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    nickname            TEXT        NOT NULL DEFAULT '',
    gender              SMALLINT    NOT NULL DEFAULT 0,
    birth_date          DATE,
    province_code       TEXT        NOT NULL DEFAULT '',
    city_code           TEXT        NOT NULL DEFAULT '',
    district_code       TEXT,
    height_cm           SMALLINT,
    education_level     SMALLINT    NOT NULL DEFAULT 0,
    occupation_category SMALLINT    NOT NULL DEFAULT 0,
    employment_status   SMALLINT    NOT NULL DEFAULT 0,
    annual_income_band  SMALLINT    NOT NULL DEFAULT 8,
    family_asset_band   SMALLINT,
    debt_band           SMALLINT    NOT NULL DEFAULT 0,
    house_status        SMALLINT    NOT NULL DEFAULT 0,
    vehicle_status      SMALLINT    NOT NULL DEFAULT 0,
    marital_status      SMALLINT    NOT NULL DEFAULT 0,
    children_count      SMALLINT    NOT NULL DEFAULT 0,
    wants_children      SMALLINT    NOT NULL DEFAULT 0,
    marriage_intent     SMALLINT    NOT NULL DEFAULT 0,
    marriage_timeline   SMALLINT    NOT NULL DEFAULT 0,
    smoking_status      SMALLINT    NOT NULL DEFAULT 0,
    drinking_status     SMALLINT    NOT NULL DEFAULT 0,
    bio                 TEXT        NOT NULL DEFAULT '',
    profile_completion  SMALLINT    NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_profiles_gender_check CHECK (gender BETWEEN 0 AND 3),
    CONSTRAINT user_profiles_height_check CHECK (height_cm IS NULL OR height_cm BETWEEN 100 AND 250),
    CONSTRAINT user_profiles_children_check CHECK (children_count BETWEEN 0 AND 10),
    CONSTRAINT user_profiles_completion_check CHECK (profile_completion BETWEEN 0 AND 100)
);
COMMENT ON TABLE user_profiles IS '一名用户一条资料；收入等敏感字段一律使用区间编码';
COMMENT ON COLUMN user_profiles.annual_income_band IS '1 <5万 2 5-10万 3 10-20万 4 20-30万 5 30-50万 6 50-100万 7 100万+ 8 不愿透露';

CREATE INDEX user_profiles_gender_idx ON user_profiles (gender);
CREATE INDEX user_profiles_birth_date_idx ON user_profiles (birth_date);
CREATE INDEX user_profiles_city_code_idx ON user_profiles (city_code);
CREATE INDEX user_profiles_height_idx ON user_profiles (height_cm);
CREATE INDEX user_profiles_education_idx ON user_profiles (education_level);
CREATE INDEX user_profiles_income_idx ON user_profiles (annual_income_band);
CREATE INDEX user_profiles_marital_idx ON user_profiles (marital_status);

CREATE TRIGGER user_profiles_set_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 择偶条件：*_hard = true 表示硬条件（不满足直接淘汰），false 参与软匹配
CREATE TABLE partner_preferences (
    user_id                     UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    target_genders              SMALLINT[]  NOT NULL DEFAULT '{}',
    min_age                     SMALLINT    NOT NULL DEFAULT 18,
    max_age                     SMALLINT    NOT NULL DEFAULT 99,
    age_hard                    BOOLEAN     NOT NULL DEFAULT TRUE,
    min_height_cm               SMALLINT    NOT NULL DEFAULT 0,
    max_height_cm               SMALLINT    NOT NULL DEFAULT 250,
    height_hard                 BOOLEAN     NOT NULL DEFAULT FALSE,
    allowed_city_codes          TEXT[]      NOT NULL DEFAULT '{}',
    city_hard                   BOOLEAN     NOT NULL DEFAULT FALSE,
    min_education_level         SMALLINT    NOT NULL DEFAULT 0,
    education_hard              BOOLEAN     NOT NULL DEFAULT FALSE,
    min_income_band             SMALLINT    NOT NULL DEFAULT 0,
    income_hard                 BOOLEAN     NOT NULL DEFAULT FALSE,
    allowed_marital_statuses    SMALLINT[]  NOT NULL DEFAULT '{}',
    marital_status_hard         BOOLEAN     NOT NULL DEFAULT FALSE,
    allowed_children_states     SMALLINT[]  NOT NULL DEFAULT '{}',
    children_hard               BOOLEAN     NOT NULL DEFAULT FALSE,
    allowed_wants_children      SMALLINT[]  NOT NULL DEFAULT '{}',
    wants_children_hard         BOOLEAN     NOT NULL DEFAULT FALSE,
    allowed_smoking_statuses    SMALLINT[]  NOT NULL DEFAULT '{}',
    smoking_hard                BOOLEAN     NOT NULL DEFAULT FALSE,
    allowed_drinking_statuses   SMALLINT[]  NOT NULL DEFAULT '{}',
    drinking_hard               BOOLEAN     NOT NULL DEFAULT FALSE,
    min_marriage_intent         SMALLINT    NOT NULL DEFAULT 0,
    marriage_intent_hard        BOOLEAN     NOT NULL DEFAULT FALSE,
    weights                     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT partner_preferences_age_check CHECK (min_age <= max_age),
    CONSTRAINT partner_preferences_height_check CHECK (min_height_cm <= max_height_cm)
);
COMMENT ON TABLE partner_preferences IS '一名用户一条择偶条件；数组字段使用 GIN 索引';

CREATE INDEX partner_preferences_target_genders_gin ON partner_preferences USING GIN (target_genders);
CREATE INDEX partner_preferences_cities_gin ON partner_preferences USING GIN (allowed_city_codes);
CREATE INDEX partner_preferences_marital_gin ON partner_preferences USING GIN (allowed_marital_statuses);
CREATE INDEX partner_preferences_children_gin ON partner_preferences USING GIN (allowed_children_states);
CREATE INDEX partner_preferences_wants_children_gin ON partner_preferences USING GIN (allowed_wants_children);
CREATE INDEX partner_preferences_smoking_gin ON partner_preferences USING GIN (allowed_smoking_statuses);
CREATE INDEX partner_preferences_drinking_gin ON partner_preferences USING GIN (allowed_drinking_statuses);

CREATE TRIGGER partner_preferences_set_updated_at BEFORE UPDATE ON partner_preferences
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 照片只保存 object_key，展示时生成签名 URL
CREATE TABLE user_photos (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    object_key        TEXT        NOT NULL,
    sort_order        SMALLINT    NOT NULL DEFAULT 0,
    is_primary        BOOLEAN     NOT NULL DEFAULT FALSE,
    moderation_status SMALLINT    NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT user_photos_moderation_check CHECK (moderation_status BETWEEN 1 AND 3)
);
CREATE INDEX user_photos_user_idx ON user_photos (user_id, sort_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX user_photos_primary_uidx ON user_photos (user_id) WHERE is_primary AND deleted_at IS NULL;
CREATE INDEX user_photos_moderation_idx ON user_photos (moderation_status) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS user_photos;
DROP TABLE IF EXISTS partner_preferences;
DROP TABLE IF EXISTS user_profiles;
