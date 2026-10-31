-- name: GetProfile :one
SELECT * FROM user_profiles WHERE user_id = $1;

-- name: CreateProfile :one
INSERT INTO user_profiles (user_id, nickname, gender, birth_date, province_code, city_code, district_code,
    height_cm, education_level, occupation_category, employment_status, annual_income_band,
    family_asset_band, debt_band, house_status, vehicle_status, marital_status, children_count,
    wants_children, marriage_intent, marriage_timeline, smoking_status, drinking_status, bio, profile_completion)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
RETURNING *;

-- name: UpdateProfile :one
UPDATE user_profiles SET
    nickname = $2,
    gender = $3,
    birth_date = $4,
    province_code = $5,
    city_code = $6,
    district_code = $7,
    height_cm = $8,
    education_level = $9,
    occupation_category = $10,
    employment_status = $11,
    annual_income_band = $12,
    family_asset_band = $13,
    debt_band = $14,
    house_status = $15,
    vehicle_status = $16,
    marital_status = $17,
    children_count = $18,
    wants_children = $19,
    marriage_intent = $20,
    marriage_timeline = $21,
    smoking_status = $22,
    drinking_status = $23,
    bio = $24,
    profile_completion = $25
WHERE user_id = $1
RETURNING *;

-- name: UpdateProfileCompletion :exec
UPDATE user_profiles SET profile_completion = $2 WHERE user_id = $1;

-- name: GetPreferences :one
SELECT * FROM partner_preferences WHERE user_id = $1;

-- name: CreatePreferences :one
INSERT INTO partner_preferences (user_id, target_genders, min_age, max_age, age_hard, min_height_cm, max_height_cm,
    height_hard, allowed_city_codes, city_hard, min_education_level, education_hard, min_income_band, income_hard,
    allowed_marital_statuses, marital_status_hard, allowed_children_states, children_hard, allowed_wants_children,
    wants_children_hard, allowed_smoking_statuses, smoking_hard, allowed_drinking_statuses, drinking_hard,
    min_marriage_intent, marriage_intent_hard, weights)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
RETURNING *;

-- name: UpdatePreferences :one
UPDATE partner_preferences SET
    target_genders = $2,
    min_age = $3,
    max_age = $4,
    age_hard = $5,
    min_height_cm = $6,
    max_height_cm = $7,
    height_hard = $8,
    allowed_city_codes = $9,
    city_hard = $10,
    min_education_level = $11,
    education_hard = $12,
    min_income_band = $13,
    income_hard = $14,
    allowed_marital_statuses = $15,
    marital_status_hard = $16,
    allowed_children_states = $17,
    children_hard = $18,
    allowed_wants_children = $19,
    wants_children_hard = $20,
    allowed_smoking_statuses = $21,
    smoking_hard = $22,
    allowed_drinking_statuses = $23,
    drinking_hard = $24,
    min_marriage_intent = $25,
    marriage_intent_hard = $26,
    weights = $27
WHERE user_id = $1
RETURNING *;
