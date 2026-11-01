<template>
  <PageShell>
    <view class="wmc-container">
      <view class="wmc-card">
        <MarketBadge market="real" />
        <text class="funnel-hint">真实候选池漏斗统计已注册、资料达标并主动加入婚恋池的用户，按你的硬条件顺序筛选。</text>
      </view>

      <view v-if="loading" class="wmc-card">
        <text class="wmc-hint">正在加载真实漏斗…</text>
      </view>
      <view v-else-if="loadError" class="wmc-card">
        <text class="wmc-error">{{ loadError }}</text>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="load">重新加载</button>
      </view>
      <template v-else-if="funnel">
        <view class="wmc-card">
          <view class="wmc-section-title">
            <text>关键指标</text>
            <text class="wmc-tag wmc-tag--muted">K 匿名</text>
          </view>
          <view class="wmc-stat-grid">
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ reciprocalText }}</text>
              <text class="wmc-stat-label">双向基础候选人</text>
            </view>
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ ratioText }}</text>
              <text class="wmc-stat-label">双向候选率</text>
            </view>
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ totalText }}</text>
              <text class="wmc-stat-label">当前入池用户</text>
            </view>
            <view class="wmc-stat">
              <text class="wmc-stat-value">{{ finalText }}</text>
              <text class="wmc-stat-label">单向符合硬条件</text>
            </view>
          </view>
        </view>

        <FunnelChart
          :steps="funnel.steps"
          market="real"
          title="真实候选池顺序漏斗"
          final-label="双向基础候选人"
          :final-text="reciprocalText"
          :note="funnel.note || '按当前条件顺序计算，改变条件顺序会改变每一步的损失展示'"
        />

        <view class="wmc-card wmc-card--flat">
          <text class="wmc-hint">
            为保护隐私，少于 {{ kThreshold }} 人的样本统一显示为 &lt;{{ kThreshold }}。双向候选率只在双向候选人数达到匿名阈值后展示。
          </text>
          <text v-if="funnel.note" class="wmc-hint note-gap">{{ funnel.note }}</text>
        </view>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 真实候选池漏斗：人数按 K 匿名处理，禁止出现「0 个人」。
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import FunnelChart from '@/components/FunnelChart.vue'
import { useAuthStore } from '@/stores/auth'
import { useDiscoveryStore } from '@/stores/discovery'
import { K_ANONYMITY_THRESHOLD } from '@/utils/constants'
import { formatPercent, maskCount } from '@/utils/format'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const discovery = useDiscoveryStore()
const loading = ref(true)
const loadError = ref('')
const kThreshold = K_ANONYMITY_THRESHOLD

const funnel = computed(() => discovery.funnel)

const reciprocalText = computed(() => (funnel.value ? maskCount(funnel.value.reciprocal_count) + ' 人' : '—'))
const finalText = computed(() => (funnel.value ? maskCount(funnel.value.final_count) + ' 人' : '—'))
const totalText = computed(() => (funnel.value ? maskCount(funnel.value.total) + ' 人' : '—'))
const ratioText = computed(() => {
  if (!funnel.value) return '—'
  if (funnel.value.reciprocal_count < K_ANONYMITY_THRESHOLD) return '人数不足 ' + K_ANONYMITY_THRESHOLD
  return formatPercent(funnel.value.reciprocal_ratio, 1)
})

onShow(async () => {
  auth.syncFromStorage()
  if (!auth.isLoggedIn) {
    uni.redirectTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/discovery/funnel') })
    return
  }
  await load()
})

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    await discovery.loadFunnel(true)
  } catch (error) {
    loadError.value = errorMessage(error, '真实漏斗加载失败')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.funnel-hint {
  display: block;
  margin-top: 20rpx;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  line-height: 1.7;
}
.retry-btn {
  margin-top: 20rpx;
}
.note-gap {
  margin-top: 12rpx;
}
</style>