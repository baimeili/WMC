<!-- 审计日志：时间、操作人、动作、对象、详情。仅可查看，不可修改。 -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listAuditLogs } from '@/api/audit'
import type { AuditLog } from '@/api/types'
import { formatDateTime, prettifyJson } from '@/utils/format'

const loading = ref(false)
const logs = ref<AuditLog[]>([])
const limit = ref(50)
const offset = ref(0)
const detailVisible = ref(false)
const detailTarget = ref<AuditLog | null>(null)

const page = computed(() => Math.floor(offset.value / limit.value) + 1)
const hasPrev = computed(() => offset.value > 0)
const hasNext = computed(() => logs.value.length >= limit.value)

async function load(): Promise<void> {
  loading.value = true
  try {
    logs.value = await listAuditLogs({ limit: limit.value, offset: offset.value })
  } catch {
    logs.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

function prevPage(): void {
  if (!hasPrev.value) {
    return
  }
  offset.value = Math.max(0, offset.value - limit.value)
  void load()
}

function nextPage(): void {
  if (!hasNext.value) {
    ElMessage.info('已经到最后一页')
    return
  }
  offset.value += limit.value
  void load()
}

function handleLimitChange(value: number): void {
  limit.value = value
  offset.value = 0
  void load()
}

function openDetail(log: AuditLog): void {
  detailTarget.value = log
  detailVisible.value = true
}

function actorText(log: AuditLog): string {
  return log.actor_user_id || '系统'
}

function targetText(log: AuditLog): string {
  if (!log.target_type && !log.target_id) {
    return '—'
  }
  return `${log.target_type || '未分类'} / ${log.target_id || '—'}`
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">审计日志</h1>
        <p class="page-desc">
          记录冻结用户、处理举报、修改基准数据、修改系统设置等敏感操作；日志只读，不包含手机号、令牌等敏感字段。
        </p>
      </div>
      <div class="page-head-actions">
        <el-select :model-value="limit" style="width: 120px" @change="handleLimitChange">
          <el-option :value="20" label="每页 20 条" />
          <el-option :value="50" label="每页 50 条" />
          <el-option :value="100" label="每页 100 条" />
        </el-select>
        <el-button size="small" @click="load">刷新</el-button>
      </div>
    </div>

    <div class="page-card">
      <el-table :data="logs" border size="default">
        <el-table-column label="时间" width="190">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作人" min-width="200">
          <template #default="{ row }">
            <span class="mono">{{ actorText(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="动作" min-width="170">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.action || '—' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="对象" min-width="220">
          <template #default="{ row }">
            <span class="mono">{{ targetText(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="详情" min-width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无审计日志" :image-size="60" />
        </template>
      </el-table>

      <div class="pager-row">
        <span class="muted">第 {{ page }} 页</span>
        <div>
          <el-button size="small" :disabled="!hasPrev" @click="prevPage">上一页</el-button>
          <el-button size="small" :disabled="!hasNext" @click="nextPage">下一页</el-button>
        </div>
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="审计详情" width="640px" top="8vh">
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="时间">{{ formatDateTime(detailTarget?.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="操作人">{{ actorText(detailTarget ?? {}) }}</el-descriptions-item>
        <el-descriptions-item label="动作">{{ detailTarget?.action || '—' }}</el-descriptions-item>
        <el-descriptions-item label="对象">{{ targetText(detailTarget ?? {}) }}</el-descriptions-item>
      </el-descriptions>
      <p class="mt-16 muted">detail：</p>
      <pre class="json-box">{{ prettifyJson(detailTarget?.detail) }}</pre>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
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
  align-items: center;
  gap: 8px;
  flex: none;
}

.mono {
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
}

.pager-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
}
</style>
