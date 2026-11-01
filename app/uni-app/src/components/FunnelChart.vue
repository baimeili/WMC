<template>
  <view class="funnel">
    <view class="funnel__header">
      <view class="funnel__head-main">
        <text class="funnel__title">{{ title }}</text>
        <text class="funnel__hint">{{ market === 'real' ? '本站真实用户池 · 双向硬筛' : '理论市场估算 · 加权样本' }}</text>
      </view>
      <text class="funnel__tag" :class="market === 'real' ? 'funnel__tag--real' : 'funnel__tag--theory'">
        {{ market === 'real' ? '真实市场' : '理论市场' }}
      </text>
    </view>

    <view class="funnel__list">
      <view
        v-for="(step, index) in steps"
        :key="step.key"
        class="funnel__item"
        :class="{ 'funnel__item--muted': !step.applied }"
      >
        <view class="funnel__row">
          <view class="funnel__label-wrap">
            <text class="funnel__index">{{ index + 1 }}</text>
            <text class="funnel__label">{{ step.label }}</text>
          </view>
          <text class="funnel__value" :class="{ 'funnel__value--muted': !step.applied }">{{ mainText(step, index) }}</text>
        </view>
        <view class="funnel__track">
          <view
            class="funnel__bar"
            :class="{ 'funnel__bar--muted': !step.applied, 'funnel__bar--real': market === 'real' }"
            :style="{ width: barWidth(step) }"
          ></view>
        </view>
        <view class="funnel__meta">
          <text class="funnel__sub">{{ subText(step, index) }}</text>
          <text
            v-if="step.applied && index > 0"
            class="funnel__drop"
            :class="'funnel__drop--' + changeTone(-step.drop_ratio)"
          >
            ↓ {{ formatPercent(step.drop_ratio, 1) }}
          </text>
        </view>
      </view>
    </view>

    <view class="funnel__final" :class="market === 'real' ? 'funnel__final--real' : ''">
      <view class="funnel__final-left">
        <text class="funnel__final-label">{{ finalLabel || '最终结果' }}</text>
        <text class="funnel__final-desc">{{ market === 'real' ? '双向都满足硬条件' : '通过全部条件后的占比' }}</text>
      </view>
      <text class="funnel__final-value wmc-num">{{ finalText }}</text>
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
  // 真实市场按「第一层」归一化，漏斗形态更直观；理论市场按总人口占比
  if (props.market === 'real') {
    const base = Number(props.steps[0] ? props.steps[0].count : 0) || 0
    ratio = base > 0 ? Number(step.count) / base : 0
  }
  if (ratio > 1) ratio = 1
  const width = Math.max(ratio * 100, 2)
  return width.toFixed(2) + '%'
}

function mainText(step: FunnelStep, index: number): string {
  if (!step.applied) return '不限制'
  if (props.market === 'real') return maskCount(step.count) + ' 人'
  return formatPercent(step.ratio, 2)
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
  border-radius: 28rpx;
  padding: 32rpx;
  box-shadow: 0 12rpx 40rpx rgba(15, 23, 42, 0.08);
  border: 1rpx solid rgba(15, 23, 42, 0.04);
  margin-bottom: 24rpx;
}
.funnel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 28rpx;
}
.funnel__head-main {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding-right: 20rpx;
}
.funnel__title {
  font-size: 32rpx;
  font-weight: 700;
  color: #0b1220;
  letter-spacing: -0.3rpx;
}
.funnel__hint {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #94a3b8;
}
.funnel__tag {
  font-size: 20rpx;
  font-weight: 600;
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  flex-shrink: 0;
}
.funnel__tag--theory {
  color: #b45309;
  background: rgba(245, 158, 11, 0.12);
}
.funnel__tag--real {
  color: #0b7f7f;
  background: rgba(14, 165, 165, 0.12);
}
.funnel__list {
  display: flex;
  flex-direction: column;
}
.funnel__item {
  padding: 20rpx 0 18rpx;
  border-bottom: 1rpx dashed rgba(15, 23, 42, 0.08);
}
.funnel__item:first-child {
  padding-top: 4rpx;
}
.funnel__item:last-child {
  border-bottom: none;
  padding-bottom: 4rpx;
}
.funnel__item--muted {
  opacity: 0.55;
}
.funnel__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}
.funnel__label-wrap {
  display: flex;
  align-items: center;
}
.funnel__index {
  width: 36rpx;
  height: 36rpx;
  border-radius: 10rpx;
  background: #eef2ff;
  color: #3355ff;
  font-size: 20rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 14rpx;
  flex-shrink: 0;
}
.funnel__label {
  font-size: 28rpx;
  font-weight: 500;
  color: #1e293b;
}
.funnel__value {
  font-size: 30rpx;
  font-weight: 700;
  color: #3355ff;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.5rpx;
}
.funnel__value--muted {
  color: #94a3b8;
  font-weight: 500;
}
.funnel__track {
  height: 14rpx;
  border-radius: 999rpx;
  background: #eef2f7;
  overflow: hidden;
}
.funnel__bar {
  height: 100%;
  border-radius: 999rpx;
  background: linear-gradient(90deg, #3355ff, #7c9bff);
  transition: width 0.3s ease;
}
.funnel__bar--muted {
  background: repeating-linear-gradient(90deg, #dbe3ee 0, #dbe3ee 12rpx, #eef2f7 12rpx, #eef2f7 24rpx);
}
.funnel__bar--real {
  background: linear-gradient(90deg, #0ea5a5, #5eead4);
}
.funnel__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10rpx;
}
.funnel__sub {
  font-size: 22rpx;
  color: #94a3b8;
}
.funnel__drop {
  font-size: 22rpx;
  font-weight: 600;
  color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
}
.funnel__drop--up {
  color: #0f9f74;
  background: rgba(18, 184, 134, 0.1);
}
.funnel__drop--flat {
  color: #64748b;
  background: #f1f5f9;
}
.funnel__final {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24rpx;
  padding: 26rpx 30rpx;
  border-radius: 24rpx;
  background: linear-gradient(135deg, #f5f8ff 0%, #eef2ff 100%);
  border: 1rpx solid rgba(51, 85, 255, 0.12);
}
.funnel__final--real {
  background: linear-gradient(135deg, #f1fdfb 0%, #e6fbf8 100%);
  border-color: rgba(14, 165, 165, 0.16);
}
.funnel__final-left {
  display: flex;
  flex-direction: column;
}
.funnel__final-label {
  font-size: 26rpx;
  font-weight: 600;
  color: #1e293b;
}
.funnel__final-desc {
  margin-top: 4rpx;
  font-size: 22rpx;
  color: #64748b;
}
.funnel__final-value {
  font-size: 44rpx;
  font-weight: 700;
  color: #3355ff;
}
.funnel__final--real .funnel__final-value {
  color: #0b7f7f;
}
.funnel__note {
  margin-top: 18rpx;
  font-size: 22rpx;
  color: #94a3b8;
  line-height: 1.7;
}
</style>