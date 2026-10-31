// 择偶条件接口。
import { get, put } from './http'
import type { Preferences, PreferencesInput } from './types'

export function getPreferences(): Promise<Preferences> {
  return get<Preferences>('/me/preferences')
}

export function updatePreferences(input: PreferencesInput): Promise<Preferences> {
  return put<Preferences>('/me/preferences', input)
}
