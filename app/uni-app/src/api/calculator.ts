// 理论婚恋市场计算器接口（游客可用）。
import { get, post, request } from './http'
import type { CalculatorRequest, CalculatorResult } from './types'

export function estimate(payload: CalculatorRequest): Promise<CalculatorResult> {
  // 游客请求不携带 Token，避免失效 Token 触发刷新
  return request<CalculatorResult>({ url: '/calculator/estimate', method: 'POST', data: payload, auth: false })
}

export function createShare(runId: string): Promise<{ share_token: string }> {
  return post<{ share_token: string }>('/calculator/share', { run_id: runId })
}

export function getSharedResult(token: string): Promise<CalculatorResult> {
  return get<CalculatorResult>('/calculator/share/' + encodeURIComponent(token))
}
