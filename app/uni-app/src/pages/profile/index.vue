<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-card">
        <EmptyState title="登录后管理个人资料" description="登录后可完善资料、设置择偶条件与价值观，并进入本站真实候选池。" action-text="去登录" @action="goLogin" />
      </view>

      <view v-else-if="loading" class="wmc-card">
        <text class="wmc-hint">正在加载个人资料…</text>
      </view>

      <template v-else-if="profile">
        <view class="wmc-card">
          <view class="profile-head">
            <view class="profile-avatar">
              <text class="profile-avatar-text">{{ avatarText }}</text>
            </view>
            <view class="profile-info">
              <text class="profile-name">{{ profile.nickname || '未填写昵称' }}</text>
              <text class="profile-meta">{{ genderText }} · {{ ageText(profile.age) }} · {{ cityDisplay(profile.city_code) || '城市未填写' }}</text>
              <text class="profile-meta">{{ educationText }} · {{ incomeText }}</text>
            </view>
            <text class="wmc-tag">{{ completion }}%</text>
          </view>
          <view class="progress">
            <view class="progress__track">
              <view class="progress__bar" :style="{ width: completion + '%' }"></view>
            </view>
            <text class="progress__label">资料完成度 {{ completion }}%</text>
          </view>
          <view class="profile-actions">
            <button class="wmc-btn wmc-btn--primary profile-actions__btn" @tap="goEdit">编辑资料</button>
            <button class="wmc-btn wmc-btn--soft profile-actions__btn" @tap="goPreferences">择偶条件</button>
          </view>
          <text v-if="completion < 80" class="wmc-hint completion-hint">资料完成度不足 80% 时无法加入真实婚恋池，可继续完善基础资料、职业、婚姻与生活习惯。</text>
        </view>

        <view v-if="errorText" class="wmc-card"><text class="wmc-error">{{ errorText }}</text></view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>我的信息</text></view>
          <view class="info-row"><text class="wmc-label">婚姻状态</text><text class="wmc-value">{{ maritalText }}</text></view>
          <view class="info-row"><text class="wmc-label">生育意愿</text><text class="wmc-value">{{ wantsText }}</text></view>
          <view class="info-row"><text class="wmc-label">结婚意愿</text><text class="wmc-value">{{ intentText }}</text></view>
          <view class="info-row"><text class="wmc-label">身高</text><text class="wmc-value">{{ heightText(profile.height_cm) }}</text></view>
          <view class="info-row"><text class="wmc-label">个人简介</text><text class="wmc-value info-row__long">{{ profile.bio || '未填写' }}</text></view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>我的入口</text></view>
          <view class="wmc-list-item" @tap="goDiscovery">
            <text class="wmc-value">本站真实候选池</text>
            <text class="wmc-link">进入</text>
          </view>
          <view class="wmc-list-item" @tap="goMatches">
            <text class="wmc-value">我的 Match</text>
            <text class="wmc-link">查看</text>
          </view>
          <view class="wmc-list-item" @tap="goCompatibility">
            <text class="wmc-value">婚恋价值观设置</text>
            <text class="wmc-link">设置</text>
          </view>
          <view class="wmc-list-item" @tap="goSettings">
            <text class="wmc-value">设置与隐私</text>
            <text class="wmc-link">进入</text>
          </view>
        </view>

        <view class="wmc-card wmc-card--flat">
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block" @tap="logout">退出登录</button>
        </view>
      </template>

      <view v-else class="wmc-card">
        <text class="wmc-error">{{ errorText || '个人资料加载失败' }}</text>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="load">重新加载</button>
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 我的：资料概览、完成度、编辑入口与真实市场入口。
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { cityDisplay, optionLabel } from '@/utils/options'
import { ageText, heightText } from '@/utils/format'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const profileStore = useProfileStore()
const loading = ref(false)
const errorText = ref('')

const loggedIn = computed(() => auth.isLoggedIn)
const profile = computed(() => profileStore.profile)
const completion = computed(() => profileStore.completion)
const avatarText = computed(() => (profile.value && profile.value.nickname ? profile.value.nickname.slice(0, 1) : '我'))
const genderText = computed(() => optionLabel('genders', profile.value ? profile.value.gender : 0, '性别未填写'))
const educationText = computed(() => optionLabel('education_levels', profile.value ? profile.value.education_level : 0, '学历未填写'))
const incomeText = computed(() => optionLabel('income_bands', profile.value ? profile.value.annual_income_band : 0, '收入未填写'))
const maritalText = computed(() => optionLabel('marital_statuses', profile.value ? profile.value.marital_status : 0, '未填写'))
const wantsText = computed(() => optionLabel('wants_children', profile.value ? profile.value.wants_children : 0, '未填写'))
const intentText = computed(() => optionLabel('marriage_intents', profile.value ? profile.value.marriage_intent : 0, '未填写'))

onShow(async () => {
  auth.syncFromStorage()
  if (!auth.isLoggedIn) return
  await load()
})

async function load() {
  loading.value = true
  errorText.value = ''
  try {
    await profileStore.loadProfile(true)
  } catch (error) {
    errorText.value = errorMessage(error, '个人资料加载失败')
  } finally {
    loading.value = false
  }
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/profile/index') })
}
function goEdit() {
  uni.navigateTo({ url: '/pages/profile/edit' })
}
function goPreferences() {
  uni.navigateTo({ url: '/pages/onboarding/preferences' })
}
function goCompatibility() {
  uni.navigateTo({ url: '/pages/onboarding/compatibility' })
}
function goDiscovery() {
  uni.navigateTo({ url: '/pages/discovery/index' })
}
function goMatches() {
  uni.navigateTo({ url: '/pages/matches/index' })
}
function goSettings() {
  uni.navigateTo({ url: '/pages/settings/index' })
}
async function logout() {
  await auth.logout()
  uni.reLaunch({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.profile-head {
  display: flex;
  align-items: center;
  gap: 24rpx;
}
.profile-avatar {
  width: 112rpx;
  height: 112rpx;
  border-radius: 20rpx;
  background: rgba(47, 107, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.profile-avatar-text {
  font-size: 44rpx;
  color: $uni-color-primary;
  font-weight: 600;
}
.profile-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  min-width: 0;
}
.profile-name {
  font-size: 34rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.profile-meta {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.progress {
  margin-top: 24rpx;
}
.progress__track {
  height: 12rpx;
  border-radius: 999rpx;
  background: #eef2f7;
  overflow: hidden;
}
.progress__bar {
  height: 100%;
  border-radius: 999rpx;
  background: $uni-color-primary;
}
.progress__label {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  color: $uni-text-color-grey;
}
.profile-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}
.profile-actions__btn {
  flex: 1;
  height: 76rpx;
  font-size: 28rpx;
}
.completion-hint {
  display: block;
  margin-top: 16rpx;
}
.info-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}
.info-row:last-child {
  border-bottom: none;
}
.info-row__long {
  flex: 1;
  text-align: right;
  margin-left: 24rpx;
  font-weight: 400;
  line-height: 1.6;
}
.retry-btn {
  margin-top: 20rpx;
}
</style>
