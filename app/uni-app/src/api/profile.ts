// 当前用户资料、协议同意、入池退池与注销接口。
import { del, get, post, put } from './http'
import type { Profile, ProfileInput } from './types'

export function getProfile(): Promise<Profile> {
  return get<Profile>('/me/profile')
}

export function updateProfile(input: ProfileInput): Promise<Profile> {
  return put<Profile>('/me/profile', input)
}

export function joinMarket(): Promise<void> {
  return post<void>('/me/market/join')
}

export function leaveMarket(): Promise<void> {
  return del<void>('/me/market/join')
}

export function recordConsent(consentType: number, policyVersion: string): Promise<void> {
  return post<void>('/me/consents', { consent_type: consentType, policy_version: policyVersion })
}

export function deleteAccount(): Promise<void> {
  return del<void>('/me/account')
}
