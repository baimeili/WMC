<template>
  <PageShell>
    <view class="wmc-container">
      <view v-if="!loggedIn" class="wmc-card">
        <EmptyState title="登录后可编辑资料" description="资料用于条件匹配与候选池筛选。" action-text="去登录" @action="goLogin" />
      </view>
      <view v-else-if="loading" class="wmc-card">
        <text class="wmc-hint">正在加载资料…</text>
      </view>
      <template v-else>
        <view class="wmc-card">
          <view class="wmc-section-title"><text>基础资料</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">昵称</text>
            <input class="wmc-input" maxlength="32" :value="draft.nickname" @input="onInput('nickname', $event)" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">性别</text>
            <OptionChips :model-value="genderValue" :options="list('genders', false)" :single="true" @update:model-value="onGender" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">出生日期</text>
            <picker mode="date" :value="draft.birth_date" @change="onBirthDate">
              <view class="wmc-picker">
                <text :class="draft.birth_date ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ draft.birth_date || '请选择出生日期' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
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
            <text class="wmc-field-label">所在城市</text>
            <picker mode="selector" :range="cityLabels" :value="cityIndex" @change="onCity">
              <view class="wmc-picker">
                <text :class="draft.city_code ? 'wmc-picker__text' : 'wmc-picker__placeholder'">{{ cityText || '请选择城市' }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">身高（cm）</text>
            <input class="wmc-input" type="number" :value="draft.height_cm" @input="onInput('height_cm', $event)" />
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>教育与职业</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">最高学历</text>
            <picker mode="selector" :range="labels('education_levels')" :value="idx('education_levels', draft.education_level)" @change="onPick('education_levels', $event, 'education_level')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('education_levels', draft.education_level, '请选择学历') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">职业类别</text>
            <picker mode="selector" :range="labels('occupations')" :value="idx('occupations', draft.occupation_category)" @change="onPick('occupations', $event, 'occupation_category')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('occupations', draft.occupation_category, '请选择职业') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">就业状态</text>
            <picker mode="selector" :range="labels('employment_statuses')" :value="idx('employment_statuses', draft.employment_status)" @change="onPick('employment_statuses', $event, 'employment_status')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('employment_statuses', draft.employment_status, '请选择就业状态') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">年收入区间</text>
            <picker mode="selector" :range="labels('income_bands', true, [8])" :value="idx('income_bands', draft.annual_income_band, true, [8])" @change="onPick('income_bands', $event, 'annual_income_band', true, [8])">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('income_bands', draft.annual_income_band, '请选择收入区间') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">家庭资产区间</text>
            <picker mode="selector" :range="labels('asset_bands')" :value="idx('asset_bands', draft.family_asset_band)" @change="onPick('asset_bands', $event, 'family_asset_band')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('asset_bands', draft.family_asset_band, '可选填') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">负债区间</text>
            <picker mode="selector" :range="labels('debt_bands', true, [6])" :value="idx('debt_bands', draft.debt_band, true, [6])" @change="onPick('debt_bands', $event, 'debt_band', true, [6])">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('debt_bands', draft.debt_band, '请选择负债区间') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>婚姻与家庭</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">婚姻状态</text>
            <picker mode="selector" :range="labels('marital_statuses', true, [4])" :value="idx('marital_statuses', draft.marital_status, true, [4])" @change="onPick('marital_statuses', $event, 'marital_status', true, [4])">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('marital_statuses', draft.marital_status, '请选择婚姻状态') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">子女数量</text>
            <picker mode="selector" :range="childCountLabels" :value="draft.children_count > 5 ? 5 : draft.children_count" @change="onChildCount">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ childCountText }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">生育意愿</text>
            <picker mode="selector" :range="labels('wants_children')" :value="idx('wants_children', draft.wants_children)" @change="onPick('wants_children', $event, 'wants_children')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('wants_children', draft.wants_children, '请选择生育意愿') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">结婚意愿</text>
            <picker mode="selector" :range="labels('marriage_intents')" :value="idx('marriage_intents', draft.marriage_intent)" @change="onPick('marriage_intents', $event, 'marriage_intent')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('marriage_intents', draft.marriage_intent, '请选择结婚意愿') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">结婚时间规划</text>
            <picker mode="selector" :range="labels('marriage_timelines')" :value="idx('marriage_timelines', draft.marriage_timeline)" @change="onPick('marriage_timelines', $event, 'marriage_timeline')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('marriage_timelines', draft.marriage_timeline, '请选择时间规划') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title"><text>生活习惯</text></view>
          <view class="wmc-field">
            <text class="wmc-field-label">吸烟习惯</text>
            <picker mode="selector" :range="labels('smoking_statuses')" :value="idx('smoking_statuses', draft.smoking_status)" @change="onPick('smoking_statuses', $event, 'smoking_status')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('smoking_statuses', draft.smoking_status, '请选择') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">饮酒习惯</text>
            <picker mode="selector" :range="labels('drinking_statuses')" :value="idx('drinking_statuses', draft.drinking_status)" @change="onPick('drinking_statuses', $event, 'drinking_status')">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('drinking_statuses', draft.drinking_status, '请选择') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">房产状态</text>
            <picker mode="selector" :range="labels('house_statuses', true, [6])" :value="idx('house_statuses', draft.house_status, true, [6])" @change="onPick('house_statuses', $event, 'house_status', true, [6])">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('house_statuses', draft.house_status, '请选择') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">车辆状态</text>
            <picker mode="selector" :range="labels('vehicle_statuses', true, [4])" :value="idx('vehicle_statuses', draft.vehicle_status, true, [4])" @change="onPick('vehicle_statuses', $event, 'vehicle_status', true, [4])">
              <view class="wmc-picker">
                <text class="wmc-picker__text">{{ optionLabel('vehicle_statuses', draft.vehicle_status, '请选择') }}</text>
                <text class="wmc-picker__arrow">选择</text>
              </view>
            </picker>
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">个人简介（最多 500 字）</text>
            <textarea class="wmc-textarea" maxlength="500" :value="draft.bio" @input="onInput('bio', $event)"></textarea>
          </view>
        </view>

        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>我的照片</text>
            <text class="wmc-tag wmc-tag--muted">上传后自动去除位置信息</text>
          </view>
          <view class="photo-grid">
            <view v-for="photo in photos" :key="photo.id" class="photo-item">
              <image class="photo-image" :src="photo.url" mode="aspectFill" />
              <text class="photo-remove" @tap="removePhoto(photo)">删除</text>
            </view>
            <view class="photo-add" @tap="choosePhoto">
              <text class="photo-add__text">{{ uploading ? '上传中…' : '+ 添加照片' }}</text>
            </view>
          </view>
          <text class="wmc-hint">照片仅用于资料展示，上传时会自动去除 EXIF 与 GPS 等信息。</text>
        </view>

        <text v-if="errorText" class="wmc-error">{{ errorText }}</text>
        <button class="wmc-btn wmc-btn--primary wmc-btn--block save-btn" :disabled="saving" @tap="save">
          {{ saving ? '保存中…' : '保存资料' }}
        </button>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block" @tap="back">取消</button>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 编辑资料：完整资料字段 + 照片上传（自动去除 EXIF），保存到 /me/profile。
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import EmptyState from '@/components/EmptyState.vue'
import OptionChips from '@/components/OptionChips.vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { deletePhoto, listPhotos, uploadPhoto } from '@/api/photos'
import {
  citiesOf,
  cityDisplay,
  ensureOptions,
  ensureRegions,
  optionLabel,
  pickerOptions,
  provinces
} from '@/utils/options'
import { errorMessage } from '@/utils/error'
import { toIntOrNull, validateBio, validateBirthDate, validateNickname, validateProfileBase } from '@/utils/validate'
import type { Photo } from '@/api/types'

