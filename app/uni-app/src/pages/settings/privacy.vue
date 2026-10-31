<template>
  <PageShell>
    <view class="wmc-container">
      <view class="wmc-card">
        <view class="wmc-section-title"><text>隐私与协议</text></view>
        <text class="wmc-hint">我们只采集完成匹配所必需的信息，不采集身份证号、精确住址、GPS 坐标、银行卡或工资流水。</text>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title"><text>1. 两个市场的区别</text></view>
        <text class="privacy-text">
          理论市场估算基于后台导入的基准人口数据集，只输出条件占比、漏斗与敏感性，不输出双向匹配率，也不是本站真实用户数量。
          本站真实市场只统计已注册、资料达标并主动加入婚恋池的用户，只有双方硬条件都满足才会进入候选池。
        </text>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title"><text>2. 数据使用与展示</text></view>
        <text class="privacy-text">
          资料、择偶条件与价值观答案仅用于匹配计算与结果解释。真实候选人数少于匿名阈值时统一显示「&lt;10」，不会展示具体人数。
          匹配分数名称固定为「条件契合度」「关系观念兼容度」「综合匹配度」，这些分数仅根据双方填写条件计算，不代表婚姻成功概率。
        </text>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title"><text>3. 照片与位置信息</text></view>
        <text class="privacy-text">
          照片上传时会自动去除 EXIF 与 GPS 信息并重新编码压缩。我们不采集精确位置，只采集到城市级别的行政区划。
        </text>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title"><text>4. 你的权利</text></view>
        <text class="privacy-text">
          你可以随时修改资料、退出婚恋池、屏蔽或举报其他用户，也可以注销账号。注销后账号立即退出婚恋池，并在保留期结束后清理数据。
          被屏蔽的用户双方不可见、不可互动、不可查看资料。
        </text>
      </view>

      <view class="wmc-card">
        <view class="wmc-section-title">
          <text>协议同意</text>
          <text class="wmc-tag wmc-tag--muted">当前版本 {{ versionsText }}</text>
        </view>
        <text v-if="!loggedIn" class="wmc-hint">登录后可记录协议同意状态。</text>
        <template v-else>
          <text class="wmc-hint">同意记录会携带当前政策版本，便于后续版本升级时重新确认。以下记录仅用于合规留痕。</text>
          <view class="consent-list">
            <button class="wmc-btn wmc-btn--soft wmc-btn--block" :disabled="submitting" @tap="record(1, 'terms_of_service')">同意用户协议（{{ versionOf('terms_of_service') }}）</button>
            <button class="wmc-btn wmc-btn--soft wmc-btn--block" :disabled="submitting" @tap="record(2, 'privacy_policy')">同意隐私政策（{{ versionOf('privacy_policy') }}）</button>
            <button class="wmc-btn wmc-btn--soft wmc-btn--block" :disabled="submitting" @tap="record(3, 'market_participation')">同意婚恋市场参与声明（{{ versionOf('market_participation') }}）</button>
          </view>
          <text v-if="message" class="wmc-error">{{ message }}</text>
        </template>
      </view>

      <view class="wmc-card wmc-card--flat">
        <text class="wmc-hint">
          本页面为产品 MVP 的隐私说明，正式上线前会由法务补充完整条款与运营主体信息。当前版本号 {{ appVersion }}。
        </text>
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 隐私与协议：说明数据使用规则，并可记录协议同意（带政策版本）。
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import { useAuthStore } from '@/stores/auth'
import { recordConsent } from '@/api/profile'
import { ensureAppConfig, metaState } from '@/utils/options'
import { errorMessage } from '@/utils/error'
import { APP_VERSION } from '@/utils/constants'

const auth = useAuthStore()
const submitting = ref(false)
const message = ref('')
const appVersion = APP_VERSION

const loggedIn = computed(() => auth.isLoggedIn)
const policyVersions = computed(() => {
  const config = metaState.appConfig
  const versions = config && config.policy_versions ? config.policy_versions : {}
  return {
    terms_of_service: versions.terms_of_service || '2026.01',
    privacy_policy: versions.privacy_policy || '2026.01',
    market_participation: versions.market_participation || '2026.01'
  }
})
const versionsText = computed(() => {
  const value = policyVersions.value
  return value.terms_of_service + ' / ' + value.privacy_policy + ' / ' + value.market_participation
})

onShow(async () => {
  auth.syncFromStorage()
  await ensureAppConfig()
})

function versionOf(key: 'terms_of_service' | 'privacy_policy' | 'market_participation'): string {
  return policyVersions.value[key]
}

async function record(consentType: number, key: 'terms_of_service' | 'privacy_policy' | 'market_participation') {
  if (submitting.value) return
  if (!auth.isLoggedIn) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }
  submitting.value = true
  message.value = ''
  try {
    await recordConsent(consentType, versionOf(key))
    uni.showToast({ title: '已记录同意状态', icon: 'none' })
  } catch (error) {
    message.value = errorMessage(error, '记录失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.privacy-text {
  display: block;
  font-size: 26rpx;
  line-height: 1.9;
  color: $uni-text-color;
}
.consent-list {
  margin-top: 20rpx;
}
.consent-list .wmc-btn {
  margin-top: 16rpx;
  font-size: 26rpx;
}
</style>
