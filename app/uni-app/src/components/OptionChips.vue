<template>
  <view class="option-chips">
    <view
      v-for="item in options"
      :key="String(item.value)"
      class="wmc-chip"
      :class="{ 'wmc-chip--active': isActive(item.value) }"
      @tap="toggle(item.value)"
    >
      {{ item.label }}
    </view>
  </view>
</template>

<script setup lang="ts">
// 多选 / 单选标签组，选项标签统一来自 /meta/options。
type ChipValue = number | string
interface ChipOption {
  value: ChipValue
  label: string
}

const props = withDefaults(
  defineProps<{
    modelValue: ChipValue[]
    options: ChipOption[]
    single?: boolean
  }>(),
  { single: false }
)

const emit = defineEmits<{ (event: 'update:modelValue', value: ChipValue[]): void }>()

function isActive(value: ChipValue): boolean {
  return props.modelValue.some((item) => String(item) === String(value))
}

function toggle(value: ChipValue) {
  if (props.single) {
    emit('update:modelValue', [value])
    return
  }
  const exists = isActive(value)
  const next = exists
    ? props.modelValue.filter((item) => String(item) !== String(value))
    : props.modelValue.concat([value])
  emit('update:modelValue', next)
}
</script>

<style lang="scss" scoped>
.option-chips {
  display: flex;
  flex-wrap: wrap;
}
</style>
