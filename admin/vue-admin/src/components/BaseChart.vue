<!-- ECharts 通用容器：负责初始化、响应式重绘与销毁。 -->
<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'

const props = withDefaults(defineProps<{ option: echarts.EChartsOption; height?: string }>(), {
  height: '320px'
})

const container = ref<HTMLDivElement | null>(null)
let chart: ReturnType<typeof echarts.init> | null = null

// 首次渲染或 option 变化时重绘。
function render(): void {
  if (!container.value) {
    return
  }
  if (!chart) {
    chart = echarts.init(container.value)
  }
  chart.setOption(props.option, true)
}

function handleResize(): void {
  chart?.resize()
}

onMounted(() => {
  render()
  window.addEventListener('resize', handleResize)
})

watch(
  () => props.option,
  () => {
    render()
  },
  { deep: true }
)

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
  chart = null
})
</script>

<template>
  <div ref="container" class="base-chart" :style="{ height }" />
</template>

<style scoped>
.base-chart {
  width: 100%;
}
</style>
