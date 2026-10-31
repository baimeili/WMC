<!-- 看板统计卡片：统一展示指标名称、数值与口径说明。 -->
<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string
    value: string
    hint?: string
    tone?: 'default' | 'primary' | 'success' | 'warning' | 'danger'
  }>(),
  {
    hint: '',
    tone: 'default'
  }
)
</script>

<template>
  <div class="stat-card" :class="'tone-' + tone">
    <p class="stat-label">{{ label }}</p>
    <p class="stat-value">{{ value }}</p>
    <p v-if="hint" class="stat-hint">{{ hint }}</p>
  </div>
</template>

<style scoped>
.stat-card {
  padding: 16px 18px;
  border: 1px solid var(--wmc-border);
  border-radius: 8px;
  background: #fff;
  border-top: 3px solid #c0c4cc;
}

.tone-primary {
  border-top-color: #409eff;
}

.tone-success {
  border-top-color: #67c23a;
}

.tone-warning {
  border-top-color: #e6a23c;
}

.tone-danger {
  border-top-color: #f56c6c;
}

.stat-label {
  margin: 0;
  color: var(--wmc-text-secondary);
  font-size: 13px;
}

.stat-value {
  margin: 8px 0 0;
  color: var(--wmc-text-primary);
  font-size: 26px;
  font-weight: 600;
  line-height: 1.2;
}

.stat-hint {
  margin: 6px 0 0;
  color: var(--wmc-text-secondary);
  font-size: 12px;
  line-height: 1.5;
}
</style>
