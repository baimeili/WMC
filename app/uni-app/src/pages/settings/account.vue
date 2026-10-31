<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-card">
        <EmptyState title="登录后管理账号" description="账号与安全相关操作需要登录。" action-text="去登录" @action="goLogin" />
      </view>

      <template v-else>
        <view class="wmc-card">
          <view class="wmc-section-title"><text>账号信息</text></view>
          <view class="wmc-list-item">
            <text class="wmc-value">用户 ID</text>
            <text class="wmc-hint">{{ maskedUserId }}</text>
          </view>
          <view class="wmc-list-item">
            <text class="wmc-value">资料完成度</text>
            <text class="wmc-hint">{{ completion }}%</text>
          </view>
          <view class="wmc-list-item">
            <text class="wmc-value">真实婚恋池状态</text>
            <text class="wmc-hint">{{ inMarketText }}</text>
          </view>
          <view v-if="summary" class="wmc-list-item">
            <text class="wmc-value">双向基础候选人</text>
            <text class="wmc-hint">{{ eligibleText }}</text>
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>婚恋池管理</text></view>
          <text class="wmc-hint">退出婚恋池后，你的资料不会再出现在其他人的真实候选池中，可随时重新加入。</text>
          <button v-if="summary && summary.in_market" class="wmc-btn wmc-btn--ghost wmc-btn--block action-btn" :disabled="acting" @tap="leaveMarketAction">退出婚恋池</button>
          <button v-else class="wmc-btn wmc-btn--soft wmc-btn--block action-btn" :disabled="acting" @tap="joinMarketAction">加入婚恋池</button>
          <text v-if="message" class="wmc-error">{{ message }}</text>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>本地数据</text></view>
          <text class="wmc-hint">清除本地缓存会退出登录并删除计算器草稿与本地缓存，不会删除服务端数据。</text>
          <button class="wmc-btn wmc-btn--soft wmc-btn--block action-btn" @tap="clearCache">清除本地缓存</button>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>危险操作</text></view>
          <text class="wmc-hint">注销账号会立即退出婚恋池，并在保留期后清理数据。请在确认前自行保存需要保留的信息。</text>
          <button class="wmc-btn wmc-btn--danger wmc-btn--block action-btn" :disabled="acting" @tap="deleteAccountAction">注销账号</button>
        </view>

        <view class="wmc-card wmc-card--flat">
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block" @tap="logout">退出登录</button>
        </view>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 账号与安全：账号信息、婚恋池管理、清除缓存、注销账号。
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { useDiscoveryStore } from '@/stores/discovery'
import { deleteAccount, joinMarket, leaveMarket } from '@/api/profile'
import { maskCount } from '@/utils/format'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const discovery = useDiscoveryStore()
const acting = ref(false)
const message = ref('')

const loggedIn = computed(() => auth.isLoggedIn)
const completion = computed(() => auth.profileCompletion)
const summary = computed(() => discovery.summary)
const maskedUserId = computed(() => {
  const id = auth.userId || ''
  if (id.length <= 8) return id || '—'
  return id.slice(0, 8) + '…'
})
const inMarketText = computed(() => {
  if (!summary.value) return '加载中…'
  return summary.value.in_market ? '已加入' : '未加入'
})
const eligibleText = computed(() => {
  if (!summary.value) return '—'
  if (summary.value.eligible_count_text) return summary.value.eligible_count_text
  return maskCount(summary.value.reciprocal_count) + ' 人'
})

onShow(async () => {
  auth.syncFromStorage()
  if (!auth.isLoggedIn) return
  try {
    await discovery.loadSummary(true)
  } catch (error) {
    // 概览失败不阻塞账号操作
  }
})

async function leaveMarketAction() {
  if (acting.value) return
  uni.showModal({
    title: '退出婚恋池',
    content: '退出后你的资料不会再出现在其他人的真实候选池中。确定继续吗？',
    confirmText: '确认退出',
    cancelText: '取消',
    success: async (response) => {
      if (!response.confirm) return
      acting.value = true
      try {
        await leaveMarket()
        uni.showToast({ title: '已退出婚恋池', icon: 'none' })
        await discovery.loadSummary(true)
      } catch (error) {
        message.value = errorMessage(error, '操作失败，请稍后重试')
      } finally {
        acting.value = false
      }
    }
  })
}

async function joinMarketAction() {
  if (acting.value) return
  acting.value = true
  message.value = ''
  try {
    await joinMarket()
    uni.showToast({ title: '已加入婚恋池', icon: 'none' })
    await discovery.loadSummary(true)
  } catch (error) {
    message.value = errorMessage(error, '加入失败，请先检查资料完成度与协议同意状态')
  } finally {
    acting.value = false
  }
}

function clearCache() {
  uni.showModal({
    title: '清除本地缓存',
    content: '将退出登录并删除本地计算器草稿与缓存，确定继续吗？',
    confirmText: '确认清除',
    cancelText: '取消',
    success: (response) => {
      if (!response.confirm) return
      try {
        uni.clearStorageSync()
      } catch (error) {
        // 忽略
      }
      auth.clearLocal()
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function deleteAccountAction() {
  if (acting.value) return
  uni.showModal({
    title: '注销账号',
    content: '注销后账号立即退出婚恋池，并在保留期后清理数据。此操作不可撤销，确定继续吗？',
    confirmText: '确认注销',
    cancelText: '取消',
    success: async (response) => {
      if (!response.confirm) return
      acting.value = true
      try {
        await deleteAccount()
        try {
          uni.clearStorageSync()
        } catch (error) {
          // 忽略
        }
        auth.clearLocal()
        uni.reLaunch({ url: '/pages/index/index' })
      } catch (error) {
        message.value = errorMessage(error, '注销失败，请稍后重试')
      } finally {
        acting.value = false
      }
    }
  })
}

async function logout() {
  await auth.logout()
  uni.reLaunch({ url: '/pages/index/index' })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/settings/account') })
}
</script>

<style lang="scss" scoped>
.action-btn {
  margin-top: 20rpx;
}
</style>
