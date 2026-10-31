<template>
  <PageShell>
    <view class="wmc-container">
      <view class="wmc-card">
        <view class="wmc-row">
          <view class="wmc-col">
            <text class="wmc-value">{{ loggedIn ? (nickname || '已登录') : '未登录' }}</text>
            <text class="wmc-hint">{{ loggedIn ? '资料完成度 ' + completion + '%' : '登录后可管理账号与隐私设置' }}</text>
          </view>
          <text v-if="!loggedIn" class="wmc-link" @tap="goLogin">去登录</text>
        </view>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title"><text>隐私与账号</text></view>
        <view class="wmc-list-item" @tap="goPrivacy">
          <text class="wmc-value">隐私与协议</text>
          <text class="wmc-link">查看</text>
        </view>
        <view class="wmc-list-item" @tap="goAccount">
          <text class="wmc-value">账号与安全</text>
          <text class="wmc-link">管理</text>
        </view>
        <view class="wmc-list-item" @tap="goProfile">
          <text class="wmc-value">个人资料</text>
          <text class="wmc-link">编辑</text>
        </view>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title"><text>关于</text></view>
        <view class="wmc-list-item" @tap="goCalculator">
          <text class="wmc-value">理论市场计算器</text>
          <text class="wmc-link">去计算</text>
        </view>
        <view class="wmc-list-item" @tap="goDiscovery">
          <text class="wmc-value">本站真实候选池</text>
          <text class="wmc-link">查看</text>
        </view>
        <view class="wmc-list-item">
          <text class="wmc-value">版本</text>
          <text class="wmc-hint">{{ appVersion }}</text>
        </view>
      </view>

      <view class="wmc-card wmc-card--flat">
        <text class="wmc-hint about-text">
          本产品不承诺婚姻成功率。理论市场数据来自基准人口数据集估算，本站真实市场数据来自已注册并主动加入婚恋池的用户，两者不能混用。
        </text>
      </view>

      <view v-if="loggedIn" class="wmc-card">
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block" @tap="logout">退出登录</button>
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 设置入口：隐私与协议、账号与安全、关于信息。
import { computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { APP_VERSION } from '@/utils/constants'

const auth = useAuthStore()
const profileStore = useProfileStore()
const appVersion = APP_VERSION
const loggedIn = computed(() => auth.isLoggedIn)
const nickname = computed(() => (profileStore.profile ? profileStore.profile.nickname : ''))
const completion = computed(() => auth.profileCompletion)

onShow(() => {
  auth.syncFromStorage()
})

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/settings/index') })
}
function goPrivacy() {
  uni.navigateTo({ url: '/pages/settings/privacy' })
}
function goAccount() {
  uni.navigateTo({ url: '/pages/settings/account' })
}
function goProfile() {
  uni.navigateTo({ url: '/pages/profile/index' })
}
function goCalculator() {
  uni.navigateTo({ url: '/pages/calculator/index' })
}
function goDiscovery() {
  uni.navigateTo({ url: '/pages/discovery/index' })
}
async function logout() {
  await auth.logout()
  uni.reLaunch({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.about-text {
  display: block;
}
</style>
