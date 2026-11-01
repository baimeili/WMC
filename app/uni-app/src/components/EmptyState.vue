<template>
  <view class="empty-state">
    <view class="empty-state__icon-wrap">
      <view class="empty-state__ring"></view>
      <view class="empty-state__icon">{{ iconText }}</view>
    </view>
    <text class="empty-state__title">{{ title }}</text>
    <text v-if="description" class="empty-state__desc">{{ description }}</text>
    <view v-if="actionText" class="empty-state__action">
      <button class="wmc-btn wmc-btn--primary wmc-btn--block" @tap="handleAction">{{ actionText }}</button>
    </view>
    <slot />
  </view>
</template>

<script setup lang="ts">
// 空状态：文案保持中性，不出现指责性表达；真实候选池空态使用冷启动说明。
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    title: string
    description?: string
    actionText?: string
    iconText?: string
  }>(),
  {
    description: '',
    actionText: '',
    iconText: '—'
  }
)

const emit = defineEmits<{ (event: 'action'): void }>()

const iconText = computed(() => props.iconText)

function handleAction() {
  emit('action')
}
</script>

<style lang="scss" scoped>
.empty-state {
  background: #ffffff;
  border-radius: 28rpx;
  padding: 64rpx 36rpx 56rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 12rpx 40rpx rgba(15, 23, 42, 0.08);
  border: 1rpx solid rgba(15, 23, 42, 0.04);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.empty-state__icon-wrap {
  position: relative;
  width: 132rpx;
  height: 132rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 28rpx;
}
.empty-state__ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, rgba(51, 85, 255, 0.16), rgba(51, 85, 255, 0.02) 70%);
  border: 1rpx dashed rgba(51, 85, 255, 0.24);
}
.empty-state__icon {
  position: relative;
  width: 84rpx;
  height: 84rpx;
  border-radius: 50%;
  background: #ffffff;
  color: #3355ff;
  font-size: 34rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 10rpx 24rpx rgba(51, 85, 255, 0.16);
}
.empty-state__title {
  font-size: 32rpx;
  font-weight: 700;
  color: #0b1220;
  margin-bottom: 12rpx;
}
.empty-state__desc {
  font-size: 24rpx;
  color: #64748b;
  line-height: 1.85;
}
.empty-state__action {
  width: 100%;
  margin-top: 36rpx;
}
</style>
