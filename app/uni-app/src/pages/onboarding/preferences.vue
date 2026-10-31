<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="择偶条件"
        subtitle="硬条件表示不满足就不会进入候选池；偏好条件用于排序与解释，不会直接淘汰对方。"
        :percent="completion"
        :show-back="true"
      />
      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载择偶条件…</text>
      </view>
      <template v-else>
        <view class="wmc-card">
          <view class="wmc-section-title"><text>基本期望</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">目标性别（可多选）</text>
            <OptionChips :model-value="draft.target_genders" :options="genderOptions" @update:model-value="onTargetGenders" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">年龄区间</text>
            <view class="range-row">
              <input class="wmc-input range-input" type="number" placeholder="最小年龄" :value="draft.min_age" @input="onInput('min_age', $event)" />
              <text class="range-sep">至</text>
              <input class="wmc-input range-input" type="number" placeholder="最大年龄" :value="draft.max_age" @input="onInput('max_age', $event)" />
            </view>
            <HardToggle v-model="draft.age_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">身高区间（cm）</text>
            <view class="range-row">
              <input class="wmc-input range-input" type="number" placeholder="最低身高" :value="draft.min_height_cm" @input="onInput('min_height_cm', $event)" />
              <text class="range-sep">至</text>
              <input class="wmc-input range-input" type="number" placeholder="最高身高" :value="draft.max_height_cm" @input="onInput('max_height_cm', $event)" />
            </view>
            <HardToggle v-model="draft.height_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">可接受城市（可多选）</text>
            <picker mode="selector" :range="cityLabels" :value="addCityIndex" @change="onAddCity">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ draft.allowed_city_codes.length > 0 ? '继续添加城市' : '请选择城市' }}</text>
                <text class="wmc-picker__arrow">添加</text>
              </view>
            </picker>
            <view v-if="draft.allowed_city_codes.length > 0" class="selected-row">
              <view v-for="code in draft.allowed_city_codes" :key="code" class="wmc-chip wmc-chip--active" @tap="removeCity(code)">
                {{ cityDisplay(code) }} · 移除
              </view>
            </view>
            <HardToggle v-model="draft.city_hard" />
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>教育与收入</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">最低学历</text>
            <picker mode="selector" :range="educationLabels" :value="educationIndex" @change="onEducation">
              <view class="wmc-picker">
                <text :class="draft.min_education_level ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ educationText || '不限制学历' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
            <HardToggle v-model="draft.education_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">最低年收入区间</text>
            <picker mode="selector" :range="incomeLabels" :value="incomeIndex" @change="onIncome">
              <view class="wmc-picker">
                <text :class="draft.min_income_band ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ incomeText || '不限制收入' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
            <HardToggle v-model="draft.income_hard" />
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>家庭与生活条件</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">可接受婚姻状态</text>
            <OptionChips :model-value="draft.allowed_marital_statuses" :options="maritalOptions" @update:model-value="onMaritalStatuses" />
            <HardToggle v-model="draft.marital_status_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">可接受子女情况</text>
            <OptionChips :model-value="draft.allowed_children_states" :options="childrenOptions" @update:model-value="onChildrenStates" />
            <HardToggle v-model="draft.children_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">可接受生育意愿</text>
            <OptionChips :model-value="draft.allowed_wants_children" :options="wantsChildrenOptions" @update:model-value="onWantsChildren" />
            <HardToggle v-model="draft.wants_children_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">可接受吸烟习惯</text>
            <OptionChips :model-value="draft.allowed_smoking_statuses" :options="smokingOptions" @update:model-value="onSmokingStatuses" />
            <HardToggle v-model="draft.smoking_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">可接受饮酒习惯</text>
            <OptionChips :model-value="draft.allowed_drinking_statuses" :options="drinkingOptions" @update:model-value="onDrinkingStatuses" />
            <HardToggle v-model="draft.drinking_hard" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">最低结婚意愿</text>
            <picker mode="selector" :range="intentLabels" :value="intentIndex" @change="onMarriageIntent">
              <view class="wmc-picker">
                <text :class="draft.min_marriage_intent ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ intentText || '不限制' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
            <HardToggle v-model="draft.marriage_intent_hard" />
          </view>
        </view>

        <text v-if="errorText" class="wmc-error">{{ errorText }}</text>
        <button class="wmc-btn wmc-btn--primary wmc-btn--block save-btn" :disabled="saving" @tap="saveNext">
          {{ saving ? '保存中…' : '保存并继续' }}
        </button>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block" :disabled="saving" @tap="saveExit">保存并退出</button>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 资料完善第 5 步：择偶条件与硬条件标记，保存到 /me/preferences。
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import StepHeader from '@/components/StepHeader.vue'
import OptionChips from '@/components/OptionChips.vue'
import HardToggle from '@/components/HardToggle.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { citiesOf, cityDisplay, cityRegions, ensureOptions, ensureRegions, optionLabel, pickerOptions, optionsOf } from '@/utils/options'
import { errorMessage } from '@/utils/error'
import { toIntOrNull, validateAgeRange, validateHeightRange } from '@/utils/validate'
import type { PreferencesInput } from '@/api/types'

