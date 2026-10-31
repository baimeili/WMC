// HTTP 请求封装：
// 1. uni.request 统一出口，页面禁止直接调用；
// 2. 注入 Authorization Bearer Token；
// 3. 401 时用 refresh token 重试一次，失败则清理登录态；
// 4. 解包统一响应 { code, message, data, request_id }。
import type { Envelope, TokenPair } from './types'
import { API_BASE_URL } from '@/utils/constants'
import { clearTokenPair, getAccessToken, getRefreshToken, setTokenPair } from '@/utils/storage'

export class ApiError extends Error {
  code: string
  status: number
  requestId: string

  constructor(code: string, message: string, status = 0, requestId = '') {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.requestId = requestId
  }
}

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

export interface RequestOptions {
  url: string
  method?: HttpMethod
  data?: unknown
  header?: Record<string, string>
  auth?: boolean
}

interface RawResponse {
  statusCode: number
  body: unknown
}

const COMMON_HEADERS: Record<string, string> = {
  'Content-Type': 'application/json'
}

function statusToCode(statusCode: number): string {
  if (statusCode === 400) return 'INVALID_ARGUMENT'
  if (statusCode === 401) return 'UNAUTHORIZED'
  if (statusCode === 403) return 'FORBIDDEN'
  if (statusCode === 404) return 'NOT_FOUND'
  if (statusCode === 409) return 'CONFLICT'
  if (statusCode === 429) return 'RATE_LIMITED'
  if (statusCode >= 500) return 'INTERNAL'
  return 'INTERNAL'
}

function statusToMessage(statusCode: number): string {
  if (statusCode === 400) return '提交的信息有误，请检查后重试'
  if (statusCode === 401) return '登录状态已失效，请重新登录'
  if (statusCode === 403) return '当前账号暂无权限执行该操作'
  if (statusCode === 404) return '请求的内容不存在'
  if (statusCode === 409) return '当前状态不允许该操作'
  if (statusCode === 429) return '操作过于频繁，请稍后再试'
  return '服务暂时不可用，请稍后重试'
}

function rawRequest(options: {
  url: string
  method: HttpMethod
  data?: unknown
  header: Record<string, string>
}): Promise<RawResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: options.url,
      method: options.method,
      data: options.data as Record<string, unknown> | string | ArrayBuffer | undefined,
      header: options.header,
      timeout: 15000,
      success: (response) => {
        resolve({ statusCode: response.statusCode || 0, body: response.data })
      },
      fail: () => {
        reject(new ApiError('NETWORK_ERROR', '网络连接失败，请检查网络后重试', 0, ''))
      }
    })
  })
}

export function parseEnvelope<T>(statusCode: number, body: unknown): T {
  const envelope = body && typeof body === 'object' ? (body as Envelope<T>) : null
  if (envelope && typeof envelope.code === 'string') {
    if (envelope.code === 'OK') {
      return envelope.data as T
    }
    throw new ApiError(envelope.code, envelope.message || statusToMessage(statusCode), statusCode, envelope.request_id || '')
  }
  if (statusCode >= 200 && statusCode < 300) {
    return body as T
  }
  throw new ApiError(statusToCode(statusCode), statusToMessage(statusCode), statusCode, '')
}

export function authHeader(): Record<string, string> {
  const token = getAccessToken()
  if (!token) return {}
  return { Authorization: 'Bearer ' + token }
}

let refreshing: Promise<boolean> | null = null

function doRefresh(): Promise<boolean> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return Promise.resolve(false)
  return rawRequest({
    url: API_BASE_URL + '/auth/refresh',
    method: 'POST',
    data: { refresh_token: refreshToken },
    header: COMMON_HEADERS
  })
    .then((response) => {
      if (response.statusCode >= 200 && response.statusCode < 300) {
        try {
          const pair = parseEnvelope<TokenPair>(response.statusCode, response.body)
          if (pair && pair.access_token) {
            setTokenPair(pair)
            return true
          }
        } catch (error) {
          return false
        }
      }
      clearTokenPair()
      return false
    })
    .catch(() => {
      clearTokenPair()
      return false
    })
}

// 多个请求同时 401 时共享同一次 refresh
function ensureRefreshed(): Promise<boolean> {
  if (!refreshing) {
    refreshing = doRefresh().then((ok) => {
      refreshing = null
      return ok
    })
  }
  return refreshing
}

export async function request<T>(options: RequestOptions): Promise<T> {
  const useAuth = options.auth !== false
  const headers: Record<string, string> = Object.assign({}, COMMON_HEADERS, options.header || {})
  if (useAuth) {
    const auth = authHeader()
    if (auth.Authorization) headers.Authorization = auth.Authorization
  }
  const response = await rawRequest({
    url: API_BASE_URL + options.url,
    method: options.method || 'GET',
    data: options.data,
    header: headers
  })
  if (response.statusCode === 401 && useAuth) {
    const refreshed = await ensureRefreshed()
    if (refreshed) {
      const retryHeaders = Object.assign({}, headers, authHeader())
      const retried = await rawRequest({
        url: API_BASE_URL + options.url,
        method: options.method || 'GET',
        data: options.data,
        header: retryHeaders
      })
      return parseEnvelope<T>(retried.statusCode, retried.body)
    }
    clearTokenPair()
    throw new ApiError('UNAUTHORIZED', '登录状态已失效，请重新登录', 401, '')
  }
  return parseEnvelope<T>(response.statusCode, response.body)
}

export function withQuery(url: string, params?: Record<string, unknown>): string {
  if (!params) return url
  const parts: string[] = []
  const keys = Object.keys(params)
  for (let i = 0; i < keys.length; i++) {
    const value = params[keys[i]]
    if (value === undefined || value === null || value === '') continue
    parts.push(encodeURIComponent(keys[i]) + '=' + encodeURIComponent(String(value)))
  }
  if (parts.length === 0) return url
  return url + (url.indexOf('?') >= 0 ? '&' : '?') + parts.join('&')
}

export function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  return request<T>({ url: withQuery(url, params), method: 'GET' })
}

export function post<T>(url: string, data?: unknown, auth = true): Promise<T> {
  return request<T>({ url, method: 'POST', data, auth })
}

export function put<T>(url: string, data?: unknown): Promise<T> {
  return request<T>({ url, method: 'PUT', data })
}

export function del<T>(url: string, data?: unknown): Promise<T> {
  return request<T>({ url, method: 'DELETE', data })
}
