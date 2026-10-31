<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="婚姻与家庭"
        subtitle="这些信息只用于条件匹配，不会展示给未匹配的用户。"
        :percent="completion"
        :show-back="true"
      />
      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载资料…</text>
      </view>
      <view v-else class="wmc-card">
        <view class="wmc-field">
          <text class="wmc-field-label">婚姻状态</text>
          <picker mode="selector" :range="maritalLabels" :value="maritalIndex" @change="onMarital">
            <view class="wmc-picker">
              <text :class="draft.marital_status ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ maritalText || '请选择婚姻状态' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">子女数量</text>
          <picker mode="selector" :range="childrenLabels" :value="childrenIndex" @change="onChildrenCount">
            <view class="wmc-picker">
              <text class="wmc-picker__text">{{ childrenText }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">生育意愿</text>
          <picker mode="selector" :range="wantsLabels" :value="wantsIndex" @change="onWantsChildren">
            <view class="wmc-picker">
              <text :class="draft.wants_children ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ wantsText || '请选择生育意愿' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">结婚意愿</text>
          <picker mode="selector" :range="intentLabels" :value="intentIndex" @change="onMarriageIntent">
            <view class="wmc-picker">
              <text :class="draft.marriage_intent ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ intentText || '请选择结婚意愿' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">结婚时间规划</text>
          <picker mode="selector" :range="timelineLabels" :value="timelineIndex" @change="onTimeline">
            <view class="wmc-picker">
              <text :class="draft.marriage_timeline ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ timelineText || '请选择时间规划' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <text v-if="errorText" class="wmc-error">{{ errorText }}</text>
        <button class="wmc-btn wmc-btn--primary wmc-btn--block save-btn" :disabled="saving" @tap="saveNext">
          {{ saving ? '保存中…' : '保存并继续' }}
        </button>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block" :disabled="saving" @tap="saveExit">保存并退出</button>
      </view>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 资料完善第 3 步：婚姻状态、子女、生育意愿、结婚意愿与时间规划。
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import StepHeader from '@/components/StepHeader.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { ensureOptions, optionLabel, pickerOptions } from '@/utils/options'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const profile = useProfileStore()
const ready = ref(false)
const saving = ref(false)
const errorText = ref('')
const draft = reactive({
  marital_status: 0,
  children_count: 0,
  wants_children: 0,
  marriage_intent: 0,
  marriage_timeline: 0
})

const completion = computed(() => profile.completion)

const maritalOptions = computed(() => pickerOptions('marital_statuses', true, [4]))
const wantsOptions = computed(() => pickerOptions('wants_children'))
const intentOptions = computed(() => pickerOptions('marriage_intents'))
const timelineOptions = computed(() => pickerOptions('marriage_timelines'))
const childCountOptions = [0, 1, 2, 3, 4, 5].map((value) => ({ value, label: value === 5 ? '5 个及以上' : value + ' 个' }))

const maritalLabels = computed(() => maritalOptions.value.map((item) => item.label))
const wantsLabels = computed(() => wantsOptions.value.map((item) => item.label))
const intentLabels = computed(() => intentOptions.value.map((item) => item.label))
const timelineLabels = computed(() => timelineOptions.value.map((item) => item.label))
const childrenLabels = childCountOptions.map((item) => item.label)

const maritalText = computed(() => optionLabel('marital_statuses', draft.marital_status, ''))
const wantsText = computed(() => optionLabel('wants_children', draft.wants_children, ''))
const intentText = computed(() => optionLabel('marriage_intents', draft.marriage_intent, ''))
const timelineText = computed(() => optionLabel('marriage_timelines', draft.marriage_timeline, ''))
const childrenText = computed(() => {
  const safe = draft.children_count > 5 ? 5 : draft.children_count
  return safe >= 5 ? '5 个及以上' : safe + ' 个'
})

function indexOfValue(options: Array<{ value: number }>, value: number): number {
  const index = options.findIndex((item) => item.value === value)
  return index < 0 ? 0 : index
}

const maritalIndex = computed(() => indexOfValue(maritalOptions.value, draft.marital_status))
const wantsIndex = computed(() => indexOfValue(wantsOptions.value, draft.wants_children))
const intentIndex = computed(() => indexOfValue(intentOptions.value, draft.marriage_intent))
const timelineIndex = computed(() => indexOfValue(timelineOptions.value, draft.marriage_timeline))
const childrenIndex = computed(() => {
  const safe = draft.children_count > 5 ? 5 : draft.children_count
  return safe < 0 ? 0 : safe
})

onShow(async () => {
  if (!auth.requireLogin('/pages/onboarding/marriage')) return
  if (ready.value) return
  await ensureOptions()
  try {
    const data = await profile.loadProfile(true)
    draft.marital_status = data.marital_status || 0
    draft.children_count = data.children_count || 0
    draft.wants_children = data.wants_children || 0
    draft.marriage_intent = data.marriage_intent || 0
    draft.marriage_timeline = data.marriage_timeline || 0
  } catch (error) {
    errorText.value = errorMessage(error, '资料加载失败，请稍后重试')
  }
  ready.value = true
})

function onMarital(event: { detail: { value: number | string } }) {
  const option = maritalOptions.value[Number(event.detail.value)]
  if (option) draft.marital_status = option.value
}
function onChildrenCount(event: { detail: { value: number | string } }) {
  draft.children_count = Number(event.detail.value)
}
function onWantsChildren(event: { detail: { value: number | string } }) {
  const option = wantsOptions.value[Number(event.detail.value)]
  if (option) draft.wants_children = option.value
}
function onMarriageIntent(event: { detail: { value: number | string } }) {
  const option = intentOptions.value[Number(event.detail.value)]
  if (option) draft.marriage_intent = option.value
}
function onTimeline(event: { detail: { value: number | string } }) {
  const option = timelineOptions.value[Number(event.detail.value)]
  if (option) draft.marriage_timeline = option.value
}

async function persist(): Promise<boolean> {
  saving.value = true
  errorText.value = ''
  try {
    await profile.updateProfile({
      marital_status: draft.marital_status,
      children_count: draft.children_count,
      wants_children: draft.wants_children,
      marriage_intent: draft.marriage_intent,
      marriage_timeline: draft.marriage_timeline
    })
    return true
  } catch (error) {
    errorText.value = errorMessage(error, '保存失败，请稍后重试')
    return false
  } finally {
    saving.value = false
  }
}

async function saveNext() {
  if (saving.value) return
  if (await persist()) uni.navigateTo({ url: '/pages/onboarding/lifestyle' })
}

async function saveExit() {
  if (saving.value) return
  if (await persist()) uni.reLaunch({ url: '/pages/profile/index' })
}
</script>

<style lang="scss" scoped>
.save-btn {
  margin-top: 16rpx;
}
</style>
