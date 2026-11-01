<template>
  <view class="page-shell">
    <slot />
  </view>
</template>

<script setup lang="ts">
// 页面外壳：只保留容器本身与安全区处理，视觉样式统一在 src/styles/global.scss
</script>

<style lang="scss" scoped>
.page-shell {
  min-height: 100vh;
  background: $uni-bg-color;
  color: $uni-text-color;
  box-sizing: border-box;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}
</style>
