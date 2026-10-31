<!-- 核心经营漏斗：访问计算器 → 完成计算器 → 注册 → 完成资料 → 加入婚恋池 → 出现 Eligible 候选 → Like → Match → 线下见面 → 建立关系。 -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getFunnel } from '@/api/analytics'
import type { FunnelStep } from '@/api/types'
import BaseChart from '@/components/BaseChart.vue'
import * as echarts from 'echarts'
import { formatNumber, formatRatio, toNumber } from '@/utils/format'

// 核心经营漏斗固定口径，顺序不可调整。
interface FunnelDefinition {
  key: string
  label: string
  description: string
}

const FUNNEL_DEFINITIONS: FunnelDefinition[] = [
  { key: 'calculator_visit', label: '访问计算器', description: '打开理论市场计算器页面' },
  { key: 'calculator_complete', label: '完成计算器', description: '提交条件并得到估算结果' },
  { key: 'register', label: '注册', description: '完成手机号注册' },
  { key: 'profile_complete', label: '完成资料', description: '资料完成度达到入池要求' },
  { key: 'market_join', label: '加入婚恋池', description: '主动同意并进入真实婚恋市场' },
  { key: 'eligible_candidate', label: '出现 Eligible 候选', description: '双向硬条件满足的候选被曝光' },
  { key: 'like', label: 'Like', description: '对候选发出喜欢' },
  { key: 'match', label: 'Match', description: '双向 Like 形成 Match' },
  { key: 'offline_meeting', label: '线下见面', description: '用户主动反馈已线下见面' },
  { key: 'relationship', label: '建立关系', description: '用户主动反馈已建立稳定关系' }
]

// 兼容后端可能使用的事件名，命中任一别名即归入对应步骤。
const KEY_ALIASES: Record<string, string[]> = {
  calculator_visit: ['calculator_visit', 'calculator_started', 'calculator_viewed', 'calculator_start', 'view_calculator'],
  calculator_complete: ['calculator_complete', 'calculator_completed'],
  register: ['register', 'registered', 'registration_completed', 'registration'],
  profile_complete: ['profile_complete', 'profile_completed'],
  market_join: ['market_join', 'market_joined', 'discovery_join'],
  eligible_candidate: ['eligible_candidate', 'discovery_viewed', 'candidate_eligible', 'eligible'],
  like: ['like', 'candidate_liked', 'likes'],
  match: ['match', 'match_created'],
  offline_meeting: ['offline_meeting', 'outcome_offline_meeting'],
  relationship: ['relationship', 'outcome_relationship']
}

interface FunnelRow extends FunnelDefinition {
  count: number
  hasData: boolean
  conversionPrev: number | null
  conversionTotal: number | null
}

const loading = ref(false)
const rawSteps = ref<FunnelStep[]>([])

