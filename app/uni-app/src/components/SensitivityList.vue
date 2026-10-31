<template>
  <view class="sensitivity">
    <view class="sensitivity__header">
      <text class="sensitivity__title">{{ title }}</text>
      <text class="sensitivity__hint">Leave-One-Out</text>
    </view>
    <view v-if="items.length === 0" class="sensitivity__empty">
      <text>暂无可分析的筛选条件，请先在计算器中设置「必须」条件。</text>
    </view>
    <view v-for="item in items" :key="item.key" class="sensitivity__item">
      <view class="sensitivity__row">
        <text class="sensitivity__label">取消「{{ item.label }}」条件</text>
        <text class="sensitivity__change" :class="'sensitivity__change--' + changeTone(item.delta_ratio)">
          {{ changeText(item) }}
        </text>
      </view>
      <view class="sensitivity__values">
        <text class="sensitivity__value">{{ baseText(item) }}</text>
        <text class="sensitivity__arrow">→</text>
        <text class="sensitivity__value sensitivity__value--strong">{{ withoutText(item) }}</text>
        <text class="sensitivity__caption">（取消该条件后）</text>
      </view>
    </view>
    <view class="sensitivity__note">
      <text>{{ note || '敏感性分析逐条移除单个条件，用于判断哪个条件对候选池影响最大。' }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
// Leave-One-Out 敏感性列表：
// 理论市场展示「占比 a → b 与变化率」；真实市场人数小于 K 时按匿名规则隐藏变化率。
import { formatChange, formatPercent, maskCount, changeTone } from '@/utils/format'
import { K_ANONYMITY_THRESHOLD } from '@/utils/constants'
import type { SensitivityItem } from '@/api/types'

type Market = 'theoretical' | 'real'

const props = withDefaults(
  defineProps<{
    items: SensitivityItem[]
    market: Market
    totalPopulation?: number
    title?: string
    note?: string
  }>(),
  {
    totalPopulation: 0,
    title: '条件敏感性分析（逐条取消）',
    note: ''
  }
)

function baseText(item: SensitivityItem): string {
  if (props.market === 'real') return maskCount(item.base_count) + ' 人'
  if (props.totalPopulation > 0) return '占比 ' + formatPercent(item.base_count / props.totalPopulation, 2)
  return '基准 ' + maskCount(item.base_count)
}

function withoutText(item: SensitivityItem): string {
  if (props.market === 'real') return maskCount(item.without_count) + ' 人'
  if (props.totalPopulation > 0) return '占比 ' + formatPercent(item.without_count / props.totalPopulation, 2)
  return '取消后 ' + maskCount(item.without_count)
}

function isMasked(item: SensitivityItem): boolean {
  return (
    props.market === 'real' &&
    (item.base_count < K_ANONYMITY_THRESHOLD || item.without_count < K_ANONYMITY_THRESHOLD)
  )
}

function changeText(item: SensitivityItem): string {
  if (isMasked(item)) return '人数不足 10，不展示变化率'
  return formatChange(item.delta_ratio)
}
</script>

<style lang="scss" scoped>
.sensitivity {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 28rpx;
  box-shadow: $wmc-shadow;
  margin-bottom: 24rpx;
}
.sensitivity__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}
.sensitivity__title {
  font-size: 30rpx;
  font-weight: 600;
}
.sensitivity__hint {
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
}
.sensitivity__empty {
  padding: 20rpx 0;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.sensitivity__item {
  padding: 20rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}
.sensitivity__item:last-of-type {
  border-bottom: none;
}
.sensitivity__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sensitivity__label {
  font-size: 28rpx;
  color: $uni-text-color;
}
.sensitivity__change {
  font-size: 26rpx;
  font-weight: 600;
  color: $uni-text-color-grey;
}
.sensitivity__change--up {
  color: $uni-color-success;
}
.sensitivity__change--down {
  color: $uni-color-error;
}
.sensitivity__values {
  display: flex;
  align-items: baseline;
  margin-top: 10rpx;
  flex-wrap: wrap;
}
.sensitivity__value {
  font-size: 26rpx;
  color: $uni-text-color-grey;
}
.sensitivity__value--strong {
  color: $uni-text-color;
  font-weight: 600;
}
.sensitivity__arrow {
  font-size: 24rpx;
  color: $uni-text-color-placeholder;
  margin: 0 12rpx;
}
.sensitivity__caption {
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
  margin-left: 12rpx;
}
.sensitivity__note {
  margin-top: 16rpx;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
</style>