const auth = useAuthStore()
const profile = useProfileStore()
const ready = ref(false)
const saving = ref(false)
const errorText = ref('')
const addCityIndex = ref(0)

const draft = reactive({
  target_genders: [] as number[],
  min_age: '',
  max_age: '',
  age_hard: false,
  min_height_cm: '',
  max_height_cm: '',
  height_hard: false,
  allowed_city_codes: [] as string[],
  city_hard: false,
  min_education_level: 0,
  education_hard: false,
  min_income_band: 0,
  income_hard: false,
  allowed_marital_statuses: [] as number[],
  marital_status_hard: false,
  allowed_children_states: [] as number[],
  children_hard: false,
  allowed_wants_children: [] as number[],
  wants_children_hard: false,
  allowed_smoking_statuses: [] as number[],
  smoking_hard: false,
  allowed_drinking_statuses: [] as number[],
  drinking_hard: false,
  min_marriage_intent: 0,
  marriage_intent_hard: false
})

const completion = computed(() => profile.completion)
const genderOptions = computed(() => pickerOptions('genders', false))
const maritalOptions = computed(() => pickerOptions('marital_statuses', true, [4]))
const childrenOptions = computed(() => optionsOf('children_states'))
const wantsChildrenOptions = computed(() => pickerOptions('wants_children'))
const educationOptions = computed(() => pickerOptions('education_levels'))
const incomeOptions = computed(() => pickerOptions('income_bands', true, [8]))
const smokingOptions = computed(() => pickerOptions('smoking_statuses'))
const drinkingOptions = computed(() => pickerOptions('drinking_statuses'))
const intentOptions = computed(() => pickerOptions('marriage_intents'))
const cityOptions = computed(() => cityRegions())
const cityLabels = computed(() => cityOptions.value.map((item) => cityDisplay(item.code)))
const educationLabels = computed(() => educationOptions.value.map((item) => item.label))
const incomeLabels = computed(() => incomeOptions.value.map((item) => item.label))
const intentLabels = computed(() => intentOptions.value.map((item) => item.label))
const educationText = computed(() => optionLabel('education_levels', draft.min_education_level, ''))
const incomeText = computed(() => optionLabel('income_bands', draft.min_income_band, ''))
const intentText = computed(() => optionLabel('marriage_intents', draft.min_marriage_intent, ''))

function indexOfValue(options: Array<{ value: number }>, value: number): number {
  const index = options.findIndex((item) => item.value === value)
  return index < 0 ? 0 : index
}

const educationIndex = computed(() => indexOfValue(educationOptions.value, draft.min_education_level))
const incomeIndex = computed(() => indexOfValue(incomeOptions.value, draft.min_income_band))
const intentIndex = computed(() => indexOfValue(intentOptions.value, draft.min_marriage_intent))

onShow(async () => {
  if (!auth.requireLogin('/pages/onboarding/preferences')) return
  if (ready.value) return
  await Promise.all([ensureOptions(), ensureRegions()])
  try {
    const data = await profile.loadPreferences(true)
    draft.target_genders = data.target_genders || []
    draft.min_age = data.min_age ? String(data.min_age) : ''
    draft.max_age = data.max_age ? String(data.max_age) : ''
    draft.age_hard = Boolean(data.age_hard)
    draft.min_height_cm = data.min_height_cm ? String(data.min_height_cm) : ''
    draft.max_height_cm = data.max_height_cm ? String(data.max_height_cm) : ''
    draft.height_hard = Boolean(data.height_hard)
    draft.allowed_city_codes = data.allowed_city_codes || []
    draft.city_hard = Boolean(data.city_hard)
    draft.min_education_level = data.min_education_level || 0
    draft.education_hard = Boolean(data.education_hard)
    draft.min_income_band = data.min_income_band || 0
    draft.income_hard = Boolean(data.income_hard)
    draft.allowed_marital_statuses = data.allowed_marital_statuses || []
    draft.marital_status_hard = Boolean(data.marital_status_hard)
    draft.allowed_children_states = data.allowed_children_states || []
    draft.children_hard = Boolean(data.children_hard)
    draft.allowed_wants_children = data.allowed_wants_children || []
    draft.wants_children_hard = Boolean(data.wants_children_hard)
    draft.allowed_smoking_statuses = data.allowed_smoking_statuses || []
    draft.smoking_hard = Boolean(data.smoking_hard)
    draft.allowed_drinking_statuses = data.allowed_drinking_statuses || []
    draft.drinking_hard = Boolean(data.drinking_hard)
    draft.min_marriage_intent = data.min_marriage_intent || 0
    draft.marriage_intent_hard = Boolean(data.marriage_intent_hard)
  } catch (error) {
    errorText.value = errorMessage(error, '择偶条件加载失败，请稍后重试')
  }
  ready.value = true
})

