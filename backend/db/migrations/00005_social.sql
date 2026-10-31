-- +goose Up
-- 互动、Match、结果反馈、Block、举报

CREATE TABLE interactions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id  UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target_user_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    action         SMALLINT    NOT NULL,
    source         SMALLINT    NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT interactions_action_check CHECK (action IN (1, 2)),
    CONSTRAINT interactions_source_check CHECK (source BETWEEN 1 AND 3),
    CONSTRAINT interactions_no_self CHECK (actor_user_id <> target_user_id)
);
COMMENT ON COLUMN interactions.action IS '1 like/2 pass；允许用户之后修改选择';

CREATE UNIQUE INDEX interactions_actor_target_uidx ON interactions (actor_user_id, target_user_id);
CREATE INDEX interactions_target_action_idx ON interactions (target_user_id, action);

CREATE TRIGGER interactions_set_updated_at BEFORE UPDATE ON interactions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 双向 Like 形成 Match：user_a_id < user_b_id + 唯一约束，保证永远不会出现两条
CREATE TABLE matches (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a_id         UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    user_b_id         UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status            SMALLINT    NOT NULL DEFAULT 1,
    matched_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at          TIMESTAMPTZ,
    confirmed_meeting BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT matches_pair_order CHECK (user_a_id < user_b_id),
    CONSTRAINT matches_pair_unique UNIQUE (user_a_id, user_b_id),
    CONSTRAINT matches_status_check CHECK (status BETWEEN 1 AND 3)
);
COMMENT ON TABLE matches IS 'Match 唯一性由 (user_a_id,user_b_id) 唯一约束 + 有序存储保证';

CREATE INDEX matches_user_a_idx ON matches (user_a_id, matched_at DESC);
CREATE INDEX matches_user_b_idx ON matches (user_b_id, matched_at DESC);
CREATE INDEX matches_confirmed_idx ON matches (confirmed_meeting) WHERE confirmed_meeting;

-- 真实结果反馈：用户主动上报，不读取任何私人聊天内容
CREATE TABLE match_outcomes (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id         UUID        NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    reporter_user_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    outcome_type     SMALLINT    NOT NULL,
    occurred_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    note             TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT match_outcomes_type_check CHECK (outcome_type BETWEEN 1 AND 8)
);
CREATE UNIQUE INDEX match_outcomes_unique ON match_outcomes (match_id, reporter_user_id, outcome_type);
CREATE INDEX match_outcomes_match_idx ON match_outcomes (match_id, occurred_at DESC);
CREATE INDEX match_outcomes_type_idx ON match_outcomes (outcome_type, occurred_at DESC);

-- Block：一旦建立，双方不可互相发现、不可查看资料、不可互动
CREATE TABLE user_blocks (
    blocker_user_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    blocked_user_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (blocker_user_id, blocked_user_id),
    CONSTRAINT user_blocks_no_self CHECK (blocker_user_id <> blocked_user_id)
);
CREATE INDEX user_blocks_blocked_idx ON user_blocks (blocked_user_id);

CREATE TABLE user_reports (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_user_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reported_user_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reason           SMALLINT    NOT NULL,
    description      TEXT        NOT NULL DEFAULT '',
    evidence_keys    JSONB       NOT NULL DEFAULT '[]'::jsonb,
    status           SMALLINT    NOT NULL DEFAULT 1,
    reviewer_id      UUID REFERENCES users (id) ON DELETE SET NULL,
    reviewed_at      TIMESTAMPTZ,
    resolution_note  TEXT        NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_reports_reason_check CHECK (reason BETWEEN 1 AND 8),
    CONSTRAINT user_reports_status_check CHECK (status BETWEEN 1 AND 4),
    CONSTRAINT user_reports_no_self CHECK (reporter_user_id <> reported_user_id)
);
CREATE INDEX user_reports_status_idx ON user_reports (status, created_at DESC);
CREATE INDEX user_reports_reported_idx ON user_reports (reported_user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_reports;
DROP TABLE IF EXISTS user_blocks;
DROP TABLE IF EXISTS match_outcomes;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS interactions;
