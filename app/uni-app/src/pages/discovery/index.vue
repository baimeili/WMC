<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-container--inner">
        <EmptyState
          title="登录后查看本站真实候选池"
          description="真实候选池只统计已注册、资料达标并主动加入婚恋池的用户。登录并完善资料后即可查看。"
          action-text="去登录"
          @action="goLogin"
        />
      </view>

      <template v-else>
        <view class="wmc-card">
          <MarketBadge market="real" />
          <view class="summary-stats">
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ reciprocalText }}</text>
              <text class="wmc-stat-label">双向基础候选人</text>
            </view>
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ oneWayText }}</text>
              <text class="wmc-stat-label">单向符合硬条件</text>
            </view>
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ completionText }}</text>
              <text class="wmc-stat-label">资料完成度</text>
            </view>
          </view>
          <view class="wmc-notice summary-note">
            双向基础候选人指双方硬条件全部满足的用户。为保护隐私，少于 10 人统一显示为 &lt;10。
          </view>
        </view>

        <view v-if="summary && summary.early_stage" class="wmc-card early-card">
          <view class="wmc-section-title">
            <text>冷启动阶段说明</text>
            <text class="wmc-tag wmc-tag--warning">早期测试</text>
          </view>
          <text class="early-text">{{ coldStartText }}</text>
        </view>

        <view v-if="summary && !summary.in_market" class="wmc-card">
          <view class="wmc-section-title">
            <text>尚未加入本站真实婚恋市场</text>
          </view>
          <text class="wmc-hint">入池后，符合你条件且你也符合对方条件的人，才会出现在双向候选池中。请先确认以下前置条件：</text>
          <view class="requirement-list">
            <view v-for="item in requirements" :key="item.label" class="requirement-item">
              <text class="requirement-item__state" :class="item.ok ? 'requirement-item__state--ok' : 'requirement-item__state--no'">
                {{ item.ok ? '已满足' : '未满足' }}
              </text>
              <text class="requirement-item__label">{{ item.label }}</text>
            </view>
          </view>
          <button
            v-if="requirementsReady"
            class="wmc-btn wmc-btn--primary wmc-btn--block join-btn"
            :disabled="joining"
            @tap="joinMarket"
          >
            {{ joining ? '加入中…' : '加入本站真实婚恋市场' }}
          </button>
          <button v-else class="wmc-btn wmc-btn--primary wmc-btn--block join-btn" @tap="goOnboarding">
            去完善资料
          </button>
          <text v-if="actionError" class="wmc-error">{{ actionError }}</text>
        </view>

        <view class="entry-grid">
          <view class="entry-item" @tap="goFunnel">
            <text class="entry-item__title">真实候选池漏斗</text>
            <text class="entry-item__desc">查看逐层筛选结果</text>
          </view>
          <view class="entry-item" @tap="goSensitivity">
            <text class="entry-item__title">条件敏感性</text>
            <text class="entry-item__desc">哪条条件影响最大</text>
          </view>
          <view class="entry-item" @tap="goMatches">
            <text class="entry-item__title">我的 Match</text>
            <text class="entry-item__desc">互相感兴趣的人</text>
          </view>
        </view>

        <view class="list-head">
          <text class="list-head__title">双向基础候选人</text>
          <text class="wmc-tag wmc-tag--muted">仅根据条件计算</text>
        </view>

        <view v-if="summaryLoading" class="wmc-card">
          <text class="wmc-hint">正在加载候选池概览…</text>
        </view>
        <view v-else-if="loadError" class="wmc-card">
          <text class="wmc-error">{{ loadError }}</text>
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="reload">重新加载</button>
        </view>
        <template v-else>
          <CandidateCard
            v-for="candidate in candidates"
            :key="candidate.user_id"
            :candidate="candidate"
            @tap="openCandidate"
          />
          <EmptyState
            v-if="candidates.length === 0"
            title="当前还没有足够的双向候选人"
            :description="coldStartText"
            action-text="去完善资料"
            @action="goOnboarding"
          />
          <view v-if="candidates.length > 0 && hasMore" class="load-more" @tap="loadMore">
            <text class="wmc-link">{{ candidatesLoading ? '加载中…' : '加载更多' }}</text>
          </view>
          <view v-if="candidates.length > 0 && !hasMore" class="load-more">
            <text class="wmc-hint">没有更多候选人了</text>
          </view>
        </template>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 本站真实候选池概览：严格区分理论市场；少于 10 人显示 <10；冷启动展示中性说明。
