// 登录态：Token 持久化到 uni storage，页面通过该 store 判断登录与引导登录。
import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import type { TokenPair } from '@/api/types'
import {
  clearTokenPair,
  getAccessToken,
  getAnonymousSessionId,
  getAuthMeta,
  setAuthMeta,
  setTokenPair
} from '@/utils/storage'

interface AuthState {
  userId: string
  role: number
  profileCompletion: number
  discoveryEnabled: boolean
  loggedIn: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => {
    const meta = getAuthMeta()
    const hasToken = Boolean(getAccessToken())
    const valid = hasToken && Boolean(meta && meta.user_id)
    return {
      userId: valid && meta ? meta.user_id : '',
      role: valid && meta ? meta.role : 0,
      profileCompletion: valid && meta ? meta.profile_completion : 0,
      discoveryEnabled: valid && meta ? meta.discovery_enabled : false,
      loggedIn: valid
    }
  },
  getters: {
    isLoggedIn: (state) => state.loggedIn,
    completion: (state) => state.profileCompletion,
    canAccessDiscovery: (state) => state.loggedIn && state.profileCompletion > 0
  },
  actions: {
    applyTokenPair(pair: TokenPair) {
      setTokenPair(pair)
      this.userId = pair.user_id || ''
      this.role = pair.role || 0
      this.profileCompletion = pair.profile_completion || 0
      this.discoveryEnabled = Boolean(pair.discovery_enabled)
      this.loggedIn = Boolean(pair.access_token && pair.user_id)
    },
    // HTTP 层刷新或清理 Token 后，同步 store 状态
    syncFromStorage() {
      const meta = getAuthMeta()
      const hasToken = Boolean(getAccessToken())
      const valid = hasToken && Boolean(meta && meta.user_id)
      this.loggedIn = valid
      this.userId = valid && meta ? meta.user_id : ''
      this.role = valid && meta ? meta.role : 0
      this.profileCompletion = valid && meta ? meta.profile_completion : 0
      this.discoveryEnabled = valid && meta ? meta.discovery_enabled : false
    },
    setProfileCompletion(value: number) {
      const safe = Math.max(0, Math.min(100, Math.round(value || 0)))
      this.profileCompletion = safe
      const meta = getAuthMeta()
      if (meta) {
        setAuthMeta({
          user_id: meta.user_id,
          role: meta.role,
          profile_completion: safe,
          discovery_enabled: meta.discovery_enabled,
          issued_at: meta.issued_at
        })
      }
    },
    async sendCode(phone: string): Promise<void> {
      await authApi.sendSmsCode(phone)
    },
    // 登录时携带匿名会话 id，便于把游客计算结果关联到账号
    async loginBySms(phone: string, code: string): Promise<TokenPair> {
      const pair = await authApi.verifySmsCode({
        phone,
        code,
        anonymous_session_id: getAnonymousSessionId()
      })
      this.applyTokenPair(pair)
      return pair
    },
    async logout(): Promise<void> {
      try {
        await authApi.logout()
      } catch (error) {
        // 退出登录失败也要清理本地登录态
      }
      this.clearLocal()
    },
    clearLocal() {
      clearTokenPair()
      this.userId = ''
      this.role = 0
      this.profileCompletion = 0
      this.discoveryEnabled = false
      this.loggedIn = false
    },
    // 需要登录的页面调用：未登录时引导到登录页，并带回跳地址
    requireLogin(redirect?: string): boolean {
      if (this.loggedIn) return true
      const query = redirect ? '?redirect=' + encodeURIComponent(redirect) : ''
      uni.navigateTo({ url: '/pages/auth/login' + query })
      return false
    }
  }
})
