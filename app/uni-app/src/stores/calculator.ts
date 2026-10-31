// 理论市场计算器：草稿（4 步表单）、计算结果与分享 token。
// 草稿持久化到本地，游客刷新页面不丢失条件。
import { defineStore } from 'pinia'
import * as calculatorApi from '@/api/calculator'
import type {
  CalculatorConditions,
  CalculatorRequest,
  CalculatorResult,
  CalculatorSelf
} from '@/api/types'
import { getAnonymousSessionId, getJson, setJson, STORAGE_KEYS } from '@/utils/storage'
import { toIntOrNull } from '@/utils/validate'

export type ConditionKey =
  | 'gender'
  | 'age'
  | 'city'
  | 'height'
  | 'education'
  | 'income'
  | 'marital_status'
  | 'children_state'
  | 'wants_children'

export type ConditionLevel = 'must' | 'prefer' | 'ignore'

export type ConditionLevels = Record<ConditionKey, ConditionLevel>

export interface CalculatorSelfDraft {
  gender: number | null
  age: string
  city_code: string
  height_cm: string
  education_level: number | null
  annual_income_band: number | null
}

export interface CalculatorConditionsDraft {
  target_genders: number[]
  min_age: string
  max_age: string
  allowed_city_codes: string[]
  min_height_cm: string
  max_height_cm: string
  min_education_level: number | null
  min_income_band: number | null
  allowed_marital_statuses: number[]
  allowed_children_states: number[]
  allowed_wants_children: number[]
}

export interface CalculatorDraft {
  self: CalculatorSelfDraft
  conditions: CalculatorConditionsDraft
  levels: ConditionLevels
}

export interface StepRow {
  key: ConditionKey
  label: string
  desc: string
}

// 条件的固定展示顺序，与后端 TheoreticalDimensions 一致
export const CONDITION_ROWS: StepRow[] = [
  { key: 'gender', label: '目标性别', desc: '对方的性别' },
  { key: 'age', label: '年龄', desc: '对方的年龄区间' },
  { key: 'city', label: '城市', desc: '对方所在城市' },
  { key: 'marital_status', label: '婚姻状态', desc: '未婚 / 离异 / 丧偶' },
  { key: 'children_state', label: '子女情况', desc: '是否已有子女' },
  { key: 'wants_children', label: '生育意愿', desc: '是否想要孩子' },
  { key: 'height', label: '身高', desc: '对方身高区间' },
  { key: 'education', label: '学历', desc: '对方最低学历' },
  { key: 'income', label: '年收入', desc: '对方最低年收入区间' }
]

export const CONDITION_LEVEL_LABELS: Record<ConditionLevel, string> = {
  must: '必须',
  prefer: '偏好',
  ignore: '不在意'
}

export function createDefaultDraft(): CalculatorDraft {
  return {
    self: {
      gender: null,
      age: '',
      city_code: '',
      height_cm: '',
      education_level: null,
      annual_income_band: null
    },
    conditions: {
      target_genders: [],
      min_age: '',
      max_age: '',
      allowed_city_codes: [],
      min_height_cm: '',
      max_height_cm: '',
      min_education_level: null,
      min_income_band: null,
      allowed_marital_statuses: [],
      allowed_children_states: [],
      allowed_wants_children: []
    },
    levels: {
      gender: 'must',
      age: 'prefer',
      city: 'prefer',
      marital_status: 'prefer',
      children_state: 'prefer',
      wants_children: 'prefer',
      height: 'prefer',
      education: 'prefer',
      income: 'prefer'
    }
  }
}

function normalizeDraft(raw: CalculatorDraft | null): CalculatorDraft {
  const fallback = createDefaultDraft()
  if (!raw || typeof raw !== 'object') return fallback
  return {
    self: Object.assign({}, fallback.self, raw.self || {}),
    conditions: Object.assign({}, fallback.conditions, raw.conditions || {}),
    levels: Object.assign({}, fallback.levels, raw.levels || {})
  }
}

interface CalculatorState {
  draft: CalculatorDraft
  result: CalculatorResult | null
  runId: string
  shareToken: string
  loading: boolean
}

interface StoredResult {
  result: CalculatorResult
  saved_at: number
}

