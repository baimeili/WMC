<template>
  <PageShell>
    <view class="wmc-container">
      <!-- 首屏：深色渐变 Hero，强调「先算清楚，再认真选择」 -->
      <view class="hero">
        <view class="hero__glow hero__glow--a"></view>
        <view class="hero__glow hero__glow--b"></view>
        <view class="hero__inner">
          <view class="hero__eyebrow">先算清楚，再认真选择</view>
          <text class="hero__title">你的择偶条件，\n会筛掉多少人？</text>
          <text class="hero__desc">
            4 步填完，得到理论市场的条件占比、顺序漏斗与条件敏感性分析。无需注册即可计算。
          </text>
          <view class="hero__meta">
            <text class="hero__meta-item">4 步</text>
            <text class="hero__meta-dot"></text>
            <text class="hero__meta-item">约 2 分钟</text>
            <text class="hero__meta-dot"></text>
            <text class="hero__meta-item">不采集精确住址与收入</text>
          </view>
          <button class="wmc-btn hero__cta" @tap="goCalculator">算算我的候选池</button>
          <view class="hero__secondary">
            <text v-if="!loggedIn" class="hero__link" @tap="goLogin">已有账号？登录 / 注册</text>
            <text v-else class="hero__link" @tap="goDiscovery">查看本站真实候选人 →</text>
          </view>
        </view>
      </view>

      <!-- 两个市场 -->
      <view class="wmc-section-head">
        <view class="wmc-section-head__main">
          <text class="wmc-section-head__title">两个市场，必须分开看</text>
          <text class="wmc-section-head__desc">数据来源不同，结论不能互相替代</text>
        </view>
      </view>

      <view class="market market--theory">
        <view class="market__head">
          <view class="market__dot"></view>
          <text class="market__name">理论市场估算</text>
          <text class="market__tag">统计估算</text>
        </view>
        <text class="market__line">「符合我条件的人，大约占目标人群多少比例？」</text>
        <text class="market__desc">基于后台导入的基准人口数据集加权计算，只能给出占比、漏斗与敏感性，不能计算双向匹配率。</text>
      </view>

      <view class="market market--real">
        <view class="market__head">
          <view class="market__dot"></view>
          <text class="market__name">本站真实用户池</text>
          <text class="market__tag">双向硬筛</text>
        </view>
        <text class="market__line">「有多少人和我双向满足硬条件？」</text>
        <text class="market__desc">只统计已注册、资料达标并主动加入婚恋池的用户；双方硬条件全部满足才算候选人。</text>
      </view>

      <!-- 计算流程 -->
      <view class="wmc-card">
        <view class="wmc-section-title">
          <text>计算流程</text>
          <text class="flow__badge">共 4 步</text>
        </view>
        <view class="flow">
          <view v-for="(item, index) in steps" :key="item.index" class="flow__row">
            <view class="flow__left">
              <view class="flow__index" :class="{ 'flow__index--last': index === steps.length - 1 }">
                <text>{{ item.index }}</text>
              </view>
              <view v-if="index < steps.length - 1" class="flow__line"></view>
            </view>
            <view class="flow__body">
              <text class="flow__title">{{ item.title }}</text>
              <text class="flow__desc">{{ item.desc }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 理性提示 -->
      <view class="wmc-notice wmc-notice--warning">
        <text>理论市场结果基于基准人口数据集做统计估算，不是本站真实用户数量，也不能计算双向匹配率。匹配分数仅根据双方填写条件计算，不代表婚姻成功概率。</text>
      </view>

      <view class="footer">
        <text class="footer__text">婚恋市场计算器 · v{{ appVersion }}</text>
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
import { useAuthStore } from '@/stores/auth'
import { APP_VERSION } from '@/utils/constants'

const auth = useAuthStore()
const appVersion = APP_VERSION
const loggedIn = computed(() => auth.isLoggedIn)

const steps = [
  { index: 1, title: '我是谁', desc: '性别、年龄、城市等基础信息，仅用于计算' },
  { index: 2, title: '我希望找谁', desc: '目标性别、年龄区间、城市与关键条件' },
  { index: 3, title: '我的硬条件', desc: '条件标记为必须 / 偏好 / 不在意' },
  { index: 4, title: '计算', desc: '理论占比、顺序漏斗与敏感性分析' }
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
/* ---------- Hero ---------- */
.hero {
  position: relative;
  overflow: hidden;
  border-radius: 40rpx;
  background: linear-gradient(150deg, #0d1533 0%, #1d2a72 46%, #3851d6 100%);
  box-shadow: 0 28rpx 60rpx rgba(13, 21, 51, 0.32);
  margin-bottom: 40rpx;
}
.hero__glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(2rpx);
  opacity: 0.55;
  pointer-events: none;
}
.hero__glow--a {
  width: 380rpx;
  height: 380rpx;
  right: -120rpx;
  top: -140rpx;
  background: radial-gradient(circle, rgba(122, 148, 255, 0.55) 0%, rgba(122, 148, 255, 0) 70%);
}
.hero__glow--b {
  width: 300rpx;
  height: 300rpx;
  left: -110rpx;
  bottom: -160rpx;
  background: radial-gradient(circle, rgba(74, 222, 210, 0.35) 0%, rgba(74, 222, 210, 0) 70%);
}
.hero__inner {
  position: relative;
  padding: 52rpx 40rpx 44rpx;
}
.hero__eyebrow {
  display: inline-flex;
  align-items: center;
  font-size: 22rpx;
  font-weight: 600;
  letter-spacing: 2rpx;
  color: #dbe3ff;
  background: rgba(255, 255, 255, 0.14);
  border: 1rpx solid rgba(255, 255, 255, 0.18);
  padding: 8rpx 20rpx;
  border-radius: 999rpx;
}
.hero__title {
  display: block;
  margin-top: 28rpx;
  font-size: 58rpx;
  line-height: 1.28;
  font-weight: 700;
  letter-spacing: -1rpx;
  color: #ffffff;
  white-space: pre-line;
}
.hero__desc {
  display: block;
  margin-top: 20rpx;
  font-size: 26rpx;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.76);
}
.hero__meta {
  display: flex;
  align-items: center;
  margin-top: 26rpx;
}
.hero__meta-item {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.62);
}
.hero__meta-dot {
  width: 6rpx;
  height: 6rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.32);
  margin: 0 16rpx;
}
.hero__cta {
  margin-top: 36rpx;
  background: #ffffff;
  color: #1d2a72;
  box-shadow: 0 18rpx 40rpx rgba(5, 12, 40, 0.35);
  font-size: 32rpx;
}
.hero__secondary {
  margin-top: 24rpx;
  text-align: center;
}
.hero__link {
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.82);
  border-bottom: 1rpx solid rgba(255, 255, 255, 0.32);
  padding-bottom: 2rpx;
}

