<template>
  <view class="market-badge" :class="['market-badge--' + market, small ? 'market-badge--small' : '']">
    <view class="market-badge__tag">{{ copy.name }}</view>
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
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  border: 1rpx solid transparent;
}
.market-badge--theoretical {
  background: rgba(245, 158, 11, 0.08);
  border-color: rgba(245, 158, 11, 0.28);
}
.market-badge--real {
  background: rgba(47, 107, 255, 0.07);
  border-color: rgba(47, 107, 255, 0.24);
}
.market-badge__tag {
  display: inline-flex;
  align-items: center;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}
.market-badge--theoretical .market-badge__tag {
  background: rgba(245, 158, 11, 0.16);
  color: #92580a;
}
.market-badge--real .market-badge__tag {
  background: rgba(47, 107, 255, 0.14);
  color: #1f4fd0;
}
.market-badge__desc {
  display: block;
  font-size: 24rpx;
  line-height: 1.7;
  color: #4a5568;
}
.market-badge--small {
  padding: 12rpx 16rpx;
}
.market-badge--small .market-badge__tag {
  margin-bottom: 0;
}
</style>
