<!-- 价值观问题配置：列表、新增、编辑、启停。问题不能写死在客户端。 -->
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createQuestion, listQuestions, updateQuestion } from '@/api/questions'
import type { CompatibilityQuestion, CompatibilityQuestionInput, QuestionOption } from '@/api/types'
import { QUESTION_TYPE_LABELS, labelOf } from '@/utils/enums'

interface QuestionFormState {
  code: string
  category: string
  title: string
  description: string
  question_type: number
  active: boolean
  sort_order: number
  options: QuestionOption[]
}

const loading = ref(false)
const questions = ref<CompatibilityQuestion[]>([])
const dialogVisible = ref(false)
const editingId = ref<string | null>(null)
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive<QuestionFormState>({
  code: '',
  category: '',
  title: '',
  description: '',
  question_type: 1,
  active: true,
  sort_order: 0,
  options: [{ value: '', label: '' }]
})
const rules: FormRules = {
  code: [{ required: true, message: '请填写问题编码', trigger: 'blur' }],
  category: [{ required: true, message: '请填问题分类', trigger: 'blur' }],
  title: [{ required: true, message: '请填写问题标题', trigger: 'blur' }]
}

async function load(): Promise<void> {
  loading.value = true
  try {
    questions.value = await listQuestions()
  } catch {
    questions.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

function resetForm(): void {
  form.code = ''
  form.category = ''
  form.title = ''
  form.description = ''
  form.question_type = 1
  form.active = true
  form.sort_order = 0
  form.options = [{ value: '', label: '' }]
}

function openCreate(): void {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(question: CompatibilityQuestion): void {
  editingId.value = question.id ?? null
  form.code = question.code ?? ''
  form.category = question.category ?? ''
  form.title = question.title ?? ''
  form.description = question.description ?? ''
  form.question_type = question.question_type ?? 1
  form.active = question.active ?? true
  form.sort_order = question.sort_order ?? 0
  form.options = (question.options ?? []).map((option) => ({ value: option.value ?? '', label: option.label ?? '' }))
  if (form.options.length === 0) {
    form.options = [{ value: '', label: '' }]
  }
  dialogVisible.value = true
}

function addOption(): void {
  form.options.push({ value: '', label: '' })
}

function removeOption(index: number): void {
  if (form.options.length <= 1) {
    ElMessage.warning('至少保留一个选项')
    return
  }
  form.options.splice(index, 1)
}

// 将行数据转换为全量更新入参（用于启停等单字段变更）。
function toInput(question: CompatibilityQuestion, active: boolean): CompatibilityQuestionInput {
  return {
    code: question.code ?? '',
    category: question.category ?? '',
    title: question.title ?? '',
    description: question.description ?? '',
    question_type: question.question_type ?? 1,
    options: (question.options ?? []).map((option) => ({ value: option.value ?? '', label: option.label ?? '' })),
    active,
    sort_order: question.sort_order ?? 0
  }
}

async function submit(): Promise<void> {
  if (!formRef.value) {
    return
  }
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  const options = form.options.map((option) => ({ value: option.value.trim(), label: option.label.trim() }))
  if (options.some((option) => !option.value || !option.label)) {
    ElMessage.warning('请完整填写每个选项的值与文案')
    return
  }
  if (options.length === 0) {
    ElMessage.warning('至少需要一个选项')
    return
  }
  const payload: CompatibilityQuestionInput = {
    code: form.code.trim(),
    category: form.category.trim(),
    title: form.title.trim(),
    description: form.description.trim() || undefined,
    question_type: Number(form.question_type),
    options,
    active: form.active,
    sort_order: Number(form.sort_order)
  }
  saving.value = true
  try {
    if (editingId.value) {
      await updateQuestion(editingId.value, payload)
      ElMessage.success('问题已更新')
    } else {
      await createQuestion(payload)
      ElMessage.success('问题已新增')
    }
    dialogVisible.value = false
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  } finally {
    saving.value = false
  }
}

async function handleToggle(question: CompatibilityQuestion): Promise<void> {
  if (!question.id) {
    ElMessage.warning('该问题缺少 id，无法启停')
    return
  }
  try {
    await updateQuestion(question.id, toInput(question, !question.active))
    ElMessage.success(question.active ? '问题已停用' : '问题已启用')
    await load()
  } catch {
    // 错误提示由 http 拦截器统一处理。
  }
}
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h1 class="page-title">价值观问题</h1>
        <p class="page-desc">
          婚恋价值观问题不能写死在客户端，必须由后台配置。问题应拆成具体生活安排，避免抽象标签；停用后不再下发给客户端。
        </p>
      </div>
      <div class="page-head-actions">
        <el-button size="small" @click="load">刷新</el-button>
        <el-button size="small" type="primary" @click="openCreate">新增问题</el-button>
      </div>
    </div>

    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="mb-12"
      title="配置建议"
      description="不要直接询问「是否女权/是否传统」等标签，应拆成婚后开支、购房首付、家务分配、育儿时间、异地迁居等可决策问题；选项需要覆盖常见安排。"
    />

    <div class="page-card">
      <el-table :data="questions" border size="default">
        <el-table-column label="排序" width="80" align="center">
          <template #default="{ row }">{{ row.sort_order ?? 0 }}</template>
        </el-table-column>
        <el-table-column label="编码" width="150">
          <template #default="{ row }"><span class="mono">{{ row.code || '—' }}</span></template>
        </el-table-column>
        <el-table-column label="分类" width="110" align="center">
          <template #default="{ row }">{{ row.category || '—' }}</template>
        </el-table-column>
        <el-table-column label="标题" min-width="240">
          <template #default="{ row }">
            <div>{{ row.title || '未命名问题' }}</div>
            <div class="muted">{{ row.description || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="90" align="center">
          <template #default="{ row }">{{ labelOf(QUESTION_TYPE_LABELS, row.question_type) }}</template>
        </el-table-column>
        <el-table-column label="选项数" width="90" align="center">
          <template #default="{ row }">{{ row.options?.length ?? 0 }}</template>
        </el-table-column>
        <el-table-column label="启用" width="100" align="center">
          <template #default="{ row }">
            <el-switch :model-value="Boolean(row.active)" @change="handleToggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="版本" width="80" align="center">
          <template #default="{ row }">{{ row.version ?? 1 }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无价值观问题，请先新增" :image-size="60" />
        </template>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑问题' : '新增问题'" width="680px" top="5vh">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="问题编码" prop="code">
          <el-input v-model="form.code" maxlength="60" placeholder="如 finance_share，创建后尽量保持不变" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-input v-model="form.category" maxlength="40" placeholder="如 财务 / 家务 / 育儿" />
        </el-form-item>
        <el-form-item label="问题标题" prop="title">
          <el-input v-model="form.title" maxlength="120" placeholder="如 婚后家庭开支如何承担？" />
        </el-form-item>
        <el-form-item label="问题说明">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="300" show-word-limit />
        </el-form-item>
        <el-form-item label="问题类型">
          <el-select v-model="form.question_type" style="width: 160px">
            <el-option :value="1" label="单选" />
            <el-option :value="2" label="多选" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" :max="999" controls-position="right" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.active" />
        </el-form-item>
        <el-form-item label="选项">
          <div class="option-list">
            <div v-for="(option, index) in form.options" :key="index" class="option-row">
              <el-input v-model="option.value" placeholder="选项值，如 A" style="width: 130px" />
              <el-input v-model="option.label" placeholder="选项文案，如 主要由男方承担" />
              <el-button link type="danger" @click="removeOption(index)">删除</el-button>
            </div>
            <el-button link type="primary" @click="addOption">添加选项</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </template>
    </el-dialog>
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
  gap: 8px;
  flex: none;
}

.mono {
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
}

.option-list {
  width: 100%;
}

.option-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
</style>
