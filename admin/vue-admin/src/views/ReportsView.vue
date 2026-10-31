<!-- 举报处理：状态筛选、查看被举报用户、填写备注并处理。 -->
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listReports, resolveReport } from '@/api/reports'
import type { Report } from '@/api/types'
import { formatDateTime } from '@/utils/format'
import {
  REPORT_REASON_LABELS,
  REPORT_STATUS_LABELS,
  REPORT_STATUS_OPTIONS,
  labelOf,
  reportStatusTagType
} from '@/utils/enums'

const router = useRouter()
const loading = ref(false)
const reports = ref<Report[]>([])
const status = ref<number | ''>('')

const resolveVisible = ref(false)
const resolveTarget = ref<Report | null>(null)
const resolving = ref(false)
const resolveForm = reactive({
  status: 3,
  note: ''
})

const RESOLVE_OPTIONS = [
  { value: 2, label: '处理中' },
  { value: 3, label: '已处理' },
  { value: 4, label: '已驳回' }
]

async function load(): Promise<void> {
  loading.value = true
  try {
    reports.value = await listReports(status.value === '' ? undefined : Number(status.value))
  } catch {
    reports.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

function handleSearch(): void {
  void load()
}

function handleReset(): void {
  status.value = ''
  void load()
}

function goUser(report: Report): void {
  if (!report.reported_user_id) {
    ElMessage.warning('该举报缺少被举报用户 id')
    return
  }
  void router.push(`/users/${report.reported_user_id}`)
}

function openResolve(report: Report): void {
  if (!report.id) {
    ElMessage.warning('该举报缺少 id，无法处理')
    return
  }
  resolveTarget.value = report
  resolveForm.status = 3
  resolveForm.note = ''
  resolveVisible.value = true
}

async function submitResolve(): Promise<void> {
  const target = resolveTarget.value
  if (!target?.id) {
    return
  }
  const note = resolveForm.note.trim()
  if (!note) {
    ElMessage.warning('请填写处理备注，备注会写入审计日志')
    return
  }
  resolving.value = true
  try {
    await resolveReport(target.id, { status: Number(resolveForm.status), note })
    ElMessage.success('举报已处理')
    resolveVisible.value = false
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    resolving.value = false
  }
}

function canResolve(report: Report): boolean {
  return report.status === 1 || report.status === 2
}
</script>

<template>
  <div class="page" v-loading="loading">
    <h1 class="page-title">举报处理</h1>
    <p class="page-desc">
      处理用户举报并填写处理备注；涉及诈骗、骚扰等严重情形时，请同时前往用户详情冻结账号。
    </p>

    <div class="page-card">
      <div class="filter-bar">
        <el-select v-model="status" placeholder="全部状态" clearable style="width: 180px">
          <el-option v-for="option in REPORT_STATUS_OPTIONS" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <el-table :data="reports" border size="default">
        <el-table-column label="被举报人" min-width="160">
          <template #default="{ row }">
            <div>{{ row.reported_nickname || '未提供昵称' }}</div>
            <div class="muted">{{ row.reported_user_id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="举报原因" width="130" align="center">
          <template #default="{ row }">{{ labelOf(REPORT_REASON_LABELS, row.reason) }}</template>
        </el-table-column>
        <el-table-column prop="description" label="举报描述" min-width="220" show-overflow-tooltip />
        <el-table-column label="举报人" min-width="140">
          <template #default="{ row }"><span class="muted">{{ row.reporter_user_id || '—' }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="reportStatusTagType(row.status)" size="small">
              {{ labelOf(REPORT_STATUS_LABELS, row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="goUser(row)">查看用户</el-button>
            <el-button v-if="canResolve(row)" link type="warning" @click="openResolve(row)">处理</el-button>
            <span v-else class="muted">已结束</span>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无举报数据" :image-size="60" />
        </template>
      </el-table>
    </div>

    <el-dialog v-model="resolveVisible" title="处理举报" width="520px">
      <p class="muted">
        被举报用户：{{ resolveTarget?.reported_nickname || resolveTarget?.reported_user_id || '—' }}；举报原因：{{
          labelOf(REPORT_REASON_LABELS, resolveTarget?.reason)
        }}
      </p>
      <el-form label-position="top" class="mt-16">
        <el-form-item label="处理结果" required>
          <el-select v-model="resolveForm.status" style="width: 100%">
            <el-option v-for="option in RESOLVE_OPTIONS" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理备注" required>
          <el-input
            v-model="resolveForm.note"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="请填写处理备注，如核实情况、采取的措施等；备注会写入审计日志"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resolveVisible = false">取消</el-button>
        <el-button type="primary" :loading="resolving" @click="submitResolve">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>
