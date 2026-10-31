<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="loading" class="wmc-card">
        <text class="wmc-hint">正在加载 Match 详情…</text>
      </view>
      <view v-else-if="loadError" class="wmc-card">
        <text class="wmc-error">{{ loadError }}</text>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="load">重新加载</button>
      </view>
      <template v-else-if="match">
        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>你们已经互相感兴趣</text>
            <text class="wmc-tag" :class="statusClass">{{ statusLabel }}</text>
          </view>
          <view class="peer-head">
            <image v-if="match.peer_photo_url" class="peer-avatar" :src="match.peer_photo_url" mode="aspectFill" />
            <view v-else class="peer-avatar peer-avatar--placeholder">
              <text class="peer-avatar-text">{{ avatarText }}</text>
            </view>
            <view class="peer-info">
              <text class="peer-name">{{ match.peer_nickname || '未填写昵称' }}</text>
              <text class="peer-meta">{{ ageText(match.peer_age) }} · {{ match.peer_city_name || '城市未填写' }}</text>
              <text class="peer-meta">匹配时间：{{ formatDateTime(match.matched_at) }}</text>
            </view>
          </view>
          <view v-if="match.confirmed_meeting" class="wmc-notice success-notice">双方均已反馈已线下见面。</view>
          <text class="match-tip">匹配分数仅根据双方填写条件计算，不代表婚姻成功概率。请理性沟通，注意个人信息安全。</text>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>聊天</text>
          </view>
          <text class="wmc-hint">{{ chatText }}</text>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>记录真实进展</text>
            <text class="wmc-tag wmc-tag--muted">可选</text>
          </view>
          <text class="wmc-hint">真实进展反馈仅用于平台改进匹配质量，不会公开给对方看具体是谁提交的。</text>
          <view class="outcome-list">
            <view
              v-for="item in outcomeOptions"
              :key="item.value"
              class="wmc-chip"
              :class="{ 'wmc-chip--active': myOutcomes.indexOf(item.value) >= 0 }"
              @tap="submitOutcome(item)"
            >
              {{ item.label }}
            </view>
          </view>
        </view>

        <view class="wmc-card">
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block" @tap="cancelMatch">取消 Match</button>
          <view class="safety-row">
            <text class="wmc-link" @tap="goReport">举报该用户</text>
            <text class="wmc-link" @tap="block">屏蔽该用户</text>
          </view>
          <text class="wmc-hint">
            取消 Match 后双方会从 Match 列表中移除；屏蔽后双方不可见、不可互动。举报内容仅用于平台处理。
          </text>
        </view>
      </template>
      <view v-else class="wmc-card">
        <EmptyState title="Match 不可用" description="该 Match 可能已取消、被屏蔽或不存在。" action-text="返回 Match 列表" @action="back" />
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// Match 详情：互相感兴趣提示、聊天未开放说明、取消 Match、举报与屏蔽、真实进展反馈。
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import EmptyState from '@/components/EmptyState.vue'
import { getMatch } from '@/api/match'
import { reportOutcome, setInteraction } from '@/api/interaction'
import { blockUser } from '@/api/safety'
import { useAuthStore } from '@/stores/auth'
import { MATCH_STATUS_LABELS } from '@/utils/constants'
import { ageText, formatDateTime } from '@/utils/format'
import { errorMessage } from '@/utils/error'
import { ensureOptions, pickerOptions } from '@/utils/options'
import type { MatchDetail } from '@/api/types'

const auth = useAuthStore()
const match = ref<MatchDetail | null>(null)
const loading = ref(true)
const loadError = ref('')
const acting = ref(false)
const matchId = ref('')

const avatarText = computed(() => (match.value && match.value.peer_nickname ? match.value.peer_nickname.slice(0, 1) : '用'))
const statusLabel = computed(() => (match.value ? MATCH_STATUS_LABELS[match.value.status] || '未知状态' : ''))
const statusClass = computed(() => {
  if (!match.value) return 'wmc-tag--muted'
  if (match.value.status === 1) return 'wmc-tag--success'
  if (match.value.status === 3) return 'wmc-tag--danger'
  return 'wmc-tag--muted'
})
const myOutcomes = computed(() => (match.value ? match.value.my_outcomes || [] : []))
const chatText = computed(() => {
  if (match.value && match.value.chat_enabled) {
    return '聊天功能由平台另行开通，当前版本暂不提供聊天入口。你可以先记录真实进展。'
  }
  return '聊天功能暂未开放。当前版本可以记录真实进展、取消 Match 或举报屏蔽。'
})
const outcomeOptions = computed(() => pickerOptions('outcome_types', false))

