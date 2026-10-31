<template>
  <PageShell>
    <view class="wmc-container">
      <view class="hero">
        <text class="hero__eyebrow">先算清楚，再认真选择</text>
        <text class="hero__title">你的择偶条件，会筛掉多少人？</text>
        <text class="hero__desc">
          用 4 步填写你的条件，得到理论市场的条件占比、顺序漏斗与条件敏感性分析。无需注册即可计算。
        </text>
        <button class="wmc-btn wmc-btn--primary wmc-btn--block hero__cta" @tap="goCalculator">算算我的候选池</button>
        <view class="hero__secondary">
          <text v-if="!loggedIn" class="wmc-link" @tap="goLogin">已有账号？登录 / 注册</text>
          <text v-else class="wmc-link" @tap="goDiscovery">查看本站真实候选人</text>
        </view>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title">
          <text>两个市场，必须分开看</text>
        </view>
        <MarketBadge market="theoretical" />
        <view class="market-gap"></view>
        <MarketBadge market="real" />
        <view class="wmc-divider"></view>
        <text class="wmc-hint">
          理论市场估算回答「符合条件的人大约占多少」，本站真实市场回答「有多少人与我双向满足硬条件」。两者数据来源不同，不能相互替代。
        </text>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title">
          <text>计算流程</text>
        </view>
        <view v-for="item in steps" :key="item.index" class="flow-item">
          <view class="flow-item__index">{{ item.index }}</view>
          <view class="flow-item__body">
            <text class="flow-item__title">{{ item.title }}</text>
            <text class="flow-item__desc">{{ item.desc }}</text>
          </view>
        </view>
      </view>

      <view class="wmc-card wmc-card--flat">
        <view class="wmc-section-title">
          <text>理性提示</text>
        </view>
        <text class="wmc-hint">
          理论市场结果基于后台导入的基准人口数据集做统计估算，不是本站真实用户数量，也不能计算双向匹配率。匹配分数仅根据双方填写条件计算，不代表婚姻成功概率。
        </text>
      </view>

      <view class="footer">
        <text class="footer__text">婚恋市场计算器 · 版本 {{ appVersion }}</text>
        <text class="footer__link" @tap="goSettings">设置与隐私</text>
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 首页：首要功能为计算器，其次为登录与本站真实候选池入口，并明确两个市场的区别。
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import { useAuthStore } from '@/stores/auth'
import { APP_VERSION } from '@/utils/constants'

const auth = useAuthStore()
const appVersion = APP_VERSION
const loggedIn = computed(() => auth.isLoggedIn)

const steps = [
  { index: 1, title: '我是谁', desc: '填写性别、年龄、城市等基础信息，仅用于计算' },
  { index: 2, title: '我希望找谁', desc: '设置目标性别、年龄区间、城市与关键条件' },
  { index: 3, title: '我的硬条件', desc: '把条件标记为必须、偏好或不在意，只有必须条件参与估算' },
  { index: 4, title: '计算', desc: '得到理论市场占比、顺序漏斗与敏感性分析' }
]

onShow(() => {
  auth.syncFromStorage()
})

function goCalculator() {
  uni.navigateTo({ url: '/pages/calculator/index' })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login' })
}

function goDiscovery() {
  uni.navigateTo({ url: '/pages/discovery/index' })
}

function goSettings() {
  uni.navigateTo({ url: '/pages/settings/index' })
}
</script>

<style lang="scss" scoped>
.hero {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 40rpx 32rpx;
  margin-bottom: 24rpx;
  box-shadow: $wmc-shadow;
}
.hero__eyebrow {
  display: inline-block;
  font-size: 24rpx;
  color: $uni-color-primary;
  background: rgba(47, 107, 255, 0.08);
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
  margin-bottom: 20rpx;
}
.hero__title {
  display: block;
  font-size: 44rpx;
  font-weight: 600;
  line-height: 1.4;
  color: $uni-text-color;
}
.hero__desc {
  display: block;
  margin-top: 16rpx;
  font-size: 26rpx;
  line-height: 1.8;
  color: $uni-text-color-grey;
}
.hero__cta {
  margin-top: 32rpx;
}
.hero__secondary {
  margin-top: 24rpx;
  text-align: center;
}
.market-gap {
  height: 16rpx;
}
.flow-item {
  display: flex;
  align-items: flex-start;
  padding: 18rpx 0;
}
.flow-item__index {
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: rgba(47, 107, 255, 0.1);
  color: $uni-color-primary;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;
  flex-shrink: 0;
}
.flow-item__body {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.flow-item__title {
  font-size: 28rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.flow-item__desc {
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
  margin-top: 4rpx;
}
.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 8rpx 40rpx;
}
.footer__text {
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
}
.footer__link {
  font-size: 22rpx;
  color: $uni-text-color-grey;
}
</style>
