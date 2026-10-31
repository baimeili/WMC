<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="教育与职业"
        subtitle="用于计算理论市场条件占比，也用于真实候选池的条件匹配。收入只采集区间，不采集精确工资。"
        :percent="completion"
        :show-back="true"
      />
      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载资料…</text>
      </view>
      <view v-else class="wmc-card">
        <view class="wmc-field">
          <text class="wmc-field-label">最高学历</text>
          <picker mode="selector" :range="educationLabels" :value="educationIndex" @change="onEducation">
            <view class="wmc-picker">
              <text :class="draft.education_level ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ educationText || '请选择学历' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">职业类别</text>
          <picker mode="selector" :range="occupationLabels" :value="occupationIndex" @change="onOccupation">
            <view class="wmc-picker">
              <text :class="draft.occupation_category ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ occupationText || '请选择职业类别' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">就业状态</text>
          <picker mode="selector" :range="employmentLabels" :value="employmentIndex" @change="onEmployment">
            <view class="wmc-picker">
              <text :class="draft.employment_status ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ employmentText || '请选择就业状态' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">年收入区间</text>
          <picker mode="selector" :range="incomeLabels" :value="incomeIndex" @change="onIncome">
            <view class="wmc-picker">
              <text :class="draft.annual_income_band ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ incomeText || '请选择年收入区间' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">家庭资产区间（可选）</text>
          <picker mode="selector" :range="assetLabels" :value="assetIndex" @change="onAsset">
            <view class="wmc-picker">
              <text :class="draft.family_asset_band ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ assetText || '不愿透露可不填' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">负债区间</text>
          <picker mode="selector" :range="debtLabels" :value="debtIndex" @change="onDebt">
            <view class="wmc-picker">
              <text :class="draft.debt_band ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ debtText || '请选择负债区间' }}</text>
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
// 资料完善第 2 步：学历、职业、就业状态、收入区间、资产与负债区间。
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
  education_level: 0,
  occupation_category: 0,
  employment_status: 0,
  annual_income_band: 0,
  family_asset_band: 0,
  debt_band: 0
})

const completion = computed(() => profile.completion)
const educationOptions = computed(() => pickerOptions('education_levels'))
const occupationOptions = computed(() => pickerOptions('occupations'))
const employmentOptions = computed(() => pickerOptions('employment_statuses'))
const incomeOptions = computed(() => pickerOptions('income_bands', true, [8]))
const assetOptions = computed(() => pickerOptions('asset_bands', true, []))
const debtOptions = computed(() => pickerOptions('debt_bands', true, [6]))

const educationLabels = computed(() => educationOptions.value.map((item) => item.label))
const occupationLabels = computed(() => occupationOptions.value.map((item) => item.label))
const employmentLabels = computed(() => employmentOptions.value.map((item) => item.label))
const incomeLabels = computed(() => incomeOptions.value.map((item) => item.label))
const assetLabels = computed(() => assetOptions.value.map((item) => item.label))
const debtLabels = computed(() => debtOptions.value.map((item) => item.label))

const educationText = computed(() => optionLabel('education_levels', draft.education_level, ''))
const occupationText = computed(() => optionLabel('occupations', draft.occupation_category, ''))
const employmentText = computed(() => optionLabel('employment_statuses', draft.employment_status, ''))
const incomeText = computed(() => optionLabel('income_bands', draft.annual_income_band, ''))
const assetText = computed(() => optionLabel('asset_bands', draft.family_asset_band, ''))
const debtText = computed(() => optionLabel('debt_bands', draft.debt_band, ''))

function indexOfValue(options: Array<{ value: number }>, value: number): number {
  const index = options.findIndex((item) => item.value === value)
  return index < 0 ? 0 : index
}

const educationIndex = computed(() => indexOfValue(educationOptions.value, draft.education_level))
const occupationIndex = computed(() => indexOfValue(occupationOptions.value, draft.occupation_category))
const employmentIndex = computed(() => indexOfValue(employmentOptions.value, draft.employment_status))
const incomeIndex = computed(() => indexOfValue(incomeOptions.value, draft.annual_income_band))
const assetIndex = computed(() => indexOfValue(assetOptions.value, draft.family_asset_band))
const debtIndex = computed(() => indexOfValue(debtOptions.value, draft.debt_band))

onShow(async () => {
  if (!auth.requireLogin('/pages/onboarding/career')) return
  if (ready.value) return
  await ensureOptions()
  try {
    const data = await profile.loadProfile(true)
    draft.education_level = data.education_level || 0
    draft.occupation_category = data.occupation_category || 0
    draft.employment_status = data.employment_status || 0
    draft.annual_income_band = data.annual_income_band || 0
    draft.family_asset_band = data.family_asset_band || 0
    draft.debt_band = data.debt_band || 0
  } catch (error) {
    errorText.value = errorMessage(error, '资料加载失败，请稍后重试')
  }
  ready.value = true
})

function onEducation(event: { detail: { value: number | string } }) {
  const option = educationOptions.value[Number(event.detail.value)]
  if (option) draft.education_level = option.value
}
function onOccupation(event: { detail: { value: number | string } }) {
  const option = occupationOptions.value[Number(event.detail.value)]
  if (option) draft.occupation_category = option.value
}
function onEmployment(event: { detail: { value: number | string } }) {
  const option = employmentOptions.value[Number(event.detail.value)]
  if (option) draft.employment_status = option.value
}
function onIncome(event: { detail: { value: number | string } }) {
  const option = incomeOptions.value[Number(event.detail.value)]
  if (option) draft.annual_income_band = option.value
}
function onAsset(event: { detail: { value: number | string } }) {
  const option = assetOptions.value[Number(event.detail.value)]
  if (option) draft.family_asset_band = option.value
}
function onDebt(event: { detail: { value: number | string } }) {
  const option = debtOptions.value[Number(event.detail.value)]
  if (option) draft.debt_band = option.value
}

async function persist(): Promise<boolean> {
  saving.value = true
  errorText.value = ''
  try {
    await profile.updateProfile({
      education_level: draft.education_level,
      occupation_category: draft.occupation_category,
      employment_status: draft.employment_status,
      annual_income_band: draft.annual_income_band,
      family_asset_band: draft.family_asset_band > 0 ? draft.family_asset_band : null,
      debt_band: draft.debt_band
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
  if (await persist()) uni.navigateTo({ url: '/pages/onboarding/marriage' })
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
