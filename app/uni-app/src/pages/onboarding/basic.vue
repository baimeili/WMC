<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="基础资料"
        subtitle="先填写昵称、性别和城市，这些是进入真实候选池的基础。资料可随时修改。"
        :percent="completion"
        :show-back="true"
      />
      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载资料…</text>
      </view>
      <view v-else class="wmc-card">
        <view class="wmc-field">
          <text class="wmc-field-label">昵称（必填）</text>
          <input class="wmc-input" maxlength="32" placeholder="例如 小林" :value="draft.nickname" @input="onInput('nickname', $event)" />
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">性别（必填）</text>
          <OptionChips :model-value="genderValue" :options="genderOptions" :single="true" @update:model-value="onGender" />
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">出生日期</text>
          <picker mode="date" :value="draft.birth_date" @change="onBirthDate">
            <view class="wmc-picker">
              <text :class="draft.birth_date ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ draft.birth_date || '请选择出生日期' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
          <text v-if="ageText" class="field-hint">{{ ageText }}</text>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">所在省份</text>
          <picker mode="selector" :range="provinceLabels" :value="provinceIndex" @change="onProvince">
            <view class="wmc-picker">
              <text :class="draft.province_code ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ provinceText || '请选择省份' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">所在城市（必填）</text>
          <picker mode="selector" :range="cityLabels" :value="cityIndex" @change="onCity">
            <view class="wmc-picker">
              <text :class="draft.city_code ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ cityText || '请选择城市' }}</text>
              <text class="wmc-picker__arrow">选择</text>
            </view>
          </picker>
        </view>
        <view class="wmc-field">
          <text class="wmc-field-label">身高（cm，可选）</text>
          <input class="wmc-input" type="number" placeholder="例如 170" :value="draft.height_cm" @input="onInput('height_cm', $event)" />
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
// 资料完善第 1 步：昵称、性别、出生日期、城市、身高。
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import StepHeader from '@/components/StepHeader.vue'
import OptionChips from '@/components/OptionChips.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { citiesOf, cityDisplay, ensureOptions, ensureRegions, metaState, optionLabel, pickerOptions, provinces } from '@/utils/options'
import { validateBirthDate, validateNickname, validateProfileBase, toIntOrNull } from '@/utils/validate'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const profile = useProfileStore()

const ready = ref(false)
const saving = ref(false)
const errorText = ref('')
const draft = reactive({
  nickname: '',
  gender: 0,
  birth_date: '',
  province_code: '',
  city_code: '',
  height_cm: ''
})

const completion = computed(() => profile.completion)
const genderOptions = computed(() => pickerOptions('genders', false))
const genderValue = computed<number[]>(() => (draft.gender > 0 ? [draft.gender] : []))
const provinceOptions = computed(() => provinces())
const provinceLabels = computed(() => provinceOptions.value.map((item) => item.name))
const cityOptions = computed(() => (draft.province_code ? citiesOf(draft.province_code) : []))
const cityLabels = computed(() => cityOptions.value.map((item) => item.name))
const provinceText = computed(() => (provinceOptions.value.find((item) => item.code === draft.province_code) || { name: '' }).name)
const cityText = computed(() => cityDisplay(draft.city_code) || cityOptions.value.find((item) => item.code === draft.city_code)?.name || '')
const provinceIndex = computed(() => {
  const index = provinceOptions.value.findIndex((item) => item.code === draft.province_code)
  return index < 0 ? 0 : index
})
const cityIndex = computed(() => {
  const index = cityOptions.value.findIndex((item) => item.code === draft.city_code)
  return index < 0 ? 0 : index
})
const ageText = computed(() => {
  if (!draft.birth_date) return ''
  const date = new Date(draft.birth_date)
  if (isNaN(date.getTime())) return ''
  const now = new Date()
  let age = now.getFullYear() - date.getFullYear()
  const monthDiff = now.getMonth() - date.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && now.getDate() < date.getDate())) age--
  return age >= 0 ? '当前年龄约 ' + age + ' 岁' : ''
})

onShow(async () => {
  if (!auth.requireLogin('/pages/onboarding/basic')) return
  if (ready.value) return
  await Promise.all([ensureOptions(), ensureRegions()])
  try {
    const data = await profile.loadProfile(true)
    draft.nickname = data.nickname || ''
    draft.gender = data.gender || 0
    draft.birth_date = data.birth_date ? String(data.birth_date).slice(0, 10) : ''
    draft.province_code = data.province_code || ''
    draft.city_code = data.city_code || ''
    draft.height_cm = data.height_cm ? String(data.height_cm) : ''
  } catch (error) {
    errorText.value = errorMessage(error, '资料加载失败，请稍后重试')
  }
  ready.value = true
})

function onInput(field: 'nickname' | 'height_cm', event: { detail: { value: string } }) {
  draft[field] = event.detail.value
}

function onGender(value: Array<number | string>) {
  draft.gender = value.length > 0 ? Number(value[0]) : 0
}

function onBirthDate(event: { detail: { value: string } }) {
  draft.birth_date = event.detail.value
}

function onProvince(event: { detail: { value: number | string } }) {
  const option = provinceOptions.value[Number(event.detail.value)]
  if (!option) return
  draft.province_code = option.code
  draft.city_code = ''
}

function onCity(event: { detail: { value: number | string } }) {
  const option = cityOptions.value[Number(event.detail.value)]
  if (option) draft.city_code = option.code
}

async function persist(): Promise<boolean> {
  const nicknameError = validateNickname(draft.nickname)
  if (nicknameError) {
    errorText.value = nicknameError
    return false
  }
  const baseError = validateProfileBase({
    nickname: draft.nickname,
    gender: draft.gender > 0 ? draft.gender : null,
    city_code: draft.city_code
  })
  if (baseError) {
    errorText.value = baseError
    return false
  }
  const birthError = validateBirthDate(draft.birth_date || null)
  if (birthError) {
    errorText.value = birthError
    return false
  }
  saving.value = true
  errorText.value = ''
  try {
    await profile.updateProfile({
      nickname: draft.nickname,
      gender: draft.gender,
      province_code: draft.province_code,
      city_code: draft.city_code,
      birth_date: draft.birth_date || null,
      height_cm: toIntOrNull(draft.height_cm)
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
  if (await persist()) {
    uni.navigateTo({ url: '/pages/onboarding/career' })
  }
}

async function saveExit() {
  if (saving.value) return
  if (await persist()) {
    uni.reLaunch({ url: '/pages/profile/index' })
  }
}
</script>

<style lang="scss" scoped>
.field-hint {
  display: block;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-top: 8rpx;
}
.save-btn {
  margin-top: 16rpx;
}
</style>
