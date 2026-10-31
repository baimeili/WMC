<template>
  <view class="score-bar">
    <view class="score-bar__head">
      <text class="score-bar__label">{{ label }}</text>
      <text class="score-bar__value">{{ formatScore(safeScore) }}</text>
    </view>
    <view class="score-bar__track">
      <view class="score-bar__fill" :style="{ width: widthText }"></view>
    </view>
    <text v-if="hint" class="score-bar__hint">{{ hint }}</text>
  </view>
</template>

<script setup lang="ts">
// 单维度得分条：用于候选人「为什么匹配」的分维度得分展示。
import { computed } from 'vue'
import { formatScore } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    label: string
    score: number
    hint?: string
  }>(),
  { hint: '' }
)

const safeScore = computed(() => {
  const value = Number(props.score)
  if (!isFinite(value)) return 0
  if (value < 0) return 0
  if (value > 1) return 1
  return value
})

const widthText = computed(() => Math.max(safeScore.value * 100, 2).toFixed(1) + '%')
</script>

<style lang="scss" scoped>
.score-bar {
  margin-bottom: 22rpx;
}
.score-bar__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10rpx;
}
.score-bar__label {
  font-size: 26rpx;
  color: $uni-text-color;
}
.score-bar__value {
  font-size: 26rpx;
  font-weight: 600;
  color: $uni-color-primary;
}
.score-bar__track {
  height: 14rpx;
  border-radius: 999rpx;
  background: #eef2f7;
  overflow: hidden;
}
.score-bar__fill {
  height: 100%;
  border-radius: 999rpx;
  background: linear-gradient(90deg, #2f6bff, #6f9bff);
}
.score-bar__hint {
  display: block;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-top: 8rpx;
  line-height: 1.6;
}
</style>
