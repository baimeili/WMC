// 基准数据集管理接口：创建、CSV 导入（含 dry_run 校验预览）、激活、回滚。
import { get, post, toArray, upload } from './http'
import type { BaselineDataset, BaselineDatasetInput, ImportReport } from './types'

// 数据集列表。
export async function listDatasets(): Promise<BaselineDataset[]> {
  const payload = await get<BaselineDataset[] | { items?: BaselineDataset[] }>('/admin/baseline/datasets')
  return toArray<BaselineDataset>(payload)
}

// 创建数据集（草稿状态）。
export function createDataset(input: BaselineDatasetInput): Promise<BaselineDataset> {
  return post<BaselineDataset>('/admin/baseline/datasets', input)
}

// CSV 导入：dry_run=true 时仅校验并返回错误与预览，不写库。
export function importCsv(datasetId: string, file: File, dryRun: boolean): Promise<ImportReport> {
  const form = new FormData()
  form.append('file', file)
  form.append('dry_run', dryRun ? 'true' : 'false')
  return upload<ImportReport>(`/admin/baseline/datasets/${datasetId}/import`, form)
}

// 激活数据集（同一时刻仅允许一个）。
export function activateDataset(datasetId: string): Promise<void> {
  return post<void>(`/admin/baseline/datasets/${datasetId}/activate`)
}

// 回滚数据集（删除记录并归档）。
export function rollbackDataset(datasetId: string): Promise<void> {
  return post<void>(`/admin/baseline/datasets/${datasetId}/rollback`)
}

// 从列表接口获取单个数据集（契约未提供单条查询端点）。
export async function getDataset(datasetId: string): Promise<BaselineDataset | undefined> {
  const datasets = await listDatasets()
  return datasets.find((dataset) => dataset.id === datasetId)
}
