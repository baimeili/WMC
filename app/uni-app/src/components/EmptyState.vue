<template>
  <view class="empty-state">
    <view class="empty-state__icon">{{ iconText }}</view>
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
  border-radius: 16rpx;
  padding: 56rpx 32rpx;
  margin-bottom: 24rpx;
  box-shadow: $wmc-shadow;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.empty-state__icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  background: #f1f5f9;
  color: $uni-text-color-placeholder;
  font-size: 36rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
}
.empty-state__title {
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 12rpx;
}
.empty-state__desc {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.8;
}
.empty-state__action {
  width: 100%;
  margin-top: 32rpx;
}
</style>
