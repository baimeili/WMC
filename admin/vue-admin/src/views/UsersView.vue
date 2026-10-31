<!-- 用户管理：分页 + 关键词/状态筛选 + 冻结/解冻（需填写原因）+ 跳转详情。 -->
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { freezeUser, listUsers, unfreezeUser } from '@/api/users'
import type { AdminUserItem } from '@/api/types'
import { formatDateTime, formatRatioOf, clampPercent } from '@/utils/format'
import {
  labelOf,
  ROLE_LABELS,
  USER_STATUS_LABELS,
  USER_STATUS_OPTIONS,
  userStatusTagType
} from '@/utils/enums'

const router = useRouter()

const loading = ref(false)
const users = ref<AdminUserItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const status = ref<number | ''>('')

// 冻结弹窗状态。
const freezeVisible = ref(false)
const freezeTarget = ref<AdminUserItem | null>(null)
const freezeReason = ref('')
const freezing = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    const result = await listUsers({
      keyword: keyword.value.trim() || undefined,
      status: status.value === '' ? undefined : Number(status.value),
      page: page.value,
      page_size: pageSize.value
    })
    users.value = result.items ?? []
    total.value = result.total ?? users.value.length
  } catch {
    users.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

onMounted(load)

function handleSearch(): void {
  page.value = 1
  void load()
}

function handleReset(): void {
  keyword.value = ''
  status.value = ''
  page.value = 1
  void load()
}

function handlePageChange(nextPage: number): void {
  page.value = nextPage
  void load()
}

function handleSizeChange(nextSize: number): void {
  pageSize.value = nextSize
  page.value = 1
  void load()
}

function goDetail(user: AdminUserItem): void {
  if (!user.id) {
    ElMessage.warning('该用户缺少 id，无法查看详情')
    return
  }
  void router.push(`/users/${user.id}`)
}

function openFreeze(user: AdminUserItem): void {
  if (!user.id) {
    ElMessage.warning('该用户缺少 id，无法冻结')
    return
  }
  freezeTarget.value = user
  freezeReason.value = ''
  freezeVisible.value = true
}

async function submitFreeze(): Promise<void> {
  const target = freezeTarget.value
  if (!target?.id) {
    return
  }
  const reason = freezeReason.value.trim()
  if (!reason) {
    ElMessage.warning('请填写冻结原因，原因会写入审计日志')
    return
  }
  freezing.value = true
  try {
    await freezeUser(target.id, reason)
    ElMessage.success('已冻结该用户')
    freezeVisible.value = false
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    freezing.value = false
  }
}

async function handleUnfreeze(user: AdminUserItem): Promise<void> {
  if (!user.id) {
    return
  }
  try {
    await ElMessageBox.confirm('解冻后该用户可正常使用平台功能，确定解冻吗？', '解冻用户', {
      type: 'warning',
      confirmButtonText: '解冻',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await unfreezeUser(user.id)
    ElMessage.success('已解冻该用户')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

function completionPercent(user: AdminUserItem): number {
  return clampPercent(user.profile_completion)
}

function completionHint(user: AdminUserItem): string {
  return formatRatioOf(user.profile_completion, 100, 0, '0%')
}
</script>

<template>
  <div class="page" v-loading="loading">
    <h1 class="page-title">用户管理</h1>
    <p class="page-desc">
      查询与冻结用户；管理端默认不可见完整手机号，冻结原因与操作人会写入审计日志。
    </p>

    <div class="page-card">
      <div class="filter-bar">
        <el-input
          v-model="keyword"
          placeholder="昵称或手机号（脱敏）关键词"
          clearable
          style="width: 240px"
          @keyup.enter="handleSearch"
        />
        <el-select v-model="status" placeholder="全部状态" clearable style="width: 160px">
          <el-option v-for="option in USER_STATUS_OPTIONS" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <el-table :data="users" border size="default">
        <el-table-column label="昵称" min-width="140">
          <template #default="{ row }">{{ row.nickname || '未设置' }}</template>
        </el-table-column>
        <el-table-column label="手机号" min-width="130">
          <template #default="{ row }">{{ row.phone_masked || '—' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="userStatusTagType(row.status)" size="small">
              {{ labelOf(USER_STATUS_LABELS, row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="120" align="center">
          <template #default="{ row }">{{ labelOf(ROLE_LABELS, row.role) }}</template>
        </el-table-column>
        <el-table-column label="入池" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.discovery_enabled" type="success" size="small">已入池</el-tag>
            <el-tag v-else type="info" size="small">未入池</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="资料完成度" width="170">
          <template #default="{ row }">
            <el-progress :percentage="completionPercent(row)" :stroke-width="8" />
            <span class="muted">{{ completionHint(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="种子用户" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_seed" type="warning" size="small">种子</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="goDetail(row)">详情</el-button>
            <el-button v-if="row.status === 3" link type="success" @click="handleUnfreeze(row)">解冻</el-button>
            <el-button v-else link type="danger" @click="openFreeze(row)">冻结</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无用户数据" :image-size="60" />
        </template>
      </el-table>

      <div class="mt-16">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          background
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-dialog v-model="freezeVisible" title="冻结用户" width="480px">
      <p class="muted">
        目标用户：{{ freezeTarget?.nickname || freezeTarget?.id || '—' }}（{{ freezeTarget?.phone_masked || '手机号不可见' }}）
      </p>
      <el-alert
        type="warning"
        :closable="false"
        class="mt-12"
        title="冻结后该用户无法登录与使用婚恋池，操作将写入审计日志。"
      />
      <el-form label-position="top" class="mt-16">
        <el-form-item label="冻结原因" required>
          <el-input
            v-model="freezeReason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请填写冻结原因（如：多次被举报骚扰、资料造假等）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="freezeVisible = false">取消</el-button>
        <el-button type="danger" :loading="freezing" @click="submitFreeze">确认冻结</el-button>
      </template>
    </el-dialog>
  </div>
</template>
