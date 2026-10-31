<template>
  <PageShell>
    <view class="wmc-container">
      <template v-if="result">
        <view class="wmc-card result-head">
          <MarketBadge market="theoretical" :show-desc="false" :small="true" />
          <text class="result-head__label">理论市场估算：符合条件的人口占比</text>
          <text class="result-head__value">{{ formatPercent(result.matched_ratio, 2) }}</text>
          <text class="result-head__desc">
            基于基准数据集「{{ result.dataset.name }}」v{{ result.dataset.version }} 的加权样本估算，不是本站真实用户数量，也不能计算双向匹配率。
          </text>
          <view v-if="result.dataset.is_demo" class="wmc-notice wmc-notice--warning demo-notice">
            当前使用演示数据集，仅用于功能演示，不构成任何现实结论。
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>数据来源与说明</text>
            <text class="wmc-tag wmc-tag--muted">{{ result.dataset.source_name || '基准数据集' }}</text>
          </view>
          <text class="wmc-hint">{{ result.dataset.disclaimer || estimateDisclaimer }}</text>
          <text class="wmc-hint">生成时间：{{ formatDateTime(result.generated_at) }}</text>
        </view>

        <FunnelChart
          :steps="result.funnel.steps"
          market="theoretical"
          title="顺序漏斗（按当前条件顺序计算）"
          final-label="最终估算占比"
          :final-text="formatPercent(result.matched_ratio, 2)"
          :note="result.funnel_note || funnelNote"
        />

        <view v-if="biggestFilter" class="wmc-card">
          <view class="wmc-section-title">
            <text>最大筛选条件</text>
            <text class="wmc-tag wmc-tag--warning">{{ biggestFilter.label }}</text>
          </view>
          <text class="biggest__value">{{ biggestFilter.label }}</text>
          <text class="wmc-hint">
            按当前条件顺序计算，该条件使通过上一层条件的人口再减少 {{ formatPercent(biggestFilter.drop_ratio, 1) }}。
          </text>
        </view>

        <view v-if="result.top_factors.length > 0" class="wmc-card">
          <view class="wmc-section-title">
            <text>Top3 影响条件</text>
          </view>
          <view v-for="(item, index) in result.top_factors" :key="item.key" class="top-item">
            <view class="top-item__rank">{{ index + 1 }}</view>
            <view class="top-item__body">
              <text class="top-item__title">{{ item.label }}</text>
              <text class="top-item__desc">
                取消该条件后，估算占比由 {{ ratioText(item.base_count) }} 变为 {{ ratioText(item.without_count) }}，变化 {{ formatChange(item.delta_ratio) }}
              </text>
            </view>
          </view>
        </view>

        <SensitivityList
          :items="result.sensitivity"
          market="theoretical"
          :total-population="result.total_population"
        />

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>下一步</text>
          </view>
          <text class="wmc-hint">
            理论市场只回答「符合条件的人大约占多少」。想知道本站真实有多少人与你双向满足硬条件，需要注册并完善资料后进入真实候选池。
          </text>
          <button class="wmc-btn wmc-btn--primary wmc-btn--block cta" @tap="goDiscovery">注册，看看本站真实候选人</button>
          <view class="wmc-btn-row cta-row">
            <button class="wmc-btn wmc-btn--ghost" :disabled="loadingShare" @tap="shareResult">
              {{ loadingShare ? '生成中…' : '分享结果' }}
            </button>
            <button class="wmc-btn wmc-btn--soft" @tap="recalculate">重新计算</button>
          </view>
          <view class="share-link" @tap="openSharePage">
            <text class="wmc-link">查看分享页</text>
          </view>
        </view>

        <view class="wmc-card wmc-card--flat">
          <text class="wmc-hint">
            以上占比与分数仅根据填写条件计算，不代表婚姻成功概率。理论市场估算与本站真实候选池是两套独立数据，请勿混用。
          </text>
        </view>
      </template>
      <EmptyState
        v-else
        title="还没有计算结果"
        description="先填写条件完成一次理论市场估算，再回来看结果。"
        action-text="去计算"
        @action="goCalculator"
      />
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 理论市场结果页：占比、顺序漏斗、最大筛选条件、Leave-One-Out 敏感性、Top3 影响条件。
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import FunnelChart from '@/components/FunnelChart.vue'
import SensitivityList from '@/components/SensitivityList.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useCalculatorStore } from '@/stores/calculator'
import { useAuthStore } from '@/stores/auth'
import { ESTIMATE_DISCLAIMER, FUNNEL_NOTE } from '@/utils/constants'
import { formatChange, formatDateTime, formatPercent } from '@/utils/format'
import { errorMessage } from '@/utils/error'
import { trackEvent } from '@/api/events'
import type { FunnelStep } from '@/api/types'

