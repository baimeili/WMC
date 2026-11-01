<template>
  <view class="sensitivity">
    <view class="sensitivity__header">
      <view class="sensitivity__head-main">
        <text class="sensitivity__title">{{ title }}</text>
        <text class="sensitivity__hint">Leave-One-Out · 逐个取消条件</text>
      </view>
      <text class="sensitivity__badge">{{ items.length }} 项</text>
    </view>

    <view v-if="items.length === 0" class="sensitivity__empty">
      <text>暂无可分析的筛选条件，请先在计算器中设置「必须」条件。</text>
    </view>

    <view v-for="(item, index) in items" :key="item.key" class="sensitivity__item">
      <view class="sensitivity__row">
        <view class="sensitivity__label-wrap">
          <text class="sensitivity__rank">{{ index + 1 }}</text>
          <text class="sensitivity__label">取消「{{ item.label }}」</text>
        </view>
        <text class="sensitivity__change" :class="'sensitivity__change--' + changeTone(item.delta_ratio)">
          {{ changeText(item) }}
        </text>
      </view>
      <view class="sensitivity__values">
        <text class="sensitivity__value wmc-num">{{ baseText(item) }}</text>
        <text class="sensitivity__arrow">→</text>
        <text class="sensitivity__value sensitivity__value--strong wmc-num">{{ withoutText(item) }}</text>
        <text class="sensitivity__caption">取消该条件后</text>
      </view>
      <view class="sensitivity__track">
        <view
          class="sensitivity__bar"
          :class="'sensitivity__bar--' + changeTone(item.delta_ratio)"
          :style="{ width: barWidth(item) }"
        ></view>
      </view>
    </view>

    <view class="sensitivity__note">{{ note || '敏感性分析逐条移除单个条件，用于判断哪个条件对候选池影响最大。' }}</view>
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
  if (props.totalPopulation > 0) return formatPercent(item.base_count / props.totalPopulation, 2)
  return '基准 ' + maskCount(item.base_count)
}

function withoutText(item: SensitivityItem): string {
  if (props.market === 'real') return maskCount(item.without_count) + ' 人'
  if (props.totalPopulation > 0) return formatPercent(item.without_count / props.totalPopulation, 2)
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

// 条形长度：取前三项的最大变化率做归一化，更直观
function barWidth(item: SensitivityItem): string {
  if (isMasked(item)) return '0%'
  const maxDelta = props.items.reduce((max, it) => {
    if (isMasked(it)) return max
    const v = Math.abs(Number(it.delta_ratio) || 0)
    return v > max ? v : max
  }, 0.0001)
  const ratio = Math.min(Math.abs(Number(item.delta_ratio) || 0) / maxDelta, 1)
  return (Math.max(ratio, 0.04) * 100).toFixed(1) + '%'
}
</script>

<style lang="scss" scoped>
.sensitivity {
  background: #ffffff;
  border-radius: 28rpx;
  padding: 32rpx;
  box-shadow: 0 12rpx 40rpx rgba(15, 23, 42, 0.08);
  border: 1rpx solid rgba(15, 23, 42, 0.04);
  margin-bottom: 24rpx;
}
.sensitivity__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20rpx;
}
.sensitivity__head-main {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding-right: 20rpx;
}
.sensitivity__title {
  font-size: 32rpx;
  font-weight: 700;
  color: #0b1220;
  letter-spacing: -0.3rpx;
}
.sensitivity__hint {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #94a3b8;
}
.sensitivity__badge {
  font-size: 20rpx;
  font-weight: 600;
  color: #64748b;
  background: #f1f5f9;
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
}
.sensitivity__empty {
  padding: 24rpx 0;
  font-size: 24rpx;
  color: #64748b;
  line-height: 1.7;
}
.sensitivity__item {
  padding: 24rpx 0;
  border-bottom: 1rpx dashed rgba(15, 23, 42, 0.08);
}
.sensitivity__item:last-of-type {
  border-bottom: none;
}
.sensitivity__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sensitivity__label-wrap {
  display: flex;
  align-items: center;
}
.sensitivity__rank {
  width: 36rpx;
  height: 36rpx;
  border-radius: 10rpx;
  background: #f1f5f9;
  color: #64748b;
  font-size: 20rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 14rpx;
  flex-shrink: 0;
}
.sensitivity__label {
  font-size: 28rpx;
  font-weight: 500;
  color: #1e293b;
}
.sensitivity__change {
  font-size: 26rpx;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: #64748b;
}
.sensitivity__change--up {
  color: #0f9f74;
}
.sensitivity__change--down {
  color: #ef4444;
}
.sensitivity__change--flat {
  color: #64748b;
}
.sensitivity__values {
  display: flex;
  align-items: baseline;
  margin-top: 12rpx;
  flex-wrap: wrap;
}
.sensitivity__value {
  font-size: 26rpx;
  color: #64748b;
}
.sensitivity__value--strong {
  color: #0b1220;
  font-weight: 700;
}
.sensitivity__arrow {
  font-size: 24rpx;
  color: #94a3b8;
  margin: 0 14rpx;
}
.sensitivity__caption {
  font-size: 22rpx;
  color: #94a3b8;
  margin-left: 14rpx;
}
.sensitivity__track {
  margin-top: 14rpx;
  height: 10rpx;
  border-radius: 999rpx;
  background: #f1f5f9;
  overflow: hidden;
}
.sensitivity__bar {
  height: 100%;
  border-radius: 999rpx;
  background: linear-gradient(90deg, #94a3b8, #cbd5e1);
}
.sensitivity__bar--up {
  background: linear-gradient(90deg, #12b886, #5eead4);
}
.sensitivity__bar--down {
  background: linear-gradient(90deg, #ef4444, #fca5a5);
}
.sensitivity__note {
  margin-top: 20rpx;
  font-size: 22rpx;
  color: #94a3b8;
  line-height: 1.7;
}
</style>
