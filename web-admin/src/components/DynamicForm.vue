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
  // 可选：指定当前订单 ID，image/file 类型上传时会触发 upload-order-material 事件
  orderId: { type: [Number, String], default: null },
})

const emit = defineEmits(['update:modelValue', 'validate', 'upload-order-material'])

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

// onImageSuccess 追加图片 URL 到 localValue
const onImageSuccess = (field, res) => {
  const urls = localValue.value[field.key] || []
  const url = res?.data?.url || res?.url || ''
  if (url) localValue.value[field.key] = [...urls, url]
}

// onImageRemove 更新 URL 列表
const onImageRemove = (field, file, fileList) => {
  localValue.value[field.key] = fileList.map(f => f.response?.data?.url || f.url || '').filter(Boolean)
}

// handleImageHttpRequest 拦截 el-upload 默认上传，改为触发父组件自定义上传
// 父组件通过 @upload-order-material 事件处理
const handleImageHttpRequest = (options, field) => {
  emit('upload-order-material', { options, field })
}

// exposeUploadedUrls 暴露给父组件，用于上传成功后追加 URL 到图片数组
// 支持单个 URL 字符串或 URL 数组
const exposeUploadedUrls = (fieldKey, urls) => {
  const existing = localValue.value[fieldKey] || []
  if (Array.isArray(urls)) {
    localValue.value[fieldKey] = [...existing, ...urls]
  } else if (urls) {
    localValue.value[fieldKey] = [...existing, urls]
  }
}
defineExpose({ validateAll, exposeUploadedUrls })
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

      <!-- image 单独处理（el-upload + http-request 拦截上传） -->
      <template v-if="field.type === 'image'">
        <el-upload
          :file-list="(localValue[field.key] || []).map((url, i) => ({ url, uid: i }))"
          :http-request="(opts) => handleImageHttpRequest(opts, field)"
          :on-remove="(file, fileList) => onImageRemove(field, file, fileList)"
          :on-error="() => ElMessage.error('图片上传失败，请重试')"
          :accept="'.jpg,.jpeg,.png'"
          list-type="picture-card"
          :multiple="field.multiple !== false"
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
