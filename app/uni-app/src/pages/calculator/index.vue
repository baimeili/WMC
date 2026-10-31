<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="算算我的候选池"
        subtitle="无需注册即可计算。理论市场结果基于基准人口数据集估算，不是本站真实用户数量。"
        :step="step"
        :steps="stepLabels"
      />

      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载选项数据…</text>
      </view>

      <template v-else>
        <view v-if="optionsError" class="wmc-notice wmc-notice--warning">{{ optionsError }}</view>

        <!-- 第 1 步：我是谁 -->
        <view v-if="step === 1" class="wmc-card">
          <view class="wmc-section-title">
            <text>我是谁</text>
            <text class="wmc-tag wmc-tag--muted">仅用于计算</text>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">性别</text>
            <OptionChips
              :model-value="selfGenderValue"
              :options="genderOptions"
              :single="true"
              @update:model-value="onSelfGender"
            />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">年龄</text>
            <input class="wmc-input" type="number" placeholder="例如 29" :value="store.draft.self.age" @input="onSelfInput('age', $event)" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">所在城市</text>
            <picker mode="selector" :range="cityLabels" :value="selfCityIndex" @change="onSelfCityChange">
              <view class="wmc-picker">
                <text :class="selfCityText ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ selfCityText || '请选择城市' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">身高（cm，可选）</text>
            <input class="wmc-input" type="number" placeholder="例如 170" :value="store.draft.self.height_cm" @input="onSelfInput('height_cm', $event)" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">学历（可选）</text>
            <picker mode="selector" :range="educationLabels" :value="selfEducationIndex" @change="onSelfEducationChange">
              <view class="wmc-picker">
                <text :class="selfEducationText ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ selfEducationText || '请选择学历' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">年收入区间（可选，不采集精确工资）</text>
            <picker mode="selector" :range="incomeLabels" :value="selfIncomeIndex" @change="onSelfIncomeChange">
              <view class="wmc-picker">
                <text :class="selfIncomeText ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ selfIncomeText || '请选择年收入区间' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
        </view>

        <!-- 第 2 步：我希望找谁 -->
        <view v-if="step === 2" class="wmc-card">
          <view class="wmc-section-title">
            <text>我希望找谁</text>
            <text class="wmc-tag wmc-tag--muted">先填写期望值</text>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">目标性别（可多选）</text>
            <OptionChips
              :model-value="store.draft.conditions.target_genders"
              :options="genderOptions"
              @update:model-value="onTargetGenders"
            />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">年龄区间</text>
            <view class="range-row">
              <input class="wmc-input range-input" type="number" placeholder="最小年龄" :value="store.draft.conditions.min_age" @input="onCondInput('min_age', $event)" />
              <text class="range-sep">至</text>
              <input class="wmc-input range-input" type="number" placeholder="最大年龄" :value="store.draft.conditions.max_age" @input="onCondInput('max_age', $event)" />
            </view>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">城市（可添加多个）</text>
            <picker mode="selector" :range="cityLabels" :value="addCityIndex" @change="onAddCity">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ store.draft.conditions.allowed_city_codes.length > 0 ? '继续添加城市' : '请选择城市' }}</text>
                <text class="wmc-picker__arrow">添加</text>
              </view>
            </picker>
            <view v-if="store.draft.conditions.allowed_city_codes.length > 0" class="selected-row">
              <view v-for="code in store.draft.conditions.allowed_city_codes" :key="code" class="wmc-chip wmc-chip--active" @tap="removeCity(code)">
                {{ cityDisplay(code) }} · 移除
              </view>
            </view>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">身高区间（cm）</text>
            <view class="range-row">
              <input class="wmc-input range-input" type="number" placeholder="最低身高" :value="store.draft.conditions.min_height_cm" @input="onCondInput('min_height_cm', $event)" />
              <text class="range-sep">至</text>
              <input class="wmc-input range-input" type="number" placeholder="最高身高" :value="store.draft.conditions.max_height_cm" @input="onCondInput('max_height_cm', $event)" />
            </view>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">最低学历</text>
            <picker mode="selector" :range="educationLabels" :value="condEducationIndex" @change="onCondEducationChange">
              <view class="wmc-picker">
                <text :class="condEducationText ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ condEducationText || '不限制学历' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">最低年收入区间</text>
            <picker mode="selector" :range="incomeLabels" :value="condIncomeIndex" @change="onCondIncomeChange">
              <view class="wmc-picker">
                <text :class="condIncomeText ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ condIncomeText || '不限制收入' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">婚姻状态（可多选）</text>
            <OptionChips
              :model-value="store.draft.conditions.allowed_marital_statuses"
              :options="maritalOptions"
              @update:model-value="onMaritalStatuses"
            />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">子女情况（可多选）</text>
            <OptionChips
              :model-value="store.draft.conditions.allowed_children_states"
              :options="childrenOptions"
              @update:model-value="onChildrenStates"
            />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">生育意愿（可多选）</text>
            <OptionChips
              :model-value="store.draft.conditions.allowed_wants_children"
              :options="wantsChildrenOptions"
              @update:model-value="onWantsChildren"
            />
          </view>
        </view>

        <!-- 第 3 步：我的硬条件 -->
        <view v-if="step === 3" class="wmc-card">
          <view class="wmc-section-title">
            <text>我的硬条件</text>
            <text class="wmc-tag wmc-tag--muted">{{ mustCount }} 项必须</text>
          </view>
          <text class="wmc-hint">
            只有标记为「必须」且已填写的条件会参与理论市场估算。「偏好」与「不在意」仅作为你的备注，不会筛掉候选。
          </text>
          <view class="wmc-divider"></view>
          <view v-for="row in conditionRows" :key="row.key" class="level-row">
            <view class="level-row__head">
              <view class="level-row__title-wrap">
                <text class="level-row__title">{{ row.label }}</text>
                <text class="level-row__value">{{ conditionSummary(row.key) }}</text>
              </view>
              <view class="level-group">
                <text
                  v-for="level in levelKeys"
                  :key="level"
                  class="level-group__item"
                  :class="{ 'level-group__item--active': store.draft.levels[row.key] === level }"
                  @tap="store.setLevel(row.key, level)"
                >
                  {{ levelLabels[level] }}
                </text>
              </view>
            </view>
            <text v-if="!store.hasConditionValue(row.key)" class="level-row__hint">尚未填写该条件，暂不参与计算。</text>
          </view>
        </view>

        <!-- 第 4 步：计算 -->
        <view v-if="step === 4" class="wmc-card">
          <view class="wmc-section-title">
            <text>确认并计算</text>
          </view>
          <MarketBadge market="theoretical" :show-desc="true" />
          <view class="wmc-divider"></view>
          <text class="wmc-label">将参与估算的「必须」条件</text>
          <view v-if="appliedRows.length === 0" class="wmc-notice wmc-notice--warning review-warning">
            还没有已填写的「必须」条件，请返回上一步设置，否则结果接近全量人口。
          </view>
          <view v-for="row in appliedRows" :key="row.key" class="review-row">
            <text class="review-row__label">{{ row.label }}</text>
            <text class="review-row__value">{{ conditionSummary(row.key) }}</text>
          </view>
          <view class="wmc-divider"></view>
          <text class="wmc-hint">
            理论市场估算只回答「符合条件的人占多少」，不能计算双向匹配率，也不代表本站真实候选人数量。计算无需注册。
          </text>
          <button class="wmc-btn wmc-btn--ghost wmc-btn--block review-account" @tap="goLogin">
            {{ loggedIn ? '已登录，计算后可查看本站真实候选人' : '先注册 / 登录，计算后查看本站真实候选人' }}
          </button>
        </view>

        <view v-if="errorText" class="wmc-error error-box">{{ errorText }}</view>

        <view class="wmc-btn-row nav-row">
          <button v-if="step > 1" class="wmc-btn wmc-btn--ghost" @tap="prevStep">上一步</button>
          <button v-if="step < 4" class="wmc-btn wmc-btn--primary" @tap="nextStep">下一步</button>
          <button v-else class="wmc-btn wmc-btn--primary" :disabled="store.loading" @tap="runEstimate">
            {{ store.loading ? '计算中…' : '开始计算' }}
          </button>
        </view>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 计算器 4 步流程：我是谁 / 我希望找谁 / 我的硬条件 / 计算。
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import StepHeader from '@/components/StepHeader.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import OptionChips from '@/components/OptionChips.vue'
import {
  CONDITION_ROWS,
  CONDITION_LEVEL_LABELS,
  useCalculatorStore
} from '@/stores/calculator'
import type {
  CalculatorConditionsDraft,
  CalculatorSelfDraft,
  ConditionKey,
  ConditionLevel
} from '@/stores/calculator'
import { useAuthStore } from '@/stores/auth'
import {
  cityDisplay,
  cityRegions,
  ensureOptions,
  ensureRegions,
  metaState,
  optionLabel,
  optionsOf,
  pickerOptions
} from '@/utils/options'
import {
  toIntOrNull,
  validateCalculatorConditions,
  validateCalculatorSelf,
  validateTargetGenders
} from '@/utils/validate'
import { errorMessage } from '@/utils/error'
import { trackEvent } from '@/api/events'

