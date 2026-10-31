<!-- 运营看板：核心经营指标、North Star、性别/城市/年龄分布与漏斗预览。 -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard } from '@/api/analytics'
import type { AdminDashboard, FunnelStep } from '@/api/types'
import BaseChart from '@/components/BaseChart.vue'
import StatCard from '@/components/StatCard.vue'
import * as echarts from 'echarts'
import { formatDecimal, formatNumber, formatRatio, formatRatioOf, toNumber } from '@/utils/format'
import { GENDER_LABELS, labelOf } from '@/utils/enums'

const router = useRouter()
const loading = ref(false)
const loadedAt = ref('')
const dashboard = ref<AdminDashboard>({})

async function loadDashboard(): Promise<void> {
  loading.value = true
  try {
    dashboard.value = await getDashboard()
    loadedAt.value = new Date().toLocaleString('zh-CN', { hour12: false })
  } catch {
    dashboard.value = {}
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)

// 性别分布饼图数据。
const genderEntries = computed(() => Object.entries(dashboard.value.gender_distribution ?? {}))
const hasGenderData = computed(() => genderEntries.value.length > 0)
const genderChartOption = computed<echarts.EChartsOption>(() => {
  const data = genderEntries.value.map(([code, count]) => ({
    name: labelOf(GENDER_LABELS, Number(code), '未填写'),
    value: toNumber(count)
  }))
  return {
    tooltip: { trigger: 'item', formatter: '{b}：{c}（{d}%）' },
    legend: { bottom: 0 },
    color: ['#409eff', '#f56c6c', '#909399', '#e6a23c'],
    series: [
      {
        name: '性别分布',
        type: 'pie',
        radius: ['40%', '66%'],
        center: ['50%', '44%'],
        label: { formatter: '{b}\n{d}%' },
        data
      }
    ]
  }
})

// 城市分布取前 10 名。
const cityRows = computed(() => (dashboard.value.city_distribution ?? []).slice(0, 10))
const hasCityData = computed(() => cityRows.value.length > 0)
const cityChartOption = computed<echarts.EChartsOption>(() => {
  const rows = [...cityRows.value].reverse()
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 10, right: 28, top: 16, bottom: 8, containLabel: true },
    xAxis: { type: 'value', minInterval: 1 },
    yAxis: { type: 'category', data: rows.map((row) => row.city_name ?? '未知城市') },
    series: [
      {
        name: '用户数',
        type: 'bar',
        barMaxWidth: 18,
        itemStyle: { color: '#409eff', borderRadius: [0, 4, 4, 0] },
        data: rows.map((row) => toNumber(row.count))
      }
    ]
  }
})

// 年龄分布柱状图。
const ageRows = computed(() => dashboard.value.age_distribution ?? [])
const hasAgeData = computed(() => ageRows.value.length > 0)
const ageChartOption = computed<echarts.EChartsOption>(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: 10, right: 16, top: 28, bottom: 8, containLabel: true },
  xAxis: { type: 'category', data: ageRows.value.map((row) => row.bucket ?? '未知') },
  yAxis: { type: 'value', minInterval: 1 },
  series: [
    {
      name: '用户数',
      type: 'bar',
      barMaxWidth: 30,
      itemStyle: { color: '#67c23a', borderRadius: [4, 4, 0, 0] },
      label: { show: true, position: 'top' },
      data: ageRows.value.map((row) => toNumber(row.count))
    }
  ]
}))

const funnelSteps = computed<FunnelStep[]>(() => dashboard.value.funnel ?? [])

