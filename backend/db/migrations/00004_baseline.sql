-- +goose Up
-- 理论婚恋市场：加权匿名样本数据集与计算记录

CREATE TABLE baseline_datasets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    version     TEXT        NOT NULL,
    source_name TEXT        NOT NULL DEFAULT '',
    source_url  TEXT        NOT NULL DEFAULT '',
    source_year SMALLINT,
    description TEXT        NOT NULL DEFAULT '',
    methodology TEXT        NOT NULL DEFAULT '',
    disclaimer  TEXT        NOT NULL DEFAULT '',
    status      SMALLINT    NOT NULL DEFAULT 1,
    active      BOOLEAN     NOT NULL DEFAULT FALSE,
    is_demo     BOOLEAN     NOT NULL DEFAULT FALSE,
    record_count BIGINT     NOT NULL DEFAULT 0,
    created_by  UUID REFERENCES users (id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT baseline_datasets_status_check CHECK (status BETWEEN 1 AND 4)
);
COMMENT ON TABLE baseline_datasets IS '基准人口数据集。理论市场只能展示估算比例，禁止展示精确人数或双向匹配率';
COMMENT ON COLUMN baseline_datasets.status IS '1 draft/2 ready/3 importing/4 archived';

-- 任意时刻只能有一个激活数据集
CREATE UNIQUE INDEX baseline_datasets_active_uidx ON baseline_datasets (active) WHERE active;
CREATE INDEX baseline_datasets_created_idx ON baseline_datasets (created_at DESC);

CREATE TRIGGER baseline_datasets_set_updated_at BEFORE UPDATE ON baseline_datasets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 加权匿名样本：weight 表示该样本代表的总体人口单位数
CREATE TABLE baseline_records (
    id                 BIGSERIAL PRIMARY KEY,
    dataset_id         UUID        NOT NULL REFERENCES baseline_datasets (id) ON DELETE CASCADE,
    gender             SMALLINT    NOT NULL,
    age                SMALLINT    NOT NULL,
    province_code      TEXT        NOT NULL DEFAULT '',
    city_code          TEXT        NOT NULL DEFAULT '',
    height_cm          SMALLINT,
    education_level    SMALLINT    NOT NULL DEFAULT 0,
    annual_income_band SMALLINT    NOT NULL DEFAULT 0,
    marital_status     SMALLINT    NOT NULL DEFAULT 0,
    children_count     SMALLINT    NOT NULL DEFAULT 0,
    wants_children     SMALLINT    NOT NULL DEFAULT 0,
    sample_weight      NUMERIC(16, 4) NOT NULL DEFAULT 1,
    CONSTRAINT baseline_records_weight_check CHECK (sample_weight > 0),
    CONSTRAINT baseline_records_age_check CHECK (age BETWEEN 18 AND 100)
);

CREATE INDEX baseline_records_dataset_idx ON baseline_records (dataset_id);
CREATE INDEX baseline_records_gender_idx ON baseline_records (dataset_id, gender);
CREATE INDEX baseline_records_age_idx ON baseline_records (dataset_id, age);
CREATE INDEX baseline_records_city_idx ON baseline_records (dataset_id, city_code);

CREATE TABLE calculator_runs (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID REFERENCES users (id) ON DELETE SET NULL,
    anonymous_session_id TEXT,
    dataset_id           UUID REFERENCES baseline_datasets (id) ON DELETE SET NULL,
    input_snapshot       JSONB       NOT NULL,
    result_snapshot      JSONB       NOT NULL,
    share_token          TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT calculator_runs_owner_check CHECK (user_id IS NOT NULL OR anonymous_session_id IS NOT NULL)
);
CREATE UNIQUE INDEX calculator_runs_share_token_uidx ON calculator_runs (share_token) WHERE share_token IS NOT NULL;
CREATE INDEX calculator_runs_user_idx ON calculator_runs (user_id, created_at DESC);
CREATE INDEX calculator_runs_created_idx ON calculator_runs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS calculator_runs;
DROP TABLE IF EXISTS baseline_records;
DROP TABLE IF EXISTS baseline_datasets;
