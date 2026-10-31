// 展示格式化：百分比、人数匿名化、时间、分数与文本截断。
import { K_ANONYMITY_THRESHOLD } from './constants'

export function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && isFinite(value)
}

// 0~1 的比例转百分比文案
export function formatPercent(ratio: number | null | undefined, digits = 1): string {
  if (!isFiniteNumber(ratio)) return '—'
  return (ratio * 100).toFixed(digits) + '%'
}

// 0~100 的百分比数值转文案
export function formatPercentValue(percent: number | null | undefined, digits = 1): string {
  if (!isFiniteNumber(percent)) return '—'
  return percent.toFixed(digits) + '%'
}

// 变化率：0.52 展示为 +52.0%
export function formatChange(delta: number | null | undefined, digits = 1): string {
  if (!isFiniteNumber(delta)) return '—'
  const percent = delta * 100
  if (Math.abs(percent) < 0.05) return '0%'
  const sign = percent > 0 ? '+' : ''
  return sign + percent.toFixed(digits) + '%'
}

// 整数千分位
export function formatInteger(value: number | null | undefined): string {
  if (!isFiniteNumber(value)) return '—'
  const rounded = Math.round(value)
  const negative = rounded < 0
  let digits = String(Math.abs(rounded))
  let out = ''
  while (digits.length > 3) {
    out = ',' + digits.slice(digits.length - 3) + out
    digits = digits.slice(0, digits.length - 3)
  }
  return (negative ? '-' : '') + digits + out
}

// 真实市场人数匿名化：小于 K 一律显示 <10，绝不显示 0 个人
export function maskCount(count: number | null | undefined, threshold = K_ANONYMITY_THRESHOLD): string {
  if (!isFiniteNumber(count)) return '—'
  if (count < threshold) return '<' + threshold
  return formatInteger(count)
}

export function peopleText(count: number | null | undefined, threshold = K_ANONYMITY_THRESHOLD): string {
  const text = maskCount(count, threshold)
  if (text === '—') return text
  return text + ' 人'
}

// 0~1 的分数转百分比
export function formatScore(score: number | null | undefined): string {
  if (!isFiniteNumber(score)) return '—'
  return (score * 100).toFixed(1) + '%'
}

export function ageText(age: number | null | undefined): string {
  if (!isFiniteNumber(age) || age <= 0) return '年龄未填写'
  return age + ' 岁'
}

export function heightText(height: number | null | undefined): string {
  if (!isFiniteNumber(height) || height <= 0) return '身高未填写'
  return height + ' cm'
}

export function truncate(text: string | null | undefined, length = 60): string {
  if (!text) return ''
  if (text.length <= length) return text
  return text.slice(0, length) + '…'
}

function pad2(value: number): string {
  return value < 10 ? '0' + value : String(value)
}

export function formatDate(value: string | null | undefined): string {
  if (!value) return '—'
  const date = new Date(value)
  if (isNaN(date.getTime())) return '—'
  return date.getFullYear() + '-' + pad2(date.getMonth() + 1) + '-' + pad2(date.getDate())
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return '—'
  const date = new Date(value)
  if (isNaN(date.getTime())) return '—'
  return (
    date.getFullYear() +
    '-' +
    pad2(date.getMonth() + 1) +
    '-' +
    pad2(date.getDate()) +
    ' ' +
    pad2(date.getHours()) +
    ':' +
    pad2(date.getMinutes())
  )
}

// 变化方向，供样式区分
export function changeTone(delta: number | null | undefined): string {
  if (!isFiniteNumber(delta)) return 'flat'
  if (delta > 0.0005) return 'up'
  if (delta < -0.0005) return 'down'
  return 'flat'
}
