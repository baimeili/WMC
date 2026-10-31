// 举报列表与处理接口。
import { get, post, toArray } from './http'
import type { Report, ResolveReportInput } from './types'

// 举报列表，可按状态筛选（1 待处理 / 2 处理中 / 3 已处理 / 4 已驳回）。
export async function listReports(status?: number): Promise<Report[]> {
  const params: Record<string, unknown> = {}
  if (status !== undefined) {
    params.status = status
  }
  const payload = await get<Report[] | { items?: Report[] }>('/admin/reports', params)
  return toArray<Report>(payload)
}

// 查询某个被举报用户的举报记录（接口无按用户筛选参数，前端过滤）。
export async function listReportsByUser(userId: string): Promise<Report[]> {
  const reports = await listReports()
  return reports.filter((report) => report.reported_user_id === userId)
}

// 处理举报：状态 2 处理中 / 3 已处理 / 4 已驳回，备注写入审计。
export function resolveReport(id: string, input: ResolveReportInput): Promise<void> {
  return post<void>(`/admin/reports/${id}/resolve`, input)
}
