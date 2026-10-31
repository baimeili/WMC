<!-- 照片审核：待审核照片列表，支持通过 / 拒绝（填写拒绝原因）。 -->
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { approvePhoto, listPhotos, rejectPhoto } from '@/api/photos'
import type { AdminPhoto } from '@/api/types'
import { formatDateTime } from '@/utils/format'
import { MODERATION_STATUS_LABELS, MODERATION_STATUS_OPTIONS, labelOf } from '@/utils/enums'

const loading = ref(false)
const photos = ref<AdminPhoto[]>([])
const status = ref<number | ''>(1)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loadError = ref('')

const rejectVisible = ref(false)
const rejectTarget = ref<AdminPhoto | null>(null)
const rejectReason = ref('')
const rejecting = ref(false)

async function load(): Promise<void> {
  loading.value = true
  loadError.value = ''
  try {
    const result = await listPhotos({
      status: status.value === '' ? undefined : Number(status.value),
      page: page.value,
      page_size: pageSize.value
    })
    photos.value = result.items ?? []
    total.value = result.total ?? photos.value.length
  } catch (error) {
    photos.value = []
    total.value = 0
    loadError.value = error instanceof Error ? error.message : '照片列表加载失败'
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
  status.value = 1
  page.value = 1
  void load()
}

function handlePageChange(nextPage: number): void {
  page.value = nextPage
  void load()
}

function photoUrl(photo: AdminPhoto): string {
  return photo.url || ''
}

function photoOwner(photo: AdminPhoto): string {
  return photo.nickname || photo.user_nickname || photo.user_id || '未知用户'
}

async function handleApprove(photo: AdminPhoto): Promise<void> {
  if (!photo.id) {
    ElMessage.warning('该照片缺少 id，无法审核')
    return
  }
  try {
    await ElMessageBox.confirm('通过后该照片可展示在候选卡片中，确定通过吗？', '通过审核', {
      type: 'warning',
      confirmButtonText: '通过',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await approvePhoto(photo.id)
    ElMessage.success('已通过该照片')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}

function openReject(photo: AdminPhoto): void {
  if (!photo.id) {
    ElMessage.warning('该照片缺少 id，无法审核')
    return
  }
  rejectTarget.value = photo
  rejectReason.value = ''
  rejectVisible.value = true
}

async function submitReject(): Promise<void> {
  const target = rejectTarget.value
  if (!target?.id) {
    return
  }
  const reason = rejectReason.value.trim()
  if (!reason) {
    ElMessage.warning('请填写拒绝原因，原因会用于后续申诉核对')
    return
  }
  rejecting.value = true
  try {
    await rejectPhoto(target.id, reason)
    ElMessage.success('已拒绝该照片')
    rejectVisible.value = false
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    rejecting.value = false
  }
}
</script>

<template>
  <div class="page" v-loading="loading">
    <h1 class="page-title">照片审核</h1>
    <p class="page-desc">
      审核用户上传的照片（后端已统一重编码去除 EXIF/GPS 信息）。通过后照片可展示在候选卡片；拒绝需填写原因。
      当前图片审核接口按运营后台约定路径实现（OpenAPI 尚未收录），接口未上线时会提示加载失败。
    </p>

    <div class="page-card">
      <div class="filter-bar">
        <el-select v-model="status" placeholder="全部状态" clearable style="width: 180px">
          <el-option v-for="option in MODERATION_STATUS_OPTIONS" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <el-alert v-if="loadError" type="error" :closable="false" class="mb-12" title="照片列表加载失败" :description="loadError" />

      <div v-if="photos.length > 0" class="photo-grid">
        <div v-for="photo in photos" :key="photo.id || photo.object_key || photoUrl(photo)" class="photo-card">
          <el-image
            v-if="photoUrl(photo)"
            :src="photoUrl(photo)"
            :preview-src-list="[photoUrl(photo)]"
            fit="cover"
            class="photo-thumb"
            preview-teleported
          />
          <div v-else class="photo-placeholder">暂无预览地址</div>
          <div class="photo-body">
            <p class="photo-owner">{{ photoOwner(photo) }}</p>
            <p class="muted">状态：{{ labelOf(MODERATION_STATUS_LABELS, photo.moderation_status ?? photo.status) }}</p>
            <p class="muted">上传时间：{{ formatDateTime(photo.created_at) }}</p>
            <div class="photo-actions">
              <el-button size="small" type="success" @click="handleApprove(photo)">通过</el-button>
              <el-button size="small" type="danger" @click="openReject(photo)">拒绝</el-button>
            </div>
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无照片数据" :image-size="80" />

      <div class="mt-16">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          background
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <el-dialog v-model="rejectVisible" title="拒绝照片" width="480px">
      <p class="muted">目标用户：{{ rejectTarget?.nickname || rejectTarget?.user_nickname || rejectTarget?.user_id || '—' }}</p>
      <el-form label-position="top" class="mt-16">
        <el-form-item label="拒绝原因" required>
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请填写拒绝原因（如：照片模糊、非本人、含违规内容等）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejecting" @click="submitReject">确认拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.photo-owner {
  margin: 0 0 6px;
  font-weight: 600;
  color: var(--wmc-text-primary);
}
</style>
