<template>
  <view class="page-shell">
    <slot />
  </view>
</template>

<script setup lang="ts">
// 页面外壳：统一页面背景，并承载全站公共样式（非 scoped，全局生效）。
</script>

<style lang="scss">
.page-shell {
  min-height: 100vh;
  background: $uni-bg-color;
  color: $uni-text-color;
  box-sizing: border-box;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}
.wmc-container {
  padding: 24rpx;
}
.wmc-card {
  background: $uni-bg-color-card;
  border-radius: 16rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
  box-shadow: $wmc-shadow;
}
.wmc-card--flat {
  box-shadow: none;
  border: 1rpx solid $uni-border-color;
}
.wmc-title {
  font-size: 36rpx;
  font-weight: 600;
  color: $uni-text-color;
  line-height: 1.4;
}
.wmc-title--sm {
  font-size: 30rpx;
}
.wmc-subtitle {
  font-size: 26rpx;
  color: $uni-text-color-grey;
  line-height: 1.6;
  margin-top: 8rpx;
}
.wmc-section {
  margin-bottom: 24rpx;
}
.wmc-section-title {
  font-size: 30rpx;
  font-weight: 600;
  margin-bottom: 16rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.wmc-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.wmc-col {
  display: flex;
  flex-direction: column;
}
.wmc-label {
  font-size: 26rpx;
  color: $uni-text-color-grey;
}
.wmc-value {
  font-size: 30rpx;
  color: $uni-text-color;
  font-weight: 500;
}
.wmc-hint {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.wmc-muted {
  color: $uni-text-color-grey;
}
.wmc-strong {
  font-weight: 600;
}
.wmc-small {
  font-size: 24rpx;
}
.wmc-divider {
  height: 1rpx;
  background: $uni-border-color;
  margin: 20rpx 0;
}
.wmc-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 88rpx;
  padding: 0 32rpx;
  border-radius: 12rpx;
  font-size: 30rpx;
  font-weight: 500;
  border: none;
  margin: 0;
  line-height: 1.2;
  box-sizing: border-box;
}
.wmc-btn::after {
  border: none;
}
.wmc-btn--primary {
  background: $uni-color-primary;
  color: #ffffff;
}
.wmc-btn--ghost {
  background: transparent;
  color: $uni-color-primary;
  border: 1rpx solid $uni-color-primary;
}
.wmc-btn--soft {
  background: rgba(47, 107, 255, 0.08);
  color: $uni-color-primary;
}
.wmc-btn--danger {
  background: transparent;
  color: $uni-color-error;
  border: 1rpx solid $uni-color-error;
}
.wmc-btn--disabled {
  background: #c8d2e2;
  color: #ffffff;
}
.wmc-btn--block {
  width: 100%;
}
.wmc-btn + .wmc-btn {
  margin-top: 20rpx;
}
.wmc-btn-row {
  display: flex;
  gap: 20rpx;
}
.wmc-btn-row .wmc-btn {
  flex: 1;
  margin-top: 0;
}
.wmc-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12rpx 24rpx;
  border-radius: 999rpx;
  border: 1rpx solid $uni-border-color;
  background: #ffffff;
  color: $uni-text-color-grey;
  font-size: 26rpx;
  margin: 0 16rpx 16rpx 0;
}
.wmc-chip--active {
  border-color: $uni-color-primary;
  background: rgba(47, 107, 255, 0.08);
  color: $uni-color-primary;
  font-weight: 500;
}
.wmc-chip--small {
  padding: 8rpx 18rpx;
  font-size: 24rpx;
}
.wmc-field {
  margin-bottom: 24rpx;
}
.wmc-field-label {
  font-size: 26rpx;
  color: $uni-text-color-grey;
  margin-bottom: 12rpx;
  display: block;
}
.wmc-input {
  height: 84rpx;
  border: 1rpx solid $uni-border-color;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  color: $uni-text-color;
  background: #ffffff;
  box-sizing: border-box;
}
.wmc-picker {
  height: 84rpx;
  border: 1rpx solid $uni-border-color;
  border-radius: 12rpx;
  padding: 0 24rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 30rpx;
  background: #ffffff;
  box-sizing: border-box;
}
.wmc-picker__text {
  color: $uni-text-color;
}
.wmc-picker__placeholder {
  color: $uni-text-color-placeholder;
}
.wmc-picker__arrow {
  color: $uni-text-color-placeholder;
  font-size: 24rpx;
}
.wmc-textarea {
  width: 100%;
  min-height: 180rpx;
  border: 1rpx solid $uni-border-color;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  box-sizing: border-box;
  background: #ffffff;
}
.wmc-tag {
  display: inline-flex;
  align-items: center;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  background: rgba(47, 107, 255, 0.08);
  color: $uni-color-primary;
  font-size: 22rpx;
  margin-right: 12rpx;
}
.wmc-tag--warning {
  background: rgba(230, 162, 60, 0.12);
  color: #b7791f;
}
.wmc-tag--danger {
  background: rgba(227, 77, 89, 0.1);
  color: $uni-color-error;
}
.wmc-tag--success {
  background: rgba(23, 166, 115, 0.1);
  color: $uni-color-success;
}
.wmc-tag--muted {
  background: #f1f5f9;
  color: $uni-text-color-grey;
}
.wmc-notice {
  background: rgba(47, 107, 255, 0.06);
  border: 1rpx solid rgba(47, 107, 255, 0.15);
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 24rpx;
  color: #33557f;
  line-height: 1.7;
}
.wmc-notice--warning {
  background: rgba(230, 162, 60, 0.08);
  border-color: rgba(230, 162, 60, 0.25);
  color: #8a5a12;
}
.wmc-notice--danger {
  background: rgba(227, 77, 89, 0.06);
  border-color: rgba(227, 77, 89, 0.2);
  color: #8c2f37;
}
.wmc-stat-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}
.wmc-stat {
  flex: 1;
  min-width: 200rpx;
  background: #f8fafc;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
}
.wmc-stat-value {
  font-size: 40rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.wmc-stat-label {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  margin-top: 6rpx;
}
.wmc-link {
  color: $uni-color-primary;
  font-size: 26rpx;
}
.wmc-error {
  color: $uni-color-error;
  font-size: 26rpx;
  line-height: 1.6;
}
.wmc-list-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 26rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}
.wmc-list-item:last-child {
  border-bottom: none;
}
.wmc-footer-actions {
  padding: 24rpx;
}
.wmc-center {
  display: flex;
  align-items: center;
  justify-content: center;
}
.wmc-flex-1 {
  flex: 1;
}
</style>
