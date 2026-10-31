// 图片审核接口。
// 说明：backend/openapi/openapi.yaml 尚未收录运营后台图片审核端点，
// 这里按运营后台约定路径实现（GET /admin/photos、POST /admin/photos/{id}/approve|reject）；
// 后端补齐后页面无需调整；接口不可用时页面会展示加载失败提示。
import { get, post } from './http'
import type { AdminPhoto, AdminPhotoPage } from './types'

// 照片列表，默认查询待审核（moderation_status = 1）。
export async function listPhotos(params: { status?: number; page?: number; page_size?: number } = {}): Promise<AdminPhotoPage> {
  const query: Record<string, unknown> = {}
  if (params.status !== undefined) {
    query.status = params.status
  }
  query.page = params.page ?? 1
  query.page_size = params.page_size ?? 20
  const payload = await get<AdminPhoto[] | AdminPhotoPage>('/admin/photos', query)
  if (Array.isArray(payload)) {
    return { items: payload, total: payload.length }
  }
  return payload ?? { items: [] }
}

// 审核通过。
export function approvePhoto(id: string): Promise<void> {
  return post<void>(`/admin/photos/${id}/approve`)
}

// 审核拒绝，需填写原因。
export function rejectPhoto(id: string, reason: string): Promise<void> {
  return post<void>(`/admin/photos/${id}/reject`, { reason })
}
