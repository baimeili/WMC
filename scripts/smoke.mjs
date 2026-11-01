#!/usr/bin/env node
/**
 * 端到端冒烟脚本（闭环验收）
 * 用法：node scripts/smoke.mjs [API_BASE]
 * 默认 API_BASE=http://127.0.0.1:8080/api/v1
 * 依赖：Node.js 18+（内置 fetch / FormData / Blob），无需安装第三方包。
 */

const API = (process.argv[2] || 'http://127.0.0.1:8080/api/v1').replace(/\/$/, '')
const results = []
let failed = 0

function record(name, ok, detail = '') {
  results.push({ name, ok, detail })
  if (!ok) failed++
  console.log((ok ? '  PASS ' : '  FAIL ') + name + (detail ? '  -> ' + detail : ''))
}

async function req(method, path, { body, token, form } = {}) {
  const headers = {}
  if (token) headers.Authorization = 'Bearer ' + token
  let payload
  if (form) {
    payload = form
  } else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }
  const resp = await fetch(API + path, { method, headers, body: payload })
  let json = null
  try { json = await resp.json() } catch (e) { json = null }
  return { status: resp.status, code: json && json.code, data: json && json.data, message: json && json.message }
}

function randomPhone() {
  // 18xxxxxxxxx 之外的号段，避免与种子用户冲突
  const tail = String(Math.floor(Math.random() * 90000000) + 10000000)
  return '135' + tail
}

async function register(phone) {
  const send = await req('POST', '/auth/sms/send', { body: { phone } })
  const verify = await req('POST', '/auth/sms/verify', { body: { phone, code: '123456' } })
  return { send, verify, token: verify.data && verify.data.access_token, userId: verify.data && verify.data.user_id }
}

const profilePayload = (gender, age, height, city = '310100') => ({
  nickname: '冒烟' + gender + age,
  gender,
  birth_date: new Date(Date.now() - age * 365.25 * 86400000).toISOString().slice(0, 10),
  province_code: '310000',
  city_code: city,
  height_cm: height,
  education_level: 3,
  occupation_category: 1,
  employment_status: 1,
  annual_income_band: 4,
  debt_band: 0,
  house_status: 2,
  vehicle_status: 1,
  marital_status: 1,
  children_count: 0,
  wants_children: 2,
  marriage_intent: 2,
  marriage_timeline: 2,
  smoking_status: 1,
  drinking_status: 1,
  bio: '端到端冒烟测试用户'
})

const preferencesPayload = (targetGender, minAge, maxAge) => ({
  target_genders: [targetGender],
  min_age: minAge,
  max_age: maxAge,
  age_hard: true,
  min_height_cm: 150,
  max_height_cm: 200,
  height_hard: false,
  allowed_city_codes: ['310100'],
  city_hard: true,
  min_education_level: 3,
  education_hard: false,
  min_income_band: 1,
  income_hard: false,
  allowed_marital_statuses: [1],
  marital_status_hard: true,
  allowed_children_states: [0],
  children_hard: true,
  allowed_wants_children: [2, 3],
  wants_children_hard: true,
  allowed_smoking_statuses: [1, 4],
  smoking_hard: false,
  allowed_drinking_statuses: [1, 2],
  drinking_hard: false,
  min_marriage_intent: 1,
  marriage_intent_hard: false,
  weights: {}
})

async function setupUser(gender, age, height, targetGender, minAge, maxAge, withPhoto) {
  const phone = randomPhone()
  const reg = await register(phone)
  record('短信验证码登录 ' + phone, reg.verify.code === 'OK', reg.verify.message || '')
  const token = reg.token
  const profile = await req('PUT', '/me/profile', { token, body: profilePayload(gender, age, height) })
  record('保存资料', profile.code === 'OK', '完成度 ' + (profile.data && profile.data.profile_completion))
  const pref = await req('PUT', '/me/preferences', { token, body: preferencesPayload(targetGender, minAge, maxAge) })
  record('保存择偶条件', pref.code === 'OK')

  const questions = await req('GET', '/meta/compatibility-questions')
  const rules = (questions.data || []).slice(0, 6).map((q) => ({
    question_id: q.id,
    own_answer: [q.options[0].value],
    accepted_answers: [q.options[0].value],
    importance: 3,
    is_hard: false
  }))
  const saved = await req('PUT', '/me/compatibility', { token, body: { rules } })
  record('保存价值观规则', saved.code === 'OK', '条数 ' + (saved.data ? saved.data.length : 0))

  for (const type of [1, 2, 3]) {
    const consent = await req('POST', '/me/consents', { token, body: { consent_type: type } })
    if (consent.code !== 'OK') record('记录同意类型 ' + type, false, consent.message || '')
  }
  record('记录三项协议同意', true)

  let photoId = null
  if (withPhoto) {
    // 1x1 PNG，用于验证上传、审核与签名 URL 链路
    const png = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==', 'base64')
    const form = new FormData()
    form.append('file', new Blob([png], { type: 'image/png' }), 'smoke.png')
    form.append('make_primary', 'true')
    const uploaded = await req('POST', '/me/photos', { token, form })
    record('上传照片', uploaded.code === 'OK', uploaded.message || '')
    photoId = uploaded.data && uploaded.data.id
  }

  const status = await req('GET', '/me/market/status', { token })
  const join = await req('POST', '/me/market/join', { token })
  record('加入真实婚恋市场', join.code === 'OK', join.message || '')
  return { phone, token, userId: reg.userId, photoId, status }
}

