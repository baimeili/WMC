// 与后端 smallint 编码一致的中文标签字典。
// 说明：正式环境可由 /meta/options 下发，管理端保留本地兜底字典，避免登录前依赖字典接口。
export interface Option {
  value: number
  label: string
}

export const ROLE_LABELS: Record<number, string> = {
  1: '普通用户',
  2: '内容审核员',
  3: '管理员',
  4: '超级管理员'
}

export const USER_STATUS_LABELS: Record<number, string> = {
  1: '待激活',
  2: '正常',
  3: '已冻结',
  4: '已注销',
  5: '注销保留期'
}

export const GENDER_LABELS: Record<number, string> = {
  0: '未填写',
  1: '男',
  2: '女',
  3: '不愿透露'
}

export const REPORT_REASON_LABELS: Record<number, string> = {
  1: '诈骗',
  2: '骚扰',
  3: '虚假资料',
  4: '盗图',
  5: '色情',
  6: '广告',
  7: '已婚冒充单身',
  8: '其他'
}

export const REPORT_STATUS_LABELS: Record<number, string> = {
  1: '待处理',
  2: '处理中',
  3: '已处理',
  4: '已驳回'
}

export const MODERATION_STATUS_LABELS: Record<number, string> = {
  1: '待审核',
  2: '已通过',
  3: '已拒绝'
}

export const DATASET_STATUS_LABELS: Record<number, string> = {
  1: '草稿',
  2: '可用',
  3: '导入中',
  4: '已归档'
}

export const QUESTION_TYPE_LABELS: Record<number, string> = {
  1: '单选',
  2: '多选'
}

export const EDUCATION_LABELS: Record<number, string> = {
  0: '未填写',
  1: '高中及以下',
  2: '中专/大专',
  3: '本科',
  4: '硕士',
  5: '博士'
}

export const MARITAL_STATUS_LABELS: Record<number, string> = {
  0: '未填写',
  1: '未婚',
  2: '离异',
  3: '丧偶',
  4: '不愿透露'
}

export const WANTS_CHILDREN_LABELS: Record<number, string> = {
  0: '未填写',
  1: '不想要孩子',
  2: '想要孩子',
  3: '看情况',
  4: '已有子女，不再生育'
}

// 安全取标签，未知编码返回兜底文案。
export function labelOf(map: Record<number, string>, code: number | null | undefined, fallback = '未知'): string {
  if (code === null || code === undefined) {
    return fallback
  }
  return map[code] ?? fallback
}

// 将编码字典转为下拉选项，按编码升序。
export function toOptions(map: Record<number, string>): Option[] {
  return Object.keys(map)
    .map((key) => Number(key))
    .filter((code) => Number.isFinite(code))
    .sort((a, b) => a - b)
    .map((code) => ({ value: code, label: map[code] }))
}

export const USER_STATUS_OPTIONS: Option[] = toOptions(USER_STATUS_LABELS)
export const REPORT_STATUS_OPTIONS: Option[] = toOptions(REPORT_STATUS_LABELS)
export const REPORT_REASON_OPTIONS: Option[] = toOptions(REPORT_REASON_LABELS)
export const MODERATION_STATUS_OPTIONS: Option[] = toOptions(MODERATION_STATUS_LABELS)

// 用户状态对应的 el-tag 颜色。
export function userStatusTagType(status: number | undefined): 'success' | 'info' | 'warning' | 'danger' {
  switch (status) {
    case 2:
      return 'success'
    case 3:
      return 'danger'
    case 5:
      return 'warning'
    default:
      return 'info'
  }
}

// 举报状态对应的 el-tag 颜色。
export function reportStatusTagType(status: number | undefined): 'success' | 'info' | 'warning' | 'danger' {
  switch (status) {
    case 2:
      return 'warning'
    case 3:
      return 'success'
    case 4:
      return 'info'
    default:
      return 'danger'
  }
}