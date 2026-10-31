// 系统参数与 Feature Flag 接口。
import { get, put } from './http'
import type { SettingsBundle } from './types'

// 读取参数与开关。
export function getSettings(): Promise<SettingsBundle> {
  return get<SettingsBundle>('/admin/settings')
}

// 按 key 更新单个参数，value 支持 boolean/number/string。
export function updateSetting(key: string, value: unknown): Promise<void> {
  return put<void>('/admin/settings', { key, value })
}