async function main() {
  console.log('=== 冒烟测试：' + API + ' ===')

  const health = await req('GET', '/health/ready'.replace('/api/v1', ''))
  // health/ready 不在 /api/v1 前缀下，单独请求
  const base = API.replace('/api/v1', '')
  const healthy = await fetch(base + '/health/ready')
  record('健康检查 /health/ready', healthy.status === 200)

  const meta = await req('GET', '/meta/options')
  record('元数据 /meta/options', meta.code === 'OK' && !!meta.data.genders)
  const config = await req('GET', '/meta/app-config')
  record('元数据 /meta/app-config', config.code === 'OK' && !!config.data.features)
  const questions = await req('GET', '/meta/compatibility-questions')
  record('价值观问题', questions.code === 'OK' && questions.data.length >= 6)

  // 游客计算器
  const anon = 'smoke-' + Date.now()
  const estimate = await req('POST', '/calculator/estimate', {
    body: { anonymous_session_id: anon, conditions: { target_genders: [2], min_age: 25, max_age: 33, min_height_cm: 163, min_education_level: 3 } }
  })
  record('理论市场估算', estimate.code === 'OK' && typeof estimate.data.matched_ratio === 'number' && estimate.data.funnel.steps.length === 9)
  const share = await req('POST', '/calculator/share', { body: { run_id: estimate.data.run_id, anonymous_session_id: anon } })
  record('生成分享 token', share.code === 'OK' && !!share.data.share_token)
  const shared = await req('GET', '/calculator/share/' + encodeURIComponent(share.data.share_token))
  record('读取分享结果', shared.code === 'OK' && shared.data.matched_ratio === estimate.data.matched_ratio)

  // 用户 A（女 28）与用户 B（男 30）
  const a = await setupUser(2, 28, 168, 1, 27, 34, true)
  const b = await setupUser(1, 30, 178, 2, 24, 32, false)

  // 管理员：审核照片
  const adminLogin = await req('POST', '/admin/auth/dev-login', { body: { phone: '13800000000' } })
  record('管理员开发登录', adminLogin.code === 'OK', adminLogin.message || '')
  const admin = adminLogin.data.access_token
  if (a.photoId) {
    const approve = await req('POST', '/admin/photos/' + a.photoId + '/approve', { token: admin })
    record('管理员审核照片通过', approve.code === 'OK', approve.message || '')
  }

  // 双向发现
  const summary = await req('GET', '/discovery/summary', { token: a.token })
  record('真实候选池概览', summary.code === 'OK' && summary.data.in_market === true, '双向 ' + (summary.data && summary.data.eligible_count_text))
  const candidates = await req('GET', '/discovery/candidates?limit=50', { token: a.token })
  const hasB = candidates.code === 'OK' && (candidates.data.items || []).some((c) => c.user_id === b.userId)
  record('候选池包含双向满足的 B', hasB)
  const funnel = await req('GET', '/discovery/funnel', { token: a.token })
  record('真实候选池漏斗', funnel.code === 'OK' && funnel.data.steps.length >= 10)
  const sens = await req('GET', '/discovery/sensitivity', { token: a.token })
  record('条件敏感性', sens.code === 'OK' && sens.data.items.length > 0)

  // Like → Match
  const likeA = await req('PUT', '/interactions/' + b.userId, { token: a.token, body: { action: 1, source: 1 } })
  record('A Like B（单向不产生 Match）', likeA.code === 'OK' && likeA.data.matched === false)
  const likeB = await req('PUT', '/interactions/' + a.userId, { token: b.token, body: { action: 1, source: 1 } })
  record('B Like A（形成 Match）', likeB.code === 'OK' && likeB.data.matched === true && !!likeB.data.match_id)
  const matchId = likeB.data.match_id
  const repeat = await req('PUT', '/interactions/' + a.userId, { token: b.token, body: { action: 1, source: 1 } })
  record('重复 Like 返回同一 Match', repeat.data.match_id === matchId)

  const matches = await req('GET', '/matches', { token: a.token })
  record('Match 列表', matches.code === 'OK' && matches.data.items.some((m) => m.id === matchId))
  const detail = await req('GET', '/matches/' + matchId, { token: a.token })
  record('Match 详情', detail.code === 'OK' && !!detail.data.peer_nickname)

  const o1 = await req('POST', '/matches/' + matchId + '/outcomes', { token: a.token, body: { outcome_type: 3 } })
  record('A 反馈线下见面', o1.code === 'OK')
  const before = await req('GET', '/matches/' + matchId, { token: a.token })
  record('单方反馈不确认见面', before.code === 'OK' && before.data.confirmed_meeting === false)
  const o2 = await req('POST', '/matches/' + matchId + '/outcomes', { token: b.token, body: { outcome_type: 3 } })
  const after = await req('GET', '/matches/' + matchId, { token: a.token })
  record('双方反馈后确认见面', o2.code === 'OK' && after.data.confirmed_meeting === true)

  // 屏蔽与举报：屏蔽后 B 必须从候选列表消失，且双向人数不增加
  const beforeBlock = await req('GET', '/discovery/candidates?limit=50', { token: a.token })
  const beforeHasB = (beforeBlock.data.items || []).some((c) => c.user_id === b.userId)
  const block = await req('POST', '/users/' + b.userId + '/block', { token: a.token })
  record('屏蔽用户', block.code === 'OK')
  const summary2 = await req('GET', '/discovery/summary', { token: a.token })
  const afterBlock = await req('GET', '/discovery/candidates?limit=50', { token: a.token })
  const afterHasB = (afterBlock.data.items || []).some((c) => c.user_id === b.userId)
  record('屏蔽前候选列表包含 B', beforeHasB)
  record(
    '屏蔽后 B 不再出现在候选列表',
    summary2.code === 'OK' && afterHasB === false && (summary2.data.reciprocal_count <= summary.data.reciprocal_count)
  )
  const unblock = await req('DELETE', '/users/' + b.userId + '/block', { token: a.token })
  record('取消屏蔽', unblock.code === 'OK')
  const report = await req('POST', '/users/' + b.userId + '/report', { token: a.token, body: { reason: 2, description: '冒烟测试' } })
  record('举报用户', report.code === 'OK')

  // 后台
  const dashboard = await req('GET', '/admin/dashboard', { token: admin })
  record('后台看板', dashboard.code === 'OK' && dashboard.data.matches_total >= 1 && !!dashboard.data.funnel.length)
  const adminFunnel = await req('GET', '/admin/funnel', { token: admin })
  record('后台经营漏斗', adminFunnel.code === 'OK' && adminFunnel.data.steps.length >= 8)
  const reports = await req('GET', '/admin/reports?status=1', { token: admin })
  record('后台举报列表', reports.code === 'OK' && reports.data.length >= 1)
  const settings = await req('GET', '/admin/settings', { token: admin })
  record('后台系统参数', settings.code === 'OK' && !!settings.data.features)
  const audit = await req('GET', '/admin/audit-logs?limit=5', { token: admin })
  record('后台审计日志', audit.code === 'OK' && audit.data.length >= 1)
  const datasets = await req('GET', '/admin/baseline/datasets', { token: admin })
  record('基准数据集列表', datasets.code === 'OK' && datasets.data.length >= 1)
  const users = await req('GET', '/admin/users?page=1&page_size=5', { token: admin })
  record('后台用户查询', users.code === 'OK' && users.data.total >= 2)

  console.log('')
  console.log('=== 结果：' + (results.length - failed) + '/' + results.length + ' 通过 ===')
  if (failed > 0) {
    console.log('失败项：')
    results.filter((r) => !r.ok).forEach((r) => console.log('  - ' + r.name + ' ' + r.detail))
    process.exit(1)
  }
}

main().catch((e) => {
  console.error('冒烟测试异常：', e)
  process.exit(1)
})