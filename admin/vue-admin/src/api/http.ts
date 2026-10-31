// axios 实例：统一鉴权、统一解包 envelope、统一错误提示。
// 所有业务 api 模块都必须通过本文件发起请求，禁止在页面里直接调用 axios。
import axios, { type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { clearStoredAuth, readStoredAuth } from '@/utils/storage'
import type { ApiEnvelope } from './types'

// 允许业务方法传入 silent，跳过全局 ElMessage 提示（用于可选接口探测）。
declare module 'axios' {
  export interface AxiosRequestConfig {
    silent?: boolean
  }
}

// 业务 API 前缀：开发环境由 vite proxy 转发到后端 8080 端口。
export const API_BASE_URL = '/api/v1'

// 成功响应码集合：契约使用 "OK"，同时兼容数值型成功码。
const SUCCESS_CODES = new Set(['OK', '0', '200'])

// 统一业务异常，页面可据此区分 404 等场景。
export class ApiError extends Error {
  readonly code: string
  readonly status?: number
  readonly requestId?: string

  constructor(code: string, message: string, status?: number, requestId?: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.requestId = requestId
  }
}

// 判断是否为「资源不存在」，用于契约未覆盖接口的降级处理。
export function isNotFound(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 404 || error.code === 'NOT_FOUND')
}

function defaultErrorMessage(status?: number): string {
  switch (status) {
    case 400:
      return '请求参数有误'
    case 401:
      return '登录已过期，请重新登录'
    case 403:
      return '没有权限执行该操作'
    case 404:
      return '接口或资源不存在'
    case 409:
      return '当前状态冲突，请刷新后重试'
    case 429:
      return '操作过于频繁，请稍后再试'
    case 500:
      return '服务端异常，请稍后重试'
    default:
      return '网络异常，请稍后重试'
  }
}

const http = axios.create({
  baseURL: API_BASE_URL,
  timeout: 20000
})

// 请求拦截：附加 Authorization: Bearer <access_token>。
http.interceptors.request.use((config) => {
  const { accessToken } = readStoredAuth()
  if (accessToken) {
    config.headers.set('Authorization', `Bearer ${accessToken}`)
  }
  return config
})

// 401 后跳转登录页，避免并发请求重复跳转。
let redirectingToLogin = false

async function redirectToLogin(): Promise<void> {
  if (redirectingToLogin) {
    return
  }
  redirectingToLogin = true
  try {
    if (!window.location.pathname.endsWith('/login')) {
      const routerModule = await import('@/router')
      await routerModule.default.replace('/login')
    }
  } catch {
    window.location.href = `${import.meta.env.BASE_URL}login`
  } finally {
    window.setTimeout(() => {
      redirectingToLogin = false
    }, 1000)
  }
}

// 响应拦截：校验 envelope；失败统一抛出 ApiError 并提示。
http.interceptors.response.use(
  (response) => {
    const envelope = response.data as Partial<ApiEnvelope<unknown>> | null | undefined
    const silent = Boolean(response.config.silent)
    if (!envelope || typeof envelope !== 'object' || typeof envelope.code !== 'string') {
      const error = new ApiError('INTERNAL', '服务返回格式异常，请稍后重试', response.status)
      if (!silent) {
        ElMessage.error(error.message)
      }
      return Promise.reject(error)
    }
    if (!SUCCESS_CODES.has(envelope.code)) {
      const error = new ApiError(envelope.code, envelope.message || '请求失败', response.status, envelope.request_id)
      if (!silent) {
        ElMessage.error(error.message)
      }
      return Promise.reject(error)
    }
    return response
  },
  (error: unknown) => {
    if (axios.isAxiosError(error)) {
      const status = error.response?.status
      const envelope = error.response?.data as Partial<ApiEnvelope<unknown>> | undefined
      const code =
        envelope?.code ??
        (status === 401
          ? 'UNAUTHORIZED'
          : status === 403
            ? 'FORBIDDEN'
            : status === 404
              ? 'NOT_FOUND'
              : status === 429
                ? 'RATE_LIMITED'
                : 'INTERNAL')
      const message = envelope?.message || defaultErrorMessage(status)
      const apiError = new ApiError(code, message, status, envelope?.request_id)
      const silent = Boolean(error.config?.silent)
      const requestUrl = String(error.config?.url ?? '')
      const isAuthRequest = requestUrl.includes('/auth/')
      if (status === 401 && !isAuthRequest) {
        clearStoredAuth()
        void redirectToLogin()
      }
      if (!silent) {
        ElMessage.error(message)
      }
      return Promise.reject(apiError)
    }
    if (error instanceof ApiError) {
      return Promise.reject(error)
    }
    const unknownError = new ApiError('INTERNAL', '网络异常，请稍后重试')
    ElMessage.error(unknownError.message)
    return Promise.reject(unknownError)
  }
)

// 统一解包 envelope：返回 data 字段。
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await http.request<ApiEnvelope<T>>(config)
  const body = response.data
  if (body && typeof body === 'object' && 'data' in body) {
    return body.data as T
  }
  return undefined as unknown as T
}

// GET 请求。
export function get<T>(url: string, params?: Record<string, unknown>, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'GET', url, params })
}

// POST 请求。
export function post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'POST', url, data })
}

// PUT 请求。
export function put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'PUT', url, data })
}

// DELETE 请求。
export function del<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'DELETE', url })
}

// 文件上传（multipart/form-data）：axios 会自动补 boundary，禁止手写 Content-Type。
export function upload<T>(url: string, form: FormData, config?: AxiosRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'POST', url, data: form, timeout: 120000 })
}

// 兼容后端返回数组或 { items: [] } 两种形态。
export function toArray<T>(payload: T[] | { items?: T[] } | null | undefined): T[] {
  if (Array.isArray(payload)) {
    return payload
  }
  if (payload && Array.isArray(payload.items)) {
    return payload.items
  }
  return []
}

export default http
