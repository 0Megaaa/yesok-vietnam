<script setup>
/**
 * DynamicForm — 动态表单渲染引擎
 *
 * 根据 fields JSON 数组通过 component :is 动态渲染 Element Plus 控件。
 * v-model 格式：{ [field.key]: value }
 *
 * fields 元素结构：
 * {
 *   key:       string                // 字段标识，用于 v-model 绑定
 *   label:     string                // 前端显示标签
 *   type:      string               // input | textarea | number | date | select | image
 *   required:  boolean               // 是否必填
 *   options?:  [{ label, value }]    // select 类型专用选项列表
 *   multiple?: boolean               // select / image 是否支持多选
 * }
 */
import { computed, ref, watch } from 'vue'
import {
  ElInput,
  ElSelect,
  ElOption,
  ElDatePicker,
  ElUpload,
  ElIcon,
  ElMessage,
} from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

// 动态组件映射表：通过 type 字符串找到对应组件
const fieldTypeComponents = {
  input: ElInput,
  textarea: ElInput,
  number: ElInput,
  phone: ElInput,
  date: ElDatePicker,
  datetime: ElDatePicker,
  select: ElSelect,
  file: null, // file 使用 el-input URL 输入独立处理
}

// 判断是否为 DatePicker 类型
const isDatePicker = (field) => field.type === 'date' || field.type === 'datetime'

// 判断是否为 File 类型
const isFile = (field) => field.type === 'file'

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  fields: { type: Array, default: () => [] },
  validateOnChange: { type: Boolean, default: true },
})

const emit = defineEmits(['update:modelValue', 'validate'])

const errors = ref({})
const localValue = ref({ ...props.modelValue })

watch(localValue, (val) => emit('update:modelValue', { ...val }), { deep: true })

// 解析字段类型对应的 el-input type 属性
const inputType = (field) => {
  if (field.type === 'textarea') return 'textarea'
  if (field.type === 'number') return 'number'
  if (field.type === 'phone') return 'number'
  return 'text'
}

// 校验单个字段
const validateField = (field) => {
  const val = localValue.value[field.key]
  const strVal = val == null ? '' : Array.isArray(val) ? val.join(',') : String(val)
  if (field.required && !strVal.trim()) {
    errors.value[field.key] = `请填写「${field.label}」`
    return false
  }
  delete errors.value[field.key]
  return true
}

// 全量校验
const validateAll = () => {
  let ok = true
  errors.value = {}
  for (const f of props.fields) {
    if (!validateField(f)) ok = false
  }
  emit('validate', { ok, errors: { ...errors.value } })
  return ok
}

// el-upload on-success：追加图片 URL 到 localValue
const onImageSuccess = (field, res) => {
  const urls = localValue.value[field.key] || []
  const url = res?.data?.url || res?.url || ''
  if (url) localValue.value[field.key] = [...urls, url]
}

// el-upload on-remove：更新 URL 列表
const onImageRemove = (field, file, fileList) => {
  localValue.value[field.key] = fileList.map(f => f.response?.data?.url || f.url || '').filter(Boolean)
}

// 暴露 validateAll 给父组件
defineExpose({ validateAll })
</script>

<template>
  <div class="dynamic-form">
    <div
      v-for="field in fields"
      :key="field.key"
      class="field-item"
    >
      <!-- 标签 -->
      <div class="field-label">
        <span>{{ field.label }}</span>
        <span v-if="field.required" class="required-mark">*</span>
      </div>

      <!-- image 单独处理（el-upload） -->
      <template v-if="field.type === 'image'">
        <el-upload
          :file-list="localValue[`_${field.key}_files`]"
          @update:file-list="(list) => { localValue[`_${field.key}_files`] = list }"
          action="/v1/admin/upload"
          list-type="picture-card"
          :multiple="field.multiple !== false"
          :on-success="(res) => onImageSuccess(field, res)"
          :on-remove="(file, fileList) => onImageRemove(field, file, fileList)"
          :on-error="() => ElMessage.error('图片上传失败，请重试')"
          accept="image/*"
        >
          <el-icon class="upload-icon"><Plus /></el-icon>
        </el-upload>
        <div class="upload-hint">支持 JPG/PNG，最多 {{ field.multiple === false ? 1 : 10 }} 张</div>
      </template>

      <!-- file 使用 el-input URL 输入 -->
      <template v-else-if="isFile(field)">
        <el-input
          v-model="localValue[field.key]"
          placeholder="请输入文件URL，或联系管理员上传"
          clearable
        />
      </template>

      <!-- date / datetime 使用 el-date-picker -->
      <template v-else-if="isDatePicker(field)">
        <el-date-picker
          v-model="localValue[field.key]"
          :type="field.type === 'datetime' ? 'datetime' : 'date'"
          :placeholder="field.type === 'datetime' ? '请选择日期时间' : '请选择日期'"
          :format="field.type === 'datetime' ? 'YYYY-MM-DD HH:mm' : 'YYYY-MM-DD'"
          :value-format="field.type === 'datetime' ? 'YYYY-MM-DD HH:mm' : 'YYYY-MM-DD'"
          style="width: 100%"
          @blur="validateOnChange && validateField(field)"
          @change="validateOnChange && validateField(field)"
        />
      </template>

      <!-- 其他类型通过 component :is 动态渲染 -->
      <component
        v-else
        :is="fieldTypeComponents[field.type] || ElInput"
        v-model="localValue[field.key]"
        :type="inputType(field)"
        :rows="field.type === 'textarea' ? 3 : undefined"
        :placeholder="`请输入${field.label}`"
        style="width: 100%"
        @blur="validateOnChange && validateField(field)"
        @change="validateOnChange && validateField(field)"
      >
        <!-- select 的选项插槽 -->
        <template v-if="field.type === 'select'">
          <el-option
            v-for="opt in (field.options || [])"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </template>
      </component>

      <!-- 错误提示 -->
      <div v-if="errors[field.key]" class="field-error">{{ errors[field.key] }}</div>
    </div>
  </div>
</template>

<style scoped>
.dynamic-form { display: flex; flex-direction: column; gap: 18px; }

.field-item { display: flex; flex-direction: column; gap: 6px; }

.field-label {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #606266;
  font-size: 13px;
  font-weight: 600;
}

.required-mark { color: #f56c6c; font-size: 14px; }

.field-error {
  color: #f56c6c;
  font-size: 11px;
  margin-top: 2px;
}

.upload-icon {
  font-size: 22px;
  color: #8c939a;
}

.upload-hint {
  color: #a0a6ab;
  font-size: 11px;
  margin-top: 4px;
}
</style>
