<!-- 系统设置：Feature Flag 开关与数值参数。保存后立即从服务端刷新，确认落库结果。 -->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSettings, updateSetting } from '@/api/settings'
import type { SettingsBundle } from '@/api/types'
import { toNumber } from '@/utils/format'

interface FeatureDefinition {
  key: string
  label: string
  description: string
}

interface NumberParamDefinition {
  key: string
  label: string
  description: string
  min: number
  max: number
  step: number
}

interface TextParamDefinition {
  key: string
  label: string
  description: string
}

interface BooleanParamDefinition {
  key: string
  label: string
  description: string
}

// Feature Flag：顺序按业务链路排列。
const FEATURE_DEFINITIONS: FeatureDefinition[] = [
  { key: 'calculator_enabled', label: '计算器', description: '游客理论市场估算入口' },
  { key: 'registration_enabled', label: '注册', description: '手机号注册开关' },
  { key: 'discovery_enabled', label: '真实婚恋池', description: '真实候选池总开关' },
  { key: 'likes_enabled', label: 'Like / Pass', description: '候选互动开关' },
  { key: 'matches_enabled', label: 'Match', description: '双向 Like 生成 Match' },
  { key: 'chat_enabled', label: '聊天', description: '第一版暂不开放' },
  { key: 'payments_enabled', label: '支付', description: '第一版暂不开放' },
  { key: 'verification_enabled', label: '认证', description: '第一版仅实现手机号认证' }
]

// 数值参数。
const NUMBER_PARAM_DEFINITIONS: NumberParamDefinition[] = [
  { key: 'k_anonymity_threshold', label: 'K 匿名阈值', description: '小于该值统一显示 <10，禁止暴露精确人数', min: 1, max: 100, step: 1 },
  { key: 'min_profile_completion', label: '入池最低资料完成度', description: '进入真实婚恋池所需完成度（%）', min: 0, max: 100, step: 5 },
  { key: 'preference_score_weight', label: '条件契合度权重', description: '综合匹配度中条件契合度占比', min: 0, max: 1, step: 0.05 },
  { key: 'compatibility_score_weight', label: '关系观念兼容度权重', description: '综合匹配度中关系观念兼容度占比', min: 0, max: 1, step: 0.05 },
  { key: 'min_valid_compatibility_questions', label: '有效价值观题最少数量', description: '不足该数量时不输出精确兼容度分数', min: 0, max: 20, step: 1 },
  { key: 'candidate_page_size', label: '候选列表默认页大小', description: '真实候选池默认每页数量', min: 1, max: 50, step: 1 },
  { key: 'max_candidate_page_size', label: '候选列表最大页大小', description: '真实候选池允许的最大每页数量', min: 1, max: 100, step: 1 },
  { key: 'sms_resend_seconds', label: '短信重发间隔（秒）', description: '同一手机号两次发送验证码的最小间隔', min: 30, max: 600, step: 10 },
  { key: 'sms_daily_limit', label: '短信每日上限', description: '同一手机号每日最多发送次数', min: 1, max: 50, step: 1 },
  { key: 'account_deletion_retention_days', label: '注销数据保留期（天）', description: '注销后保留期结束再删除或匿名化', min: 0, max: 365, step: 1 }
]

// 布尔参数。
const BOOLEAN_PARAM_DEFINITIONS: BooleanParamDefinition[] = [
  { key: 'demo_dataset_enabled', label: '允许 DEMO 数据集', description: '仅开发环境可开启；生产必须关闭，避免演示数据冒充真实统计' }
]

// 协议版本号（字符串参数）。
const TEXT_PARAM_DEFINITIONS: TextParamDefinition[] = [
  { key: 'policy_version_terms', label: '用户协议版本', description: '用户同意记录中保存的版本号' },
  { key: 'policy_version_privacy', label: '隐私政策版本', description: '隐私政策版本号' },
  { key: 'policy_version_market', label: '婚恋市场参与声明版本', description: '加入真实婚恋池时同意的声明版本' }
]

const loading = ref(false)
const saving = ref(false)
const features = reactive<Record<string, boolean>>({})
const numbers = reactive<Record<string, number>>({})
const booleans = reactive<Record<string, boolean>>({})
const texts = reactive<Record<string, string>>({})
const savedSnapshot = ref('')

function snapshot(): string {
  return JSON.stringify({
    features: { ...features },
    numbers: { ...numbers },
    booleans: { ...booleans },
    texts: { ...texts }
  })
}

const dirty = computed(() => snapshot() !== savedSnapshot.value)
const weightSum = computed(() => toNumber(numbers.preference_score_weight) + toNumber(numbers.compatibility_score_weight))
const weightValid = computed(() => Math.abs(weightSum.value - 1) < 0.001)
const kThresholdWarning = computed(() => toNumber(numbers.k_anonymity_threshold) < 10)

// 将接口返回值同步到本地表单。
function applyBundle(bundle: SettingsBundle): void {
  const featureMap = bundle.features ?? {}
  for (const definition of FEATURE_DEFINITIONS) {
    features[definition.key] = Boolean(featureMap[definition.key])
  }
  const settingMap = bundle.settings ?? {}
  for (const definition of NUMBER_PARAM_DEFINITIONS) {
    numbers[definition.key] = toNumber(settingMap[definition.key], definition.min)
  }
  for (const definition of BOOLEAN_PARAM_DEFINITIONS) {
    booleans[definition.key] = Boolean(settingMap[definition.key])
  }
  for (const definition of TEXT_PARAM_DEFINITIONS) {
    const value = settingMap[definition.key]
    texts[definition.key] = value === undefined || value === null ? '' : String(value)
  }
  savedSnapshot.value = snapshot()
}

