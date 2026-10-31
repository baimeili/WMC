<!-- 基准数据集详情：元数据、理论与真实区分提示、激活与回滚入口。 -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { activateDataset, getDataset, rollbackDataset } from '@/api/baseline'
import type { BaselineDataset } from '@/api/types'
import { formatDateTime, formatNumber } from '@/utils/format'
import { DATASET_STATUS_LABELS, labelOf } from '@/utils/enums'

const route = useRoute()
const router = useRouter()

const datasetId = computed(() => String(route.params.id ?? ''))
const loading = ref(false)
const dataset = ref<BaselineDataset | null>(null)
const loadError = ref('')

async function load(): Promise<void> {
  if (!datasetId.value) {
    loadError.value = '缺少数据集 id'
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const found = await getDataset(datasetId.value)
    if (found) {
      dataset.value = found
    } else {
      dataset.value = null
      loadError.value = '未找到该基准数据集'
    }
  } catch {
    dataset.value = null
    loadError.value = '加载数据集详情失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function handleActivate(): Promise<void> {
  const current = dataset.value
  if (!current?.id) {
    return
  }
  const demoHint = current.is_demo ? '该数据集为 DEMO 演示数据，生产环境禁止激活。' : ''
  try {
    await ElMessageBox.confirm(
      `激活后该数据集将成为当前唯一生效的基准数据集，用于理论市场估算。${demoHint}`,
      '激活数据集',
      {
        type: current.is_demo ? 'error' : 'warning',
        confirmButtonText: '确认激活',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }
  try {
    await activateDataset(current.id)
    ElMessage.success('已激活该数据集')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

async function handleRollback(): Promise<void> {
  const current = dataset.value
  if (!current?.id) {
    return
  }
  try {
    await ElMessageBox.confirm('回滚将删除该数据集的全部基准记录并归档数据集，操作不可恢复，确定继续吗？', '回滚数据集', {
      type: 'error',
      confirmButtonText: '确认回滚',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await rollbackDataset(current.id)
    ElMessage.success('已回滚该数据集')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

function goBack(): void {
  void router.push('/baseline')
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">数据集详情</h1>
        <p class="page-desc">数据集 ID：{{ datasetId || '—' }}；基准数据仅用于理论市场统计估算。</p>
      </div>
      <div class="page-head-actions">
        <el-button size="small" @click="goBack">返回列表</el-button>
        <el-button v-if="dataset && !dataset.active" size="small" type="success" @click="handleActivate">激活</el-button>
        <el-button v-if="dataset" size="small" type="danger" @click="handleRollback">回滚</el-button>
      </div>
    </div>

    <el-alert
      type="warning"
      :closable="false"
      show-icon
      class="mb-12"
      title="理论市场为统计估算，必须与本站真实市场严格区分"
      description="该数据集只用于估算「符合条件人口占比」，只能展示「约 X%」；禁止对外宣称精确人数或双向匹配率，也不得与本站真实用户池数据混用。"
    />

    <el-result v-if="loadError && !dataset" icon="warning" title="无法加载数据集" :sub-title="loadError">
      <template #extra>
        <el-button type="primary" @click="load">重试</el-button>
        <el-button @click="goBack">返回列表</el-button>
      </template>
    </el-result>

    <template v-else-if="dataset">
      <div class="page-card">
        <div class="card-title-row">
          <h2 class="card-title">{{ dataset.name || '未命名数据集' }}</h2>
          <div class="tag-row">
            <el-tag v-if="dataset.is_demo" type="warning" size="small">DEMO 演示数据</el-tag>
            <el-tag :type="dataset.active ? 'success' : 'info'" size="small">
              {{ dataset.active ? '当前生效' : '未激活' }}
            </el-tag>
            <el-tag type="info" size="small">{{ labelOf(DATASET_STATUS_LABELS, dataset.status) }}</el-tag>
          </div>
        </div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="版本">{{ dataset.version || '—' }}</el-descriptions-item>
          <el-descriptions-item label="记录数">{{ formatNumber(dataset.record_count) }}</el-descriptions-item>
          <el-descriptions-item label="来源名称">{{ dataset.source_name || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="来源年份">{{ dataset.source_year ?? '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="来源链接" :span="2">
            <el-link v-if="dataset.source_url" type="primary" :href="dataset.source_url" target="_blank" rel="noopener noreferrer">
              {{ dataset.source_url }}
            </el-link>
            <span v-else class="muted">未填写</span>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(dataset.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            <span class="pre-wrap">{{ dataset.description || '未填写' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="统计方法" :span="2">
            <span class="pre-wrap">{{ dataset.methodology || '未填写' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="免责声明" :span="2">
            <span class="pre-wrap">{{ dataset.disclaimer || '未填写' }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </template>
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

.tag-row {
  display: flex;
  gap: 8px;
}
</style>
