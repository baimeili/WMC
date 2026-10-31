// 婚恋价值观规则接口。
import { get, put } from './http'
import type { CompatibilityRule, CompatibilityRuleInput } from './types'

export function getRules(): Promise<CompatibilityRule[]> {
  return get<CompatibilityRule[]>('/me/compatibility')
}

export function updateRules(rules: CompatibilityRuleInput[]): Promise<void> {
  return put<void>('/me/compatibility', { rules })
}