function goFunnel(): void {
  void router.push('/funnel')
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">数据看板</h1>
        <p class="page-desc">
          本站真实运营数据；理论市场数据来自基准数据集统计估算，两者严格区分，禁止对外宣称精确人数或双向匹配率。
        </p>
      </div>
      <div class="page-head-actions">
        <span class="muted">更新时间：{{ loadedAt || '—' }}</span>
        <el-button size="small" @click="loadDashboard">刷新</el-button>
      </div>
    </div>

    <el-alert
      type="warning"
      :closable="false"
      show-icon
      class="mb-12"
      title="理论与真实必须区分"
      description="本页指标均来自本站真实用户行为与真实婚恋池；理论市场为统计估算，不得与本站数据混用，也不得对外宣称精确人数或双向匹配率。"
    />

    <div class="page-card">
      <h2 class="card-title">用户规模与活跃</h2>
      <div class="stat-grid">
        <StatCard label="总注册人数" :value="formatNumber(dashboard.total_users)" hint="累计注册，不含已注销" tone="primary" />
        <StatCard label="近 7 天新增" :value="formatNumber(dashboard.new_users_7d)" hint="近 7 天完成注册" />
        <StatCard label="DAU" :value="formatNumber(dashboard.dau)" hint="今日活跃用户" />
        <StatCard label="WAU" :value="formatNumber(dashboard.wau)" hint="近 7 天活跃用户" />
        <StatCard label="MAU" :value="formatNumber(dashboard.mau)" hint="近 30 天活跃用户" />
        <StatCard
          label="资料完成率"
          :value="formatRatioOf(dashboard.profile_completed_users, dashboard.total_users)"
          :hint="`已完成资料 ${formatNumber(dashboard.profile_completed_users)} 人`"
          tone="success"
        />
        <StatCard label="入池人数" :value="formatNumber(dashboard.market_users)" hint="已加入真实婚恋池" tone="success" />
      </div>
    </div>

    <div class="page-card">
      <h2 class="card-title">匹配与线下结果</h2>
      <div class="stat-grid">
        <StatCard label="Like 数量" :value="formatNumber(dashboard.interactions_like)" hint="用户发出的 Like 总数" />
        <StatCard label="Match 数量" :value="formatNumber(dashboard.matches_total)" hint="双向 Like 形成的 Match" />
        <StatCard label="Match Rate" :value="formatRatio(dashboard.match_rate)" hint="Match / Like，口径由后端统计" />
        <StatCard label="线下见面报告数" :value="formatNumber(dashboard.offline_meeting_reports)" hint="用户主动反馈已线下见面" />
        <StatCard label="双方确认见面数" :value="formatNumber(dashboard.confirmed_meetings)" hint="双方均反馈线下见面" tone="success" />
        <StatCard label="建立关系数" :value="formatNumber(dashboard.relationship_reports)" hint="反馈已建立稳定关系" tone="success" />
        <StatCard
          label="North Star"
          :value="formatDecimal(toNumber(dashboard.north_star), 2, '0.00')"
          hint="每 1000 MAU 的双方确认线下见面数"
          tone="primary"
        />
      </div>
      <p class="muted mt-12">
        North Star 口径：双方确认线下见面数 ÷ MAU × 1000；是判断产品是否真正创造婚恋价值的关键指标，不读取任何聊天内容。
      </p>
    </div>

    <div class="chart-grid">
      <div class="page-card">
        <h2 class="card-title">性别分布</h2>
        <BaseChart v-if="hasGenderData" :option="genderChartOption" height="300px" />
        <el-empty v-else description="暂无性别分布数据" :image-size="60" />
      </div>
      <div class="page-card">
        <h2 class="card-title">城市分布（Top 10）</h2>
        <BaseChart v-if="hasCityData" :option="cityChartOption" height="300px" />
        <el-empty v-else description="暂无城市分布数据" :image-size="60" />
      </div>
      <div class="page-card">
        <h2 class="card-title">年龄分布</h2>
        <BaseChart v-if="hasAgeData" :option="ageChartOption" height="300px" />
        <el-empty v-else description="暂无年龄分布数据" :image-size="60" />
      </div>
    </div>

    <div class="page-card">
      <div class="card-title-row">
        <h2 class="card-title">核心经营漏斗（预览）</h2>
        <el-button link type="primary" @click="goFunnel">查看完整漏斗</el-button>
      </div>
      <el-table v-if="funnelSteps.length > 0" :data="funnelSteps" size="small" border>
        <el-table-column prop="key" label="步骤标识" min-width="160" />
        <el-table-column prop="label" label="步骤" min-width="160" />
        <el-table-column label="人数" min-width="110" align="right">
          <template #default="{ row }">{{ formatNumber(row.count) }}</template>
        </el-table-column>
        <el-table-column label="转化率" min-width="110" align="right">
          <template #default="{ row }">{{ formatRatio(row.conversion) }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无漏斗数据，等待后端统计任务产出" :image-size="60" />
    </div>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-head-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: none;
}
</style>
