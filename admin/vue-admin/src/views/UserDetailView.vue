<!-- 用户详情：资料摘要、入池状态、举报记录，并可冻结/解冻。 -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { findUserInList, freezeUser, getUserDetail, unfreezeUser } from '@/api/users'
import { listReportsByUser } from '@/api/reports'
import { isNotFound } from '@/api/http'
import type { AdminUserDetail, Report } from '@/api/types'
import { clampPercent, formatDateTime, formatRatioOf } from '@/utils/format'
import {
  EDUCATION_LABELS,
  GENDER_LABELS,
  MARITAL_STATUS_LABELS,
  REPORT_REASON_LABELS,
  REPORT_STATUS_LABELS,
  ROLE_LABELS,
  USER_STATUS_LABELS,
  WANTS_CHILDREN_LABELS,
  labelOf,
  reportStatusTagType,
  userStatusTagType
} from '@/utils/enums'

const route = useRoute()
const router = useRouter()

const id = computed(() => String(route.params.id ?? ''))
const loading = ref(false)
const user = ref<AdminUserDetail | null>(null)
const reports = ref<Report[]>([])
const loadError = ref('')
const detailFallback = ref(false)

const freezeVisible = ref(false)
const freezeReason = ref('')
const freezing = ref(false)

async function loadUser(): Promise<void> {
  if (!id.value) {
    loadError.value = '缺少用户 id'
    return
  }
  loading.value = true
  loadError.value = ''
  detailFallback.value = false
  try {
    user.value = normalizeDetail(await getUserDetail(id.value))
  } catch (error) {
    if (isNotFound(error)) {
      // 契约未提供单用户详情端点时的兜底：按 id 检索用户列表。
      try {
        const fallback = await findUserInList(id.value)
        if (fallback) {
          user.value = normalizeDetail(fallback)
          detailFallback.value = true
        } else {
          user.value = null
          loadError.value = '未找到该用户，或后端尚未提供单用户详情接口。'
        }
      } catch {
        user.value = null
        loadError.value = '加载用户详情失败，请稍后重试。'
      }
    } else {
      user.value = null
      loadError.value = error instanceof Error ? error.message : '加载用户详情失败'
    }
  } finally {
    loading.value = false
  }
}

async function loadReports(): Promise<void> {
  if (!id.value) {
    return
  }
  try {
    reports.value = await listReportsByUser(id.value)
  } catch {
    reports.value = []
  }
}

// 后端当前返回扁平资料字段，统一归一化为 profile 供模板展示。
function normalizeDetail(detail: AdminUserDetail): AdminUserDetail {
  if (detail.profile) {
    return detail
  }
  const hasFlatProfile =
    detail.age !== undefined ||
    detail.gender !== undefined ||
    detail.city_code !== undefined ||
    detail.education_level !== undefined ||
    detail.marital_status !== undefined
  if (!hasFlatProfile) {
    return detail
  }
  return {
    ...detail,
    profile: {
      nickname: detail.nickname,
      gender: detail.gender,
      age: detail.age,
      city_code: detail.city_code,
      education_level: detail.education_level,
      marital_status: detail.marital_status,
      profile_completion: detail.profile_completion
    }
  }
}

onMounted(async () => {
  await loadUser()
  await loadReports()
})

const completion = computed(() => clampPercent(user.value?.profile_completion ?? user.value?.profile?.profile_completion))

