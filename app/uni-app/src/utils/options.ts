// 选项字典与行政区划：统一从 GET /meta/options、/meta/regions 获取，避免两端文案分裂。
import { reactive } from 'vue'
import { getAppConfig, getOptions, getRegions } from '@/api/meta'
import type { AppConfig, OptionDict, OptionItem, Region } from '@/api/types'
import {
  getCachedAppConfig,
  getCachedOptions,
  getCachedRegions,
  setCachedAppConfig,
  setCachedOptions,
  setCachedRegions
} from './storage'

const CACHE_TTL = 24 * 60 * 60 * 1000

export const metaState = reactive({
  options: {} as OptionDict,
  regions: [] as Region[],
  appConfig: null as AppConfig | null,
  optionsLoaded: false,
  regionsLoaded: false,
  appConfigLoaded: false,
  optionsLoading: false,
  regionsLoading: false,
  error: ''
})

let optionsPromise: Promise<boolean> | null = null
let regionsPromise: Promise<boolean> | null = null
let appConfigPromise: Promise<boolean> | null = null

function cacheValid(ts: number | undefined): boolean {
  return typeof ts === 'number' && Date.now() - ts < CACHE_TTL
}

export async function ensureOptions(force = false): Promise<boolean> {
  if (!force && metaState.optionsLoaded) return true
  if (optionsPromise) return optionsPromise
  const cached = getCachedOptions()
  if (!force && cached && cacheValid(cached.ts) && cached.data) {
    metaState.options = cached.data
    metaState.optionsLoaded = true
    return true
  }
  metaState.optionsLoading = true
  const task = getOptions()
    .then((data) => {
      metaState.options = data || {}
      metaState.optionsLoaded = true
      metaState.error = ''
      setCachedOptions(metaState.options)
      return true
    })
    .catch((error: Error) => {
      metaState.error = error.message || '选项加载失败'
      return false
    })
    .then((ok) => {
      metaState.optionsLoading = false
      optionsPromise = null
      return ok
    })
  optionsPromise = task
  return task
}

export async function ensureRegions(force = false): Promise<boolean> {
  if (!force && metaState.regionsLoaded) return true
  if (regionsPromise) return regionsPromise
  const cached = getCachedRegions()
  if (!force && cached && cacheValid(cached.ts) && cached.data) {
    metaState.regions = cached.data
    metaState.regionsLoaded = true
    return true
  }
  metaState.regionsLoading = true
  const task = getRegions()
    .then((data) => {
      metaState.regions = data || []
      metaState.regionsLoaded = true
      setCachedRegions(metaState.regions)
      return true
    })
    .catch((error: Error) => {
      metaState.error = error.message || '行政区划加载失败'
      return false
    })
    .then((ok) => {
      metaState.regionsLoading = false
      regionsPromise = null
      return ok
    })
  regionsPromise = task
  return task
}

export async function ensureAppConfig(force = false): Promise<boolean> {
  if (!force && metaState.appConfigLoaded) return true
  if (appConfigPromise) return appConfigPromise
  const cached = getCachedAppConfig()
  if (!force && cached && cacheValid(cached.ts) && cached.data) {
    metaState.appConfig = cached.data as unknown as AppConfig
    metaState.appConfigLoaded = true
    return true
  }
  const task = getAppConfig()
    .then((data) => {
      metaState.appConfig = data
      metaState.appConfigLoaded = true
      setCachedAppConfig(data as unknown as Record<string, unknown>)
      return true
    })
    .catch(() => false)
    .then((ok) => {
      appConfigPromise = null
      return ok
    })
  appConfigPromise = task
  return task
}

export function optionsOf(dict: string): OptionItem[] {
  return metaState.options[dict] || []
}

// 去掉编码 0（未填写）与指定编码，用于选择器
export function pickerOptions(dict: string, skipZero = true, skipValues: number[] = []): OptionItem[] {
  return optionsOf(dict).filter((item) => {
    if (skipZero && item.value === 0) return false
    if (skipValues.indexOf(item.value) >= 0) return false
    return true
  })
}

export function optionLabel(dict: string, value: number | string | null | undefined, fallback = '未填写'): string {
  if (value === null || value === undefined || value === '') return fallback
  const list = optionsOf(dict)
  if (list.length === 0) return fallback
  for (let i = 0; i < list.length; i++) {
    if (String(list[i].value) === String(value)) return list[i].label
  }
  return fallback
}

export function optionText(dict: string, values: Array<number | string> | null | undefined): string {
  if (!values || values.length === 0) return '不限制'
  const parts: string[] = []
  for (let i = 0; i < values.length; i++) {
    parts.push(optionLabel(dict, values[i], String(values[i])))
  }
  return parts.join('、')
}

function findRegion(code: string | null | undefined): Region | null {
  if (!code) return null
  for (let i = 0; i < metaState.regions.length; i++) {
    if (metaState.regions[i].code === code) return metaState.regions[i]
  }
  return null
}

export function provinces(): Region[] {
  return metaState.regions.filter((item) => item.level === 1)
}

export function citiesOf(parentCode: string): Region[] {
  return metaState.regions.filter((item) => item.level === 2 && item.parent_code === parentCode)
}

export function cityRegions(): Region[] {
  return metaState.regions.filter((item) => item.level === 2)
}

export function regionName(code: string | null | undefined): string {
  const region = findRegion(code)
  return region ? region.name : ''
}

// 城市展示：省 · 市
export function cityDisplay(code: string | null | undefined): string {
  const region = findRegion(code)
  if (!region) return ''
  const parent = findRegion(region.parent_code)
  return parent ? parent.name + ' · ' + region.name : region.name
}

export function cityDisplayOrBlank(code: string | null | undefined): string {
  return cityDisplay(code) || '未填写'
}

export function provinceOfCity(cityCode: string | null | undefined): string {
  const region = findRegion(cityCode)
  if (!region || !region.parent_code) return ''
  return region.parent_code
}
