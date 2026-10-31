// 管理端鉴权状态：Token 持久化在 localStorage，角色决定能否进入运营后台。
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import type { TokenPair } from '@/api/types'
import { clearStoredAuth, readStoredAuth, writeStoredAuth } from '@/utils/storage'

// 允许进入运营后台的角色编码：2 内容审核员 / 3 管理员 / 4 超级管理员。
export const ADMIN_ROLES: number[] = [2, 3, 4]

export const useAuthStore = defineStore('auth', () => {
  const stored = readStoredAuth()

  const accessToken = ref(stored.accessToken)
  const refreshToken = ref(stored.refreshToken)
  const role = ref(stored.role)
  const userId = ref(stored.userId)
  const profileCompletion = ref(0)

  const isLoggedIn = computed(() => Boolean(accessToken.value))
  const canAccessAdmin = computed(() => ADMIN_ROLES.includes(role.value))

  // 保存登录响应中的令牌与角色信息。
  function applyTokenPair(pair: TokenPair): void {
    if (!pair.access_token || !pair.refresh_token) {
      throw new Error('登录响应缺少访问令牌，请稍后重试')
    }
    accessToken.value = pair.access_token
    refreshToken.value = pair.refresh_token
    role.value = pair.role ?? 0
    userId.value = pair.user_id ?? ''
    profileCompletion.value = pair.profile_completion ?? 0
    writeStoredAuth({
      accessToken: accessToken.value,
      refreshToken: refreshToken.value,
      role: role.value,
      userId: userId.value
    })
  }

  // 手机号 + 短信验证码登录。
  async function loginWithSms(phone: string, code: string): Promise<void> {
    const pair = await authApi.verifySmsCode(phone, code)
    applyTokenPair(pair)
  }

  // 管理端开发登录（生产环境不可用）。
  async function loginWithDev(phone: string): Promise<void> {
    const pair = await authApi.devLogin(phone)
    applyTokenPair(pair)
  }

  // 退出登录：先尽力吊销服务端令牌，再清理本地状态。
  async function logout(): Promise<void> {
    try {
      await authApi.logout()
    } catch {
      // 服务端吊销失败不阻塞本地退出。
    } finally {
      clear()
    }
  }

  // 清理本地登录态。
  function clear(): void {
    accessToken.value = ''
    refreshToken.value = ''
    role.value = 0
    userId.value = ''
    profileCompletion.value = 0
    clearStoredAuth()
  }

  return {
    accessToken,
    refreshToken,
    role,
    userId,
    profileCompletion,
    isLoggedIn,
    canAccessAdmin,
    applyTokenPair,
    loginWithSms,
    loginWithDev,
    logout,
    clear
  }
})
