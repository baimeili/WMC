<!-- 基准数据集：列表、创建、CSV 导入（dry_run 校验预览）、激活、回滚。
     理论市场数据为统计估算，页面显著提示禁止对外宣称精确人数/双向匹配率。 -->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { activateDataset, createDataset, importCsv, listDatasets, rollbackDataset } from '@/api/baseline'
import type { BaselineDataset, BaselineDatasetInput, ImportReport } from '@/api/types'
import { formatCell, formatDateTime, formatNumber } from '@/utils/format'
import { DATASET_STATUS_LABELS, labelOf } from '@/utils/enums'

const router = useRouter()

const loading = ref(false)
const datasets = ref<BaselineDataset[]>([])

// 创建数据集表单。
interface CreateFormState {
  name: string
  version: string
  source_name: string
  source_url: string
  source_year: number | null
  description: string
  methodology: string
  disclaimer: string
  is_demo: boolean
}

const createVisible = ref(false)
const createFormRef = ref<FormInstance>()
const creating = ref(false)
const createForm = reactive<CreateFormState>({
  name: '',
  version: '',
  source_name: '',
  source_url: '',
  source_year: null,
  description: '',
  methodology: '',
  disclaimer: '',
  is_demo: false
})
const createRules: FormRules = {
  name: [{ required: true, message: '请填写数据集名称', trigger: 'blur' }],
  version: [{ required: true, message: '请填写版本号', trigger: 'blur' }]
}

// CSV 导入状态。
const importVisible = ref(false)
const importTarget = ref<BaselineDataset | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const importReport = ref<ImportReport | null>(null)
const dryRunning = ref(false)
const importing = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    datasets.value = await listDatasets()
  } catch {
    datasets.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

function resetCreateForm(): void {
  createForm.name = ''
  createForm.version = ''
  createForm.source_name = ''
  createForm.source_url = ''
  createForm.source_year = null
  createForm.description = ''
  createForm.methodology = ''
  createForm.disclaimer = ''
  createForm.is_demo = false
}

function openCreate(): void {
  resetCreateForm()
  createVisible.value = true
}

async function submitCreate(): Promise<void> {
  if (!createFormRef.value) {
    return
  }
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }
  creating.value = true
  try {
    const payload: BaselineDatasetInput = {
      name: createForm.name.trim(),
      version: createForm.version.trim(),
      source_name: createForm.source_name.trim() || undefined,
      source_url: createForm.source_url.trim() || undefined,
      source_year: createForm.source_year ?? null,
      description: createForm.description.trim() || undefined,
      methodology: createForm.methodology.trim() || undefined,
      disclaimer: createForm.disclaimer.trim() || undefined,
      is_demo: createForm.is_demo
    }
    await createDataset(payload)
    ElMessage.success('数据集已创建（草稿），请继续导入 CSV')
    createVisible.value = false
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    creating.value = false
  }
}

function openImport(dataset: BaselineDataset): void {
  importTarget.value = dataset
  selectedFile.value = null
  importReport.value = null
  importVisible.value = true
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function triggerFilePick(): void {
  fileInput.value?.click()
}

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  importReport.value = null
  if (!file) {
    selectedFile.value = null
    return
  }
  if (!file.name.toLowerCase().endsWith('.csv')) {
    ElMessage.warning('仅支持 CSV 文件')
    selectedFile.value = null
    input.value = ''
    return
  }
  selectedFile.value = file
}

// dry_run：仅校验并预览，不写库。
async function runDryRun(): Promise<void> {
  const target = importTarget.value
  const file = selectedFile.value
  if (!target?.id || !file) {
    ElMessage.warning('请先选择 CSV 文件')
    return
  }
  dryRunning.value = true
  try {
    importReport.value = await importCsv(target.id, file, true)
    ElMessage.success('校验完成，请查看校验结果与数据预览')
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    dryRunning.value = false
  }
}

// 正式导入：校验通过后写库。
async function submitImport(): Promise<void> {
  const target = importTarget.value
  const file = selectedFile.value
  if (!target?.id || !file) {
    ElMessage.warning('请先选择 CSV 文件')
    return
  }
  importing.value = true
  try {
    importReport.value = await importCsv(target.id, file, false)
    ElMessage.success('CSV 导入完成')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    importing.value = false
  }
}

