// 管理端通用格式化工具，保证各页面数字与时间口径一致。

// 将任意输入安全转换为数字，非法值返回兜底值。
export function toNumber(value: unknown, fallback = 0): number {
  if (value === null || value === undefined || value === '') {
    return fallback
  }
  const result = Number(value)
  return Number.isFinite(result) ? result : fallback
}

// 整数千分位格式。
export function formatNumber(value: number | null | undefined, fallback = '—'): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return fallback
  }
  return Math.round(value).toLocaleString('zh-CN')
}

// 保留指定小数位。
export function formatDecimal(value: number | null | undefined, digits = 2, fallback = '—'): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return fallback
  }
  return value.toFixed(digits)
}

// 比率格式化：小于等于 1 视为 0-1 的小数，大于 1 视为已经是百分数。
export function formatRatio(value: number | null | undefined, digits = 2, fallback = '—'): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return fallback
  }
  const percent = Math.abs(value) <= 1 ? value * 100 : value
  return `${percent.toFixed(digits)}%`
}

// 由分子分母计算百分比，分母为 0 时返回占位符。
export function formatRatioOf(numerator: number | null | undefined, denominator: number | null | undefined, digits = 2, fallback = '—'): string {
  const top = toNumber(numerator)
  const bottom = toNumber(denominator)
  if (bottom <= 0) {
    return fallback
  }
  return formatRatio(top / bottom, digits, fallback)
}

// 时间格式化，无法解析时返回占位符。
export function formatDateTime(value: string | null | undefined, fallback = '—'): string {
  if (!value) {
    return fallback
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return fallback
  }
  return date.toLocaleString('zh-CN', { hour12: false })
}

// 仅日期格式化。
export function formatDate(value: string | null | undefined, fallback = '—'): string {
  if (!value) {
    return fallback
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return fallback
  }
  return date.toLocaleDateString('zh-CN')
}

// 百分比数值夹取，供 el-progress 使用。
export function clampPercent(value: number | null | undefined): number {
  const result = toNumber(value)
  return Math.min(100, Math.max(0, Math.round(result)))
}

// JSON 美化展示，用于审计日志详情等场景。
export function prettifyJson(value: unknown, fallback = '—'): string {
  if (value === null || value === undefined || value === '') {
    return fallback
  }
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
      return value
    }
  }
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return fallback
  }
}

// 表格动态列单元格格式化（CSV 预览等）。
export function formatCell(value: unknown): string {
  if (value === null || value === undefined) {
    return '—'
  }
  if (typeof value === 'object') {
    return prettifyJson(value)
  }
  return String(value)
}

// 当前时间（本地时区）字符串，用于「刷新时间」展示。
export function formatNow(): string {
  return new Date().toLocaleString('zh-CN', { hour12: false })
}
