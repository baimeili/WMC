// 元数据接口：选项字典、行政区划、价值观问题、前端配置。
import { get } from './http'
import type { AppConfig, CompatibilityQuestion, OptionDict, Region } from './types'

export function getOptions(): Promise<OptionDict> {
  return get<OptionDict>('/meta/options')
}

export function getRegions(): Promise<Region[]> {
  return get<Region[]>('/meta/regions')
}

export function getCompatibilityQuestions(): Promise<CompatibilityQuestion[]> {
  return get<CompatibilityQuestion[]>('/meta/compatibility-questions')
}

export function getAppConfig(): Promise<AppConfig> {
  return get<AppConfig>('/meta/app-config')
}
