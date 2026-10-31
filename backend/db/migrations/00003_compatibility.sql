-- +goose Up
-- 婚恋价值观问题体系：问题不写死在客户端，后台可增改停用

CREATE TABLE compatibility_questions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          TEXT        NOT NULL,
    category      TEXT        NOT NULL,
    title         TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    question_type SMALLINT    NOT NULL DEFAULT 1,
    options       JSONB       NOT NULL DEFAULT '[]'::jsonb,
    active        BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order    SMALLINT    NOT NULL DEFAULT 0,
    version       SMALLINT    NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT compatibility_questions_type_check CHECK (question_type BETWEEN 1 AND 2)
);
CREATE UNIQUE INDEX compatibility_questions_code_uidx ON compatibility_questions (code);
CREATE INDEX compatibility_questions_active_idx ON compatibility_questions (active, sort_order);

CREATE TRIGGER compatibility_questions_set_updated_at BEFORE UPDATE ON compatibility_questions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 用户价值观规则：重要度 0 不在意 ~ 5 非常重要；is_hard 为硬条件
CREATE TABLE user_compatibility_rules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    question_id      UUID        NOT NULL REFERENCES compatibility_questions (id) ON DELETE CASCADE,
    own_answer       JSONB       NOT NULL DEFAULT '[]'::jsonb,
    accepted_answers JSONB       NOT NULL DEFAULT '[]'::jsonb,
    importance       SMALLINT    NOT NULL DEFAULT 3,
    is_hard          BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_compatibility_rules_importance_check CHECK (importance BETWEEN 0 AND 5)
);
CREATE UNIQUE INDEX user_compatibility_rules_user_question_uidx ON user_compatibility_rules (user_id, question_id);
CREATE INDEX user_compatibility_rules_question_idx ON user_compatibility_rules (question_id);

CREATE TRIGGER user_compatibility_rules_set_updated_at BEFORE UPDATE ON user_compatibility_rules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 第一版 12 道婚恋价值观问题（财务/房屋/家务/育儿/职业/异地/父母/彩礼/同住/消费/冲突/社交）
INSERT INTO compatibility_questions (code, category, title, description, question_type, options, sort_order) VALUES
('finance_household', '财务', '婚后家庭开支如何承担？', '日常开销、公共支出如何分摊', 1,
 '[{"value":"male","label":"主要由男方"},{"value":"female","label":"主要由女方"},{"value":"income_ratio","label":"按收入比例"},{"value":"aa","label":"大致 AA"},{"value":"joint_account","label":"全部进入共同账户"},{"value":"other","label":"其他"}]', 10),
('house_downpayment', '房屋', '购房首付如何承担？', '首付款、产权与还贷安排', 1,
 '[{"value":"male","label":"主要由男方"},{"value":"female","label":"主要由女方"},{"value":"split","label":"双方按比例"},{"value":"joint","label":"共同出资并共同署名"},{"value":"one_party_full","label":"一方全款"},{"value":"other","label":"其他"}]', 20),
('chores', '家务', '家务如何分配？', '日常家务与清洁安排', 1,
 '[{"value":"equal","label":"平均分担"},{"value":"one_main","label":"以一方为主，另一方辅助"},{"value":"outsource","label":"请家政"},{"value":"flexible","label":"按时间灵活分配"},{"value":"other","label":"其他"}]', 30),
('parenting_time', '育儿', '孩子出生后谁承担主要育儿时间？', '包括夜间照料、接送与日常陪伴', 1,
 '[{"value":"mother","label":"母亲为主"},{"value":"father","label":"父亲为主"},{"value":"both","label":"双方平均"},{"value":"grandparents","label":"由祖辈协助"},{"value":"other","label":"其他"}]', 40),
('career_conflict', '职业', '如果双方事业发生冲突，应如何处理？', '涉及迁居、加班、出国等重大安排', 1,
 '[{"value":"one_yields","label":"一方主动让步"},{"value":"take_turns","label":"轮流让步"},{"value":"negotiate","label":"共同协商后再定"},{"value":"independent","label":"各自发展互不迁就"},{"value":"other","label":"其他"}]', 50),
('relocation', '异地', '双方工作城市不同，是否接受迁居？', '婚前或婚后的工作城市选择', 1,
 '[{"value":"i_relocate","label":"愿意迁居对方城市"},{"value":"partner_relocate","label":"希望对方迁居我所在城市"},{"value":"keep_long_distance","label":"都不迁居，维持异地"},{"value":"depends","label":"视情况而定"}]', 60),
('elder_care', '父母', '婚后双方父母养老责任如何承担？', '经济支持与照护分工', 1,
 '[{"value":"own_parents","label":"各自负责各自父母"},{"value":"share_all","label":"共同承担双方父母"},{"value":"one_dominant","label":"以一方为主"},{"value":"other","label":"其他"}]', 70),
('bride_price', '彩礼嫁妆', '你如何看待彩礼和嫁妆？', '具体安排而非立场表态', 1,
 '[{"value":"local_custom","label":"按当地习俗执行"},{"value":"symbolic","label":"双方协商象征性金额"},{"value":"not_needed","label":"不需要彩礼嫁妆"},{"value":"startup_fund","label":"全部作为新家庭启动资金"},{"value":"other","label":"其他"}]', 80),
('live_with_parents', '父母同住', '是否接受婚后长期与父母同住？', '长期同住而非短期探访', 1,
 '[{"value":"accept","label":"完全接受"},{"value":"short_term","label":"短期可以，长期不接受"},{"value":"reject","label":"不接受"},{"value":"depends","label":"视具体情况"}]', 90),
('consumption_model', '消费', '你更倾向哪种家庭消费模式？', '收入归集与大额支出决策', 1,
 '[{"value":"joint_account","label":"共同账户统一管理"},{"value":"separate_split","label":"各自管理，共同开支均摊"},{"value":"one_manages","label":"一方主导管理"},{"value":"ratio","label":"按收入比例分担"},{"value":"other","label":"其他"}]', 100),
('conflict_resolution', '冲突', '出现严重争论后，你通常希望如何解决？', '冲突后的沟通方式', 1,
 '[{"value":"same_day","label":"当天说清楚"},{"value":"cool_down","label":"先冷静再谈"},{"value":"third_party","label":"请第三方调解"},{"value":"avoid","label":"回避不提"}]', 110),
('social_boundary', '社交', '伴侣与异性的社交边界应该是什么？', '日常社交与单独相处', 1,
 '[{"value":"normal","label":"保持正常社交即可"},{"value":"notify","label":"需提前告知"},{"value":"joint","label":"共同参与"},{"value":"avoid_alone","label":"尽量避免单独相处"}]', 120);

-- +goose Down
DROP TABLE IF EXISTS user_compatibility_rules;
DROP TABLE IF EXISTS compatibility_questions;
