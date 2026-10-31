// 表单校验工具：返回 null 表示通过，否则返回中文提示。
// 不引入正则，避免跨端转义差异。

export function isBlank(value: unknown): boolean {
  return value === null || value === undefined || String(value).trim() === ''
}

function isAllDigits(value: string): boolean {
  if (value.length === 0) return false
  for (let i = 0; i < value.length; i++) {
    const code = value.charCodeAt(i)
    if (code < 48 || code > 57) return false
  }
  return true
}

export function isPhone(phone: string): boolean {
  if (!phone || phone.length !== 11) return false
  if (phone.charAt(0) !== '1') return false
  const second = phone.charAt(1)
  if (second < '3' || second > '9') return false
  return isAllDigits(phone)
}

export function isSmsCode(code: string): boolean {
  return code.length === 6 && isAllDigits(code)
}

export function validatePhone(phone: string): string | null {
  if (isBlank(phone)) return '请输入手机号'
  if (!isPhone(phone)) return '请输入正确的中国大陆手机号'
  return null
}

export function validateSmsCodeInput(code: string): string | null {
  if (isBlank(code)) return '请输入验证码'
  if (!isSmsCode(code)) return '验证码为 6 位数字'
  return null
}

export function validateNickname(nickname: string): string | null {
  const value = (nickname || '').trim()
  if (!value) return '请填写昵称'
  if (value.length > 32) return '昵称不能超过 32 个字符'
  return null
}

export function toNumberOrNull(value: unknown): number | null {
  if (value === null || value === undefined || String(value).trim() === '') return null
  const num = Number(value)
  if (isNaN(num)) return null
  return num
}

export function toIntOrNull(value: unknown): number | null {
  const num = toNumberOrNull(value)
  if (num === null) return null
  return Math.round(num)
}

export function validateAgeRange(min: number | null, max: number | null): string | null {
  if (min !== null && (min < 18 || min > 100)) return '最小年龄需在 18 到 100 岁之间'
  if (max !== null && (max < 18 || max > 100)) return '最大年龄需在 18 到 100 岁之间'
  if (min !== null && max !== null && min > max) return '最小年龄不能大于最大年龄'
  return null
}

export function validateHeightRange(min: number | null, max: number | null): string | null {
  if (min !== null && (min < 100 || min > 250)) return '身高需在 100 到 250 cm 之间'
  if (max !== null && (max < 100 || max > 250)) return '身高需在 100 到 250 cm 之间'
  if (min !== null && max !== null && min > max) return '最低身高不能高于最高身高'
  return null
}

export function validateBirthDate(value: string | null): string | null {
  if (!value) return null
  const date = new Date(value)
  if (isNaN(date.getTime())) return '出生日期格式不正确'
  const now = new Date()
  if (date.getTime() > now.getTime()) return '出生日期不能晚于今天'
  let age = now.getFullYear() - date.getFullYear()
  const monthDiff = now.getMonth() - date.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && now.getDate() < date.getDate())) age--
  if (age < 18) return '本站仅面向 18 岁以上用户'
  if (age > 100) return '请检查出生日期是否正确'
  return null
}

// 计算器第 1 步：我是谁
export function validateCalculatorSelf(self: {
  age: string
  height_cm: string
}): string | null {
  const age = toIntOrNull(self.age)
  if (age !== null && (age < 18 || age > 100)) return '年龄需在 18 到 100 岁之间'
  const height = toIntOrNull(self.height_cm)
  if (height !== null && (height < 100 || height > 250)) return '身高需在 100 到 250 cm 之间'
  return null
}

// 计算器第 2 步：我希望找谁（范围类条件）
export function validateCalculatorConditions(input: {
  min_age: string
  max_age: string
  min_height_cm: string
  max_height_cm: string
}): string | null {
  const ageError = validateAgeRange(toIntOrNull(input.min_age), toIntOrNull(input.max_age))
  if (ageError) return ageError
  const heightError = validateHeightRange(toIntOrNull(input.min_height_cm), toIntOrNull(input.max_height_cm))
  if (heightError) return heightError
  return null
}

// 资料必填项：昵称、性别、城市
export function validateProfileBase(draft: {
  nickname: string
  gender: number | null
  city_code: string
}): string | null {
  const nicknameError = validateNickname(draft.nickname)
  if (nicknameError) return nicknameError
  if (draft.gender === null || draft.gender === undefined || draft.gender <= 0) return '请选择性别'
  if (isBlank(draft.city_code)) return '请选择所在城市'
  return null
}

export function validateBio(bio: string): string | null {
  if (bio && bio.length > 500) return '个人简介不能超过 500 个字符'
  return null
}

// 计算器第 2 步至少选择一个目标性别
export function validateTargetGenders(genders: number[]): string | null {
  if (!genders || genders.length === 0) return '请至少选择一个目标性别'
  return null
}
