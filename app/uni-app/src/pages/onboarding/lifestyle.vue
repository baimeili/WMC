<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="生活习惯"
        subtitle="生活习惯会影响关系观念匹配度，是否接受由双方条件决定。"
        :percent="completion"
        :show-back="true"
      />
      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载资料…</text>
      </view>
      <view v-else class="wmc-card">
        <view class="wmc-field">
          <text class="wmc-field-label">吸烟习惯</text>
          <picker mode="selector" :range="smokingLabels" :value="smokingIndex" @change="onSmoking">
            <view class="wmc-picker">
              <text :class="draft.smoking_status ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ smokingText || '请选择吸烟习惯' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">饮酒习惯</text>
          <picker mode="selector" :range="drinkingLabels" :value="drinkingIndex" @change="onDrinking">
            <view class="wmc-picker">
              <text :class="draft.drinking_status ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ drinkingText || '请选择饮酒习惯' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">房产状态</text>
          <picker mode="selector" :range="houseLabels" :value="houseIndex" @change="onHouse">
            <view class="wmc-picker">
              <text :class="draft.house_status ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ houseText || '请选择房产状态' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">车辆状态</text>
          <picker mode="selector" :range="vehicleLabels" :value="vehicleIndex" @change="onVehicle">
            <view class="wmc-picker">
              <text :class="draft.vehicle_status ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ vehicleText || '请选择车辆状态' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">个人简介（最多 500 字）</text>
          <textarea
            class="wmc-textarea"
            maxlength="500"
            placeholder="简单介绍自己，避免填写手机号、微信号等联系方式"
            :value="draft.bio"
            @input="onBio"
          ></textarea>
          <text class="bio-count">{{ draft.bio.length }} / 500</text>
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
// 资料完善第 4 步：吸烟、饮酒、房产、车辆与个人简介。
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import StepHeader from '@/components/StepHeader.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { ensureOptions, optionLabel, pickerOptions } from '@/utils/options'
import { errorMessage } from '@/utils/error'
import { validateBio } from '@/utils/validate'

const auth = useAuthStore()
const profile = useProfileStore()
const ready = ref(false)
const saving = ref(false)
const errorText = ref('')
const draft = reactive({
  smoking_status: 0,
  drinking_status: 0,
  house_status: 0,
  vehicle_status: 0,
  bio: ''
})

const completion = computed(() => profile.completion)
const smokingOptions = computed(() => pickerOptions('smoking_statuses'))
const drinkingOptions = computed(() => pickerOptions('drinking_statuses'))
const houseOptions = computed(() => pickerOptions('house_statuses', true, [6]))
const vehicleOptions = computed(() => pickerOptions('vehicle_statuses', true, [4]))
const smokingLabels = computed(() => smokingOptions.value.map((item) => item.label))
const drinkingLabels = computed(() => drinkingOptions.value.map((item) => item.label))
const houseLabels = computed(() => houseOptions.value.map((item) => item.label))
const vehicleLabels = computed(() => vehicleOptions.value.map((item) => item.label))
const smokingText = computed(() => optionLabel('smoking_statuses', draft.smoking_status, ''))
const drinkingText = computed(() => optionLabel('drinking_statuses', draft.drinking_status, ''))
const houseText = computed(() => optionLabel('house_statuses', draft.house_status, ''))
const vehicleText = computed(() => optionLabel('vehicle_statuses', draft.vehicle_status, ''))

function indexOfValue(options: Array<{ value: number }>, value: number): number {
  const index = options.findIndex((item) => item.value === value)
  return index < 0 ? 0 : index
}

const smokingIndex = computed(() => indexOfValue(smokingOptions.value, draft.smoking_status))
const drinkingIndex = computed(() => indexOfValue(drinkingOptions.value, draft.drinking_status))
const houseIndex = computed(() => indexOfValue(houseOptions.value, draft.house_status))
const vehicleIndex = computed(() => indexOfValue(vehicleOptions.value, draft.vehicle_status))

onShow(async () => {
  if (!auth.requireLogin('/pages/onboarding/lifestyle')) return
  if (ready.value) return
  await ensureOptions()
  try {
    const data = await profile.loadProfile(true)
    draft.smoking_status = data.smoking_status || 0
    draft.drinking_status = data.drinking_status || 0
    draft.house_status = data.house_status || 0
    draft.vehicle_status = data.vehicle_status || 0
    draft.bio = data.bio || ''
  } catch (error) {
    errorText.value = errorMessage(error, '资料加载失败，请稍后重试')
  }
  ready.value = true
})

function onSmoking(event: { detail: { value: number | string } }) {
  const option = smokingOptions.value[Number(event.detail.value)]
  if (option) draft.smoking_status = option.value
}
function onDrinking(event: { detail: { value: number | string } }) {
  const option = drinkingOptions.value[Number(event.detail.value)]
  if (option) draft.drinking_status = option.value
}
function onHouse(event: { detail: { value: number | string } }) {
  const option = houseOptions.value[Number(event.detail.value)]
  if (option) draft.house_status = option.value
}
function onVehicle(event: { detail: { value: number | string } }) {
  const option = vehicleOptions.value[Number(event.detail.value)]
  if (option) draft.vehicle_status = option.value
}
function onBio(event: { detail: { value: string } }) {
  draft.bio = event.detail.value
}

async function persist(): Promise<boolean> {
  const bioError = validateBio(draft.bio)
  if (bioError) {
    errorText.value = bioError
    return false
  }
  saving.value = true
  errorText.value = ''
  try {
    await profile.updateProfile({
      smoking_status: draft.smoking_status,
      drinking_status: draft.drinking_status,
      house_status: draft.house_status,
      vehicle_status: draft.vehicle_status,
      bio: draft.bio
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
  if (await persist()) uni.navigateTo({ url: '/pages/onboarding/preferences' })
}

async function saveExit() {
  if (saving.value) return
  if (await persist()) uni.reLaunch({ url: '/pages/profile/index' })
}
</script>

<style lang="scss" scoped>
.bio-count {
  display: block;
  text-align: right;
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
  margin-top: 8rpx;
}
.save-btn {
  margin-top: 16rpx;
}
</style>
