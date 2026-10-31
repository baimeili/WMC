-- +goose Up
-- 补充索引：面向候选池与漏斗查询

-- 婚恋池用户快速过滤
CREATE INDEX users_discovery_active_idx ON users (id)
    WHERE status = 2 AND discovery_enabled;

-- 资料完成度过滤（入池门槛）
CREATE INDEX user_profiles_completion_idx ON user_profiles (profile_completion);

-- 候选池常按「性别 + 城市 + 年龄」组合筛选
CREATE INDEX user_profiles_gender_city_idx ON user_profiles (gender, city_code);
CREATE INDEX user_profiles_gender_income_idx ON user_profiles (gender, annual_income_band);

-- 价值观规则批量加载
CREATE INDEX user_compatibility_rules_user_idx ON user_compatibility_rules (user_id);

-- 事件按日聚合
CREATE INDEX product_events_time_idx ON product_events (occurred_at DESC);

-- +goose Down
DROP INDEX IF EXISTS product_events_time_idx;
DROP INDEX IF EXISTS user_compatibility_rules_user_idx;
DROP INDEX IF EXISTS user_profiles_gender_income_idx;
DROP INDEX IF EXISTS user_profiles_gender_city_idx;
DROP INDEX IF EXISTS user_profiles_completion_idx;
DROP INDEX IF EXISTS users_discovery_active_idx;
