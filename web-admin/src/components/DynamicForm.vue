<script setup>
/**
 * DynamicForm — 动态表单渲染引擎
 *
 * 根据 fields JSON 数组通过 component :is 动态渲染 Element Plus 控件。
 * v-model 格式：{ [field.key]: value }
 *
 * fields 元素结构：
 * {
 *   key:            string                // 字段标识，用于 v-model 绑定
 *   label:          string                // 前端显示标签
 *   type:           string               // input | textarea | number | date | select | image | file
 *   required:       boolean               // 是否必填
 *   required_when:  { key, value }        // 条件必填
 *   visible_when:   { key, value }        // 条件显示
 *   options?:       [{ label, value }]    // select 类型专用选项列表
 *   multiple?:      boolean               // select / image 是否支持多选
 * }
 *
 * image / file 类型由父组件接管上传，通过 @upload-order-material 事件触发。
 * 选择文件后立即 emit，父组件调用接口后调用 onSuccess(url) 写回 localValue。
 */
import { ref, watch } from 'vue'
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
}

// image / file 类型统一渲染为上传组件，由父组件处理上传
const fieldTypeComponentsForNormal = {
  input: ElInput,
  textarea: ElInput,
  number: ElInput,
  phone: ElInput,
  select: ElSelect,
}

// 判断是否为 DatePicker 类型
const isDatePicker = (field) => field.type === 'date' || field.type === 'datetime'

// 判断是否为上传类型（image 或 file）
const isUploadField = (field) => field.type === 'image' || field.type === 'file'

// 判断条件是否满足
// condition 格式：{ key: "payment_type", value: "full" }
// value 支持数组：{ key: "x", value: ["a","b"] }
const matchCondition = (condition) => {
  if (!condition) return true

  const key = condition.key
  const expected = condition.value
  const actual = localValue.value[key]

  if (Array.isArray(expected)) {
    return expected.includes(actual)
  }

  return actual === expected
}

// 字段在当前表单值下是否可见
const isFieldVisible = (field) => {
  if (!field.visible_when) return true
  return matchCondition(field.visible_when)
}

// 字段在当前表单值下是否为必填（required=true 或满足 required_when 条件）
const isFieldRequired = (field) => {
  if (field.required === true) return true
  if (field.required_when) return matchCondition(field.required_when)
  return false
}

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  fields: { type: Array, default: () => [] },
  validateOnChange: { type: Boolean, default: true },
})

const emit = defineEmits(['update:modelValue', 'validate', 'upload-order-material'])

const errors = ref({})
const localValue = ref({ ...props.modelValue })

watch(localValue, (val) => emit('update:modelValue', { ...val }), { deep: true })

// 条件显示变化时，清除不可见字段的校验错误
watch(localValue, () => {
  for (const field of props.fields) {
    if (!isFieldVisible(field)) {
      delete errors.value[field.key]
    }
  }
}, { deep: true })

// 解析字段类型对应的 el-input type 属性
const inputType = (field) => {
  if (field.type === 'textarea') return 'textarea'
  if (field.type === 'number') return 'number'
  if (field.type === 'phone') return 'number'
  return 'text'
}

// 统一把值转为数组
const asArray = (value) => {
  if (!value) return []
  if (Array.isArray(value)) return value
  return [value]
}

// 校验单个字段
const validateField = (field) => {
  // 不可见字段跳过校验
  if (!isFieldVisible(field)) {
    delete errors.value[field.key]
    return true
  }

  const val = localValue.value[field.key]

  let hasValue = false
  if (Array.isArray(val)) {
    hasValue = val.filter(Boolean).length > 0
  } else {
    hasValue = val !== null && val !== undefined && String(val).trim() !== ''
  }

  if (isFieldRequired(field) && !hasValue) {
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
    if (!isFieldVisible(f)) continue
    if (!validateField(f)) ok = false
  }
  emit('validate', { ok, errors: { ...errors.value } })
  return ok
}