const store = useCalculatorStore()
const auth = useAuthStore()

const step = ref(1)
const ready = ref(false)
const errorText = ref('')
const optionsError = ref('')
const addCityIndex = ref(0)

const stepLabels = ['我是谁', '我希望找谁', '我的硬条件', '计算']
const conditionRows = CONDITION_ROWS
const levelKeys: ConditionLevel[] = ['must', 'prefer', 'ignore']
const levelLabels = CONDITION_LEVEL_LABELS

const loggedIn = computed(() => auth.isLoggedIn)

const genderOptions = computed(() => pickerOptions('genders', false))
const maritalOptions = computed(() => pickerOptions('marital_statuses', true, [4]))
const childrenOptions = computed(() => optionsOf('children_states'))
const wantsChildrenOptions = computed(() => pickerOptions('wants_children'))
const educationOptions = computed(() => pickerOptions('education_levels'))
const incomeOptions = computed(() => pickerOptions('income_bands', true, [8]))
const cityOptions = computed(() => cityRegions())
const cityLabels = computed(() => cityOptions.value.map((item) => cityDisplay(item.code)))

const educationLabels = computed(() => educationOptions.value.map((item) => item.label))
const incomeLabels = computed(() => incomeOptions.value.map((item) => item.label))

const selfGenderValue = computed<number[]>(() => {
  const value = store.draft.self.gender
  return value === null || value === undefined || value <= 0 ? [] : [value]
})

