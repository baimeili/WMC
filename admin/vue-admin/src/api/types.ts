// 管理端接口类型定义。
// 以 backend/openapi/openapi.yaml 为契约来源；字段统一声明为可选，
// 便于后端裁剪字段时页面降级展示而不是直接报错。

// 统一响应信封：{ code, message, data, request_id }。
export interface ApiEnvelope<T> {
  code: string
  message?: string
  data?: T
  request_id?: string
}

// 登录令牌对。
export interface TokenPair {
  access_token?: string
  refresh_token?: string
  expires_in?: number
  user_id?: string
  role?: number
  profile_completion?: number
  discovery_enabled?: boolean
}

// 漏斗步骤（看板与经营漏斗共用）。
export interface FunnelStep {
  key?: string
  label?: string
  count?: number
  conversion?: number
}

// 运营看板数据。
export interface AdminDashboard {
  total_users?: number
  new_users_7d?: number
  dau?: number
  wau?: number
  mau?: number
  profile_completed_users?: number
  market_users?: number
  gender_distribution?: Record<string, number>
  city_distribution?: Array<{ city_name?: string; count?: number }>
  age_distribution?: Array<{ bucket?: string; count?: number }>
  interactions_like?: number
  matches_total?: number
  match_rate?: number
  offline_meeting_reports?: number
  confirmed_meetings?: number
  relationship_reports?: number
  north_star?: number
  funnel?: FunnelStep[]
}

// 用户列表项。
export interface AdminUserItem {
  id?: string
  nickname?: string
  phone_masked?: string
  status?: number
  role?: number
  discovery_enabled?: boolean
  profile_completion?: number
  is_seed?: boolean
  created_at?: string
}

// 用户分页结果。
export interface AdminUserPage {
  items?: AdminUserItem[]
  total?: number
  page?: number
  page_size?: number
}

// 用户资料摘要（详情页展示，契约暂未单列 schema，字段保持宽松）。
export interface AdminUserProfileSummary {
  nickname?: string
  gender?: number
  age?: number
  birth_date?: string
  province_code?: string
  city_code?: string
  city_name?: string
  height_cm?: number
  education_level?: number
  occupation_category?: number
  marital_status?: number
  children_count?: number
  wants_children?: number
  marriage_intent?: number
  bio?: string
  profile_completion?: number
  updated_at?: string
}

// 用户详情：在列表项基础上补充资料摘要、最近登录与举报记录。
export interface AdminUserDetail extends AdminUserItem {
  last_login_at?: string
  profile?: AdminUserProfileSummary
  reports?: Report[]
}

// 举报记录。
export interface Report {
  id?: string
  reporter_user_id?: string
  reported_user_id?: string
  reported_nickname?: string
  reason?: number
  description?: string
  status?: number
  resolution_note?: string
  created_at?: string
}

// 处理举报入参：status 2 处理中 / 3 已处理 / 4 已驳回。
export interface ResolveReportInput {
  status: number
  note?: string
}

// 照片审核记录（运营后台图片审核接口）。
export interface AdminPhoto {
  id?: string
  user_id?: string
  nickname?: string
  user_nickname?: string
  url?: string
  object_key?: string
  moderation_status?: number
  status?: number
  is_primary?: boolean
  created_at?: string
}

// 照片分页结果。
export interface AdminPhotoPage {
  items?: AdminPhoto[]
  total?: number
  page?: number
  page_size?: number
}

// 基准数据集。
export interface BaselineDataset {
  id?: string
  name?: string
  version?: string
  source_name?: string
  source_url?: string
  source_year?: number | null
  description?: string
  methodology?: string
  disclaimer?: string
  status?: number
  active?: boolean
  is_demo?: boolean
  record_count?: number
  created_at?: string
}

// 创建基准数据集入参。
export interface BaselineDatasetInput {
  name: string
  version: string
  source_name?: string
  source_url?: string
  source_year?: number | null
  description?: string
  methodology?: string
  disclaimer?: string
  is_demo?: boolean
}

// CSV 导入结果（dry_run 时 imported=false，仅返回校验错误与预览）。
export interface ImportReport {
  total_rows?: number
  valid_rows?: number
  invalid_rows?: number
  errors?: string[]
  imported?: boolean
  preview?: Array<Record<string, unknown>>
}

// 价值观问题选项。
export interface QuestionOption {
  value: string
  label: string
}

// 价值观问题。
export interface CompatibilityQuestion {
  id?: string
  code?: string
  category?: string
  title?: string
  description?: string
  question_type?: number
  options?: QuestionOption[]
  active?: boolean
  sort_order?: number
  version?: number
}

// 价值观问题新增/修改入参。
export interface CompatibilityQuestionInput {
  code: string
  category: string
  title: string
  description?: string
  question_type: number
  options: QuestionOption[]
  active: boolean
  sort_order: number
}

// 系统参数与 Feature Flag。
export interface SettingsBundle {
  features?: Record<string, boolean>
  settings?: Record<string, unknown>
}

// 审计日志。
export interface AuditLog {
  id?: number
  actor_user_id?: string | null
  action?: string
  target_type?: string
  target_id?: string
  detail?: unknown
  created_at?: string
}

// 用户列表查询参数。
export interface AdminUserQuery {
  keyword?: string
  status?: number
  page?: number
  page_size?: number
}

// 审计日志查询参数。
export interface AuditLogQuery {
  limit?: number
  offset?: number
}