/* ---------- 市场卡片 ---------- */
.market {
  position: relative;
  background: #ffffff;
  border-radius: 28rpx;
  padding: 30rpx 32rpx 32rpx 38rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 10rpx 32rpx rgba(15, 23, 42, 0.06);
  border: 1rpx solid rgba(15, 23, 42, 0.05);
  overflow: hidden;
}
.market::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 8rpx;
}
.market--theory::before {
  background: linear-gradient(180deg, #fbbf24, #f59e0b);
}
.market--real::before {
  background: linear-gradient(180deg, #2dd4bf, #0ea5a5);
}
.market__head {
  display: flex;
  align-items: center;
}
.market__dot {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  margin-right: 14rpx;
}
.market--theory .market__dot {
  background: #f59e0b;
  box-shadow: 0 0 0 8rpx rgba(245, 158, 11, 0.14);
}
.market--real .market__dot {
  background: #0ea5a5;
  box-shadow: 0 0 0 8rpx rgba(14, 165, 165, 0.14);
}
.market__name {
  font-size: 30rpx;
  font-weight: 700;
  color: #0b1220;
}
.market__tag {
  margin-left: 16rpx;
  font-size: 20rpx;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
}
.market--theory .market__tag {
  color: #b45309;
  background: rgba(245, 158, 11, 0.12);
}
.market--real .market__tag {
  color: #0b7f7f;
  background: rgba(14, 165, 165, 0.12);
}
.market__line {
  display: block;
  margin-top: 18rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #1e293b;
  line-height: 1.5;
}
.market__desc {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #64748b;
  line-height: 1.8;
}

/* ---------- 流程时间线 ---------- */
.flow__badge {
  font-size: 22rpx;
  color: #64748b;
  background: #f1f5f9;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  font-weight: 500;
}
.flow__row {
  display: flex;
  align-items: stretch;
}
.flow__left {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 64rpx;
  flex-shrink: 0;
}
.flow__index {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #3355ff, #5a7bff);
  color: #ffffff;
  font-size: 24rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 6rpx 14rpx rgba(51, 85, 255, 0.28);
}
.flow__index--last {
  background: linear-gradient(135deg, #0ea5a5, #2dd4bf);
  box-shadow: 0 6rpx 14rpx rgba(14, 165, 165, 0.28);
}
.flow__line {
  flex: 1;
  width: 2rpx;
  background: linear-gradient(180deg, rgba(51, 85, 255, 0.35), rgba(51, 85, 255, 0.08));
  margin: 8rpx 0;
}
.flow__body {
  flex: 1;
  padding: 4rpx 0 32rpx 24rpx;
  display: flex;
  flex-direction: column;
}
.flow__title {
  font-size: 28rpx;
  font-weight: 600;
  color: #0b1220;
}
.flow__desc {
  margin-top: 6rpx;
  font-size: 24rpx;
  color: #64748b;
  line-height: 1.7;
}

/* ---------- 底部 ---------- */
.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 8rpx 16rpx;
}
.footer__text {
  font-size: 22rpx;
  color: #94a3b8;
}
.footer__link {
  font-size: 22rpx;
  color: #64748b;
}
</style>
