<template>
  <PageShell>
    <view class="wmc-container">
      <view class="wmc-card">
        <MarketBadge market="real" />
        <text class="sens-hint">逐条取消你的硬条件，观察双向候选人数的变化，用于判断哪个条件对候选池影响最大。</text>
      </view>

      <view v-if="loading" class="wmc-card">
        <text class="wmc-hint">正在计算条件敏感性…</text>
      </view>
      <view v-else-if="loadError" class="wmc-card">
        <text class="wmc-error">{{ loadError }}</text>
        <button class="wmc-btn wmc-btn--ghost wmc-btn--block retry-btn" @tap="load">重新加载</button>
      </view>
      <template v-else-if="sensitivity">
        <SensitivityList :items="sensitivity.items" market="real" title="真实候选池条件敏感性（Leave-One-Out）" />
        <view class="wmc-card wmc-card--flat">
          <text class="wmc-hint">
            为保护隐私，少于 {{ kThreshold }} 人的样本统一显示为 &lt;{{ kThreshold }}，并隐藏变化率。取消某个条件后人数反而下降，说明该条件同时在限制对方对你的选择。
          </text>
          <text v-if="sensitivity.note" class="wmc-hint note-gap">{{ sensitivity.note }}</text>
        </view>
      </template>
    </view>
  </PageShell>
</template>

<script setup lang="ts">
// 真实候选池条件敏感性：人数 K 匿名，低于阈值不展示变化率。
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import PageShell from '@/components/PageShell.vue'
import MarketBadge from '@/components/MarketBadge.vue'
import SensitivityList from '@/components/SensitivityList.vue'
import { useAuthStore } from '@/stores/auth'
import { useDiscoveryStore } from '@/stores/discovery'
import { K_ANONYMITY_THRESHOLD } from '@/utils/constants'
import { errorMessage } from '@/utils/error'

const auth = useAuthStore()
const discovery = useDiscoveryStore()
const loading = ref(true)
const loadError = ref('')
const kThreshold = K_ANONYMITY_THRESHOLD
const sensitivity = computed(() => discovery.sensitivity)

onShow(async () => {
  auth.syncFromStorage()
  if (!auth.isLoggedIn) {
    uni.redirectTo({ url: '/pages/auth/login?redirect=' + encodeURIComponent('/pages/discovery/sensitivity') })
    return
  }
  await load()
})

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    await discovery.loadSensitivity(true)
  } catch (error) {
    loadError.value = errorMessage(error, '条件敏感性加载失败')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.sens-hint {
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