async function load(): Promise<void> {
  loading.value = true
  try {
    applyBundle(await getSettings())
  } catch {
    // 错误提示由 http 拦截器统一处理；保留当前已填内容。
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function save(): Promise<void> {
  if (!dirty.value) {
    ElMessage.info('没有需要保存的修改')
    return
  }
  if (!weightValid.value) {
    ElMessage.warning('条件契合度权重与关系观念兼容度权重之和必须为 1')
    return
  }
  try {
    await ElMessageBox.confirm('保存后新参数立即生效，可能影响线上行为，确定保存吗？', '保存系统参数', {
      type: 'warning',
      confirmButtonText: '保存',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  saving.value = true
  try {
    // PUT /admin/settings 单 key 更新，逐个提交保证顺序与审计可读。
    for (const definition of FEATURE_DEFINITIONS) {
      await updateSetting(definition.key, features[definition.key])
    }
    for (const definition of NUMBER_PARAM_DEFINITIONS) {
      await updateSetting(definition.key, toNumber(numbers[definition.key]))
    }
    for (const definition of BOOLEAN_PARAM_DEFINITIONS) {
      await updateSetting(definition.key, booleans[definition.key])
    }
    for (const definition of TEXT_PARAM_DEFINITIONS) {
      await updateSetting(definition.key, texts[definition.key])
    }
    // 保存后从服务端重新拉取，确保展示值与实际生效值一致。
    await load()
    ElMessage.success('参数已保存，并已刷新为服务端最新值')
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">系统设置</h1>
        <p class="page-desc">
          修改 Feature Flag 与业务参数无需重新发版。保存后立即生效并自动刷新；所有修改写入审计日志。
        </p>
      </div>
      <div class="page-head-actions">
        <span v-if="dirty" class="dirty-hint">有未保存的修改</span>
        <el-button size="small" @click="load">刷新</el-button>
        <el-button size="small" type="primary" :loading="saving" @click="save">保存全部修改</el-button>
      </div>
    </div>

    <el-alert
      type="warning"
      :closable="false"
      show-icon
      class="mb-12"
      title="参数会影响线上行为"
      description="关闭开关会立即影响用户可用功能；K 匿名阈值建议不低于 10，权重之和必须为 1。理论市场相关参数仅影响统计估算，不得用于宣称精确人数或双向匹配率。"
    />

    <div class="page-card">
      <h2 class="card-title">Feature Flag</h2>
      <div class="feature-grid">
        <div v-for="definition in FEATURE_DEFINITIONS" :key="definition.key" class="feature-item">
          <div class="feature-text">
            <p class="feature-label">{{ definition.label }}</p>
            <p class="muted">{{ definition.description }}</p>
            <p class="muted mono">{{ definition.key }}</p>
          </div>
          <el-switch v-model="features[definition.key]" />
        </div>
      </div>
    </div>

    <div class="page-card">
      <h2 class="card-title">数值参数</h2>
      <el-form label-width="220px" label-position="right">
        <el-form-item v-for="definition in NUMBER_PARAM_DEFINITIONS" :key="definition.key" :label="definition.label">
          <div class="param-row">
            <el-input-number
              v-model="numbers[definition.key]"
              :min="definition.min"
              :max="definition.max"
              :step="definition.step"
              controls-position="right"
            />
            <span class="muted">{{ definition.description }}</span>
          </div>
        </el-form-item>
      </el-form>
      <el-alert
        v-if="!weightValid"
        type="error"
        :closable="false"
        class="mb-12"
        :title="'权重之和当前为 ' + weightSum.toFixed(2) + '，必须为 1'"
        description="条件契合度权重 + 关系观念兼容度权重 = 1，否则综合匹配度口径不一致。"
      />
      <el-alert
        v-if="kThresholdWarning"
        type="warning"
        :closable="false"
        class="mb-12"
        title="K 匿名阈值偏低"
        description="阈值小于 10 时存在小样本反向识别风险，建议保持不低于 10（页面统一显示 <10 类占位）。"
      />
      <p class="muted mt-12">权重之和：{{ weightSum.toFixed(2) }}；有效价值观题不足时客户端展示「关系观念数据不足」，不生成虚假精确分数。</p>
    </div>

    <div class="page-card">
      <h2 class="card-title">其他开关与协议版本</h2>
      <div class="feature-grid">
        <div v-for="definition in BOOLEAN_PARAM_DEFINITIONS" :key="definition.key" class="feature-item">
          <div class="feature-text">
            <p class="feature-label">{{ definition.label }}</p>
            <p class="muted">{{ definition.description }}</p>
            <p class="muted mono">{{ definition.key }}</p>
          </div>
          <el-switch v-model="booleans[definition.key]" />
        </div>
      </div>
      <el-form label-width="220px" class="mt-16">
        <el-form-item v-for="definition in TEXT_PARAM_DEFINITIONS" :key="definition.key" :label="definition.label">
          <div class="param-row">
            <el-input v-model="texts[definition.key]" style="width: 200px" maxlength="40" />
            <span class="muted">{{ definition.description }}</span>
          </div>
        </el-form-item>
      </el-form>
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
  gap: 8px;
  flex: none;
}

.dirty-hint {
  color: #e6a23c;
  font-size: 12px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 14px;
}

.feature-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  border: 1px solid var(--wmc-border);
  border-radius: 8px;
}

.feature-label {
  margin: 0 0 4px;
  font-weight: 600;
}

.param-row {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.mono {
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
}
</style>