// 把本地已有 URL 转为 el-upload 需要的 file-list 格式
const getUploadFileList = (field) => {
  return asArray(localValue.value[field.key])
    .filter(Boolean)
    .map((url, index) => ({
      name: `${field.key}_${index + 1}`,
      url,
      uid: `${field.key}_${index}`,
    }))
}

// 上传成功后追加 URL 到 localValue
// multiple=false 时替换，单图模式；multiple=true 时追加
const appendUploadedUrl = (field, url) => {
  if (!url) return

  if (field.multiple === false) {
    localValue.value[field.key] = url
  } else {
    const current = asArray(localValue.value[field.key])
    localValue.value[field.key] = [...current, url]
  }

  delete errors.value[field.key]
}

// 移除文件后更新 URL 列表
const removeUploadedUrl = (field, file, fileList) => {
  const urls = (fileList || [])
    .map((item) => item.response?.url || item.response?.data?.url || item.url || '')
    .filter(Boolean)

  if (field.multiple === false) {
    localValue.value[field.key] = urls[0] || ''
  } else {
    localValue.value[field.key] = urls
  }

  if (props.validateOnChange) validateField(field)
}

// 选择文件后立即触发父组件上传事件
// auto-upload="false" 时 on-change 在选文件后触发，然后 emit 给父组件
const handleUploadChange = (uploadFile, uploadFiles, field) => {
  const rawFile = uploadFile?.raw
  if (!rawFile) {
    ElMessage.error('请选择要上传的文件')
    return
  }

  const fileName = (rawFile.name || uploadFile.name || '').toLowerCase()
  if (!/\.(jpg|jpeg|png)$/.test(fileName)) {
    ElMessage.error('仅支持 JPG、JPEG、PNG 格式')
    return
  }

  emit('upload-order-material', {
    file: rawFile,
    field,
    onSuccess: (url) => {
      appendUploadedUrl(field, url)
    },
    onError: (message) => {
      ElMessage.error(message || '上传失败，请重试')
      // 移除刚添加的文件（el-upload 已选中，file-list 会显示，取消选中即可）
    },
  })
}

// 暴露给父组件：父组件上传成功后主动写入 URL（兼容新旧两种写法）
const exposeUploadedUrls = (fieldKey, urls) => {
  const field = props.fields.find((item) => item.key === fieldKey)
  if (!field) return

  if (Array.isArray(urls)) {
    if (field.multiple === false) {
      localValue.value[fieldKey] = urls[0] || ''
    } else {
      localValue.value[fieldKey] = [...asArray(localValue.value[fieldKey]), ...urls]
    }
  } else if (urls) {
    appendUploadedUrl(field, urls)
  }

  delete errors.value[fieldKey]
}

defineExpose({ validateAll, exposeUploadedUrls })
</script>

<template>
  <div class="dynamic-form">
    <div
      v-for="field in fields"
      v-show="isFieldVisible(field)"
      :key="field.key"
      class="field-item"
    >
      <!-- 标签 -->
      <div class="field-label">
        <span>{{ field.label }}</span>
        <span v-if="isFieldRequired(field)" class="required-mark">*</span>
      </div>

      <!-- image / file 类型：统一渲染为上传组件，由父组件处理上传 -->
      <template v-if="isUploadField(field)">
        <el-upload
          action="#"
          :auto-upload="false"
          :file-list="getUploadFileList(field)"
          :on-change="(uploadFile, uploadFiles) => handleUploadChange(uploadFile, uploadFiles, field)"
          :on-remove="(file, fileList) => removeUploadedUrl(field, file, fileList)"
          :accept="'.jpg,.jpeg,.png'"
          list-type="picture-card"
          :multiple="field.multiple !== false"
          :limit="field.multiple === false ? 1 : 10"
        >
          <el-icon class="upload-icon"><Plus /></el-icon>
        </el-upload>
        <div class="upload-hint">
          支持 JPG/PNG，最多 {{ field.multiple === false ? 1 : 10 }} 张
        </div>
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
        :is="fieldTypeComponentsForNormal[field.type] || ElInput"
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