const auth = useAuthStore()
const profileStore = useProfileStore()
const loading = ref(true)
const saving = ref(false)
const uploading = ref(false)
const errorText = ref('')
const photos = ref<Photo[]>([])

const draft = reactive({
  nickname: '',
  gender: 0,
  birth_date: '',
  province_code: '',
  city_code: '',
  height_cm: '',
  education_level: 0,
  occupation_category: 0,
  employment_status: 0,
  annual_income_band: 0,
  family_asset_band: 0,
  debt_band: 0,
  marital_status: 0,
  children_count: 0,
  wants_children: 0,
  marriage_intent: 0,
  marriage_timeline: 0,
  smoking_status: 0,
  drinking_status: 0,
  house_status: 0,
  vehicle_status: 0,
  bio: ''
})

const loggedIn = computed(() => auth.isLoggedIn)
const genderValue = computed<number[]>(() => (draft.gender > 0 ? [draft.gender] : []))
const provinceOptions = computed(() => provinces())
const provinceLabels = computed(() => provinceOptions.value.map((item) => item.name))
const cityOptions = computed(() => (draft.province_code ? citiesOf(draft.province_code) : []))
const cityLabels = computed(() => cityOptions.value.map((item) => item.name))
const provinceText = computed(() => (provinceOptions.value.find((item) => item.code === draft.province_code) || { name: '' }).name)
const cityText = computed(() => cityDisplay(draft.city_code) || (cityOptions.value.find((item) => item.code === draft.city_code) || { name: '' }).name)
const provinceIndex = computed(() => {
  const index = provinceOptions.value.findIndex((item) => item.code === draft.province_code)
  return index < 0 ? 0 : index
})
const cityIndex = computed(() => {
  const index = cityOptions.value.findIndex((item) => item.code === draft.city_code)
  return index < 0 ? 0 : index
})
const childCountOptions = [0, 1, 2, 3, 4, 5].map((value) => ({ value, label: value === 5 ? '5 个及以上' : value + ' 个' }))
const childCountLabels = childCountOptions.map((item) => item.label)
const childCountText = computed(() => (draft.children_count >= 5 ? '5 个及以上' : draft.children_count + ' 个'))

