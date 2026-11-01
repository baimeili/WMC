<template>
  <view class="candidate" @tap.stop="handleTap">
    <view class="candidate__head">
      <image
        v-if="candidate.primary_photo_url"
        class="candidate__avatar"
        :src="candidate.primary_photo_url"
        mode="aspectFill"
      />
      <view v-else class="candidate__avatar candidate__avatar--placeholder">
        <text class="candidate__avatar-text">{{ avatarText }}</text>
      </view>
      <view class="candidate__info">
        <view class="candidate__name-row">
          <text class="candidate__name">{{ candidate.nickname || '未填写昵称' }}</text>
          <text class="candidate__age">{{ ageText(candidate.age) }}</text>
          <text v-if="candidate.my_action === 1" class="wmc-tag">已喜欢</text>
          <text v-else-if="candidate.my_action === 2" class="wmc-tag wmc-tag--muted">已跳过</text>
        </view>
        <view class="candidate__meta-row">
          <text class="candidate__meta">{{ candidate.city_name || '城市未填写' }}</text>
          <text class="candidate__meta-dot"></text>
          <text class="candidate__meta">{{ educationLabel }}</text>
          <text class="candidate__meta-dot"></text>
          <text class="candidate__meta">{{ heightText(candidate.height_cm) }}</text>
        </view>
        <text class="candidate__bio">{{ candidate.bio || '还没有填写简介' }}</text>
      </view>
      <text class="candidate__arrow">›</text>
    </view>

    <view class="candidate__scores">
      <view class="candidate__score">
        <text class="candidate__score-value wmc-num">{{ formatScore(candidate.preference_score) }}</text>
        <text class="candidate__score-label">条件契合度</text>
      </view>
      <view class="candidate__score">
        <text v-if="candidate.compatibility_insufficient" class="candidate__score-value candidate__score-value--muted">数据不足</text>
        <text v-else class="candidate__score-value wmc-num">{{ formatScore(candidate.compatibility_score) }}</text>
        <text class="candidate__score-label">关系观念兼容度</text>
      </view>
      <view class="candidate__score candidate__score--primary">
        <text class="candidate__score-value candidate__score-value--primary wmc-num">{{ formatScore(candidate.overall_score) }}</text>
        <text class="candidate__score-label candidate__score-label--primary">综合匹配度</text>
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
const emit = defineEmits<{ (event: 'select', candidate: CandidateCardItem): void }>()

const avatarText = computed(() => {
  const name = props.candidate.nickname || '用'
  return name.slice(0, 1)
})
const educationLabel = computed(() => optionLabel('education_levels', props.candidate.education_level, '学历未填写'))

function handleTap() {
  emit('select', props.candidate)
}
</script>

<style lang="scss" scoped>
.candidate {
  background: #ffffff;
  border-radius: 28rpx;
  padding: 30rpx;
  margin-bottom: 22rpx;
  box-shadow: 0 12rpx 40rpx rgba(15, 23, 42, 0.08);
  border: 1rpx solid rgba(15, 23, 42, 0.04);
}
.candidate__head {
  display: flex;
  align-items: flex-start;
}
.candidate__avatar {
  width: 104rpx;
  height: 104rpx;
  border-radius: 26rpx;
  flex-shrink: 0;
  background: #eef2f7;
}
.candidate__avatar--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #dbe4ff 0%, #eef2ff 100%);
}
.candidate__avatar-text {
  font-size: 40rpx;
  font-weight: 700;
  color: #3355ff;
}
.candidate__info {
  flex: 1;
  padding-left: 22rpx;
  min-width: 0;
}
.candidate__name-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
.candidate__name {
  font-size: 32rpx;
  font-weight: 700;
  color: #0b1220;
  margin-right: 14rpx;
}
.candidate__age {
  font-size: 24rpx;
  color: #64748b;
  background: #f1f5f9;
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  margin-right: 12rpx;
}
.candidate__meta-row {
  display: flex;
  align-items: center;
  margin-top: 12rpx;
  flex-wrap: wrap;
}
.candidate__meta {
  font-size: 24rpx;
  color: #64748b;
}
.candidate__meta-dot {
  width: 6rpx;
  height: 6rpx;
  border-radius: 50%;
  background: #cbd5e1;
  margin: 0 12rpx;
}
.candidate__bio {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #94a3b8;
  line-height: 1.6;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.candidate__arrow {
  font-size: 40rpx;
  color: #cbd5e1;
  line-height: 1;
  margin-left: 8rpx;
}
.candidate__scores {
  display: flex;
  gap: 16rpx;
  margin-top: 26rpx;
}
.candidate__score {
  flex: 1;
  background: #f8fafc;
  border-radius: 20rpx;
  padding: 20rpx 18rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  border: 1rpx solid rgba(15, 23, 42, 0.04);
}
.candidate__score--primary {
  background: linear-gradient(150deg, #eef2ff 0%, #e0e7ff 100%);
  border-color: rgba(51, 85, 255, 0.16);
}
.candidate__score-value {
  font-size: 34rpx;
  font-weight: 700;
  color: #334155;
}
.candidate__score-value--primary {
  color: #3355ff;
}
.candidate__score-value--muted {
  font-size: 24rpx;
  color: #94a3b8;
  font-weight: 500;
}
.candidate__score-label {
  margin-top: 6rpx;
  font-size: 20rpx;
  color: #94a3b8;
}
.candidate__score-label--primary {
  color: #4f67d8;
}
.candidate__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 22rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid rgba(15, 23, 42, 0.06);
}
.candidate__disclaimer {
  font-size: 20rpx;
  color: #94a3b8;
  flex: 1;
  padding-right: 16rpx;
}
.candidate__more {
  font-size: 24rpx;
  font-weight: 600;
  color: #3355ff;
}
</style>