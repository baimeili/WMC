// 本地存储封装：登录态、匿名会话、计算结果草稿与元数据缓存。
import type { OptionDict, Region, TokenPair } from '@/api/types'

const PREFIX = 'wmc:'

export const STORAGE_KEYS = {
  accessToken: PREFIX + 'auth:access_token',
  refreshToken: PREFIX + 'auth:refresh_token',
  authMeta: PREFIX + 'auth:meta',
  anonymousSessionId: PREFIX + 'anonymous_session_id',
  metaOptions: PREFIX + 'meta:options',
  metaRegions: PREFIX + 'meta:regions',
  metaAppConfig: PREFIX + 'meta:app_config',
  calculatorDraft: PREFIX + 'calculator:draft',
  calculatorLastResult: PREFIX + 'calculator:last_result',
  calculatorLastShare: PREFIX + 'calculator:last_share'
}

export interface AuthMeta {
  user_id: string
  role: number
  profile_completion: number
  discovery_enabled: boolean
  issued_at: number
}

export interface CachedData<T> {
  ts: number
  data: T
}

function safeGet(key: string): unknown {
  try {
    return uni.getStorageSync(key)
  } catch (error) {
    return null
  }
}

function safeSet(key: string, value: unknown): void {
  try {
    uni.setStorageSync(key, value)
  } catch (error) {
    // 本地存储失败不阻塞业务流程
  }
}

function safeRemove(key: string): void {
  try {
    uni.removeStorageSync(key)
  } catch (error) {
    // 忽略
  }
}

export function getString(key: string): string {
  const value = safeGet(key)
  if (value === null || value === undefined || value === '') return ''
  return String(value)
}

export function setString(key: string, value: string): void {
  safeSet(key, value)
}

export function removeValue(key: string): void {
  safeRemove(key)
}

export function getJson<T>(key: string, fallback: T): T {
  const value = safeGet(key)
  if (value === null || value === undefined || value === '') return fallback
  if (typeof value === 'object') return value as T
  try {
    return JSON.parse(String(value)) as T
  } catch (error) {
    return fallback
  }
}

export function setJson(key: string, value: unknown): void {
  safeSet(key, value)
}

export function getAccessToken(): string {
  return getString(STORAGE_KEYS.accessToken)
}

export function getRefreshToken(): string {
  return getString(STORAGE_KEYS.refreshToken)
}

export function setTokenPair(pair: TokenPair | null): void {
  if (!pair) {
    clearTokenPair()
    return
  }
  setString(STORAGE_KEYS.accessToken, pair.access_token || '')
  setString(STORAGE_KEYS.refreshToken, pair.refresh_token || '')
  setAuthMeta({
    user_id: pair.user_id || '',
    role: pair.role || 0,
    profile_completion: pair.profile_completion || 0,
    discovery_enabled: Boolean(pair.discovery_enabled),
    issued_at: Date.now()
  })
}

export function clearTokenPair(): void {
  safeRemove(STORAGE_KEYS.accessToken)
  safeRemove(STORAGE_KEYS.refreshToken)
  safeRemove(STORAGE_KEYS.authMeta)
}

export function getAuthMeta(): AuthMeta | null {
  return getJson<AuthMeta | null>(STORAGE_KEYS.authMeta, null)
}

export function setAuthMeta(meta: AuthMeta): void {
  setJson(STORAGE_KEYS.authMeta, meta)
}

// 生成 UUID v4，用于匿名会话与请求幂等标识
export function createUuid(): string {
  const globalCrypto = typeof crypto !== 'undefined' ? crypto : null
  if (globalCrypto && typeof globalCrypto.randomUUID === 'function') {
    return globalCrypto.randomUUID()
  }
  const chars = '0123456789abcdef'
  let out = ''
  for (let i = 0; i < 36; i++) {
    if (i === 8 || i === 13 || i === 18 || i === 23) {
      out += '-'
    } else if (i === 14) {
      out += '4'
    } else if (i === 19) {
      const r = Math.floor(Math.random() * 4) + 8
      out += chars.charAt(r)
    } else {
      out += chars.charAt(Math.floor(Math.random() * 16))
    }
  }
  return out
}

// 匿名会话 id：游客计算理论市场时用于关联登录后的账号
export function getAnonymousSessionId(): string {
  let id = getString(STORAGE_KEYS.anonymousSessionId)
  if (!id) {
    id = createUuid()
    setString(STORAGE_KEYS.anonymousSessionId, id)
  }
  return id
}

export function getCachedOptions(): CachedData<OptionDict> | null {
  return getJson<CachedData<OptionDict> | null>(STORAGE_KEYS.metaOptions, null)
}

export function setCachedOptions(data: OptionDict): void {
  setJson(STORAGE_KEYS.metaOptions, { ts: Date.now(), data })
}

export function getCachedRegions(): CachedData<Region[]> | null {
  return getJson<CachedData<Region[]> | null>(STORAGE_KEYS.metaRegions, null)
}

export function setCachedRegions(data: Region[]): void {
  setJson(STORAGE_KEYS.metaRegions, { ts: Date.now(), data })
}

export function getCachedAppConfig(): CachedData<Record<string, unknown>> | null {
  return getJson<CachedData<Record<string, unknown>> | null>(STORAGE_KEYS.metaAppConfig, null)
}

export function setCachedAppConfig(data: Record<string, unknown>): void {
  setJson(STORAGE_KEYS.metaAppConfig, { ts: Date.now(), data })
}

export function generateRequestId(): string {
  return createUuid()
}
