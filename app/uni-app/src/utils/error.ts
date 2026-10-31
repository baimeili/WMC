// 统一错误文案：页面只展示中文提示，不直接暴露后端内部信息。
import { ApiError } from '@/api/http'

export function errorMessage(error: unknown, fallback = '操作失败，请稍后重试'): string {
  if (error instanceof ApiError) {
    return error.message || fallback
  }
  if (error instanceof Error) {
    return error.message || fallback
  }
  return fallback
}

export function isUnauthorized(error: unknown): boolean {
  return error instanceof ApiError && error.code === 'UNAUTHORIZED'
}

export function isConflict(error: unknown): boolean {
  return error instanceof ApiError && (error.code === 'CONFLICT' || error.code === 'PROFILE_INCOMPLETE')
}

export function isFeatureDisabled(error: unknown): boolean {
  return error instanceof ApiError && error.code === 'FEATURE_DISABLED'
}

export function isNotFound(error: unknown): boolean {
  return error instanceof ApiError && error.code === 'NOT_FOUND'
}
