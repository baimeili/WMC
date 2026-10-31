// 运营看板与核心经营漏斗接口。
import { get } from './http'
import type { AdminDashboard, FunnelStep } from './types'

// 运营看板（含 North Star 指标）。
export function getDashboard(): Promise<AdminDashboard> {
  return get<AdminDashboard>('/admin/dashboard')
}

// 核心经营漏斗：契约中由看板接口的 funnel 字段下发。
// 若后端未来提供独立漏斗端点，只需在此处替换数据来源，页面无需调整。
export async function getFunnel(): Promise<FunnelStep[]> {
  const dashboard = await getDashboard()
  return dashboard.funnel ?? []
}
