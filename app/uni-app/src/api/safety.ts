// 屏蔽与举报接口。
import { del, post } from './http'
import type { ReportPayload } from './types'

export function blockUser(userId: string): Promise<void> {
  return post<void>('/users/' + encodeURIComponent(userId) + '/block')
}

export function unblockUser(userId: string): Promise<void> {
  return del<void>('/users/' + encodeURIComponent(userId) + '/block')
}

export function reportUser(userId: string, payload: ReportPayload): Promise<void> {
  return post<void>('/users/' + encodeURIComponent(userId) + '/report', payload)
}
