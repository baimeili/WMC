// Like / Pass 与真实结果反馈接口。
import { post, put } from './http'
import type { InteractionResult, OutcomePayload } from './types'

// action: 1 喜欢，2 跳过；source: 1 卡片，2 详情，3 Match
export function setInteraction(userId: string, action: number, source = 2): Promise<InteractionResult> {
  return put<InteractionResult>('/interactions/' + encodeURIComponent(userId), { action, source })
}

export function reportOutcome(matchId: string, payload: OutcomePayload): Promise<void> {
  return post<void>('/matches/' + encodeURIComponent(matchId) + '/outcomes', payload)
}
