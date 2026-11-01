<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="loading" class="wmc-card">
        <text class="wmc-hint">正在读取分享结果…</text>
      </view>

      <template v-else-if="result">
        <!-- 结果 Hero -->
        <view class="result-hero">
          <view class="result-hero__glow"></view>
          <view class="result-hero__inner">
            <view class="result-hero__badge">理论市场估算</view>
            <text class="result-hero__label">符合条件的人口占比</text>
            <text class="result-hero__value wmc-num">{{ formatPercent(result.matched_ratio, 2) }}</text>
            <text class="result-hero__desc">
              基于基准数据集「{{ result.dataset.name }}」v{{ result.dataset.version }} 的加权样本估算，不是本站真实用户数量。
            </text>
            <view v-if="result.dataset.is_demo" class="result-hero__demo">演示数据集 · 仅用于功能演示</view>
          </view>
        </view>

        <view class="wmc-notice wmc-notice--warning share-tip">
          <text>该结果由他人分享，仅包含条件估算，不含个人身份信息；理论市场不能计算双向匹配率。</text>
        </view>

        <FunnelChart
          :steps="result.funnel.steps"
          market="theoretical"
          title="条件顺序漏斗"
          final-label="最终估算占比"
          :final-text="formatPercent(result.matched_ratio, 2)"
          :note="result.funnel_note || funnelNote"
        />

        <view v-if="result.top_factors.length > 0" class="wmc-card">
          <view class="wmc-section-title">
            <text>Top3 影响条件</text>
            <text class="wmc-tag wmc-tag--warning">最影响候选池</text>
          </view>
          <view v-for="(item, index) in result.top_factors" :key="item.key" class="rank">
            <view class="rank__badge" :class="'rank__badge--' + (index + 1)">{{ index + 1 }}</view>
            <view class="rank__body">
              <text class="rank__title">{{ item.label }}</text>
              <text class="rank__desc">取消该条件后，估算占比变化 {{ formatChange(item.delta_ratio) }}</text>
            </view>
            <text class="rank__delta wmc-num">{{ formatChange(item.delta_ratio) }}</text>
          </view>
        </view>

        <view class="cta-card">
          <text class="cta-card__title">算算你自己的候选池</text>
          <text class="cta-card__desc">无需注册即可计算理论市场估算；注册后可查看本站真实双向候选人。</text>
          <button class="wmc-btn wmc-btn--primary wmc-btn--block" @tap="goCalculator">算算我的候选池</button>
        </view>

        <view class="wmc-card wmc-card--flat">
          <text class="wmc-hint">
            以上占比仅根据填写条件计算，不代表婚姻成功概率。理论市场估算与本站真实候选池是两套独立数据。
          </text>
        </view>
      </template>

      <EmptyState
        v-else
        title="分享结果不可用"
        description="链接可能已过期或不存在。你可以直接填写条件，得到自己的理论市场估算。"
        action-text="去计算"
        @action="goCalculator"
      />
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 分享结果页：按分享 token 读取理论市场估算结果，只读展示。
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import FunnelChart from '@/components/FunnelChart.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useCalculatorStore } from '@/stores/calculator'
import { FUNNEL_NOTE } from '@/utils/constants'
import { formatChange, formatPercent } from '@/utils/format'
import type { CalculatorResult } from '@/api/types'

const store = useCalculatorStore()
const result = ref<CalculatorResult | null>(null)
const loading = ref(true)
const funnelNote = FUNNEL_NOTE

onLoad(async (options) => {
  const token = options && options.token ? decodeURIComponent(options.token) : ''
  if (!token) {
    loading.value = false
    return
  }
  try {
    result.value = await store.loadSharedResult(token)
  } catch (error) {
    result.value = null
  } finally {
    loading.value = false
  }
})

function goCalculator() {
  uni.reLaunch({ url: '/pages/calculator/index' })
}
</script>

<style lang="scss" scoped>
/* 深色结果 Hero：突出核心数字 */
.result-hero {
  position: relative;
  overflow: hidden;
  border-radius: 40rpx;
  background: linear-gradient(150deg, #0d1533 0%, #1d2a72 52%, #3851d6 100%);
  box-shadow: 0 28rpx 60rpx rgba(13, 21, 51, 0.32);
  margin-bottom: 24rpx;
}
.result-hero__glow {
  position: absolute;
  width: 420rpx;
  height: 420rpx;
  right: -150rpx;
  top: -170rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(122, 148, 255, 0.5) 0%, rgba(122, 148, 255, 0) 70%);
}
.result-hero__inner {
  position: relative;
  padding: 40rpx 36rpx 44rpx;
}
.result-hero__badge {
  display: inline-flex;
  align-items: center;
  font-size: 22rpx;
  font-weight: 600;
  color: #ffe9c2;
  background: rgba(245, 158, 11, 0.18);
  border: 1rpx solid rgba(245, 158, 11, 0.35);
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
}
.result-hero__label {
  display: block;
  margin-top: 26rpx;
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.7);
}
.result-hero__value {
  display: block;
  margin-top: 12rpx;
  font-size: 96rpx;
  line-height: 1.1;
  font-weight: 700;
  letter-spacing: -2rpx;
  color: #ffffff;
}
.result-hero__desc {
  display: block;
  margin-top: 18rpx;
  font-size: 24rpx;
  line-height: 1.75;
  color: rgba(255, 255, 255, 0.62);
}
.result-hero__demo {
  display: inline-flex;
  margin-top: 22rpx;
  font-size: 22rpx;
  color: #ffe9c2;
  background: rgba(255, 255, 255, 0.1);
  border: 1rpx solid rgba(255, 255, 255, 0.2);
  padding: 8rpx 20rpx;
  border-radius: 999rpx;
}
.share-tip {
  margin-bottom: 24rpx;
}

/* Top3 排名 */
.rank {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx dashed rgba(15, 23, 42, 0.08);
}
.rank:last-child {
  border-bottom: none;
}
.rank__badge {
  width: 48rpx;
  height: 48rpx;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  font-weight: 700;
  color: #64748b;
  background: #f1f5f9;
  margin-right: 20rpx;
  flex-shrink: 0;
}
.rank__badge--1 {
  color: #b45309;
  background: linear-gradient(135deg, #fde68a, #fbbf24);
}
.rank__badge--2 {
  color: #475569;
  background: linear-gradient(135deg, #e2e8f0, #cbd5e1);
}
.rank__badge--3 {
  color: #9a3412;
  background: linear-gradient(135deg, #fed7aa, #fdba74);
}
.rank__body {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.rank__title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1e293b;
}
.rank__desc {
  margin-top: 4rpx;
  font-size: 22rpx;
  color: #94a3b8;
}
.rank__delta {
  font-size: 26rpx;
  font-weight: 700;
  color: #0f9f74;
}

/* 底部 CTA */
.cta-card {
  border-radius: 32rpx;
  padding: 36rpx 32rpx;
  margin-bottom: 24rpx;
  background: linear-gradient(160deg, #ffffff 0%, #eef2ff 100%);
  border: 1rpx solid rgba(51, 85, 255, 0.14);
  box-shadow: 0 12rpx 40rpx rgba(15, 23, 42, 0.06);
}
.cta-card__title {
  display: block;
  font-size: 32rpx;
  font-weight: 700;
  color: #0b1220;
}
.cta-card__desc {
  display: block;
  margin: 12rpx 0 28rpx;
  font-size: 24rpx;
  line-height: 1.75;
  color: #64748b;
}
</style>
