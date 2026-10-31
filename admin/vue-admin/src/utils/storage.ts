// 管理端登录态本地存储。
// 独立成模块，避免 pinia store 与 axios 实例互相引用造成循环依赖。
export const ACCESS_TOKEN_KEY = 'wmc_admin_access_token'
export const REFRESH_TOKEN_KEY = 'wmc_admin_refresh_token'
export const ROLE_KEY = 'wmc_admin_role'
export const USER_ID_KEY = 'wmc_admin_user_id'

export interface StoredAuth {
  accessToken: string
  refreshToken: string
  role: number
  userId: string
}

function safeGet(key: string): string {
  try {
    return localStorage.getItem(key) ?? ''
  } catch {
    return ''
  }
}

function safeSet(key: string, value: string): void {
  try {
    localStorage.setItem(key, value)
  } catch {
    // 隐私模式下 localStorage 可能不可写，登录态仅保留在内存中。
  }
}

function safeRemove(key: string): void {
  try {
    localStorage.removeItem(key)
  } catch {
    // 忽略清理失败，不影响后续跳转登录页。
  }
}

// 读取本地登录态，缺失字段回退为空值。
export function readStoredAuth(): StoredAuth {
  const role = Number.parseInt(safeGet(ROLE_KEY), 10)
  return {
    accessToken: safeGet(ACCESS_TOKEN_KEY),
    refreshToken: safeGet(REFRESH_TOKEN_KEY),
    role: Number.isFinite(role) ? role : 0,
    userId: safeGet(USER_ID_KEY)
  }
}

// 写入登录态。
export function writeStoredAuth(auth: StoredAuth): void {
  safeSet(ACCESS_TOKEN_KEY, auth.accessToken)
  safeSet(REFRESH_TOKEN_KEY, auth.refreshToken)
  safeSet(ROLE_KEY, String(auth.role))
  safeSet(USER_ID_KEY, auth.userId)
}

// 清空登录态。
export function clearStoredAuth(): void {
  safeRemove(ACCESS_TOKEN_KEY)
  safeRemove(REFRESH_TOKEN_KEY)
  safeRemove(ROLE_KEY)
  safeRemove(USER_ID_KEY)
}