export const useCalculatorStore = defineStore('calculator', {
  state: (): CalculatorState => {
    const stored = getJson<StoredResult | null>(STORAGE_KEYS.calculatorLastResult, null)
    return {
      draft: normalizeDraft(getJson<CalculatorDraft | null>(STORAGE_KEYS.calculatorDraft, null)),
      result: stored && stored.result ? stored.result : null,
      runId: stored && stored.result ? stored.result.run_id : '',
      shareToken: stored && stored.result && stored.result.share_token ? stored.result.share_token : '',
      loading: false
    }
  },
  getters: {
    mustCount: (state) => {
      const levels = state.draft.levels
      const keys = Object.keys(levels) as ConditionKey[]
      let count = 0
      for (let i = 0; i < keys.length; i++) {
        if (levels[keys[i]] === 'must') count++
      }
      return count
    }
  },
  actions: {
    persistDraft() {
      setJson(STORAGE_KEYS.calculatorDraft, this.draft)
    },
    resetDraft() {
      this.draft = createDefaultDraft()
      this.result = null
      this.runId = ''
      this.shareToken = ''
      setJson(STORAGE_KEYS.calculatorDraft, this.draft)
      setJson(STORAGE_KEYS.calculatorLastResult, null)
    },
    setSelf(patch: Partial<CalculatorSelfDraft>) {
      this.draft.self = Object.assign({}, this.draft.self, patch)
      this.persistDraft()
    },
    setConditions(patch: Partial<CalculatorConditionsDraft>) {
      this.draft.conditions = Object.assign({}, this.draft.conditions, patch)
      this.persistDraft()
    },
    setLevel(key: ConditionKey, level: ConditionLevel) {
      this.draft.levels = Object.assign({}, this.draft.levels, { [key]: level })
      this.persistDraft()
    },
    hasConditionValue(key: ConditionKey): boolean {
      const conditions = this.draft.conditions
      if (key === 'gender') return conditions.target_genders.length > 0
      if (key === 'age') return conditions.min_age !== '' || conditions.max_age !== ''
      if (key === 'city') return conditions.allowed_city_codes.length > 0
      if (key === 'height') return conditions.min_height_cm !== '' || conditions.max_height_cm !== ''
      if (key === 'education') return conditions.min_education_level !== null && conditions.min_education_level !== undefined && conditions.min_education_level > 0
      if (key === 'income') return conditions.min_income_band !== null && conditions.min_income_band !== undefined && conditions.min_income_band > 0
      if (key === 'marital_status') return conditions.allowed_marital_statuses.length > 0
      if (key === 'children_state') return conditions.allowed_children_states.length > 0
      if (key === 'wants_children') return conditions.allowed_wants_children.length > 0
      return false
    },
    buildConditions(): CalculatorConditions {
      const conditions = this.draft.conditions
      const levels = this.draft.levels
      const out: CalculatorConditions = {}
      if (levels.gender === 'must' && conditions.target_genders.length > 0) {
        out.target_genders = conditions.target_genders.slice()
      }
      if (levels.age === 'must') {
        const minAge = toIntOrNull(conditions.min_age)
        const maxAge = toIntOrNull(conditions.max_age)
        if (minAge !== null) out.min_age = minAge
        if (maxAge !== null) out.max_age = maxAge
      }
      if (levels.city === 'must' && conditions.allowed_city_codes.length > 0) {
        out.allowed_city_codes = conditions.allowed_city_codes.slice()
      }
      if (levels.height === 'must') {
        const minHeight = toIntOrNull(conditions.min_height_cm)
        const maxHeight = toIntOrNull(conditions.max_height_cm)
        if (minHeight !== null) out.min_height_cm = minHeight
        if (maxHeight !== null) out.max_height_cm = maxHeight
      }
      if (levels.education === 'must') {
        const education = conditions.min_education_level
        if (education !== null && education !== undefined && education > 0) out.min_education_level = education
      }
      if (levels.income === 'must') {
        const income = conditions.min_income_band
        if (income !== null && income !== undefined && income > 0) out.min_income_band = income
      }
      if (levels.marital_status === 'must' && conditions.allowed_marital_statuses.length > 0) {
        out.allowed_marital_statuses = conditions.allowed_marital_statuses.slice()
      }
      if (levels.children_state === 'must' && conditions.allowed_children_states.length > 0) {
        out.allowed_children_states = conditions.allowed_children_states.slice()
      }
      if (levels.wants_children === 'must' && conditions.allowed_wants_children.length > 0) {
        out.allowed_wants_children = conditions.allowed_wants_children.slice()
      }
      return out
    },
    buildRequest(): CalculatorRequest {
      const self: CalculatorSelf = {}
      const draftSelf = this.draft.self
      if (draftSelf.gender !== null && draftSelf.gender > 0) self.gender = draftSelf.gender
      const age = toIntOrNull(draftSelf.age)
      if (age !== null) self.age = age
      if (draftSelf.city_code) self.city_code = draftSelf.city_code
      const height = toIntOrNull(draftSelf.height_cm)
      if (height !== null) self.height_cm = height
      if (draftSelf.education_level !== null && draftSelf.education_level > 0) self.education_level = draftSelf.education_level
      if (draftSelf.annual_income_band !== null && draftSelf.annual_income_band > 0) {
        self.annual_income_band = draftSelf.annual_income_band
      }
      return {
        anonymous_session_id: getAnonymousSessionId(),
        self,
        conditions: this.buildConditions()
      }
    },
    async runEstimate(): Promise<CalculatorResult> {
      this.loading = true
      try {
        const data = await calculatorApi.estimate(this.buildRequest())
        this.result = data
        this.runId = data.run_id || ''
        this.shareToken = data.share_token || ''
        this.persistResult()
        return data
      } finally {
        this.loading = false
      }
    },
    persistResult() {
      if (!this.result) return
      const stored: StoredResult = { result: this.result, saved_at: Date.now() }
      setJson(STORAGE_KEYS.calculatorLastResult, stored)
    },
    loadLastResult(): CalculatorResult | null {
      if (this.result) return this.result
      const stored = getJson<StoredResult | null>(STORAGE_KEYS.calculatorLastResult, null)
      if (stored && stored.result) {
        this.result = stored.result
        this.runId = stored.result.run_id || ''
        this.shareToken = stored.result.share_token || ''
        return stored.result
      }
      return null
    },
    async ensureShareToken(): Promise<string> {
      if (this.shareToken) return this.shareToken
      if (!this.result || !this.result.run_id) throw new Error('请先完成一次计算')
      const data = await calculatorApi.createShare(this.result.run_id)
      this.shareToken = data.share_token
      if (this.result) {
        this.result = Object.assign({}, this.result, { share_token: data.share_token })
        this.persistResult()
      }
      return this.shareToken
    },
    async loadSharedResult(token: string): Promise<CalculatorResult> {
      return calculatorApi.getSharedResult(token)
    }
  }
})
