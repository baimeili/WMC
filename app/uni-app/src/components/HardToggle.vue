<template>
  <view class="hard-toggle" @tap="toggle">
    <view class="hard-toggle__box" :class="{ 'hard-toggle__box--checked': modelValue }">
      <view v-if="modelValue" class="hard-toggle__dot"></view>
    </view>
    <view class="hard-toggle__body">
      <text class="hard-toggle__label">{{ label }}</text>
      <text v-if="hint" class="hard-toggle__hint">{{ hint }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
// 硬条件开关：开启后对方不满足该条件会直接不进入候选池。
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    label?: string
    hint?: string
  }>(),
  {
    label: '视为硬条件',
    hint: '不满足时直接不进入候选池'
  }
)

const emit = defineEmits<{ (event: 'update:modelValue', value: boolean): void }>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<style lang="scss" scoped>
.hard-toggle {
  display: flex;
  align-items: flex-start;
  padding: 12rpx 0;
}
.hard-toggle__box {
  width: 34rpx;
  height: 34rpx;
  border-radius: 8rpx;
  border: 1rpx solid $uni-border-color;
  background: #ffffff;
  margin-right: 16rpx;
  margin-top: 2rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.hard-toggle__box--checked {
  background: $uni-color-primary;
  border-color: $uni-color-primary;
}
.hard-toggle__dot {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  background: #ffffff;
}
.hard-toggle__body {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.hard-toggle__label {
  font-size: 26rpx;
  color: $uni-text-color;
}
.hard-toggle__hint {
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
  margin-top: 4rpx;
}
</style>