function list(dict: string, skipZero = true, skipValues: number[] = []) {
  return pickerOptions(dict, skipZero, skipValues)
}

function labels(dict: string, skipZero = true, skipValues: number[] = []): string[] {
  return list(dict, skipZero, skipValues).map((item) => item.label)
}

function idx(dict: string, value: number, skipZero = true, skipValues: number[] = []): number {
  const options = list(dict, skipZero, skipValues)
  const index = options.findIndex((item) => item.value === Number(value))
  return index < 0 ? 0 : index
}

onShow(async () => {
  if (!auth.requireLogin('/pages/profile/edit')) return
  if (!loading.value) return
  await Promise.all([ensureOptions(), ensureRegions()])
  try {
    const data = await profileStore.loadProfile(true)
    draft.nickname = data.nickname || ''
    draft.gender = data.gender || 0
    draft.birth_date = data.birth_date ? String(data.birth_date).slice(0, 10) : ''
    draft.province_code = data.province_code || ''
    draft.city_code = data.city_code || ''
    draft.height_cm = data.height_cm ? String(data.height_cm) : ''
    draft.education_level = data.education_level || 0
    draft.occupation_category = data.occupation_category || 0
    draft.employment_status = data.employment_status || 0
    draft.annual_income_band = data.annual_income_band || 0
    draft.family_asset_band = data.family_asset_band || 0
    draft.debt_band = data.debt_band || 0
    draft.marital_status = data.marital_status || 0
    draft.children_count = data.children_count || 0
    draft.wants_children = data.wants_children || 0
    draft.marriage_intent = data.marriage_intent || 0
    draft.marriage_timeline = data.marriage_timeline || 0
    draft.smoking_status = data.smoking_status || 0
    draft.drinking_status = data.drinking_status || 0
    draft.house_status = data.house_status || 0
    draft.vehicle_status = data.vehicle_status || 0
    draft.bio = data.bio || ''
    await loadPhotos()
  } catch (error) {
    errorText.value = errorMessage(error, '资料加载失败')
  } finally {
    loading.value = false
  }
})

async function loadPhotos() {
  try {
    photos.value = await listPhotos()
  } catch (error) {
    photos.value = []
  }
}

