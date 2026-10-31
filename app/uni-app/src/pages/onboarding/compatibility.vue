<template>
  <PageShell>
    <view class="wmc-container">
      <StepHeader
        title="婚恋价值观"
        subtitle="价值观问题用于计算关系观念兼容度。未填写的问题不计入分母。"
        :percent="completion"
        :show-back="true"
      />
      <view v-if="!ready" class="wmc-card">
        <text class="wmc-hint">正在加载价值观问题…</text>
      </view>
      <template v-else>
        <view class="wmc-card">
          <view class="wmc-row">
            <text class="wmc-value">有效价值观题：{{ validCount }} 道</text>
            <text class="wmc-tag wmc-tag--muted">至少 6 道</text>
          </view>
          <text class="wmc-hint">
            双方都填写的问题才会计入关系观念兼容度。有效题目不足 6 道时，页面会显示「关系观念数据不足」，不输出精确分数。
          </text>
          <view v-if="validCount < 6" class="wmc-notice wmc-notice--warning compat-notice">
            当前有效题目不足 6 道，关系观念兼容度将显示「关系观念数据不足」。
          </view>
        </view>

        <view v-for="(item, index) in ruleDrafts" :key="item.question_id" class="wmc-card">
          <view class="question-head">
            <text class="question-index">第 {{ index + 1 }} 题</text>
            <text class="wmc-tag">{{ item.category || '价值观' }}</text>
          </view>
          <text class="question-title">{{ item.title }}</text>
          <text v-if="item.description" class="question-desc">{{ item.description }}</text>
          <view class="wmc-divider"></view>
          <view class="wmc-field">
            <text class="wmc-field-label">我的答案</text>
            <OptionChips
              :model-value="item.own_answer"
              :options="item.options"
              :single="item.question_type === 1"
              @update:model-value="onOwnAnswer(index, $event)"
            />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">我可以接受对方选择</text>
            <OptionChips :model-value="item.accepted_answers" :options="item.options" @update:model-value="onAcceptedAnswers(index, $event)" />
          </view>
          <view class="wmc-field">
            <text class="wmc-field-label">在意程度</text>
            <view class="importance-row">
              <text
                v-for="level in importanceOptions"
                :key="level.value"
                class="importance-item"
                :class="{ 'importance-item--active': item.importance === level.value }"
                @tap="setImportance(index, level.value)"
              >
                {{ level.label }}
              </text>
            </view>
          </view>
          <HardToggle v-model="item.is_hard" label="价值观硬条件" hint="对方答案不在可接受范围时直接不进入候选池" />
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
// 资料完善第 6 步：婚恋价值观问题与可接受答案，保存到 /me/compatibility。
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import StepHeader from '@/components/StepHeader.vue'
import OptionChips from '@/components/OptionChips.vue'
import HardToggle from '@/components/HardToggle.vue'
import { getCompatibilityQuestions } from '@/api/meta'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { ensureOptions } from '@/utils/options'
import { errorMessage } from '@/utils/error'
import type { CompatibilityQuestion, CompatibilityRuleInput } from '@/api/types'

interface RuleDraft {
  question_id: string
  category: string
  title: string
  description: string
  question_type: number
  options: Array<{ value: string; label: string }>
  own_answer: string[]
  accepted_answers: string[]
  importance: number
  is_hard: boolean
}

const auth = useAuthStore()
const profile = useProfileStore()
const ready = ref(false)
const saving = ref(false)
const errorText = ref('')
const ruleDrafts = ref<RuleDraft[]>([])

const importanceOptions = [
  { value: 0, label: '不在意' },
  { value: 1, label: '略重要' },
  { value: 2, label: '一般' },
  { value: 3, label: '比较重要' },
  { value: 4, label: '重要' },
  { value: 5, label: '非常重要' }
]

const completion = computed(() => profile.completion)
const validCount = computed(() => ruleDrafts.value.filter((item) => item.own_answer.length > 0).length)

onShow(async () => {
  if (!auth.requireLogin('/pages/onboarding/compatibility')) return
  if (ready.value) return
  await ensureOptions()
  try {
    const questions = await getCompatibilityQuestions()
    const rules = await profile.loadRules(true)
    const sorted = (questions || []).slice().sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
    ruleDrafts.value = sorted.map((question) => buildDraft(question, rules))
  } catch (error) {
    errorText.value = errorMessage(error, '价值观问题加载失败，请稍后重试')
  }
  ready.value = true
})

function buildDraft(question: CompatibilityQuestion, rules: Array<{ question_id: string; own_answer: string[]; accepted_answers: string[]; importance: number; is_hard: boolean }>): RuleDraft {
  const existing = rules.find((item) => item.question_id === question.id)
  return {
    question_id: question.id,
    category: question.category || '',
    title: question.title,
    description: question.description || '',
    question_type: question.question_type || 1,
    options: question.options || [],
    own_answer: existing && existing.own_answer ? existing.own_answer.slice() : [],
    accepted_answers: existing && existing.accepted_answers ? existing.accepted_answers.slice() : [],
    importance: existing && typeof existing.importance === 'number' ? existing.importance : 3,
    is_hard: Boolean(existing && existing.is_hard)
  }
}

function onOwnAnswer(index: number, value: Array<number | string>) {
  const item = ruleDrafts.value[index]
  if (!item) return
  item.own_answer = value.map((entry) => String(entry))
  if (item.accepted_answers.length === 0) {
    item.accepted_answers = item.own_answer.slice()
  }
}

function onAcceptedAnswers(index: number, value: Array<number | string>) {
  const item = ruleDrafts.value[index]
  if (!item) return
  item.accepted_answers = value.map((entry) => String(entry))
}

function setImportance(index: number, value: number) {
  const item = ruleDrafts.value[index]
  if (item) item.importance = value
}

function buildPayload(): CompatibilityRuleInput[] {
  return ruleDrafts.value
    .filter((item) => item.own_answer.length > 0)
    .map((item) => ({
      question_id: item.question_id,
      own_answer: item.own_answer,
      accepted_answers: item.accepted_answers.length > 0 ? item.accepted_answers : item.own_answer.slice(),
      importance: item.importance,
      is_hard: item.is_hard
    }))
}

async function persist(): Promise<boolean> {
  saving.value = true
  errorText.value = ''
  try {
    await profile.saveRules(buildPayload())
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
  if (await persist()) uni.reLaunch({ url: '/pages/discovery/index' })
}

async function saveExit() {
  if (saving.value) return
  if (await persist()) uni.reLaunch({ url: '/pages/profile/index' })
}
</script>

<style lang="scss" scoped>
.compat-notice {
  margin-top: 16rpx;
}
.question-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}
.question-index {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}
.question-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: $uni-text-color;
  line-height: 1.6;
}
.question-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.importance-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.importance-item {
  padding: 10rpx 20rpx;
  border-radius: 8rpx;
  border: 1rpx solid $uni-border-color;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  background: #ffffff;
}
.importance-item--active {
  border-color: $uni-color-primary;
  background: rgba(47, 107, 255, 0.08);
  color: $uni-color-primary;
}
.save-btn {
  margin-top: 16rpx;
}
</style>
