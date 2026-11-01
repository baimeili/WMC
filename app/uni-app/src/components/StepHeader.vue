<template>
  <view class="step-header">
    <view class="step-header__top">
      <view v-if="showBack" class="step-header__back" @tap="handleBack">‹</view>
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
      <text class="step-header__percent wmc-num">完成度 {{ safePercent }}%</text>
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
  margin-bottom: 24rpx;
}
.step-header__top {
  display: flex;
  align-items: center;
}
.step-header__back {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #ffffff;
  color: #334155;
  font-size: 40rpx;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 18rpx;
  box-shadow: 0 6rpx 18rpx rgba(15, 23, 42, 0.08);
  flex-shrink: 0;
}
.step-header__title-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.step-header__title {
  font-size: 38rpx;
  font-weight: 700;
  color: #0b1220;
  letter-spacing: -0.5rpx;
}
.step-header__subtitle {
  margin-top: 6rpx;
  font-size: 24rpx;
  color: #64748b;
  line-height: 1.6;
}
.step-header__badge {
  font-size: 22rpx;
  font-weight: 600;
  color: #3355ff;
  background: rgba(51, 85, 255, 0.08);
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  flex-shrink: 0;
}
.step-header__progress {
  margin-top: 22rpx;
  display: flex;
  align-items: center;
}
.step-header__track {
  flex: 1;
  height: 12rpx;
  border-radius: 999rpx;
  background: #e7ecf5;
  overflow: hidden;
}
.step-header__bar {
  height: 100%;
  border-radius: 999rpx;
  background: linear-gradient(90deg, #3355ff, #7c9bff);
  transition: width 0.3s ease;
}
.step-header__percent {
  margin-left: 18rpx;
  font-size: 22rpx;
  font-weight: 600;
  color: #3355ff;
  flex-shrink: 0;
}
.step-header__steps {
  display: flex;
  align-items: flex-start;
  margin-top: 26rpx;
}
.step-header__step {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0.45;
}
.step-header__step--active {
  opacity: 1;
}
.step-header__dot {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: #e2e8f0;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
}
.step-header__step--active .step-header__dot {
  background: linear-gradient(135deg, #3355ff, #5a7bff);
  color: #ffffff;
  box-shadow: 0 6rpx 14rpx rgba(51, 85, 255, 0.28);
}
.step-header__dot-text {
  font-size: 22rpx;
  font-weight: 700;
}
.step-header__dot-label {
  margin-top: 8rpx;
  font-size: 20rpx;
  color: #64748b;
  text-align: center;
}
</style>
