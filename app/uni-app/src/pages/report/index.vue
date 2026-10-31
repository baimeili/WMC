<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-card">
        <EmptyState title="登录后提交举报" description="举报需要登录，以便平台跟进处理。" action-text="去登录" @action="goLogin" />
      </view>

      <view v-else-if="!targetUserId" class="wmc-card">
        <EmptyState
          title="缺少举报对象"
          description="请从候选人详情或 Match 详情页面进入举报，以便平台核实具体用户。"
          action-text="返回"
          @action="back"
        />
      </view>

      <template v-else>
        <view class="wmc-card">
          <view class="wmc-section-title"><text>举报用户</text></view>
          <text class="wmc-hint">
            举报内容仅用于平台处理，不会向对方透露举报人身份。恶意举报会被记录并可能影响账号使用。
          </text>
        </view>

        <view class="wmc-card">
          <view class="wmc-field">
            <text class="wmc-field-label">举报原因</text>
            <picker mode="selector" :range="reasonLabels" :value="reasonIndex" @change="onReason">
              <view class="wmc-picker">
                <text :class="selectedReason ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ selectedReasonText || '请选择举报原因' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">补充说明（可选，最多 1000 字）</text>
            <textarea
              class="wmc-textarea"
              maxlength="1000"
              placeholder="请客观描述发生的情况，便于平台核实"
              :value="description"
              @input="onDescription"
            ></textarea>
          </view>
          <text v-if="errorText" class="wmc-error">{{ errorText }}</text>
          <button class="wmc-btn wmc-btn--primary wmc-btn--block submit-btn" :disabled="submitting" @tap="submit">
            {{ submitting ? '提交中…' : '提交举报' }}
          </button>
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block" :disabled="submitting" @tap="blockAndBack">举报并屏蔽该用户</button>
        </view>

        <view class="wmc-card wmc-card--flat">
          <text class="wmc-hint">举报记录会进入平台处理队列；如涉及诈骗、人身安全等紧急情况，请同时向公安机关报案。</text>
        </view>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 举报页：从候选人详情或 Match 详情携带 target 用户进入，提交举报并可同时屏蔽。
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { reportUser, blockUser } from '@/api/safety'
import { ensureOptions, optionLabel, pickerOptions } from '@/utils/options'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const targetUserId = ref('')
const description = ref('')
const submitting = ref(false)
const errorText = ref('')
const reasonValue = ref(0)

const loggedIn = computed(() => auth.isLoggedIn)
const reasonOptions = computed(() => pickerOptions('report_reasons', false))
const reasonLabels = computed(() => reasonOptions.value.map((item) => item.label))
const reasonIndex = computed(() => {
  const index = reasonOptions.value.findIndex((item) => item.value === reasonValue.value)
  return index < 0 ? 0 : index
})
const selectedReason = computed(() => reasonValue.value > 0)
const selectedReasonText = computed(() => optionLabel('report_reasons', reasonValue.value, ''))

onLoad(async (options) => {
  targetUserId.value = options && options.target ? decodeURIComponent(options.target) : ''
  auth.syncFromStorage()
  await ensureOptions()
})

function onReason(event: { detail: { value: number | string } }) {
  const option = reasonOptions.value[Number(event.detail.value)]
  if (option) reasonValue.value = option.value
}

function onDescription(event: { detail: { value: string } }) {
  description.value = event.detail.value
}

async function submit() {
  if (submitting.value) return
  if (!targetUserId.value) return
  if (reasonValue.value <= 0) {
    errorText.value = '请选择举报原因'
    return
  }
  submitting.value = true
  errorText.value = ''
  try {
    await reportUser(targetUserId.value, {
      reason: reasonValue.value,
      description: description.value || undefined
    })
    uni.showToast({ title: '举报已提交', icon: 'none' })
    setTimeout(() => {
      uni.navigateBack({ delta: 1 })
    }, 400)
  } catch (error) {
    errorText.value = errorMessage(error, '举报提交失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

async function blockAndBack() {
  if (submitting.value) return
  if (reasonValue.value <= 0) {
    errorText.value = '请先选择举报原因'
    return
  }
  submitting.value = true
  errorText.value = ''
  try {
    await reportUser(targetUserId.value, {
      reason: reasonValue.value,
      description: description.value || undefined
    })
    await blockUser(targetUserId.value)
    uni.showToast({ title: '已举报并屏蔽', icon: 'none' })
    setTimeout(() => {
      uni.navigateBack({ delta: 1 })
    }, 400)
  } catch (error) {
    errorText.value = errorMessage(error, '操作失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

function back() {
  uni.navigateBack({ delta: 1 })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/report/index') })
}
</script>

<style lang="scss" scoped>
.submit-btn {
  margin-top: 16rpx;
}
</style>