const selfCityText = computed(() => cityDisplay(store.draft.self.city_code))
const selfCityIndex = computed(() => {
  const index = cityOptions.value.findIndex((item) => item.code === store.draft.self.city_code)
  return index < 0 ? 0 : index
})

const selfEducationText = computed(() => optionLabel('education_levels', store.draft.self.education_level, ''))
const selfEducationIndex = computed(() => {
  const index = educationOptions.value.findIndex((item) => item.value === store.draft.self.education_level)
  return index < 0 ? 0 : index
})

const selfIncomeText = computed(() => optionLabel('income_bands', store.draft.self.annual_income_band, ''))
const selfIncomeIndex = computed(() => {
  const index = incomeOptions.value.findIndex((item) => item.value === store.draft.self.annual_income_band)
  return index < 0 ? 0 : index
})

const condEducationText = computed(() => optionLabel('education_levels', store.draft.conditions.min_education_level, ''))
const condEducationIndex = computed(() => {
  const index = educationOptions.value.findIndex((item) => item.value === store.draft.conditions.min_education_level)
  return index < 0 ? 0 : index
})

const condIncomeText = computed(() => optionLabel('income_bands', store.draft.conditions.min_income_band, ''))
const condIncomeIndex = computed(() => {
  const index = incomeOptions.value.findIndex((item) => item.value === store.draft.conditions.min_income_band)
  return index < 0 ? 0 : index
})

const mustCount = computed(() => store.mustCount)
const appliedRows = computed(() => {
  return conditionRows.filter((row) => store.draft.levels[row.key] === 'must' && store.hasConditionValue(row.key))
})

