<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-card">
        <EmptyState title="登录后查看我的 Match" description="只有互相感兴趣的人才会出现在这里。" action-text="去登录" @action="goLogin" />
      </view>

      <template v-else>
        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>我的 Match</text>
            <text class="wmc-tag">互相感兴趣</text>
          </view>
          <text class="wmc-hint">你们已经互相感兴趣，才会出现在这里。匹配结果仅根据双方填写条件计算，不代表婚姻成功概率。</text>
          <view class="wmc-notice chat-notice">
            聊天功能暂未开放。当前版本可以查看 Match 详情、记录真实进展、取消 Match 或举报屏蔽。
          </view>
        </view>

        <view v-if="!loggedIn" class="wmc-card"></view>
        <view v-else-if="loading" class="wmc-card">
          <text class="wmc-hint">正在加载 Match 列表…</text>
        </view>
        <view v-else-if="loadError" class="wmc-card">
          <text class="wmc-error">{{ loadError }}</text>
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="load">重新加载</button>
        </view>
        <template v-else>
          <view v-for="match in matches" :key="match.id" class="match-card" @tap="openMatch(match)">
            <view class="match-head">
              <image v-if="match.peer_photo_url" class="match-avatar" :src="match.peer_photo_url" mode="aspectFill" />
              <view v-else class="match-avatar match-avatar--placeholder">
                <text class="match-avatar-text">{{ (match.peer_nickname || '用').slice(0, 1) }}</text>
              </view>
              <view class="match-info">
                <view class="match-name-row">
                  <text class="match-name">{{ match.peer_nickname || '未填写昵称' }}</text>
                  <text class="match-age">{{ ageText(match.peer_age) }}</text>
                </view>
                <text class="match-meta">{{ match.peer_city_name || '城市未填写' }}</text>
                <text class="match-meta">匹配时间：{{ formatDateTime(match.matched_at) }}</text>
              </view>
              <text class="wmc-tag" :class="statusClass(match.status)">{{ statusLabel(match.status) }}</text>
            </view>
            <view class="match-footer">
              <text v-if="match.confirmed_meeting" class="wmc-tag wmc-tag--success">双方已确认线下见面</text>
              <text class="wmc-link">查看详情</text>
            </view>
          </view>

          <EmptyState
            v-if="matches.length === 0"
            title="还没有互相感兴趣的人"
            description="继续浏览双向基础候选人，互相喜欢后就会生成 Match。"
            action-text="去看看候选人"
            @action="goDiscovery"
          />
          <view v-if="matches.length > 0 && hasMore" class="load-more" @tap="loadMore">
            <text class="wmc-link">{{ loadingMore ? '加载中…' : '加载更多' }}</text>
          </view>
        </template>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// Match 列表：互相感兴趣的用户；聊天功能下线时显示「聊天功能暂未开放」。
import { computed, ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import EmptyState from '@/components/EmptyState.vue'
import { getMatches } from '@/api/match'
import { useAuthStore } from '@/stores/auth'
import { MATCH_STATUS_LABELS } from '@/utils/constants'
import { ageText, formatDateTime } from '@/utils/format'
import { errorMessage } from '@/utils/error'
import type { MatchSummary } from '@/api/types'

const auth = useAuthStore()
const loggedIn = computed(() => auth.isLoggedIn)
const loading = ref(false)
const loadingMore = ref(false)
const loadError = ref('')
const matches = ref<MatchSummary[]>([])
const nextCursor = ref('')
const hasMore = ref(true)

onShow(async () => {
  auth.syncFromStorage()
  if (!auth.isLoggedIn) return
  await load()
})

onReachBottom(() => {
  if (hasMore.value && !loadingMore.value) loadMore()
})

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const page = await getMatches({ limit: 20 })
    matches.value = page.items || []
    nextCursor.value = page.next_cursor || ''
    hasMore.value = Boolean(page.has_more)
  } catch (error) {
    loadError.value = errorMessage(error, 'Match 列表加载失败')
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const page = await getMatches({ cursor: nextCursor.value || undefined, limit: 20 })
    matches.value = matches.value.concat(page.items || [])
    nextCursor.value = page.next_cursor || ''
    hasMore.value = Boolean(page.has_more)
  } catch (error) {
    uni.showToast({ title: errorMessage(error, '加载更多失败'), icon: 'none' })
  } finally {
    loadingMore.value = false
  }
}

function statusLabel(status: number): string {
  return MATCH_STATUS_LABELS[status] || '未知状态'
}

function statusClass(status: number): string {
  if (status === 1) return 'wmc-tag--success'
  if (status === 3) return 'wmc-tag--danger'
  return 'wmc-tag--muted'
}

function openMatch(match: MatchSummary) {
  uni.navigateTo({ url: '/pages/matches/detail?id=' + encodeURIComponent(match.id) })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/matches/index') })
}

function goDiscovery() {
  uni.navigateTo({ url: '/pages/discovery/index' })
}
</script>

<style lang="scss" scoped>
.chat-notice {
  margin-top: 20rpx;
}
.match-card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-shadow: $wmc-shadow;
}
.match-head {
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.match-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 16rpx;
  background: #eef2f7;
  flex-shrink: 0;
}
.match-avatar--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(47, 107, 255, 0.1);
}
.match-avatar-text {
  font-size: 38rpx;
  color: $uni-color-primary;
  font-weight: 600;
}
.match-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
  min-width: 0;
}
.match-name-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}
.match-name {
  font-size: 30rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.match-age {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.match-meta {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.match-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16rpx;
}
.load-more {
  text-align: center;
  padding: 20rpx 0 40rpx;
}
.retry-btn {
  margin-top: 20rpx;
}
</style>
