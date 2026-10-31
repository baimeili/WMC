// 产品事件上报：失败不影响业务流程。
import { getAnonymousSessionId } from '@/utils/storage'
import { post } from './http'

export function trackEvent(eventName: string, properties?: Record<string, unknown>): void {
  const payload = {
    event_name: eventName,
    anonymous_id: getAnonymousSessionId(),
    properties: properties || {}
  }
  post<void>('/events', payload, false).catch(() => {
    // 事件上报失败直接忽略
  })
}
