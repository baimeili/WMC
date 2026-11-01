<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-card">
        <EmptyState title="登录后查看候选人详情" description="候选人详情只对已登录用户开放。" action-text="去登录" @action="goLogin" />
      </view>

      <view v-else-if="loading" class="wmc-card">
        <text class="wmc-hint">正在加载候选人资料…</text>
      </view>
      <view v-else-if="loadError" class="wmc-card">
        <text class="wmc-error">{{ loadError }}</text>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="load">重新加载</button>
      </view>
      <view v-else-if="!candidate" class="wmc-card">
        <EmptyState title="候选人不可见" description="该用户可能已退出婚恋池、被屏蔽或资料不可见。" action-text="返回候选池" @action="back" />
      </view>

      <template v-else>
        <view class="wmc-card profile-card">
          <view class="profile-head">
            <image v-if="candidate.primary_photo_url" class="profile-avatar" :src="candidate.primary_photo_url" mode="aspectFill" />
            <view v-else class="profile-avatar profile-avatar--placeholder">
              <text class="profile-avatar-text">{{ avatarText }}</text>
            </view>
            <view class="profile-info">
              <view class="profile-name-row">
                <text class="profile-name">{{ candidate.nickname || '未填写昵称' }}</text>
                <text class="profile-age">{{ ageText(candidate.age) }}</text>
              </view>
              <text class="profile-meta">{{ candidate.city_name || '城市未填写' }} · {{ educationLabel }}</text>
              <text class="profile-meta">{{ heightText(candidate.height_cm) }} · {{ maritalLabel }} · {{ wantsChildrenLabel }}</text>
            </view>
          </view>
          <text v-if="candidate.bio" class="profile-bio">{{ candidate.bio }}</text>
          <view class="market-row">
            <MarketBadge market="real" :show-desc="false" :small="true" />
            <text class="wmc-hint">该用户已加入本站真实婚恋市场</text>
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>匹配分数</text>
            <text class="wmc-tag wmc-tag--muted">仅根据条件计算</text>
          </view>
          <view class="score-grid">
            <view class="score-cell">
              <text class="score-cell__value">{{ formatScore(candidate.preference_score) }}</text>
              <text class="score-cell__label">条件契合度</text>
            </view>
            <view class="score-cell">
              <text v-if="candidate.compatibility_insufficient" class="score-cell__value score-cell__value--muted">数据不足</text>
              <text v-else class="score-cell__value">{{ formatScore(candidate.compatibility_score) }}</text>
              <text class="score-cell__label">关系观念兼容度</text>
            </view>
            <view class="score-cell">
              <text class="score-cell__value score-cell__value--primary">{{ formatScore(candidate.overall_score) }}</text>
              <text class="score-cell__label">综合匹配度</text>
            </view>
          </view>
          <text v-if="candidate.compatibility_insufficient" class="wmc-notice wmc-notice--warning compat-notice">
            关系观念数据不足：双方共同填写的有效价值观题不足 6 道，暂不输出精确的关系观念兼容度。
          </text>
          <text class="score-disclaimer">{{ candidate.disclaimer || defaultDisclaimer }}</text>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>为什么匹配</text>
            <text class="wmc-link" @tap="toggleWhy">{{ showWhy ? '收起' : '展开' }}</text>
          </view>
          <text class="wmc-hint">解释仅包含硬条件满足情况与分类维度得分，不代表婚姻成功概率。</text>
          <view v-if="showWhy" class="why-body">
            <view v-if="safeReasons.length > 0" class="why-block">
              <text class="why-block__title">硬条件满足项</text>
              <view v-for="(reason, index) in safeReasons" :key="index" class="why-item">
                <text class="why-item__dot">·</text>
                <text class="why-item__text">{{ reason }}</text>
              </view>
            </view>
            <view v-if="safeConflicts.length > 0" class="why-block">
              <text class="why-block__title">硬条件冲突项</text>
              <text class="why-item__text">你与该用户的硬条件存在 {{ safeConflicts.length }} 项冲突</text>
              <view v-for="(conflict, index) in safeConflicts" :key="index" class="why-item">
                <text class="why-item__dot">·</text>
                <text class="why-item__text">{{ conflict }}</text>
              </view>
            </view>
            <view v-if="safeDimensions.length > 0" class="why-block">
              <text class="why-block__title">各分类维度得分</text>
              <ScoreBar v-for="dimension in safeDimensions" :key="dimension.key || dimension.name" :label="dimension.label || dimension.name" :score="dimension.score" />
            </view>
            <view v-if="safeReasons.length === 0 && safeDimensions.length === 0" class="why-block">
              <text class="wmc-hint">暂无更细的解释数据。</text>
            </view>
            <text class="why-disclaimer">{{ defaultDisclaimer }}</text>
          </view>
        </view>

        <view class="wmc-card">
          <button class="wmc-btn wmc-btn--primary wmc-btn--block" :disabled="acting" @tap="like">喜欢</button>
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block" :disabled="acting" @tap="pass">跳过</button>
          <text v-if="actionMessage" class="wmc-error action-message">{{ actionMessage }}</text>
        </view>

        <view class="wmc-card wmc-card--flat">
          <view class="wmc-section-title">
            <text>安全操作</text>
          </view>
          <view class="safety-row">
            <text class="wmc-link" @tap="goReport">举报该用户</text>
            <text class="wmc-link" @tap="block">屏蔽该用户</text>
          </view>
          <text class="wmc-hint">屏蔽后双方不会再互相看到或互动；举报内容仅用于平台处理，不会向对方透露举报人。</text>
        </view>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 候选人详情：展开「为什么匹配」，展示硬条件满足项、分类维度得分与免责声明。
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import ScoreBar from '@/components/ScoreBar.vue'
import { getCandidate } from '@/api/discovery'
import { setInteraction } from '@/api/interaction'
import { blockUser } from '@/api/safety'
import { useAuthStore } from '@/stores/auth'
import { optionLabel } from '@/utils/options'
import { ageText, formatScore, heightText } from '@/utils/format'
import { errorMessage } from '@/utils/error'
import { RESULT_DISCLAIMER } from '@/utils/constants'
import { trackEvent } from '@/api/events'
import type { CandidateDetail } from '@/api/types'