import { computed, ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import CandidateCard from '@/components/CandidateCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { useDiscoveryStore } from '@/stores/discovery'
import { joinMarket as joinMarketApi } from '@/api/profile'
import { COLD_START_MESSAGE, K_ANONYMITY_THRESHOLD, K_ANONYMITY_NOTE } from '@/utils/constants'
import { maskCount } from '@/utils/format'
import { errorMessage } from '@/utils/error'
import type { CandidateCardItem } from '@/api/types'

const auth = useAuthStore()
const discovery = useDiscoveryStore()

const loading = ref(false)
const summaryLoading = computed(() => discovery.summaryLoading)
const candidatesLoading = computed(() => discovery.candidatesLoading)
const loadError = ref('')
const actionError = ref('')
const joining = ref(false)

const loggedIn = computed(() => auth.isLoggedIn)
const summary = computed(() => discovery.summary)
const candidates = computed(() => discovery.candidates)
const hasMore = computed(() => discovery.hasMore)

const reciprocalText = computed(() => {
  if (!summary.value) return '—'
  if (summary.value.eligible_count_text) return summary.value.eligible_count_text
  return maskCount(summary.value.reciprocal_count)
})

const oneWayText = computed(() => {
  if (!summary.value) return '—'
  return maskCount(summary.value.one_way_count)
})

const completionText = computed(() => {
  if (!summary.value) return '—'
  return (summary.value.profile_completion || 0) + '%'
})

const coldStartText = computed(() => {
  if (summary.value && summary.value.early_stage_message) return summary.value.early_stage_message
  return COLD_START_MESSAGE
})

const requirements = computed(() => {
  if (!summary.value) return []
  const req = summary.value.requirements || { min_profile_completion: 0, age_ok: true, consent_ok: true, account_ok: true }
  return [
    { label: '资料完成度达到 ' + (req.min_profile_completion || 0) + '%（当前 ' + (summary.value.profile_completion || 0) + '%）', ok: (summary.value.profile_completion || 0) >= (req.min_profile_completion || 0) },
    { label: '年龄符合要求（18 岁以上）', ok: Boolean(req.age_ok) },
    { label: '已同意用户协议、隐私政策与婚恋市场参与声明', ok: Boolean(req.consent_ok) },
    { label: '账号状态正常', ok: Boolean(req.account_ok) }
  ]
})

const requirementsReady = computed(() => {
  if (!summary.value) return false
  const req = summary.value.requirements || { min_profile_completion: 0, age_ok: true, consent_ok: true, account_ok: true }
  return (
    (summary.value.profile_completion || 0) >= (req.min_profile_completion || 0) &&
    Boolean(req.age_ok) &&
    Boolean(req.consent_ok) &&
    Boolean(req.account_ok)
  )
})

onShow(async () => {
  auth.syncFromStorage()
  if (!auth.isLoggedIn) return
  await loadSummary()
  if (summary.value && summary.value.in_market) {
    await discovery.loadCandidates(true)
  }
})

onReachBottom(() => {
  if (!loggedIn.value) return
  if (discovery.hasMore && !discovery.candidatesLoading) {
    discovery.loadCandidates(false)
  }
})

async function loadSummary() {
  loadError.value = ''
  try {
    await discovery.loadSummary(true)
  } catch (error) {
    loadError.value = errorMessage(error, '候选池概览加载失败')
  }
}

async function reload() {
  await loadSummary()
  if (summary.value && summary.value.in_market) {
    await discovery.loadCandidates(true)
  }
}

async function loadMore() {
  if (discovery.candidatesLoading || !discovery.hasMore) return
  await discovery.loadCandidates(false)
}

async function joinMarket() {
  if (joining.value) return
  joining.value = true
  actionError.value = ''
  try {
    await joinMarketApi()
    uni.showToast({ title: '已加入真实婚恋市场', icon: 'none' })
    await loadSummary()
    await discovery.loadCandidates(true)
  } catch (error) {
    actionError.value = errorMessage(error, '加入失败，请稍后重试')
  } finally {
    joining.value = false
  }
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/discovery/index') })
}

function goOnboarding() {
  uni.navigateTo({ url: '/pages/onboarding/basic' })
}

function goFunnel() {
  uni.navigateTo({ url: '/pages/discovery/funnel' })
}

function goSensitivity() {
  uni.navigateTo({ url: '/pages/discovery/sensitivity' })
}

function goMatches() {
  uni.navigateTo({ url: '/pages/matches/index' })
}

function openCandidate(candidate: CandidateCardItem) {
  uni.navigateTo({ url: '/pages/candidate/detail?id=' + encodeURIComponent(candidate.user_id) })
}
</script>

<style lang="scss" scoped>
.wmc-container--inner {
  padding-top: 24rpx;
}
.summary-stats {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}
.summary-note {
  margin-top: 20rpx;
}
.early-card {
  border: 1rpx solid rgba(245, 158, 11, 0.3);
}
.early-text {
  font-size: 26rpx;
  line-height: 1.8;
  color: #7c4d0a;
}
.requirement-list {
  margin: 16rpx 0 8rpx;
}
.requirement-item {
  display: flex;
  align-items: center;
  padding: 12rpx 0;
}
.requirement-item__state {
  font-size: 22rpx;
  border-radius: 6rpx;
  padding: 4rpx 12rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
}
.requirement-item__state--ok {
  background: rgba(23, 166, 115, 0.1);
  color: $uni-color-success;
}
.requirement-item__state--no {
  background: rgba(227, 77, 89, 0.08);
  color: $uni-color-error;
}
.requirement-item__label {
  font-size: 26rpx;
  color: $uni-text-color;
  line-height: 1.6;
}
.join-btn {
  margin-top: 16rpx;
}
.entry-grid {
  display: flex;
  gap: 16rpx;
  margin-bottom: 24rpx;
}
.entry-item {
  flex: 1;
  background: #ffffff;
  border-radius: 12rpx;
  padding: 20rpx 16rpx;
  box-shadow: $wmc-shadow;
  display: flex;
  flex-direction: column;
}
.entry-item__title {
  font-size: 26rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.entry-item__desc {
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-top: 6rpx;
  line-height: 1.5;
}
.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8rpx 4rpx 16rpx;
}
.list-head__title {
  font-size: 30rpx;
  font-weight: 600;
}
.load-more {
  text-align: center;
  padding: 20rpx 0 40rpx;
}
.retry-btn {
  margin-top: 20rpx;
}
</style>
