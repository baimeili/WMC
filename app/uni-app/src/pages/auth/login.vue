<template>
  <PageShell>
    <view class="wmc-container">
      <view class="login-head">
        <text class="login-head__title">手机号登录 / 注册</text>
        <text class="login-head__desc">
          首次登录会自动创建账号。登录后可完善资料，查看本站真实双向候选人。
        </text>
      </view>

      <view class="wmc-card">
        <view class="wmc-field">
          <text class="wmc-field-label">手机号</text>
          <input
            class="wmc-input"
            type="number"
            maxlength="11"
            placeholder="请输入中国大陆手机号"
            :value="phone"
            @input="onPhoneInput"
          />
        </view>

        <view class="wmc-field">
          <text class="wmc-field-label">验证码</text>
          <view class="code-row">
            <input
              class="wmc-input code-input"
              type="number"
              maxlength="6"
              placeholder="6 位数字验证码"
              :value="code"
              @input="onCodeInput"
            />
            <button class="wmc-btn wmc-btn--soft code-btn" :class="{ 'wmc-btn--disabled': counting }" @tap="sendCode">
              {{ counting ? countdown + 's 后重发' : '获取验证码' }}
            </button>
          </view>
        </view>

        <view v-if="isDev" class="wmc-notice wmc-notice--warning">
          开发环境验证码固定为 123456，生产环境会通过短信下发。
        </view>

        <view class="consent" @tap="toggleConsent">
          <view class="consent__box" :class="{ 'consent__box--checked': agreed }">
            <view v-if="agreed" class="consent__dot"></view>
          </view>
          <text class="consent__text">
            我已阅读并同意
            <text class="wmc-link" @tap.stop="goPrivacy">《用户协议与隐私政策》</text>
            ，并理解匹配结果仅根据双方填写条件计算，不代表婚姻成功概率。
          </text>
        </view>

        <text v-if="errorText" class="wmc-error">{{ errorText }}</text>

        <button class="wmc-btn wmc-btn--primary wmc-btn--block login-btn" :disabled="submitting" @tap="submit">
          {{ submitting ? '登录中…' : '登录 / 注册' }}
        </button>
      </view>

      <view class="login-foot">
        <text class="wmc-hint">登录代表你已年满 18 岁。手机号仅用于登录，不会在个人主页公开展示。</text>
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 登录页：手机号 + 验证码；开发环境提示固定验证码 123456。
import { ref, computed } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from '@/utils/error'
import { validatePhone, validateSmsCodeInput } from '@/utils/validate'
import { trackEvent } from '@/api/events'

const auth = useAuthStore()
const phone = ref('')
const code = ref('')
const agreed = ref(false)
const submitting = ref(false)
const counting = ref(false)
const countdown = ref(60)
const errorText = ref('')
const redirect = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const isDev = computed(() => {
  const env = (import.meta as unknown as { env?: Record<string, string | undefined> }).env
  return !env || env.MODE !== 'production'
})

onLoad((options) => {
  redirect.value = options && options.redirect ? decodeURIComponent(options.redirect) : ''
})

onUnload(() => {
  stopTimer()
})

function stopTimer() {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
}

function onPhoneInput(event: { detail: { value: string } }) {
  phone.value = event.detail.value
}

function onCodeInput(event: { detail: { value: string } }) {
  code.value = event.detail.value
}

function toggleConsent() {
  agreed.value = !agreed.value
}

function goPrivacy() {
  uni.navigateTo({ url: '/pages/settings/privacy' })
}

async function sendCode() {
  if (counting.value) return
  const phoneError = validatePhone(phone.value)
  if (phoneError) {
    errorText.value = phoneError
    return
  }
  errorText.value = ''
  try {
    await auth.sendCode(phone.value)
    uni.showToast({ title: '验证码已发送', icon: 'none' })
    counting.value = true
    countdown.value = 60
    stopTimer()
    timer = setInterval(() => {
      countdown.value = countdown.value - 1
      if (countdown.value <= 0) {
        counting.value = false
        stopTimer()
      }
    }, 1000)
  } catch (error) {
    errorText.value = errorMessage(error, '验证码发送失败，请稍后重试')
  }
}

async function submit() {
  if (submitting.value) return
  const phoneError = validatePhone(phone.value)
  if (phoneError) {
    errorText.value = phoneError
    return
  }
  const codeError = validateSmsCodeInput(code.value)
  if (codeError) {
    errorText.value = codeError
    return
  }
  if (!agreed.value) {
    errorText.value = '请先阅读并同意用户协议与隐私政策'
    return
  }
  errorText.value = ''
  submitting.value = true
  try {
    const pair = await auth.loginBySms(phone.value, code.value)
    trackEvent('login_success', { role: pair.role })
    const target = redirect.value
    const completion = pair.profile_completion || 0
    if (target) {
      uni.redirectTo({ url: target })
    } else if (completion < 80) {
      uni.redirectTo({ url: '/pages/onboarding/basic' })
    } else {
      uni.redirectTo({ url: '/pages/discovery/index' })
    }
  } catch (error) {
    errorText.value = errorMessage(error, '登录失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-head {
  padding: 24rpx 8rpx 28rpx;
}
.login-head__title {
  display: block;
  font-size: 40rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.login-head__desc {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  line-height: 1.8;
  color: $uni-text-color-grey;
}
.code-row {
  display: flex;
  gap: 16rpx;
}
.code-input {
  flex: 1;
}
.code-btn {
  width: 240rpx;
  flex-shrink: 0;
  font-size: 26rpx;
  height: 84rpx;
}
.consent {
  display: flex;
  align-items: flex-start;
  margin: 24rpx 0;
}
.consent__box {
  width: 36rpx;
  height: 36rpx;
  border-radius: 8rpx;
  border: 1rpx solid $uni-border-color;
  margin-right: 16rpx;
  margin-top: 4rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: #ffffff;
}
.consent__box--checked {
  background: $uni-color-primary;
  border-color: $uni-color-primary;
}
.consent__dot {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  background: #ffffff;
}
.consent__text {
  flex: 1;
  font-size: 22rpx;
  line-height: 1.7;
  color: $uni-text-color-grey;
}
.login-btn {
  margin-top: 24rpx;
}
.login-foot {
  padding: 8rpx 8rpx 40rpx;
}
</style>