async function load(): Promise<void> {
  loading.value = true
  try {
    rawSteps.value = await getFunnel()
  } catch {
    rawSteps.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

// 将接口返回的步骤按固定口径归位；缺失步骤标记为待接入。
const baseRows = computed<Array<FunnelDefinition & { count: number; hasData: boolean }>>(() => {
  const raw = rawSteps.value
  const used = new Set<number>()
  return FUNNEL_DEFINITIONS.map((definition, index) => {
    const aliases = KEY_ALIASES[definition.key] ?? [definition.key]
    let matchedIndex = raw.findIndex(
      (step, stepIndex) => !used.has(stepIndex) && typeof step.key === 'string' && aliases.includes(step.key)
    )
    if (matchedIndex < 0 && index < raw.length && !used.has(index)) {
      matchedIndex = index
    }
    if (matchedIndex < 0) {
      return { ...definition, count: 0, hasData: false }
    }
    used.add(matchedIndex)
    const step = raw[matchedIndex]
    return {
      key: definition.key,
      label: step?.label || definition.label,
      description: definition.description,
      count: toNumber(step?.count),
      hasData: step !== undefined && step.count !== undefined
    }
  })
})

// 计算每一步较上一步与相对首步的转化率。
const rows = computed<FunnelRow[]>(() => {
  const base = baseRows.value
  const first = base.length > 0 ? base[0].count : 0
  let previous: number | null = null
  return base.map((row) => {
    const conversionPrev = previous !== null && previous > 0 ? row.count / previous : null
    const conversionTotal = first > 0 ? row.count / first : null
    previous = row.hasData ? row.count : previous
    return { ...row, conversionPrev, conversionTotal }
  })
})

const availableRows = computed(() => rows.value.filter((row) => row.hasData))
const firstRow = computed(() => availableRows.value[0])
const lastRow = computed(() => availableRows.value[availableRows.value.length - 1])
const overallConversion = computed<number | null>(() => {
  const first = firstRow.value
  const last = lastRow.value
  if (!first || !last || first.count <= 0) {
    return null
  }
  return last.count / first.count
})
// 找出相对上一步流失最多的环节。
const biggestDrop = computed(() => {
  let result: { label: string; drop: number } | null = null
  for (const row of rows.value) {
    if (!row.hasData || row.conversionPrev === null) {
      continue
    }
    const drop = 1 - row.conversionPrev
    if (!result || drop > result.drop) {
      result = { label: row.label, drop }
    }
  }
  return result
})

const chartRows = computed(() => [...availableRows.value].reverse())
const chartOption = computed<echarts.EChartsOption>(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: 12, right: 40, top: 16, bottom: 12, containLabel: true },
  xAxis: { type: 'value', minInterval: 1 },
  yAxis: { type: 'category', data: chartRows.value.map((row) => row.label) },
  series: [
    {
      name: '人数',
      type: 'bar',
      barMaxWidth: 20,
      itemStyle: { color: '#409eff', borderRadius: [0, 4, 4, 0] },
      label: { show: true, position: 'right' },
      data: chartRows.value.map((row) => row.count)
    }
  ]
}))
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">核心经营漏斗</h1>
        <p class="page-desc">
          从计算器访问到建立关系的完整经营闭环，每一步展示人数与转化率；数据来自本站产品事件，与理论市场估算严格区分。
        </p>
      </div>
      <el-button size="small" @click="load">刷新</el-button>
    </div>

    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="mb-12"
      title="口径说明"
      description="各步骤按触发过对应产品事件的用户去重统计；小样本（小于 K 匿名阈值）不展示精确人数，统一以 <10 占位；理论市场数据不参与本页计算。"
    />

    <div class="stat-grid">
      <div class="page-card summary-card">
        <p class="summary-label">漏斗首步人数</p>
        <p class="summary-value">{{ firstRow ? formatNumber(firstRow.count) : '—' }}</p>
        <p class="muted">{{ firstRow?.label ?? '暂无数据' }}</p>
      </div>
      <div class="page-card summary-card">
        <p class="summary-label">漏斗末步人数</p>
        <p class="summary-value">{{ lastRow ? formatNumber(lastRow.count) : '—' }}</p>
        <p class="muted">{{ lastRow?.label ?? '暂无数据' }}</p>
      </div>
      <div class="page-card summary-card">
        <p class="summary-label">总体转化率</p>
        <p class="summary-value">{{ formatRatio(overallConversion) }}</p>
        <p class="muted">末步人数 ÷ 首步人数</p>
      </div>
      <div class="page-card summary-card">
        <p class="summary-label">流失最大的环节</p>
        <p class="summary-value">{{ biggestDrop ? formatRatio(biggestDrop.drop) : '—' }}</p>
        <p class="muted">{{ biggestDrop?.label ?? '暂无数据' }}</p>
      </div>
    </div>

    <div class="page-card">
      <h2 class="card-title">漏斗明细</h2>
      <el-table :data="rows" border size="default">
        <el-table-column type="index" label="#" width="52" align="center" />
        <el-table-column label="步骤" min-width="220">
          <template #default="{ row }">
            <div class="step-label">{{ row.label }}</div>
            <div class="muted">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column label="人数" width="120" align="right">
          <template #default="{ row }">
            <span v-if="row.hasData">{{ formatNumber(row.count) }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="较上一步" width="120" align="right">
          <template #default="{ row }">
            <span v-if="row.hasData && row.conversionPrev !== null">{{ formatRatio(row.conversionPrev) }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="相对首步" width="120" align="right">
          <template #default="{ row }">
            <span v-if="row.hasData && row.conversionTotal !== null">{{ formatRatio(row.conversionTotal) }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="流失率" width="120" align="right">
          <template #default="{ row }">
            <span v-if="row.hasData && row.conversionPrev !== null">{{ formatRatio(1 - row.conversionPrev) }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.hasData" type="success" size="small">已接入</el-tag>
            <el-tag v-else type="info" size="small">待接入</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="page-card">
      <h2 class="card-title">漏斗人数分布</h2>
      <BaseChart v-if="chartRows.length > 0" :option="chartOption" height="360px" />
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

.summary-card {
  padding: 16px 18px;
}

.summary-label {
  margin: 0;
  color: var(--wmc-text-secondary);
  font-size: 13px;
}

.summary-value {
  margin: 8px 0 4px;
  font-size: 26px;
  font-weight: 600;
  color: var(--wmc-text-primary);
}

.step-label {
  font-weight: 500;
  color: var(--wmc-text-primary);
}
</style>
