// 本站真实候选池：概览、真实漏斗、条件敏感性与候选人分页。
import { defineStore } from 'pinia'
import * as discoveryApi from '@/api/discovery'
import type {
  CandidateCardItem,
  DiscoverySummary,
  FunnelResult,
  SensitivityResult
} from '@/api/types'

interface DiscoveryState {
  summary: DiscoverySummary | null
  funnel: FunnelResult | null
  sensitivity: SensitivityResult | null
  candidates: CandidateCardItem[]
  nextCursor: string
  hasMore: boolean
  summaryLoading: boolean
  funnelLoading: boolean
  sensitivityLoading: boolean
  candidatesLoading: boolean
  error: string
}

export const useDiscoveryStore = defineStore('discovery', {
  state: (): DiscoveryState => ({
    summary: null,
    funnel: null,
    sensitivity: null,
    candidates: [],
    nextCursor: '',
    hasMore: true,
    summaryLoading: false,
    funnelLoading: false,
    sensitivityLoading: false,
    candidatesLoading: false,
    error: ''
  }),
  getters: {
    inMarket: (state) => (state.summary ? state.summary.in_market : false),
    // 概览返回值优先使用后端已做 K 匿名处理的文本
    eligibleText: (state) => (state.summary && state.summary.eligible_count_text ? state.summary.eligible_count_text : '')
  },
  actions: {
    async loadSummary(force = false): Promise<DiscoverySummary> {
      if (this.summary && !force) return this.summary
      this.summaryLoading = true
      this.error = ''
      try {
        const data = await discoveryApi.getSummary()
        this.summary = data
        return data
      } catch (error) {
        this.error = error instanceof Error ? error.message : '候选池加载失败'
        throw error
      } finally {
        this.summaryLoading = false
      }
    },
    async loadFunnel(force = false): Promise<FunnelResult> {
      if (this.funnel && !force) return this.funnel
      this.funnelLoading = true
      try {
        const data = await discoveryApi.getFunnel()
        this.funnel = data
        return data
      } finally {
        this.funnelLoading = false
      }
    },
    async loadSensitivity(force = false): Promise<SensitivityResult> {
      if (this.sensitivity && !force) return this.sensitivity
      this.sensitivityLoading = true
      try {
        const data = await discoveryApi.getSensitivity()
        this.sensitivity = data
        return data
      } finally {
        this.sensitivityLoading = false
      }
    },
    async loadCandidates(reset = false): Promise<CandidateCardItem[]> {
      if (this.candidatesLoading) return this.candidates
      if (reset) {
        this.candidates = []
        this.nextCursor = ''
        this.hasMore = true
      }
      if (!this.hasMore) return this.candidates
      this.candidatesLoading = true
      try {
        const page = await discoveryApi.getCandidates({
          cursor: this.nextCursor || undefined,
          limit: 20
        })
        const items = page.items || []
        this.candidates = reset ? items : this.candidates.concat(items)
        this.nextCursor = page.next_cursor || ''
        this.hasMore = Boolean(page.has_more)
        return this.candidates
      } finally {
        this.candidatesLoading = false
      }
    },
    clearCandidates() {
      this.candidates = []
      this.nextCursor = ''
      this.hasMore = true
    }
  }
})