async function handleUnfreeze(): Promise<void> {
  if (!id.value) {
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
    await unfreezeUser(id.value)
    ElMessage.success('已解冻该用户')
    await loadUser()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

function openFreeze(): void {
  freezeReason.value = ''
  freezeVisible.value = true
}

async function submitFreeze(): Promise<void> {
  if (!id.value) {
    return
  }
  const reason = freezeReason.value.trim()
  if (!reason) {
    ElMessage.warning('请填写冻结原因，原因会写入审计日志')
    return
  }
  freezing.value = true
  try {
    await freezeUser(id.value, reason)
    ElMessage.success('已冻结该用户')
    freezeVisible.value = false
    await loadUser()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    freezing.value = false
  }
}

function goReports(): void {
  void router.push('/reports')
}

function goBack(): void {
  void router.push('/users')
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">用户详情</h1>
        <p class="page-desc">用户 ID：{{ id || '—' }}；管理端默认不可见完整手机号，敏感操作写入审计日志。</p>
      </div>
      <div class="page-head-actions">
        <el-button size="small" @click="goBack">返回列表</el-button>
        <el-button size="small" @click="goReports">举报处理</el-button>
      </div>
    </div>

    <el-result v-if="loadError && !user" icon="warning" title="无法加载用户" :sub-title="loadError">
      <template #extra>
        <el-button type="primary" @click="loadUser">重试</el-button>
        <el-button @click="goBack">返回用户列表</el-button>
      </template>
    </el-result>

    <template v-else-if="user">
      <el-alert
        v-if="detailFallback"
        type="info"
        :closable="false"
        show-icon
        class="mb-12"
        title="当前展示基础信息"
        description="后端尚未提供单用户详情端点（OpenAPI 未覆盖），资料摘要可能不完整；接口补齐后本页自动展示完整数据。"
      />

      <div class="page-card">
        <div class="card-title-row">
          <h2 class="card-title">账号信息</h2>
          <div>
            <el-button v-if="user.status === 3" size="small" type="success" @click="handleUnfreeze">解冻用户</el-button>
            <el-button v-else size="small" type="danger" @click="openFreeze">冻结用户</el-button>
          </div>
        </div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="昵称">{{ user.nickname || user.profile?.nickname || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ user.phone_masked || '不可见（隐私保护）' }}</el-descriptions-item>
          <el-descriptions-item label="用户 ID">{{ user.id || '—' }}</el-descriptions-item>
          <el-descriptions-item label="角色">{{ user.role_label || labelOf(ROLE_LABELS, user.role) }}</el-descriptions-item>
          <el-descriptions-item label="账号状态">
            <el-tag :type="userStatusTagType(user.status)" size="small">
              {{ user.status_label || labelOf(USER_STATUS_LABELS, user.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="入池状态">
            <el-tag v-if="user.discovery_enabled" type="success" size="small">已加入真实婚恋池</el-tag>
            <el-tag v-else type="info" size="small">未加入婚恋池</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="资料完成度">
            <el-progress :percentage="completion" :stroke-width="8" style="max-width: 220px" />
            <span class="muted">{{ formatRatioOf(user.profile_completion, 100, 0, '0%') }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="种子用户">{{ user.is_seed ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ formatDateTime(user.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="最近登录">{{ formatDateTime(user.last_login_at) }}</el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-if="!user.discovery_enabled"
          type="info"
          :closable="false"
          class="mt-16"
          title="该用户未加入真实婚恋池"
          description="入池需同时满足：账户状态正常、资料完成度达到阈值、年满 18 岁、主动同意婚恋市场参与声明。"
        />
      </div>

      <div class="page-card">
        <h2 class="card-title">资料摘要</h2>
        <el-descriptions v-if="user.profile" :column="2" border>
          <el-descriptions-item label="性别">{{ labelOf(GENDER_LABELS, user.profile.gender) }}</el-descriptions-item>
          <el-descriptions-item label="年龄">{{ user.profile.age ?? '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="城市">{{ user.profile.city_name || user.profile.city_code || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="身高">
            {{ user.profile.height_cm ? user.profile.height_cm + ' cm' : '未填写' }}
          </el-descriptions-item>
          <el-descriptions-item label="学历">{{ labelOf(EDUCATION_LABELS, user.profile.education_level) }}</el-descriptions-item>
          <el-descriptions-item label="婚姻状态">{{ labelOf(MARITAL_STATUS_LABELS, user.profile.marital_status) }}</el-descriptions-item>
          <el-descriptions-item label="子女数量">{{ user.profile.children_count ?? '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="生育意愿">{{ labelOf(WANTS_CHILDREN_LABELS, user.profile.wants_children) }}</el-descriptions-item>
          <el-descriptions-item label="资料更新时间" :span="2">{{ formatDateTime(user.profile.updated_at) }}</el-descriptions-item>
          <el-descriptions-item label="个人简介" :span="2">
            <span class="pre-wrap">{{ user.profile.bio || '未填写' }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <el-empty v-else description="暂无资料摘要（后端详情接口未返回 profile 字段）" :image-size="60" />
      </div>

      <div class="page-card">
        <div class="card-title-row">
          <h2 class="card-title">举报记录</h2>
          <el-button link type="primary" @click="goReports">前往举报处理</el-button>
        </div>
        <el-table v-if="reports.length > 0" :data="reports" border size="small">
          <el-table-column label="举报原因" width="140">
            <template #default="{ row }">{{ labelOf(REPORT_REASON_LABELS, row.reason) }}</template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="220" show-overflow-tooltip />
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
        </el-table>
        <el-empty v-else description="该用户暂无举报记录" :image-size="60" />
      </div>
    </template>

    <el-dialog v-model="freezeVisible" title="冻结用户" width="480px">
      <el-alert
        type="warning"
        :closable="false"
        title="冻结后该用户无法登录与使用婚恋池，操作会写入审计日志。"
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
</style>