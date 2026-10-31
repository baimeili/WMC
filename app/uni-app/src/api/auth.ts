// 注册登录相关接口。
import { post } from './http'
import type { TokenPair } from './types'

export function sendSmsCode(phone: string): Promise<void> {
  return post<void>('/auth/sms/send', { phone })
}

export function verifySmsCode(payload: {
  phone: string
  code: string
  anonymous_session_id?: string
}): Promise<TokenPair> {
  return post<TokenPair>('/auth/sms/verify', payload)
}

export function logout(): Promise<void> {
  return post<void>('/auth/logout')
}
