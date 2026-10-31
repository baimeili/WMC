<template>
  <view class="candidate" @tap="handleTap">
    <view class="candidate__head">
      <image v-if="candidate.primary_photo_url" class="candidate__avatar" :src="candidate.primary_photo_url" mode="aspectFill" />
      <view v-else class="candidate__avatar candidate__avatar--placeholder">
        <text class="candidate__avatar-text">{{ avatarText }}</text>
      </view>
      <view class="candidate__info">
        <view class="candidate__name-row">
          <text class="candidate__name">{{ candidate.nickname || '未填写昵称' }}</text>
          <text class="candidate__age">{{ ageText(candidate.age) }}</text>
        </view>
        <text class="candidate__meta">{{ candidate.city_name || '城市未填写' }} · {{ educationLabel }}</text>
        <text class="candidate__meta">{{ heightText(candidate.height_cm) }} · {{ maritalLabel }}</text>
      </view>
      <text v-if="candidate.my_action === 1" class="wmc-tag">已喜欢</text>
      <text v-else-if="candidate.my_action === 2" class="wmc-tag wmc-tag--muted">已跳过</text>
    </view>
    <view class="candidate__scores">
      <view class="candidate__score">
        <text class="candidate__score-value">{{ formatScore(candidate.preference_score) }}</text>
        <text class="candidate__score-label">条件契合度</text>
      </view>
      <view class="candidate__score">
        <text v-if="candidate.compatibility_insufficient" class="candidate__score-value candidate__score-value--muted">数据不足</text>
        <text v-else class="candidate__score-value">{{ formatScore(candidate.compatibility_score) }}</text>
        <text class="candidate__score-label">关系观念兼容度</text>
      </view>
      <view class="candidate__score">
        <text class="candidate__score-value candidate__score-value--primary">{{ formatScore(candidate.overall_score) }}</text>
        <text class="candidate__score-label">综合匹配度</text>
      </view>
    </view>
    <view class="candidate__footer">
      <text class="candidate__disclaimer">仅根据双方填写条件计算，不代表婚姻成功概率</text>
      <text class="candidate__more">查看详情</text>
    </view>
  </view>
</template>

<script setup lang="ts">
// 候选人卡片：仅展示本站真实市场候选，分数名称固定为三类兼容度指标。
import { computed } from 'vue'
import type { CandidateCardItem } from '@/api/types'
import { optionLabel } from '@/utils/options'
import { ageText, formatScore, heightText } from '@/utils/format'

const props = defineProps<{ candidate: CandidateCardItem }>()
const emit = defineEmits<{ (event: 'tap', candidate: CandidateCardItem): void }>()

const avatarText = computed(() => {
  const name = props.candidate.nickname || '用'
  return name.slice(0, 1)
})
const educationLabel = computed(() => optionLabel('education_levels', props.candidate.education_level, '学历未填写'))
const maritalLabel = computed(() => optionLabel('marital_statuses', props.candidate.marital_status, '婚姻状态未填写'))

function handleTap() {
  emit('tap', props.candidate)
}
</script>

<style lang="scss" scoped>
.candidate {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-shadow: $wmc-shadow;
}
.candidate__head {
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.candidate__avatar {
  width: 100rpx;
  height: 100rpx;
  border-radius: 16rpx;
  background: #eef2f7;
}
.candidate__avatar--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(47, 107, 255, 0.1);
}
.candidate__avatar-text {
  font-size: 40rpx;
  color: $uni-color-primary;
  font-weight: 600;
}
.candidate__info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
  min-width: 0;
}
.candidate__name-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}
.candidate__name {
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
  max-width: 320rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.candidate__age {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.candidate__meta {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.candidate__scores {
  display: flex;
  margin-top: 20rpx;
  background: #f8fafc;
  border-radius: 12rpx;
  padding: 16rpx 8rpx;
}
.candidate__score {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4rpx;
}
.candidate__score-value {
  font-size: 30rpx;
  font-weight: 600;
  color: $uni-text-color;
}
.candidate__score-value--primary {
  color: $uni-color-primary;
}
.candidate__score-value--muted {
  font-size: 24rpx;
  color: $uni-text-color-placeholder;
}
.candidate__score-label {
  font-size: 22rpx;
  color: $uni-text-color-grey;
}
.candidate__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16rpx;
}
.candidate__disclaimer {
  font-size: 20rpx;
  color: $uni-text-color-placeholder;
  flex: 1;
}
.candidate__more {
  font-size: 24rpx;
  color: $uni-color-primary;
  margin-left: 16rpx;
}
</style>