function onInput(field: 'min_age' | 'max_age' | 'min_height_cm' | 'max_height_cm', event: { detail: { value: string } }) {
  draft[field] = event.detail.value
}

function onTargetGenders(value: Array<number | string>) {
  draft.target_genders = value.map((item) => Number(item))
}
function onMaritalStatuses(value: Array<number | string>) {
  draft.allowed_marital_statuses = value.map((item) => Number(item))
}
function onChildrenStates(value: Array<number | string>) {
  draft.allowed_children_states = value.map((item) => Number(item))
}
function onWantsChildren(value: Array<number | string>) {
  draft.allowed_wants_children = value.map((item) => Number(item))
}
function onSmokingStatuses(value: Array<number | string>) {
  draft.allowed_smoking_statuses = value.map((item) => Number(item))
}
function onDrinkingStatuses(value: Array<number | string>) {
  draft.allowed_drinking_statuses = value.map((item) => Number(item))
}

function onAddCity(event: { detail: { value: number | string } }) {
  const option = cityOptions.value[Number(event.detail.value)]
  if (!option) return
  if (draft.allowed_city_codes.indexOf(option.code) >= 0) return
  draft.allowed_city_codes = draft.allowed_city_codes.concat([option.code])
}
function removeCity(code: string) {
  draft.allowed_city_codes = draft.allowed_city_codes.filter((item) => item !== code)
}
function onEducation(event: { detail: { value: number | string } }) {
  const option = educationOptions.value[Number(event.detail.value)]
  if (option) draft.min_education_level = option.value
}
function onIncome(event: { detail: { value: number | string } }) {
  const option = incomeOptions.value[Number(event.detail.value)]
  if (option) draft.min_income_band = option.value
}
function onMarriageIntent(event: { detail: { value: number | string } }) {
  const option = intentOptions.value[Number(event.detail.value)]
  if (option) draft.min_marriage_intent = option.value
}

function buildInput(): PreferencesInput {
  return {
    target_genders: draft.target_genders,
    min_age: toIntOrNull(draft.min_age) || 0,
    max_age: toIntOrNull(draft.max_age) || 0,
    age_hard: draft.age_hard,
    min_height_cm: toIntOrNull(draft.min_height_cm) || 0,
    max_height_cm: toIntOrNull(draft.max_height_cm) || 0,
    height_hard: draft.height_hard,
    allowed_city_codes: draft.allowed_city_codes,
    city_hard: draft.city_hard,
    min_education_level: draft.min_education_level,
    education_hard: draft.education_hard,
    min_income_band: draft.min_income_band,
    income_hard: draft.income_hard,
    allowed_marital_statuses: draft.allowed_marital_statuses,
    marital_status_hard: draft.marital_status_hard,
    allowed_children_states: draft.allowed_children_states,
    children_hard: draft.children_hard,
    allowed_wants_children: draft.allowed_wants_children,
    wants_children_hard: draft.wants_children_hard,
    allowed_smoking_statuses: draft.allowed_smoking_statuses,
    smoking_hard: draft.smoking_hard,
    allowed_drinking_statuses: draft.allowed_drinking_statuses,
    drinking_hard: draft.drinking_hard,
    min_marriage_intent: draft.min_marriage_intent,
    marriage_intent_hard: draft.marriage_intent_hard
  }
}

async function persist(): Promise<boolean> {
  const ageError = validateAgeRange(toIntOrNull(draft.min_age), toIntOrNull(draft.max_age))
  if (ageError) {
    errorText.value = ageError
    return false
  }
  const heightError = validateHeightRange(toIntOrNull(draft.min_height_cm), toIntOrNull(draft.max_height_cm))
  if (heightError) {
    errorText.value = heightError
    return false
  }
  saving.value = true
  errorText.value = ''
  try {
    await profile.savePreferences(buildInput())
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
  if (await persist()) uni.navigateTo({ url: '/pages/onboarding/compatibility' })
}

async function saveExit() {
  if (saving.value) return
  if (await persist()) uni.reLaunch({ url: '/pages/profile/index' })
}
</script>

<style lang="scss" scoped>
.range-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.range-input {
  flex: 1;
}
.range-sep {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.selected-row {
  display: flex;
  flex-wrap: wrap;
  margin-top: 16rpx;
}
.save-btn {
  margin-top: 16rpx;
}
</style>
