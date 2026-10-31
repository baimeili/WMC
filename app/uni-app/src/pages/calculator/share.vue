<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="loading" class="wmc-card">
        <text class="wmc-hint">正在读取分享结果…</text>
      </view>

      <template v-else-if="result">
        <view class="wmc-card share-head">
          <MarketBadge market="theoretical" :show-desc="false" :small="true" />
          <text class="share-head__tip">该结果由他人分享，仅包含条件估算，不含个人身份信息。</text>
          <text class="share-head__label">理论市场估算：符合条件的人口占比</text>
          <text class="share-head__value">{{ formatPercent(result.matched_ratio, 2) }}</text>
          <text class="share-head__desc">
            基于基准数据集「{{ result.dataset.name }}」v{{ result.dataset.version }} 估算，不是本站真实用户数量，也不能计算双向匹配率。
          </text>
        </view>

        <FunnelChart
          :steps="result.funnel.steps"
          market="theoretical"
          title="顺序漏斗（按当前条件顺序计算）"
          final-label="最终估算占比"
          :final-text="formatPercent(result.matched_ratio, 2)"
          :note="result.funnel_note || funnelNote"
        />

        <view v-if="result.top_factors.length > 0" class="wmc-card">
          <view class="wmc-section-title">
            <text>Top3 影响条件</text>
          </view>
          <view v-for="(item, index) in result.top_factors" :key="item.key" class="top-item">
            <text class="top-item__rank">{{ index + 1 }}</text>
            <view class="top-item__body">
              <text class="top-item__title">{{ item.label }}</text>
              <text class="top-item__desc">
                取消该条件后，估算占比变化 {{ formatChange(item.delta_ratio) }}
              </text>
            </view>
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>算算你自己的候选池</text>
          </view>
          <text class="wmc-hint">无需注册即可计算理论市场估算；注册后可查看本站真实双向候选人。</text>
          <button class="wmc-btn wmc-btn--primary wmc-btn--block share-cta" @tap="goCalculator">算算我的候选池</button>
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
import MarketBadge from '@/components/MarketBadge.vue'
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
.share-head {
  text-align: center;
}
.share-head__tip {
  display: block;
  margin-top: 20rpx;
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
}
.share-head__label {
  display: block;
  margin-top: 20rpx;
  font-size: 26rpx;
  color: $uni-text-color-grey;
}
.share-head__value {
  display: block;
  margin: 12rpx 0;
  font-size: 72rpx;
  font-weight: 600;
  color: $uni-color-primary;
  line-height: 1.1;
}
.share-head__desc {
  display: block;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
  text-align: left;
}
.top-item {
  display: flex;
  align-items: flex-start;
  padding: 16rpx 0;
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
}
.top-item__desc {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
  margin-top: 4rpx;
}
.share-cta {
  margin-top: 20rpx;
}
</style>
