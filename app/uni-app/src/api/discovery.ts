// 本站真实候选池接口。
import { get } from './http'
import type { CandidateDetail, CandidatePage, DiscoverySummary, FunnelResult, SensitivityResult } from './types'

export function getSummary(): Promise<DiscoverySummary> {
  return get<DiscoverySummary>('/discovery/summary')
}

export function getFunnel(): Promise<FunnelResult> {
  return get<FunnelResult>('/discovery/funnel')
}

export function getSensitivity(): Promise<SensitivityResult> {
  return get<SensitivityResult>('/discovery/sensitivity')
}

export function getCandidates(params?: { cursor?: string; limit?: number }): Promise<CandidatePage> {
  return get<CandidatePage>('/discovery/candidates', params as Record<string, unknown>)
}

export function getCandidate(userId: string): Promise<CandidateDetail> {
  return get<CandidateDetail>('/discovery/candidates/' + encodeURIComponent(userId))
}
