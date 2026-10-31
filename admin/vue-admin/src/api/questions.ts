// 婚恋价值观问题配置接口。
import { get, post, put, toArray } from './http'
import type { CompatibilityQuestion, CompatibilityQuestionInput } from './types'

// 查询全部价值观问题（含停用）。
export async function listQuestions(): Promise<CompatibilityQuestion[]> {
  const payload = await get<CompatibilityQuestion[] | { items?: CompatibilityQuestion[] }>('/admin/questions')
  return toArray<CompatibilityQuestion>(payload)
}

// 新增问题。
export function createQuestion(input: CompatibilityQuestionInput): Promise<CompatibilityQuestion> {
  return post<CompatibilityQuestion>('/admin/questions', input)
}

// 修改问题（全量更新，包含启停）。
export function updateQuestion(id: string, input: CompatibilityQuestionInput): Promise<CompatibilityQuestion> {
  return put<CompatibilityQuestion>(`/admin/questions/${id}`, input)
}
