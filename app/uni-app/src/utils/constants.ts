// 全局常量与产品固定文案。
// 产品硬性规则：理论市场与本站真实市场严格区分；真实候选数小于 K 匿名阈值一律显示 <10。

function resolveApiBaseUrl(): string {
  const env = (import.meta as unknown as { env?: Record<string, string | undefined> }).env
  const configured = env && env.VITE_API_BASE_URL ? String(env.VITE_API_BASE_URL) : ''
  if (configured) {
    return configured.replace(/\/$/, '')
  }
  // H5 默认同源访问 /api/v1，开发环境由 vite 代理到 http://127.0.0.1:8080
  return '/api/v1'
}

export const API_BASE_URL = resolveApiBaseUrl()

// K 匿名阈值：真实市场少于该人数一律显示 <10
export const K_ANONYMITY_THRESHOLD = 10

export const COLD_START_MESSAGE =
  '本站目前仍处于早期测试阶段，你当前已有 <10 名双向基础候选人，随着目标城市用户增加，候选池会自动更新'

export const RESULT_DISCLAIMER = '以上分数仅根据双方填写条件计算，不代表婚姻成功概率。'

export const ESTIMATE_DISCLAIMER = '理论市场结果为统计估算，不是精确人数，也不代表本站真实用户数量。'

export const INSUFFICIENT_COMPATIBILITY_TEXT = '关系观念数据不足'

export const FUNNEL_NOTE = '按当前条件顺序计算，改变条件顺序会改变每一步的损失展示'

export const K_ANONYMITY_NOTE = '为保护隐私，少于 10 人的样本统一显示为 <10。'

export const SCORE_LABELS = {
  preference: '条件契合度',
  compatibility: '关系观念兼容度',
  overall: '综合匹配度'
}

export const MARKET_COPY = {
  theoretical: {
    name: '理论市场估算',
    short: '理论市场',
    desc: '基于后台导入的基准人口数据集做统计估算，只有条件占比与敏感性分析，不能计算双向匹配率。'
  },
  real: {
    name: '本站真实用户池',
    short: '本站真实市场',
    desc: '只统计已注册、资料达标并主动加入婚恋池的用户，双向硬条件全部满足才计入候选人。'
  }
} as const

export type MarketType = keyof typeof MARKET_COPY

export const APP_VERSION = '1.0.0'

export const MATCH_STATUS_LABELS: Record<number, string> = {
  1: '进行中',
  2: '已取消',
  3: '已屏蔽'
}

export const REPORT_STATUS_TAIL = '我们会在收到举报后按社区规则处理，处理结果不会向被举报人透露你的身份。'
