// 审计日志接口。
import { get, toArray } from './http'
import type { AuditLog, AuditLogQuery } from './types'

// 查询审计日志（limit/offset 分页，倒序）。
export async function listAuditLogs(query: AuditLogQuery = {}): Promise<AuditLog[]> {
  const payload = await get<AuditLog[] | { items?: AuditLog[] }>('/admin/audit-logs', {
    limit: query.limit ?? 50,
    offset: query.offset ?? 0
  })
  return toArray<AuditLog>(payload)
}