onLoad(async (options) => {
  matchId.value = options && options.id ? decodeURIComponent(options.id) : ''
  auth.syncFromStorage()
  if (!auth.isLoggedIn) {
    uni.redirectTo({ url: '/pages/auth/login' })
    return
  }
  await ensureOptions()
  await load()
})

async function load() {
  if (!matchId.value) {
    loadError.value = '缺少 Match 标识'
    loading.value = false
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    match.value = await getMatch(matchId.value)
  } catch (error) {
    loadError.value = errorMessage(error, 'Match 详情加载失败')
  } finally {
    loading.value = false
  }
}

async function submitOutcome(item: { value: number; label: string }) {
  if (!match.value || acting.value) return
  const confirmed = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: '记录真实进展',
      content: '将记录「' + item.label + '」，是否继续？',
      confirmText: '记录',
      cancelText: '取消',
      success: (response) => resolve(Boolean(response.confirm))
    })
  })
  if (!confirmed) return
  acting.value = true
  try {
    await reportOutcome(match.value.id, { outcome_type: item.value })
    uni.showToast({ title: '已记录', icon: 'none' })
    await load()
  } catch (error) {
    uni.showToast({ title: errorMessage(error, '记录失败'), icon: 'none' })
  } finally {
    acting.value = false
  }
}

function cancelMatch() {
  if (!match.value || acting.value) return
  uni.showModal({
    title: '取消 Match',
    content: '取消后双方会从 Match 列表中移除，且不再互相可见于 Match。确定继续吗？',
    confirmText: '确认取消',
    cancelText: '返回',
    success: async (response) => {
      if (!response.confirm) return
      acting.value = true
      try {
        await setInteraction(match.value!.peer_user_id, 2, 3)
        uni.showToast({ title: '已取消 Match', icon: 'none' })
        uni.navigateBack({ delta: 1 })
      } catch (error) {
        uni.showToast({ title: errorMessage(error, '取消失败'), icon: 'none' })
      } finally {
        acting.value = false
      }
    }
  })
}

function goReport() {
  if (!match.value) return
  uni.navigateTo({ url: '/pages/report/index?target=' + encodeURIComponent(match.value.peer_user_id) })
}

function block() {
  if (!match.value) return
  uni.showModal({
    title: '屏蔽该用户',
    content: '屏蔽后双方不会再互相看到或互动，确定继续吗？',
    confirmText: '确认屏蔽',
    cancelText: '取消',
    success: async (response) => {
      if (!response.confirm) return
      try {
        await blockUser(match.value!.peer_user_id)
        uni.showToast({ title: '已屏蔽该用户', icon: 'none' })
        uni.navigateBack({ delta: 1 })
      } catch (error) {
        uni.showToast({ title: errorMessage(error, '屏蔽失败'), icon: 'none' })
      }
    }
  })
}

function back() {
  uni.navigateBack({ delta: 1 })
}
</script>

<style lang="scss" scoped>
.peer-head {
  display: flex;
  align-items: center;
  gap: 24rpx;
  margin-top: 8rpx;
}
.peer-avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 20rpx;
  background: #eef2f7;
  flex-shrink: 0;
}
.peer-avatar--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(47, 107, 255, 0.1);
}
.peer-avatar-text {
  font-size: 44rpx;
  color: $uni-color-primary;
  font-weight: 600;
}
.peer-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.peer-name {
  font-size: 34rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.peer-meta {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.success-notice {
  margin-top: 20rpx;
}
.match-tip {
  display: block;
  margin-top: 20rpx;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.outcome-list {
  display: flex;
  flex-wrap: wrap;
  margin-top: 16rpx;
}
.safety-row {
  display: flex;
  gap: 40rpx;
  margin: 24rpx 0 12rpx;
}
.retry-btn {
  margin-top: 20rpx;
}
</style>
