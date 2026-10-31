// 用户查询与冻结/解冻接口。
import { get, post } from './http'
import type { AdminUserDetail, AdminUserPage, AdminUserQuery } from './types'

// 分页查询用户。
export function listUsers(query: AdminUserQuery): Promise<AdminUserPage> {
  const params: Record<string, unknown> = {}
  if (query.keyword) {
    params.keyword = query.keyword
  }
  if (query.status !== undefined && !Number.isNaN(query.status)) {
    params.status = query.status
  }
  params.page = query.page ?? 1
  params.page_size = query.page_size ?? 20
  return get<AdminUserPage>('/admin/users', params)
}

// 冻结用户，必须填写原因（后端写审计日志）。
export function freezeUser(id: string, reason: string): Promise<void> {
  return post<void>(`/admin/users/${id}/freeze`, { reason })
}

// 解冻用户。
export function unfreezeUser(id: string): Promise<void> {
  return post<void>(`/admin/users/${id}/unfreeze`)
}

// 用户详情：契约未定义单用户端点，优先探测 /admin/users/{id}。
export function getUserDetail(id: string): Promise<AdminUserDetail> {
  return get<AdminUserDetail>(`/admin/users/${id}`, undefined, { silent: true })
}

// 详情端点不可用时的兜底：按关键字检索列表并匹配用户 id。
export async function findUserInList(id: string): Promise<AdminUserDetail | undefined> {
  const page = await listUsers({ keyword: id, page: 1, page_size: 50 })
  const matched = (page.items ?? []).find((item) => item.id === id)
  return matched as AdminUserDetail | undefined
}
