<template>
  <view class="step-header">
    <view class="step-header__top">
      <view v-if="showBack" class="step-header__back" @tap="handleBack">返回</view>
      <view class="step-header__title-wrap">
        <text class="step-header__title">{{ title }}</text>
        <text v-if="subtitle" class="step-header__subtitle">{{ subtitle }}</text>
      </view>
      <text v-if="badge" class="step-header__badge">{{ badge }}</text>
    </view>
    <view v-if="percent !== null" class="step-header__progress">
      <view class="step-header__track">
        <view class="step-header__bar" :style="{ width: safePercent + '%' }"></view>
      </view>
      <text class="step-header__percent">完成度 {{ safePercent }}%</text>
    </view>
    <view v-else-if="steps.length > 0" class="step-header__steps">
      <view
        v-for="item in stepItems"
        :key="item.index"
        class="step-header__step"
        :class="{ 'step-header__step--active': step !== null && item.index <= step }"
      >
        <view class="step-header__dot">
          <text class="step-header__dot-text">{{ item.index }}</text>
        </view>
        <text class="step-header__dot-label">{{ item.label }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
// 步骤头：用于计算器 4 步流程与资料完善完成度展示。
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    percent?: number | null
    showBack?: boolean
    step?: number | null
    steps?: string[]
    badge?: string
  }>(),
  {
    subtitle: '',
    percent: null,
    showBack: false,
    step: null,
    steps: () => [],
    badge: ''
  }
)

const safePercent = computed(() => {
  if (props.percent === null || props.percent === undefined) return 0
  const value = Math.round(props.percent)
  if (value < 0) return 0
  if (value > 100) return 100
  return value
})

const stepItems = computed(() => {
  return props.steps.map((label, index) => ({ index: index + 1, label }))
})

function handleBack() {
  uni.navigateBack({
    delta: 1,
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}
</script>

<style lang="scss" scoped>
.step-header {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx 28rpx;
  margin-bottom: 24rpx;
  box-shadow: $wmc-shadow;
}
.step-header__top {
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.step-header__back {
  font-size: 26rpx;
  color: $uni-color-primary;
  padding: 8rpx 12rpx;
  border-radius: 8rpx;
  background: rgba(47, 107, 255, 0.08);
}
.step-header__title-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.step-header__title {
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.step-header__subtitle {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  margin-top: 6rpx;
  line-height: 1.6;
}
.step-header__badge {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  background: #f1f5f9;
  border-radius: 8rpx;
  padding: 6rpx 16rpx;
}
.step-header__progress {
  margin-top: 20rpx;
}
.step-header__track {
  height: 12rpx;
  border-radius: 999rpx;
  background: #eef2f7;
  overflow: hidden;
}
.step-header__bar {
  height: 100%;
  border-radius: 999rpx;
  background: $uni-color-primary;
  transition: width 0.2s ease;
}
.step-header__percent {
  display: block;
  text-align: right;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  margin-top: 10rpx;
}
.step-header__steps {
  display: flex;
  margin-top: 20rpx;
  gap: 12rpx;
}
.step-header__step {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
}
.step-header__dot {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: #eef2f7;
  display: flex;
  align-items: center;
  justify-content: center;
}
.step-header__step--active .step-header__dot {
  background: $uni-color-primary;
}
.step-header__dot-text {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.step-header__step--active .step-header__dot-text {
  color: #ffffff;
}
.step-header__dot-label {
  font-size: 22rpx;
  color: $uni-text-color-grey;
  text-align: center;
}
.step-header__step--active .step-header__dot-label {
  color: $uni-color-primary;
}
</style>
