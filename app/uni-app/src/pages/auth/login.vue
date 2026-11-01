<template>
  <PageShell>
    <view class="wmc-container login">
      <view class="login__brand">
        <view class="login__logo">W</view>
        <text class="login__title">手机号登录 / 注册</text>
        <text class="login__subtitle">首次登录自动创建账号，之后可完善资料并加入真实婚恋市场</text>
      </view>

      <view class="wmc-card login__card">
        <view class="wmc-field">
          <text class="wmc-field-label">手机号</text>
          <view class="login__input-wrap">
            <text class="login__prefix">+86</text>
            <input
              class="login__input"
              type="number"
              maxlength="11"
              placeholder="请输入中国大陆手机号"
              :value="phone"
              @input="onPhoneInput"
            />
          </view>
        </view>

        <view class="wmc-field">
          <text class="wmc-field-label">验证码</text>
          <view class="login__code-row">
            <view class="login__input-wrap login__input-wrap--code">
              <input
                class="login__input"
                type="number"
                maxlength="6"
                placeholder="6 位数字验证码"
                :value="code"
                @input="onCodeInput"
              />
            </view>
            <button class="login__code-btn" :class="{ 'login__code-btn--disabled': counting }" @tap="sendCode">
              {{ counting ? countdown + 's' : '获取验证码' }}
            </button>
          </view>
        </view>

        <view v-if="isDev" class="wmc-notice login__dev">
          <text>开发环境验证码固定为 123456，生产环境会通过短信下发。</text>
        </view>

        <view class="consent" @tap="toggleConsent">
          <view class="consent__box" :class="{ 'consent__box--checked': agreed }">
            <view v-if="agreed" class="consent__tick"></view>
          </view>
          <text class="consent__text">
            我已阅读并同意
            <text class="consent__link" @tap.stop="goPrivacy">《用户协议与隐私政策》</text>
            ，并理解匹配结果仅根据双方填写条件计算，不代表婚姻成功概率。
          </text>
        </view>

        <text v-if="errorText" class="wmc-error login__error">{{ errorText }}</text>

        <button class="wmc-btn wmc-btn--primary wmc-btn--block login__submit" :disabled="submitting" @tap="submit">
          {{ submitting ? '登录中…' : '登录 / 注册' }}
        </button>
      </view>

      <view class="login__foot">
        <text>登录代表你已年满 18 岁。手机号仅用于登录，不会在个人主页公开展示。</text>
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
.login {
  padding-top: 40rpx;
}
.login__brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24rpx 0 40rpx;
}
.login__logo {
  width: 108rpx;
  height: 108rpx;
  border-radius: 32rpx;
  background: linear-gradient(145deg, #101a3f 0%, #3355ff 100%);
  color: #ffffff;
  font-size: 48rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 18rpx 36rpx rgba(35, 50, 140, 0.32);
}
.login__title {
  margin-top: 26rpx;
  font-size: 40rpx;
  font-weight: 700;
  color: #0b1220;
  letter-spacing: -0.5rpx;
}
.login__subtitle {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #64748b;
  text-align: center;
  line-height: 1.7;
  padding: 0 48rpx;
}
.login__card {
  padding: 40rpx 32rpx 36rpx;
}
.login__input-wrap {
  display: flex;
  align-items: center;
  height: 96rpx;
  border: 1rpx solid rgba(15, 23, 42, 0.14);
  border-radius: 20rpx;
  background: #ffffff;
  padding: 0 26rpx;
  box-sizing: border-box;
}
.login__input-wrap:focus-within {
  border-color: rgba(51, 85, 255, 0.55);
  box-shadow: 0 0 0 6rpx rgba(51, 85, 255, 0.1);
}
.login__prefix {
  font-size: 28rpx;
  font-weight: 600;
  color: #334155;
  padding-right: 18rpx;
  margin-right: 18rpx;
  border-right: 1rpx solid rgba(15, 23, 42, 0.1);
}
.login__input {
  flex: 1;
  height: 100%;
  font-size: 32rpx;
  color: #0b1220;
  letter-spacing: 1rpx;
}
.login__code-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.login__input-wrap--code {
  flex: 1;
}
.login__code-btn {
  width: 220rpx;
  height: 96rpx;
  border-radius: 20rpx;
  background: rgba(51, 85, 255, 0.08);
  color: #3355ff;
  font-size: 26rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  margin: 0;
  line-height: 1.2;
}
.login__code-btn::after {
  border: none;
}
.login__code-btn--disabled {
  color: #94a3b8;
  background: #f1f5f9;
}
.login__dev {
  margin-bottom: 24rpx;
}
.consent {
  display: flex;
  align-items: flex-start;
  padding: 8rpx 0 4rpx;
}
.consent__box {
  width: 34rpx;
  height: 34rpx;
  border-radius: 10rpx;
  border: 2rpx solid #cbd5e1;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16rpx;
  margin-top: 2rpx;
  flex-shrink: 0;
  transition: all 0.12s ease;
}
.consent__box--checked {
  background: #3355ff;
  border-color: #3355ff;
}
.consent__tick {
  width: 14rpx;
  height: 8rpx;
  border-left: 3rpx solid #ffffff;
  border-bottom: 3rpx solid #ffffff;
  transform: rotate(-45deg) translateY(-2rpx);
}
.consent__text {
  flex: 1;
  font-size: 22rpx;
  line-height: 1.7;
  color: #64748b;
}
.consent__link {
  color: #3355ff;
  font-weight: 500;
}
.login__error {
  margin-top: 8rpx;
}
.login__submit {
  margin-top: 28rpx;
}
.login__foot {
  padding: 28rpx 48rpx 8rpx;
  text-align: center;
  font-size: 22rpx;
  color: #94a3b8;
  line-height: 1.8;
}
</style>