const auth = useAuthStore()
const candidate = ref<CandidateDetail | null>(null)

// 后端可能省略空数组字段，这里统一兜底，避免模板取 .length 报错
const safeReasons = computed(() => (candidate.value && candidate.value.reasons ? candidate.value.reasons : []))
const safeConflicts = computed(() => (candidate.value && candidate.value.conflicts ? candidate.value.conflicts : []))
const safeDimensions = computed(() => (candidate.value && candidate.value.dimensions ? candidate.value.dimensions : []))
const loading = ref(true)
const loadError = ref('')
const showWhy = ref(true)
const acting = ref(false)
const actionMessage = ref('')
const defaultDisclaimer = RESULT_DISCLAIMER
const userId = ref('')

const loggedIn = computed(() => auth.isLoggedIn)
const avatarText = computed(() => {
  const name = candidate.value ? candidate.value.nickname : '用'
  return (name || '用').slice(0, 1)
})
const educationLabel = computed(() => (candidate.value ? optionLabel('education_levels', candidate.value.education_level, '学历未填写') : ''))
const maritalLabel = computed(() => (candidate.value ? optionLabel('marital_statuses', candidate.value.marital_status, '婚姻状态未填写') : ''))
const wantsChildrenLabel = computed(() => (candidate.value ? optionLabel('wants_children', candidate.value.wants_children, '生育意愿未填写') : ''))

onLoad((options) => {
  userId.value = options && options.id ? decodeURIComponent(options.id) : ''
  auth.syncFromStorage()
  if (auth.isLoggedIn) load()
  else loading.value = false
})

async function load() {
  if (!userId.value) {
    loadError.value = '缺少候选人标识'
    loading.value = false
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    candidate.value = await getCandidate(userId.value)
  } catch (error) {
    loadError.value = errorMessage(error, '候选人详情加载失败')
  } finally {
    loading.value = false
  }
}

function toggleWhy() {
  showWhy.value = !showWhy.value
}