const previewColumns = computed<string[]>(() => {
  const rows = importReport.value?.preview ?? []
  const first = rows[0]
  return first ? Object.keys(first) : []
})

// 拒绝非法 CSV 行的校验错误。
const importErrors = computed<string[]>(() => importReport.value?.errors ?? [])

async function handleActivate(dataset: BaselineDataset): Promise<void> {
  if (!dataset.id) {
    return
  }
  const demoHint = dataset.is_demo ? '该数据集标记为 DEMO（演示数据），生产环境禁止激活。' : ''
  try {
    await ElMessageBox.confirm(
      `激活后该数据集将成为当前唯一生效的基准数据集，用于理论市场估算。${demoHint}`,
      '激活数据集',
      {
        type: dataset.is_demo ? 'error' : 'warning',
        confirmButtonText: '确认激活',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }
  try {
    await activateDataset(dataset.id)
    ElMessage.success('已激活该数据集')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

async function handleRollback(dataset: BaselineDataset): Promise<void> {
  if (!dataset.id) {
    return
  }
  try {
    await ElMessageBox.confirm(
      '回滚将删除该数据集的全部基准记录并归档数据集，操作不可恢复，确定继续吗？',
      '回滚数据集',
      {
        type: 'error',
        confirmButtonText: '确认回滚',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }
  try {
    await rollbackDataset(dataset.id)
    ElMessage.success('已回滚该数据集')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

function goDetail(dataset: BaselineDataset): void {
  if (!dataset.id) {
    return
  }
  void router.push(`/baseline/${dataset.id}`)
}

function statusTagType(status: number | undefined): 'success' | 'warning' | 'info' {
  if (status === 2) {
    return 'success'
  }
  if (status === 3) {
    return 'warning'
  }
  return 'info'
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">基准数据集</h1>
        <p class="page-desc">
          管理理论婚恋市场的基准人口数据集：创建、CSV 导入校验、激活、回滚。同一时刻仅允许一个激活的数据集。
        </p>
      </div>
      <div class="page-head-actions">
        <el-button size="small" @click="load">刷新</el-button>
        <el-button size="small" type="primary" @click="openCreate">新建数据集</el-button>
      </div>
    </div>

    <el-alert
      type="warning"
      :closable="false"
      show-icon
      class="mb-12"
      title="理论市场为统计估算，必须与本站真实市场严格区分"
      description="基准数据集来自公开统计/调查的加权匿名样本，只用于估算「符合条件人口占比」。页面只能展示「约 X%」，禁止对外宣称精确人数或双向匹配率；本站真实匹配率只来自真实用户池。"
    />
    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="mb-12"
      title="DEMO 标记说明"
      description="标记为 DEMO 的数据集仅用于开发联调；生产环境必须关闭 demo_dataset_enabled 并禁止激活演示数据，避免把虚构数据冒充真实统计。"
    />

    <div class="page-card">
      <el-table :data="datasets" border size="default">
        <el-table-column label="名称" min-width="200">
          <template #default="{ row }">
            <div class="dataset-name">{{ row.name || '未命名数据集' }}</div>
            <div class="muted">{{ row.description || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="版本" width="100" align="center">
          <template #default="{ row }">{{ row.version || '—' }}</template>
        </el-table-column>
        <el-table-column label="来源" min-width="180">
          <template #default="{ row }">
            <div>{{ row.source_name || '未填写' }}</div>
            <div class="muted">{{ row.source_year ? row.source_year + ' 年' : '年份未填写' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="记录数" width="110" align="right">
          <template #default="{ row }">{{ formatNumber(row.record_count) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ labelOf(DATASET_STATUS_LABELS, row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="激活" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.active" type="success" size="small">当前生效</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="DEMO" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_demo" type="warning" size="small">DEMO</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="goDetail(row)">详情</el-button>
            <el-button link type="primary" @click="openImport(row)">导入 CSV</el-button>
            <el-button v-if="!row.active" link type="success" @click="handleActivate(row)">激活</el-button>
            <el-button link type="danger" @click="handleRollback(row)">回滚</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无基准数据集，请先创建并导入 CSV" :image-size="60" />
        </template>
      </el-table>
    </div>

    <el-dialog v-model="createVisible" title="新建基准数据集" width="640px" top="6vh">
      <el-alert
        type="warning"
        :closable="false"
        class="mb-12"
        title="理论市场数据只能用于统计估算"
        description="请填写真实来源、方法与免责声明；禁止在页面或文案中宣称精确人数或双向匹配率。"
      />
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="110px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="如：2025 年全国婚恋市场抽样基准" />
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="createForm.version" placeholder="如：v2025.1" />
        </el-form-item>
        <el-form-item label="来源名称">
          <el-input v-model="createForm.source_name" placeholder="如：某公开人口普查抽样数据" />
        </el-form-item>
        <el-form-item label="来源链接">
          <el-input v-model="createForm.source_url" placeholder="https://" />
        </el-form-item>
        <el-form-item label="来源年份">
          <el-input-number v-model="createForm.source_year" :min="1990" :max="2100" controls-position="right" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="2" maxlength="300" show-word-limit />
        </el-form-item>
        <el-form-item label="统计方法">
          <el-input v-model="createForm.methodology" type="textarea" :rows="2" maxlength="500" show-word-limit placeholder="如：按性别、年龄、城市分层的加权抽样" />
        </el-form-item>
        <el-form-item label="免责声明">
          <el-input v-model="createForm.disclaimer" type="textarea" :rows="2" maxlength="500" show-word-limit placeholder="提示理论市场为统计估算，禁止当作精确人数" />
        </el-form-item>
        <el-form-item label="DEMO 标记">
          <el-switch v-model="createForm.is_demo" />
          <span class="muted inline-block mt-8">演示数据仅供开发联调，生产环境禁止激活</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建草稿</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="CSV 导入" width="760px" top="6vh">
      <p class="muted">
        目标数据集：{{ importTarget?.name || '—' }}（{{ importTarget?.version || '版本未填写' }}）。支持 dry_run 预览校验错误，确认无误后再正式导入。
      </p>
      <div class="import-toolbar">
        <input ref="fileInput" type="file" accept=".csv,text/csv" class="hidden-file" @change="onFileChange" />
        <el-button @click="triggerFilePick">选择 CSV 文件</el-button>
        <span class="muted">{{ selectedFile ? selectedFile.name : '未选择文件' }}</span>
        <el-button :disabled="!selectedFile" :loading="dryRunning" @click="runDryRun">校验预览（dry_run）</el-button>
        <el-button type="primary" :disabled="!selectedFile" :loading="importing" @click="submitImport">正式导入</el-button>
      </div>

      <template v-if="importReport">
        <el-divider content-position="left">导入结果</el-divider>
        <el-descriptions :column="4" border size="small">
          <el-descriptions-item label="总行数">{{ formatNumber(importReport.total_rows) }}</el-descriptions-item>
          <el-descriptions-item label="有效行">{{ formatNumber(importReport.valid_rows) }}</el-descriptions-item>
          <el-descriptions-item label="无效行">{{ formatNumber(importReport.invalid_rows) }}</el-descriptions-item>
          <el-descriptions-item label="是否写库">{{ importReport.imported ? '已写库' : '未写库（预览）' }}</el-descriptions-item>
        </el-descriptions>

        <template v-if="importErrors.length > 0">
          <p class="text-danger mt-16">校验错误（{{ importErrors.length }} 条）：</p>
          <el-scrollbar max-height="160px">
            <p v-for="(error, index) in importErrors" :key="index" class="import-error">{{ error }}</p>
          </el-scrollbar>
        </template>

        <template v-if="previewColumns.length > 0">
          <p class="mt-16">数据预览（前 {{ importReport.preview?.length ?? 0 }} 行）：</p>
          <el-table :data="importReport.preview ?? []" border size="small" max-height="260">
            <el-table-column v-for="column in previewColumns" :key="column" :label="column" min-width="110">
              <template #default="{ row }">{{ formatCell(row[column]) }}</template>
            </el-table-column>
          </el-table>
        </template>
      </template>

      <template #footer>
        <el-button @click="importVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-head-actions {
  display: flex;
  gap: 8px;
  flex: none;
}

.dataset-name {
  font-weight: 600;
}

.import-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
}

.import-error {
  margin: 0 0 6px;
  color: #f56c6c;
  font-size: 12px;
  line-height: 1.6;
  word-break: break-all;
}
</style>
