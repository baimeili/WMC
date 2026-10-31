<template>
  <view class="funnel">
    <view class="funnel__header">
      <text class="funnel__title">{{ title }}</text>
      <text class="funnel__tag" :class="market === 'real' ? 'funnel__tag--real' : ''">
        {{ market === 'real' ? '本站真实用户' : '理论市场估算' }}
      </text>
    </view>
    <view class="funnel__rows">
      <view v-for="(step, index) in steps" :key="step.key" class="funnel__row">
        <view class="funnel__row-top">
          <text class="funnel__label">{{ index + 1 }}. {{ step.label }}</text>
          <text class="funnel__main">{{ mainText(step, index) }}</text>
        </view>
        <view class="funnel__track">
          <view
            class="funnel__bar"
            :class="{ 'funnel__bar--muted': !step.applied, 'funnel__bar--real': market === 'real' }"
            :style="{ width: barWidth(step) }"
          ></view>
        </view>
        <view class="funnel__row-bottom">
          <text class="funnel__sub">{{ subText(step, index) }}</text>
          <text v-if="step.applied && index > 0" class="funnel__drop" :class="'funnel__drop--' + changeTone(-step.drop_ratio)">
            较上一层减少 {{ formatPercent(step.drop_ratio, 1) }}
          </text>
        </view>
      </view>
    </view>
    <view class="funnel__final">
      <text class="funnel__final-label">{{ finalLabel }}</text>
      <text class="funnel__final-value">{{ finalText }}</text>
    </view>
    <view class="funnel__note">{{ note || '按当前条件顺序计算，改变条件顺序会改变每一步的损失展示' }}</view>
  </view>
</template>

<script setup lang="ts">
// 顺序漏斗：理论市场只展示占比（禁止精确人数与双向匹配率）；真实市场展示 K 匿名后的人数。
import { formatPercent, maskCount, changeTone } from '@/utils/format'
import type { FunnelStep } from '@/api/types'

type Market = 'theoretical' | 'real'

const props = withDefaults(
  defineProps<{
    steps: FunnelStep[]
    market: Market
    title?: string
    finalLabel?: string
    finalText?: string
    note?: string
  }>(),
  {
    title: '条件顺序漏斗',
    finalLabel: '',
    finalText: '',
    note: ''
  }
)

function barWidth(step: FunnelStep): string {
  if (!step.applied) return '100%'
  let ratio = Number(step.ratio)
  if (!isFinite(ratio) || ratio < 0) ratio = 0
  if (ratio > 1) ratio = 1
  const width = Math.max(ratio * 100, 2)
  return width.toFixed(2) + '%'
}

function mainText(step: FunnelStep, index: number): string {
  if (!step.applied) return '不限制'
  if (props.market === 'real') return maskCount(step.count) + ' 人'
  return '占比 ' + formatPercent(step.ratio, 2)
}

function subText(step: FunnelStep, index: number): string {
  if (!step.applied) return '未设置该条件，沿用上一层结果'
  if (index === 0) {
    return props.market === 'real' ? '当前真实用户池中满足该条件' : '占基准人口样本的比例'
  }
  return props.market === 'real' ? '通过该条件后的候选人数' : '占基准人口样本的比例'
}
</script>

<style lang="scss" scoped>
.funnel {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 28rpx;
  box-shadow: $wmc-shadow;
  margin-bottom: 24rpx;
}
.funnel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}
.funnel__title {
  font-size: 30rpx;
  font-weight: 600;
}
.funnel__tag {
  font-size: 22rpx;
  color: #92580a;
  background: rgba(245, 158, 11, 0.12);
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}
.funnel__tag--real {
  color: #1f4fd0;
  background: rgba(47, 107, 255, 0.1);
}
.funnel__row {
  margin-bottom: 26rpx;
}
.funnel__row-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10rpx;
}
.funnel__label {
  font-size: 28rpx;
  color: $uni-text-color;
}
.funnel__main {
  font-size: 28rpx;
  font-weight: 600;
  color: $uni-color-primary;
}
.funnel__track {
  height: 18rpx;
  border-radius: 999rpx;
  background: #eef2f7;
  overflow: hidden;
}
.funnel__bar {
  height: 100%;
  border-radius: 999rpx;
  background: linear-gradient(90deg, #2f6bff, #6f9bff);
}
.funnel__bar--muted {
  background: #dbe3ee;
}
.funnel__bar--real {
  background: linear-gradient(90deg, #17a673, #4fc79b);
}
.funnel__row-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8rpx;
}
.funnel__sub {
  font-size: 22rpx;
  color: $uni-text-color-grey;
}
.funnel__drop {
  font-size: 22rpx;
  color: $uni-color-error;
}
.funnel__drop--up {
  color: $uni-color-success;
}
.funnel__final {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8rpx;
  padding: 20rpx 24rpx;
  border-radius: 12rpx;
  background: #f8fafc;
}
.funnel__final-label {
  font-size: 26rpx;
  color: $uni-text-color-grey;
}
.funnel__final-value {
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.funnel__note {
  margin-top: 18rpx;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
</style>