async function like() {
  if (!candidate.value || acting.value) return
  acting.value = true
  actionMessage.value = ''
  try {
    const result = await setInteraction(candidate.value.user_id, 1, 2)
    trackEvent('candidate_like', { matched: result.matched })
    if (result.matched && result.match_id) {
      uni.showModal({
        title: '你们已经互相感兴趣',
        content: '已生成 Match，可以在「我的 Match」中查看。',
        confirmText: '查看 Match',
        cancelText: '稍后',
        success: (response) => {
          if (response.confirm) {
            uni.navigateTo({ url: '/pages/matches/detail?id=' + encodeURIComponent(result.match_id || '') })
          } else {
            uni.navigateBack({ delta: 1 })
          }
        }
      })
    } else {
      uni.showToast({ title: '已记录你的兴趣', icon: 'none' })
      uni.navigateBack({ delta: 1 })
    }
  } catch (error) {
    actionMessage.value = errorMessage(error, '操作失败，请稍后重试')
  } finally {
    acting.value = false
  }
}

async function pass() {
  if (!candidate.value || acting.value) return
  acting.value = true
  actionMessage.value = ''
  try {
    await setInteraction(candidate.value.user_id, 2, 2)
    uni.showToast({ title: '已跳过该用户', icon: 'none' })
    uni.navigateBack({ delta: 1 })
  } catch (error) {
    actionMessage.value = errorMessage(error, '操作失败，请稍后重试')
  } finally {
    acting.value = false
  }
}

function goReport() {
  if (!candidate.value) return
  uni.navigateTo({ url: '/pages/report/index?target=' + encodeURIComponent(candidate.value.user_id) })
}

function block() {
  if (!candidate.value || acting.value) return
  uni.showModal({
    title: '屏蔽该用户',
    content: '屏蔽后双方不会再互相看到或互动，确定继续吗？',
    confirmText: '确认屏蔽',
    cancelText: '取消',
    success: async (response) => {
      if (!response.confirm) return
      acting.value = true
      try {
        await blockUser(candidate.value!.user_id)
        uni.showToast({ title: '已屏蔽该用户', icon: 'none' })
        uni.navigateBack({ delta: 1 })
      } catch (error) {
        actionMessage.value = errorMessage(error, '屏蔽失败，请稍后重试')
      } finally {
        acting.value = false
      }
    }
  })
}

function back() {
  uni.navigateBack({ delta: 1 })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/discovery/index') })
}
</script>

<style lang="scss" scoped>
.profile-card {
  padding: 28rpx;
}
.profile-head {
  display: flex;
  align-items: center;
  gap: 24rpx;
}
.profile-avatar {
  width: 128rpx;
  height: 128rpx;
  border-radius: 20rpx;
  background: #eef2f7;
  flex-shrink: 0;
}
.profile-avatar--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(47, 107, 255, 0.1);
}
.profile-avatar-text {
  font-size: 48rpx;
  color: $uni-color-primary;
  font-weight: 600;
}
.profile-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.profile-name-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}
.profile-name {
  font-size: 36rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.profile-age {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.profile-meta {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.profile-bio {
  display: block;
  margin-top: 20rpx;
  font-size: 26rpx;
  line-height: 1.8;
  color: $uni-text-color;
}
.market-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-top: 20rpx;
}
.score-grid {
  display: flex;
  gap: 16rpx;
}
.score-cell {
  flex: 1;
  background: #f8fafc;
  border-radius: 12rpx;
  padding: 20rpx 8rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6rpx;
}
.score-cell__value {
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.score-cell__value--primary {
  color: $uni-color-primary;
}
.score-cell__value--muted {
  font-size: 24rpx;
  color: $uni-text-color-placeholder;
}
.score-cell__label {
  font-size: 22rpx;
  color: $uni-text-color-grey;
}
.compat-notice {
  margin-top: 20rpx;
}
.score-disclaimer {
  display: block;
  margin-top: 20rpx;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.why-body {
  margin-top: 12rpx;
}
.why-block {
  margin-top: 16rpx;
}
.why-block__title {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 10rpx;
}
.why-item {
  display: flex;
  align-items: flex-start;
  margin-bottom: 8rpx;
}
.why-item__dot {
  color: $uni-color-primary;
  margin-right: 8rpx;
}
.why-item__text {
  flex: 1;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.why-disclaimer {
  display: block;
  margin-top: 16rpx;
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
  line-height: 1.7;
}
.action-message {
  display: block;
  margin-top: 12rpx;
}
.safety-row {
  display: flex;
  gap: 40rpx;
  margin-bottom: 12rpx;
}
.retry-btn {
  margin-top: 20rpx;
}
</style>