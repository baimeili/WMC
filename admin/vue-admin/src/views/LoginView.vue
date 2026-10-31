<!-- 运营后台登录：手机号 + 短信验证码；开发环境额外提供管理端开发登录。 -->
<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sendSmsCode } from '@/api/auth'
import { ApiError } from '@/api/http'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const activeTab = ref('sms')
const phone = ref('')
const code = ref('')
const devPhone = ref('')
const sending = ref(false)
const submitting = ref(false)
const countdown = ref(0)

let countdownTimer: ReturnType<typeof setInterval> | null = null

const phonePattern = /^1[3-9]\d{9}$/
const canSendCode = computed(() => phonePattern.test(phone.value.trim()) && countdown.value === 0 && !sending.value)
const canSmsLogin = computed(
  () => phonePattern.test(phone.value.trim()) && /^\d{4,6}$/.test(code.value.trim()) && !submitting.value
)
const canDevLogin = computed(() => phonePattern.test(devPhone.value.trim()) && !submitting.value)

// 发送验证码后进入 60 秒重发倒计时。
function startCountdown(): void {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      stopCountdown()
    }
  }, 1000)
}

function stopCountdown(): void {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

onBeforeUnmount(stopCountdown)

async function handleSendCode(): Promise<void> {
  if (!canSendCode.value) {
    ElMessage.warning('请输入正确的中国大陆手机号')
    return
  }
  sending.value = true
  try {
    await sendSmsCode(phone.value.trim())
    ElMessage.success('验证码已发送。开发环境为固定验证码，请查看后端开发说明。')
    startCountdown()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    sending.value = false
  }
}

// 登录成功后的统一处理：校验后台角色并跳转目标页。
async function finishLogin(): Promise<void> {
  if (!auth.canAccessAdmin) {
    auth.clear()
    ElMessage.error('当前账号没有运营后台权限，需要内容审核员及以上角色')
    return
  }
  ElMessage.success('登录成功')
  const redirect =
    typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
      ? route.query.redirect
      : '/dashboard'
  await router.replace(redirect)
}

async function handleSmsLogin(): Promise<void> {
  if (!canSmsLogin.value) {
    ElMessage.warning('请填写正确的手机号与验证码')
    return
  }
  submitting.value = true
  try {
    await auth.loginWithSms(phone.value.trim(), code.value.trim())
    await finishLogin()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    submitting.value = false
  }
}

async function handleDevLogin(): Promise<void> {
  if (!canDevLogin.value) {
    ElMessage.warning('请输入管理员手机号')
    return
  }
  submitting.value = true
  try {
    await auth.loginWithDev(devPhone.value.trim())
    await finishLogin()
  } catch (error) {
    if (error instanceof ApiError) {
      if (error.status === 404 || error.code === 'NOT_FOUND') {
        ElMessage.warning('生产环境已关闭开发登录，请使用手机号 + 短信验证码登录')
      } else if (error.status === 403 || error.code === 'FORBIDDEN') {
        ElMessage.error('该手机号不是管理员账号，无法登录运营后台')
      } else {
        ElMessage.error(error.message || '开发登录失败，请稍后重试')
      }
    } else {
      ElMessage.error('开发登录失败，请稍后重试')
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <section class="login-intro">
      <p class="intro-eyebrow">WMC ADMIN</p>
      <h1>婚恋市场计算器<br />运营后台</h1>
      <p class="intro-desc">
        统一管理用户、举报、照片、基准数据集、价值观问题与系统参数，查看真实运营漏斗与 North Star 指标。
      </p>
      <ul class="intro-list">
        <li>理论市场：基准数据集统计估算，只能表述为「约占多少」</li>
        <li>真实市场：本站注册用户的双向硬条件匹配结果</li>
        <li>管理端默认不可见完整手机号，敏感操作写入审计日志</li>
      </ul>
      <p class="intro-warning">理论市场为统计估算，禁止对外宣称精确人数或双向匹配率。</p>
    </section>

    <section class="login-panel">
      <el-card shadow="never" class="login-card">
        <h2 class="login-title">登录</h2>
        <p class="muted">仅内容审核员、管理员、超级管理员可进入运营后台</p>
        <el-tabs v-model="activeTab" class="login-tabs">
          <el-tab-pane label="手机号登录" name="sms">
            <el-form label-position="top" @submit.prevent>
              <el-form-item label="手机号">
                <el-input v-model="phone" maxlength="11" placeholder="请输入中国大陆手机号" />
              </el-form-item>
              <el-form-item label="短信验证码">
                <div class="code-row">
                  <el-input v-model="code" maxlength="6" placeholder="请输入验证码" />
                  <el-button :disabled="!canSendCode" :loading="sending" @click="handleSendCode">
                    {{ countdown > 0 ? countdown + ' 秒后重发' : '发送验证码' }}
                  </el-button>
                </div>
              </el-form-item>
              <el-button
                type="primary"
                class="submit-btn"
                :loading="submitting"
                :disabled="!canSmsLogin"
                @click="handleSmsLogin"
              >
                登录运营后台
              </el-button>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="开发登录" name="dev">
            <el-alert
              type="info"
              :closable="false"
              title="仅开发环境可用"
              description="生产环境该接口返回 404。请使用已完成管理员初始化的手机号登录。"
            />
            <el-form label-position="top" class="mt-16" @submit.prevent>
              <el-form-item label="管理员手机号">
                <el-input v-model="devPhone" maxlength="11" placeholder="已是管理员的手机号" />
              </el-form-item>
              <el-button
                type="primary"
                class="submit-btn"
                :loading="submitting"
                :disabled="!canDevLogin"
                @click="handleDevLogin"
              >
                开发登录
              </el-button>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </el-card>
      <p class="panel-footnote">登录即代表知晓运营数据保密要求，所有敏感操作将写入审计日志。</p>
    </section>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  min-height: 100vh;
  background: #f5f7fa;
}

.login-intro {
  flex: 1.1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 56px 64px;
  color: #fff;
  background: linear-gradient(150deg, #1f2d3d 0%, #253b52 55%, #2f4f6f 100%);
}

.intro-eyebrow {
  margin: 0 0 18px;
  color: #7fb3e0;
  font-size: 12px;
  letter-spacing: 4px;
}

.login-intro h1 {
  margin: 0 0 20px;
  font-size: 34px;
  line-height: 1.35;
  font-weight: 600;
}

.intro-desc {
  max-width: 460px;
  margin: 0 0 24px;
  color: #c7d3e0;
  font-size: 14px;
  line-height: 1.8;
}

.intro-list {
  max-width: 460px;
  margin: 0 0 28px;
  padding-left: 18px;
  color: #d7e1ec;
  font-size: 13px;
  line-height: 2;
}

.intro-warning {
  max-width: 460px;
  margin: 0;
  padding: 12px 16px;
  border-left: 3px solid #e6a23c;
  border-radius: 4px;
  background: rgba(230, 162, 60, 0.12);
  color: #f3d19e;
  font-size: 13px;
  line-height: 1.7;
}

.login-panel {
  flex: 0.9;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 48px 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  border-radius: 10px;
}

.login-title {
  margin: 0 0 6px;
  font-size: 22px;
}

.login-tabs {
  margin-top: 16px;
}

.code-row {
  display: flex;
  gap: 10px;
  width: 100%;
}

.code-row .el-input {
  flex: 1;
}

.submit-btn {
  width: 100%;
  margin-top: 4px;
}

.panel-footnote {
  margin: 18px 0 0;
  color: var(--wmc-text-secondary);
  font-size: 12px;
}

@media (max-width: 960px) {
  .login-page {
    flex-direction: column;
  }

  .login-intro {
    padding: 36px 28px;
  }

  .login-intro h1 {
    font-size: 26px;
  }
}
</style>
