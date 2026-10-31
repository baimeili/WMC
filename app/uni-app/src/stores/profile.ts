// 我的资料、择偶条件与价值观规则。
import { defineStore } from 'pinia'
import * as compatibilityApi from '@/api/compatibility'
import * as preferenceApi from '@/api/preference'
import * as profileApi from '@/api/profile'
import type {
  CompatibilityRule,
  CompatibilityRuleInput,
  Preferences,
  PreferencesInput,
  Profile,
  ProfileInput
} from '@/api/types'
import { useAuthStore } from './auth'

interface ProfileState {
  profile: Profile | null
  preferences: Preferences | null
  rules: CompatibilityRule[]
  loading: boolean
}

// 资料对象转提交结构（剔除服务端计算字段）
export function toProfileInput(profile: Profile): ProfileInput {
  return {
    nickname: profile.nickname,
    gender: profile.gender,
    city_code: profile.city_code,
    birth_date: profile.birth_date,
    province_code: profile.province_code,
    district_code: profile.district_code,
    height_cm: profile.height_cm,
    education_level: profile.education_level,
    occupation_category: profile.occupation_category,
    employment_status: profile.employment_status,
    annual_income_band: profile.annual_income_band,
    family_asset_band: profile.family_asset_band,
    debt_band: profile.debt_band,
    house_status: profile.house_status,
    vehicle_status: profile.vehicle_status,
    marital_status: profile.marital_status,
    children_count: profile.children_count,
    wants_children: profile.wants_children,
    marriage_intent: profile.marriage_intent,
    marriage_timeline: profile.marriage_timeline,
    smoking_status: profile.smoking_status,
    drinking_status: profile.drinking_status,
    bio: profile.bio
  }
}

function mergeProfileInput(base: ProfileInput, patch: Partial<ProfileInput>): ProfileInput {
  const target = Object.assign({}, base) as Record<string, unknown>
  const keys = Object.keys(patch) as Array<keyof ProfileInput>
  for (let i = 0; i < keys.length; i++) {
    const key = keys[i]
    const value = patch[key]
    if (value !== undefined) target[String(key)] = value
  }
  return target as unknown as ProfileInput
}

export const useProfileStore = defineStore('profile', {
  state: (): ProfileState => ({
    profile: null,
    preferences: null,
    rules: [],
    loading: false
  }),
  getters: {
    completion: (state) => (state.profile ? state.profile.profile_completion : 0),
    nickname: (state) => (state.profile ? state.profile.nickname : ''),
    cityCode: (state) => (state.profile ? state.profile.city_code : '')
  },
  actions: {
    async loadProfile(force = false): Promise<Profile> {
      if (this.profile && !force) return this.profile
      this.loading = true
      try {
        const data = await profileApi.getProfile()
        this.profile = data
        useAuthStore().setProfileCompletion(data.profile_completion || 0)
        return data
      } finally {
        this.loading = false
      }
    },
    async updateProfile(patch: Partial<ProfileInput>): Promise<Profile> {
      let base: ProfileInput
      if (this.profile) {
        base = toProfileInput(this.profile)
      } else {
        const current = await this.loadProfile()
        base = toProfileInput(current)
      }
      const payload = mergeProfileInput(base, patch)
      if (!payload.nickname || !payload.gender || !payload.city_code) {
        throw new Error('请先填写昵称、性别和所在城市')
      }
      const data = await profileApi.updateProfile(payload)
      this.profile = data
      useAuthStore().setProfileCompletion(data.profile_completion || 0)
      return data
    },
    async loadPreferences(force = false): Promise<Preferences> {
      if (this.preferences && !force) return this.preferences
      const data = await preferenceApi.getPreferences()
      this.preferences = data
      return data
    },
    async savePreferences(input: PreferencesInput): Promise<Preferences> {
      const data = await preferenceApi.updatePreferences(input)
      this.preferences = data
      return data
    },
    async loadRules(force = false): Promise<CompatibilityRule[]> {
      if (this.rules.length > 0 && !force) return this.rules
      const data = await compatibilityApi.getRules()
      this.rules = data || []
      return this.rules
    },
    async saveRules(rules: CompatibilityRuleInput[]): Promise<void> {
      await compatibilityApi.updateRules(rules)
      await this.loadRules(true)
    },
    async loadAll(force = false): Promise<void> {
      this.loading = true
      try {
        await this.loadProfile(force)
        await this.loadPreferences(force)
        await this.loadRules(force)
      } finally {
        this.loading = false
      }
    }
  }
})
