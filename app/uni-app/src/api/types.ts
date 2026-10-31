// 与 backend/openapi/openapi.yaml 对齐的客户端契约类型。
// 当前仓库尚未产出 openapi-typescript 生成物，此文件作为客户端唯一类型来源，字段按页面需要裁剪。

export interface Envelope<T = unknown> {
  code: string
  message: string
  data: T
  request_id: string
}

export interface OptionItem {
  value: number
  label: string
}

export type OptionDict = Record<string, OptionItem[]>

export interface Region {
  code: string
  name: string
  level: number
  parent_code: string | null
}

export interface AppConfig {
  features: Record<string, boolean>
  settings: Record<string, unknown>
  policy_versions: Record<string, string>
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
  user_id: string
  role: number
  profile_completion: number
  discovery_enabled: boolean
}

export interface CompatibilityQuestionOption {
  value: string
  label: string
}

export interface CompatibilityQuestion {
  id: string
  code: string
  category: string
  title: string
  description: string
  question_type: number
  options: CompatibilityQuestionOption[]
  active: boolean
  sort_order: number
  version: number
}

export interface CompatibilityRule {
  question_id: string
  question_code: string
  own_answer: string[]
  accepted_answers: string[]
  importance: number
  is_hard: boolean
}

export interface CompatibilityRuleInput {
  question_id: string
  own_answer: string[]
  accepted_answers: string[]
  importance: number
  is_hard: boolean
}

export interface Profile {
  user_id: string
  nickname: string
  gender: number
  birth_date: string | null
  age: number | null
  province_code: string
  city_code: string
  district_code: string | null
  height_cm: number | null
  education_level: number
  occupation_category: number
  employment_status: number
  annual_income_band: number
  family_asset_band: number | null
  debt_band: number
  house_status: number
  vehicle_status: number
  marital_status: number
  children_count: number
  wants_children: number
  marriage_intent: number
  marriage_timeline: number
  smoking_status: number
  drinking_status: number
  bio: string
  profile_completion: number
  updated_at: string
}

export interface ProfileInput {
  nickname: string
  gender: number
  city_code: string
  birth_date?: string | null
  province_code?: string
  district_code?: string | null
  height_cm?: number | null
  education_level?: number
  occupation_category?: number
  employment_status?: number
  annual_income_band?: number
  family_asset_band?: number | null
  debt_band?: number
  house_status?: number
  vehicle_status?: number
  marital_status?: number
  children_count?: number
  wants_children?: number
  marriage_intent?: number
  marriage_timeline?: number
  smoking_status?: number
  drinking_status?: number
  bio?: string
}

export interface Preferences {
  target_genders?: number[]
  min_age?: number
  max_age?: number
  age_hard?: boolean
  min_height_cm?: number
  max_height_cm?: number
  height_hard?: boolean
  allowed_city_codes?: string[]
  city_hard?: boolean
  min_education_level?: number
  education_hard?: boolean
  min_income_band?: number
  income_hard?: boolean
  allowed_marital_statuses?: number[]
  marital_status_hard?: boolean
  allowed_children_states?: number[]
  children_hard?: boolean
  allowed_wants_children?: number[]
  wants_children_hard?: boolean
  allowed_smoking_statuses?: number[]
  smoking_hard?: boolean
  allowed_drinking_statuses?: number[]
  drinking_hard?: boolean
  min_marriage_intent?: number
  marriage_intent_hard?: boolean
  weights?: Record<string, number>
}

export type PreferencesInput = Preferences

export interface Photo {
  id: string
  url: string
  sort_order: number
  is_primary: boolean
  moderation_status: number
}

export interface CalculatorSelf {
  gender?: number
  age?: number
  city_code?: string
  height_cm?: number | null
  education_level?: number
  annual_income_band?: number
}

export interface CalculatorConditions {
  target_genders?: number[]
  min_age?: number | null
  max_age?: number | null
  allowed_city_codes?: string[]
  min_height_cm?: number | null
  max_height_cm?: number | null
  min_education_level?: number | null
  min_income_band?: number | null
  allowed_marital_statuses?: number[]
  allowed_children_states?: number[]
  allowed_wants_children?: number[]
}

export interface CalculatorRequest {
  anonymous_session_id?: string
  self?: CalculatorSelf
  conditions: CalculatorConditions
}

export interface DatasetInfo {
  id: string
  name: string
  version: string
  source_name: string
  is_demo: boolean
  disclaimer: string
}

export interface FunnelStep {
  key: string
  label: string
  count: number
  ratio: number
  drop_ratio: number
  applied: boolean
}

export interface SensitivityItem {
  key: string
  label: string
  base_count: number
  without_count: number
  delta_ratio: number
  added_count: number
}

export interface CalculatorFunnel {
  total_population: number
  final_count: number
  final_ratio: number
  steps: FunnelStep[]
}

export interface CalculatorResult {
  run_id: string
  dataset: DatasetInfo
  total_population: number
  matched_count: number
  matched_ratio: number
  funnel: CalculatorFunnel
  sensitivity: SensitivityItem[]
  top_factors: SensitivityItem[]
  funnel_note: string
  share_token: string | null
  generated_at: string
}

export interface FunnelResult {
  steps: FunnelStep[]
  total: number
  final_count: number
  final_ratio: number
  reciprocal_count: number
  reciprocal_ratio: number
  note: string
}

export interface SensitivityResult {
  items: SensitivityItem[]
  note: string
}

export interface DiscoveryRequirements {
  min_profile_completion: number
  age_ok: boolean
  consent_ok: boolean
  account_ok: boolean
}

export interface DiscoverySummary {
  in_market: boolean
  eligible_count_text: string
  eligible_count: number
  one_way_count: number
  reciprocal_count: number
  early_stage: boolean
  early_stage_message: string
  profile_completion: number
  requirements: DiscoveryRequirements
}

export interface CandidateCardItem {
  user_id: string
  nickname: string
  age: number
  city_code: string
  city_name: string
  height_cm: number | null
  education_level: number
  occupation_category: number
  marital_status: number
  wants_children: number
  primary_photo_url: string | null
  bio: string
  preference_score: number
  compatibility_score: number
  overall_score: number
  compatibility_insufficient: boolean
  my_action: number | null
}

export interface CandidatePage {
  items: CandidateCardItem[]
  next_cursor: string
  has_more: boolean
}

export interface CandidateDimension {
  name: string
  score: number
}

export interface CandidateDetail extends CandidateCardItem {
  reasons: string[]
  conflicts: string[]
  dimensions: CandidateDimension[]
  disclaimer: string
}

export interface MatchSummary {
  id: string
  peer_user_id: string
  peer_nickname: string
  peer_age: number
  peer_city_name: string
  peer_photo_url: string | null
  matched_at: string
  status: number
  confirmed_meeting: boolean
}

export interface MatchPage {
  items: MatchSummary[]
  next_cursor: string
  has_more: boolean
}

export interface MatchDetail extends MatchSummary {
  my_outcomes: number[]
  peer_outcomes: number[]
  chat_enabled: boolean
}

export interface InteractionResult {
  matched: boolean
  match_id: string | null
}

export interface ReportPayload {
  reason: number
  description?: string
  evidence_keys?: string[]
}

export interface OutcomePayload {
  outcome_type: number
  occurred_at?: string | null
  note?: string
}
