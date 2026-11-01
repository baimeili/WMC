<template>
  <view class="market-badge" :class="['market-badge--' + market, small ? 'market-badge--small' : '']">
    <view class="market-badge__head">
      <view class="market-badge__dot"></view>
      <text class="market-badge__tag">{{ copy.name }}</text>
    </view>
    <text v-if="showDesc" class="market-badge__desc">{{ copy.desc }}</text>
    <slot />
  </view>
</template>

<script setup lang="ts">
// 市场标识：明确区分「理论市场估算」与「本站真实用户池」，禁止混用。
import { computed } from 'vue'
import { MARKET_COPY } from '@/utils/constants'

type Market = 'theoretical' | 'real'

const props = withDefaults(
  defineProps<{
    market: Market
    showDesc?: boolean
    small?: boolean
  }>(),
  { showDesc: true, small: false }
)

const copy = computed(() => MARKET_COPY[props.market])
</script>

<style lang="scss" scoped>
.market-badge {
  border-radius: 22rpx;
  padding: 22rpx 26rpx;
  border: 1rpx solid transparent;
}
.market-badge--theoretical {
  background: rgba(245, 158, 11, 0.07);
  border-color: rgba(245, 158, 11, 0.24);
}
.market-badge--real {
  background: rgba(14, 165, 165, 0.07);
  border-color: rgba(14, 165, 165, 0.24);
}
.market-badge__head {
  display: flex;
  align-items: center;
}
.market-badge__dot {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  margin-right: 14rpx;
  flex-shrink: 0;
}
.market-badge--theoretical .market-badge__dot {
  background: #f59e0b;
  box-shadow: 0 0 0 8rpx rgba(245, 158, 11, 0.14);
}
.market-badge--real .market-badge__dot {
  background: #0ea5a5;
  box-shadow: 0 0 0 8rpx rgba(14, 165, 165, 0.14);
}
.market-badge__tag {
  font-size: 26rpx;
  font-weight: 700;
  letter-spacing: 0.3rpx;
}
.market-badge--theoretical .market-badge__tag {
  color: #b45309;
}
.market-badge--real .market-badge__tag {
  color: #0b7f7f;
}
.market-badge__desc {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.75;
  color: #475569;
}
.market-badge--small {
  padding: 12rpx 20rpx;
  border-radius: 999rpx;
}
.market-badge--small .market-badge__tag {
  font-size: 22rpx;
}
</style>
