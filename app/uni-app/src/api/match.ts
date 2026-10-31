// Match 列表与详情接口。
import { get } from './http'
import type { MatchDetail, MatchPage } from './types'

export function getMatches(params?: { cursor?: string; limit?: number }): Promise<MatchPage> {
  return get<MatchPage>('/matches', params as Record<string, unknown>)
}

export function getMatch(matchId: string): Promise<MatchDetail> {
  return get<MatchDetail>('/matches/' + encodeURIComponent(matchId))
}
