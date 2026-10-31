// 登录与令牌相关接口。
import { post } from './http'
import type { TokenPair } from './types'

// 发送短信验证码。
export function sendSmsCode(phone: string): Promise<void> {
  return post<void>('/auth/sms/send', { phone })
}

// 校验验证码并登录（首次自动注册）。
export function verifySmsCode(phone: string, code: string): Promise<TokenPair> {
  return post<TokenPair>('/auth/sms/verify', { phone, code })
}

// 管理端开发登录：生产环境返回 404，因此使用 silent 由登录页自行提示。
export function devLogin(phone: string): Promise<TokenPair> {
  return post<TokenPair>('/admin/auth/dev-login', { phone }, { silent: true })
}

// 退出登录并吊销 Refresh Token。
export function logout(): Promise<void> {
  return post<void>('/auth/logout')
}

// 刷新 Access Token（Refresh Token 轮换）。
export function refreshTokens(refreshToken: string): Promise<TokenPair> {
  return post<TokenPair>('/auth/refresh', { refresh_token: refreshToken })
}