const store = useCalculatorStore()
const auth = useAuthStore()
const result = computed(() => store.result)
const loadingShare = ref(false)
const estimateDisclaimer = ESTIMATE_DISCLAIMER
const funnelNote = FUNNEL_NOTE

onLoad(() => {
  const data = store.loadLastResult()
  auth.syncFromStorage()
  if (data) {
    trackEvent('calculator_result_view', { run_id: data.run_id })
  }
})

const biggestFilter = computed<FunnelStep | null>(() => {
  if (!result.value) return null
  const steps = result.value.funnel.steps.filter((item) => item.applied)
  let best: FunnelStep | null = null
  for (let i = 0; i < steps.length; i++) {
    if (!best || steps[i].drop_ratio > best.drop_ratio) best = steps[i]
  }
  return best
})

function ratioText(count: number): string {
  if (!result.value || !result.value.total_population) return '—'
  return formatPercent(count / result.value.total_population, 2)
}

function goCalculator() {
  uni.navigateTo({ url: '/pages/calculator/index' })
}

function recalculate() {
  uni.navigateTo({ url: '/pages/calculator/index' })
}

function goDiscovery() {
  if (auth.isLoggedIn) {
    uni.navigateTo({ url: '/pages/discovery/index' })
    return
  }
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/discovery/index') })
}

function buildShareLink(token: string): string {
  let origin = ''
  if (typeof window !== 'undefined' && window.location) {
    origin = window.location.origin
  }
  return origin + '/#/pages/calculator/share?token=' + token
}

async function shareResult() {
  if (!result.value || loadingShare.value) return
  loadingShare.value = true
  try {
    const token = await store.ensureShareToken()
    const link = buildShareLink(token)
    uni.showModal({
      title: '分享链接已生成',
      content: '将复制结果分享链接。分享页只包含条件估算结果，不包含个人身份信息。',
      confirmText: '复制链接',
      cancelText: '取消',
      success: (response) => {
        if (!response.confirm) return
        uni.setClipboardData({
          data: link,
          success: () => {
            uni.showToast({ title: '链接已复制', icon: 'none' })
          }
        })
      }
    })
    trackEvent('calculator_share_link_created', { run_id: result.value.run_id })
  } catch (error) {
    uni.showToast({ title: errorMessage(error, '分享链接生成失败'), icon: 'none' })
  } finally {
    loadingShare.value = false
  }
}

async function openSharePage() {
  if (!result.value || loadingShare.value) return
  loadingShare.value = true
  try {
    const token = await store.ensureShareToken()
    uni.navigateTo({ url: '/pages/calculator/share?token=' + encodeURIComponent(token) })
  } catch (error) {
    uni.showToast({ title: errorMessage(error, '分享页打开失败'), icon: 'none' })
  } finally {
    loadingShare.value = false
  }
}
</script>

<style lang="scss" scoped>
.result-head {
  text-align: center;
}
.result-head__label {
  display: block;
  margin-top: 24rpx;
  font-size: 26rpx;
  color: $uni-text-color-grey;
}
.result-head__value {
  display: block;
  margin: 12rpx 0;
  font-size: 72rpx;
  font-weight: 600;
  color: $uni-color-primary;
  line-height: 1.1;
}
.result-head__desc {
  display: block;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
  text-align: left;
}
.demo-notice {
  margin-top: 20rpx;
  text-align: left;
}
.biggest__value {
  display: block;
  font-size: 40rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 10rpx;
}
.top-item {
  display: flex;
  align-items: flex-start;
  padding: 18rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}
.top-item:last-child {
  border-bottom: none;
}
.top-item__rank {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: rgba(47, 107, 255, 0.1);
  color: $uni-color-primary;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;
  flex-shrink: 0;
}
.top-item__body {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.top-item__title {
  font-size: 28rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.top-item__desc {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
  margin-top: 4rpx;
}
.cta {
  margin-top: 24rpx;
}
.cta-row {
  margin-top: 20rpx;
}
.share-link {
  margin-top: 24rpx;
  text-align: center;
}
</style>