onShow(async () => {
  auth.syncFromStorage()
  if (ready.value) return
  const [optionsOk, regionsOk] = await Promise.all([ensureOptions(), ensureRegions()])
  if (!optionsOk || !regionsOk) {
    optionsError.value = metaState.error || '选项数据加载失败，请检查网络后重试'
  }
  ready.value = true
})

function onSelfInput(field: keyof CalculatorSelfDraft, event: { detail: { value: string } }) {
  const patch: Partial<CalculatorSelfDraft> = {}
  if (field === 'age' || field === 'height_cm' || field === 'city_code') {
    patch[field] = event.detail.value
  }
  store.setSelf(patch)
}

function onSelfGender(value: Array<number | string>) {
  const first = value.length > 0 ? Number(value[0]) : 0
  store.setSelf({ gender: first > 0 ? first : null })
}

function onSelfCityChange(event: { detail: { value: number | string } }) {
  const option = cityOptions.value[Number(event.detail.value)]
  if (option) store.setSelf({ city_code: option.code })
}

function onSelfEducationChange(event: { detail: { value: number | string } }) {
  const option = educationOptions.value[Number(event.detail.value)]
  if (option) store.setSelf({ education_level: option.value })
}

function onSelfIncomeChange(event: { detail: { value: number | string } }) {
  const option = incomeOptions.value[Number(event.detail.value)]
  if (option) store.setSelf({ annual_income_band: option.value })
}

function onCondInput(field: keyof CalculatorConditionsDraft, event: { detail: { value: string } }) {
  const patch: Partial<CalculatorConditionsDraft> = {}
  if (
    field === 'min_age' ||
    field === 'max_age' ||
    field === 'min_height_cm' ||
    field === 'max_height_cm'
  ) {
    patch[field] = event.detail.value
  }
  store.setConditions(patch)
}

function onTargetGenders(value: Array<number | string>) {
  store.setConditions({ target_genders: value.map((item) => Number(item)) })
}

function onMaritalStatuses(value: Array<number | string>) {
  store.setConditions({ allowed_marital_statuses: value.map((item) => Number(item)) })
}

function onChildrenStates(value: Array<number | string>) {
  store.setConditions({ allowed_children_states: value.map((item) => Number(item)) })
}

function onWantsChildren(value: Array<number | string>) {
  store.setConditions({ allowed_wants_children: value.map((item) => Number(item)) })
}

function onAddCity(event: { detail: { value: number | string } }) {
  const option = cityOptions.value[Number(event.detail.value)]
  if (!option) return
  const current = store.draft.conditions.allowed_city_codes
  if (current.indexOf(option.code) >= 0) return
  store.setConditions({ allowed_city_codes: current.concat([option.code]) })
}

function removeCity(code: string) {
  const current = store.draft.conditions.allowed_city_codes.filter((item) => item !== code)
  store.setConditions({ allowed_city_codes: current })
}

function onCondEducationChange(event: { detail: { value: number | string } }) {
  const option = educationOptions.value[Number(event.detail.value)]
  if (option) store.setConditions({ min_education_level: option.value })
}

function onCondIncomeChange(event: { detail: { value: number | string } }) {
  const option = incomeOptions.value[Number(event.detail.value)]
  if (option) store.setConditions({ min_income_band: option.value })
}

function rangeText(minText: string, maxText: string, unit: string): string {
  const min = toIntOrNull(minText)
  const max = toIntOrNull(maxText)
  if (min === null && max === null) return '不限制'
  if (min !== null && max !== null) return min + ' - ' + max + ' ' + unit
  if (min !== null) return '不低于 ' + min + ' ' + unit
  return '不高于 ' + max + ' ' + unit
}