function onInput(field: 'nickname' | 'height_cm' | 'bio', event: { detail: { value: string } }) {
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

function onPick(dict: string, event: { detail: { value: number | string } }, field: string, skipZero = true, skipValues: number[] = []) {
  const options = list(dict, skipZero, skipValues)
  const option = options[Number(event.detail.value)]
  if (!option) return
  ;(draft as unknown as Record<string, number>)[field] = option.value
}

function onChildCount(event: { detail: { value: number | string } }) {
  draft.children_count = Number(event.detail.value)
}

async function choosePhoto() {
  if (uploading.value) return
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: async (response) => {
      const filePath = response.tempFilePaths && response.tempFilePaths.length > 0 ? response.tempFilePaths[0] : ''
      if (!filePath) return
      uploading.value = true
      try {
        await uploadPhoto(filePath, photos.value.length === 0)
        uni.showToast({ title: '照片已上传', icon: 'none' })
        await loadPhotos()
      } catch (error) {
        uni.showToast({ title: errorMessage(error, '照片上传失败'), icon: 'none' })
      } finally {
        uploading.value = false
      }
    }
  })
}

function removePhoto(photo: Photo) {
  uni.showModal({
    title: '删除照片',
    content: '删除后不可恢复，确定继续吗？',
    confirmText: '删除',
    cancelText: '取消',
    success: async (response) => {
      if (!response.confirm) return
      try {
        await deletePhoto(photo.id)
        await loadPhotos()
      } catch (error) {
        uni.showToast({ title: errorMessage(error, '删除失败'), icon: 'none' })
      }
    }
  })
}

async function save() {
  if (saving.value) return
  const nicknameError = validateNickname(draft.nickname)
  if (nicknameError) {
    errorText.value = nicknameError
    return
  }
  const baseError = validateProfileBase({
    nickname: draft.nickname,
    gender: draft.gender > 0 ? draft.gender : null,
    city_code: draft.city_code
  })
  if (baseError) {
    errorText.value = baseError
    return
  }
  const birthError = validateBirthDate(draft.birth_date || null)
  if (birthError) {
    errorText.value = birthError
    return
  }
  const bioError = validateBio(draft.bio)
  if (bioError) {
    errorText.value = bioError
    return
  }
  saving.value = true
  errorText.value = ''
  try {
    await profileStore.updateProfile({
      nickname: draft.nickname,
      gender: draft.gender,
      province_code: draft.province_code,
      city_code: draft.city_code,
      birth_date: draft.birth_date || null,
      height_cm: toIntOrNull(draft.height_cm),
      education_level: draft.education_level,
      occupation_category: draft.occupation_category,
      employment_status: draft.employment_status,
      annual_income_band: draft.annual_income_band,
      family_asset_band: draft.family_asset_band > 0 ? draft.family_asset_band : null,
      debt_band: draft.debt_band,
      marital_status: draft.marital_status,
      children_count: draft.children_count,
      wants_children: draft.wants_children,
      marriage_intent: draft.marriage_intent,
      marriage_timeline: draft.marriage_timeline,
      smoking_status: draft.smoking_status,
      drinking_status: draft.drinking_status,
      house_status: draft.house_status,
      vehicle_status: draft.vehicle_status,
      bio: draft.bio
    })
    uni.showToast({ title: '资料已保存', icon: 'none' })
    setTimeout(() => {
      uni.navigateBack({ delta: 1 })
    }, 300)
  } catch (error) {
    errorText.value = errorMessage(error, '保存失败，请稍后重试')
  } finally {
    saving.value = false
  }
}

function back() {
  uni.navigateBack({ delta: 1 })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/profile/edit') })
}
</script>

<style lang="scss" scoped>
.photo-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-bottom: 16rpx;
}
.photo-item {
  position: relative;
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  overflow: hidden;
  background: #eef2f7;
}
.photo-image {
  width: 100%;
  height: 100%;
}
.photo-remove {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  text-align: center;
  font-size: 22rpx;
  color: #ffffff;
  background: rgba(0, 0, 0, 0.45);
  padding: 6rpx 0;
}
.photo-add {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  border: 1rpx dashed $uni-border-color;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafbfd;
}
.photo-add__text {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.save-btn {
  margin-top: 16rpx;
}
</style>