function conditionSummary(key: ConditionKey): string {
  const conditions = store.draft.conditions
  if (key === 'gender') {
    if (conditions.target_genders.length === 0) return '未填写'
    return conditions.target_genders.map((value) => optionLabel('genders', value, String(value))).join('、')
  }
  if (key === 'age') return rangeText(conditions.min_age, conditions.max_age, '岁')
  if (key === 'city') {
    if (conditions.allowed_city_codes.length === 0) return '未填写'
    return conditions.allowed_city_codes.map((code) => cityDisplay(code)).join('、')
  }
  if (key === 'height') return rangeText(conditions.min_height_cm, conditions.max_height_cm, 'cm')
  if (key === 'education') {
    const value = conditions.min_education_level
    return value === null || value === undefined || value <= 0 ? '未填写' : '最低 ' + optionLabel('education_levels', value, String(value))
  }
  if (key === 'income') {
    const value = conditions.min_income_band
    return value === null || value === undefined || value <= 0 ? '未填写' : '最低 ' + optionLabel('income_bands', value, String(value))
  }
  if (key === 'marital_status') {
    if (conditions.allowed_marital_statuses.length === 0) return '未填写'
    return conditions.allowed_marital_statuses.map((value) => optionLabel('marital_statuses', value, String(value))).join('、')
  }
  if (key === 'children_state') {
    if (conditions.allowed_children_states.length === 0) return '未填写'
    return conditions.allowed_children_states.map((value) => optionLabel('children_states', value, String(value))).join('、')
  }
  if (key === 'wants_children') {
    if (conditions.allowed_wants_children.length === 0) return '未填写'
    return conditions.allowed_wants_children.map((value) => optionLabel('wants_children', value, String(value))).join('、')
  }
  return '未填写'
}

function validateStep(current: number): string {
  if (current === 1) {
    const selfError = validateCalculatorSelf(store.draft.self)
    if (selfError) return selfError
  }
  if (current === 2) {
    const genderError = validateTargetGenders(store.draft.conditions.target_genders)
    if (genderError) return genderError
    const conditionError = validateCalculatorConditions(store.draft.conditions)
    if (conditionError) return conditionError
  }
  if (current === 3) {
    if (appliedRows.value.length === 0) return '请至少设置一个已填写的「必须」条件'
  }
  return ''
}

function nextStep() {
  const message = validateStep(step.value)
  if (message) {
    errorText.value = message
    return
  }
  errorText.value = ''
  step.value = step.value + 1
}

function prevStep() {
  errorText.value = ''
  if (step.value > 1) step.value = step.value - 1
}

async function runEstimate() {
  const message = validateStep(3)
  if (message) {
    errorText.value = message
    return
  }
  errorText.value = ''
  try {
    const result = await store.runEstimate()
    trackEvent('calculator_estimate_submitted', {
      must_count: store.mustCount,
      matched_ratio: result.matched_ratio
    })
    uni.navigateTo({ url: '/pages/calculator/result' })
  } catch (error) {
    errorText.value = errorMessage(error, '计算失败，请稍后重试')
  }
}

function goLogin() {
  if (loggedIn.value) {
    uni.navigateTo({ url: '/pages/discovery/index' })
    return
  }
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/discovery/index') })
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
.level-row {
  padding: 20rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}
.level-row:last-child {
  border-bottom: none;
}
.level-row__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}
.level-row__title-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.level-row__title {
  font-size: 28rpx;
  color: $uni-text-color;
  font-weight: 500;
}
.level-row__value {
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-top: 4rpx;
  line-height: 1.5;
}
.level-row__hint {
  display: block;
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
  margin-top: 8rpx;
}
.level-group {
  display: flex;
  border: 1rpx solid $uni-border-color;
  border-radius: 10rpx;
  overflow: hidden;
  flex-shrink: 0;
}
.level-group__item {
  padding: 10rpx 20rpx;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  background: #ffffff;
}
.level-group__item--active {
  background: $uni-color-primary;
  color: #ffffff;
}
.review-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}
.review-row:last-of-type {
  border-bottom: none;
}
.review-row__label {
  font-size: 26rpx;
  color: $uni-text-color-grey;
  flex-shrink: 0;
  margin-right: 24rpx;
}
.review-row__value {
  font-size: 26rpx;
  color: $uni-text-color;
  text-align: right;
  flex: 1;
  line-height: 1.6;
}
.review-warning {
  margin-top: 16rpx;
}
.review-account {
  margin-top: 24rpx;
  font-size: 26rpx;
}
.error-box {
  padding: 0 8rpx 16rpx;
}
.nav-row {
  margin-top: 8rpx;
}
</style>
