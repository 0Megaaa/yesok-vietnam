<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { get, post, ORIGIN_URL } from '@/api/request'
import { uploadOrderMaterial } from '@/api/order'
import { useClientStore } from '@/store/client'

const client = useClientStore()

const mode = ref('')
const serviceId = ref('')
const orderId = ref('')
const actionName = ref('')

const loading = ref(true)
const submitting = ref(false)

const serviceInfo = ref(null)
const orderInfo = ref(null)
const currentAction = ref(null)

const fields = ref([])
const formData = ref({})
const errors = ref({})
const localFileMap = ref({})
const uploadingMap = ref({})

const pageTitle = computed(() => {
  if (mode.value === 'service') return '填写预约信息'
  return currentAction.value?.button_label || '填写信息'
})

const isServiceMode = computed(() => mode.value === 'service')

const unwrapResponse = (res) => res?.data ?? res ?? {}

const normalizeFieldOptions = (options = []) => {
  if (!Array.isArray(options)) return []
  return options.map((opt) => {
    if (typeof opt === 'string') return { label: opt, value: opt }
    return {
      label: opt.label ?? opt.name ?? opt.value ?? '',
      value: opt.value ?? opt.label ?? opt.name ?? '',
    }
  })
}

const normalizeFields = (rawFields = []) => {
  if (typeof rawFields === 'string') {
    try {
      rawFields = JSON.parse(rawFields)
    } catch (e) {
      rawFields = []
    }
  }

  if (!Array.isArray(rawFields)) return []

  return rawFields
    .map((f) => {
      const key = f.key || f.name || ''
      return {
        ...f,
        key,
        name: key,
        label: f.label || key,
        type: f.type || 'text',
        required: !!f.required,
        placeholder: f.placeholder || `请输入${f.label || key}`,
        options: normalizeFieldOptions(f.options || []),
      }
    })
    .filter((f) => f.key)
}

const initFormData = (list = []) => {
  const data = {}
  const err = {}
  list.forEach((field) => {
    data[field.key] = ''
    err[field.key] = ''
  })
  formData.value = data
  errors.value = err
  localFileMap.value = {}
  uploadingMap.value = {}
}

const getOptionLabel = (field) => {
  const value = formData.value[field.key]
  const option = (field.options || []).find((item) => item.value === value)
  return option?.label || value || field.placeholder || `请选择${field.label}`
}

const matchCondition = (condition) => {
  if (!condition) return true
  const expected = condition.value
  const actual = formData.value[condition.key]
  if (Array.isArray(expected)) return expected.includes(actual)
  return actual === expected
}

const isFieldVisible = (field) => {
  if (!field.visible_when) return true
  return matchCondition(field.visible_when)
}

const isFieldRequired = (field) => {
  if (field.required === true) return true
  if (field.required_when) return matchCondition(field.required_when)
  return false
}

const validateField = (field) => {
  if (!isFieldVisible(field)) return true

  const val = formData.value[field.key]

  if (!isFieldRequired(field)) {
    errors.value[field.key] = ''
    return true
  }

  if (field.type === 'image' || field.type === 'file') {
    if (!val || typeof val !== 'string' || !val.trim()) {
      errors.value[field.key] = `请上传${field.label}`
      return false
    }
    errors.value[field.key] = ''
    return true
  }

  const str = (val || '').toString().trim()
  if (!str) {
    errors.value[field.key] = `请填写${field.label}`
    return false
  }

  errors.value[field.key] = ''
  return true
}

const validateAll = () => {
  let ok = true
  fields.value.forEach((field) => {
    if (!validateField(field)) ok = false
  })
  return ok
}

const isLocalTempFile = (url) => {
  if (!url || typeof url !== 'string') return false
  return (
    url.startsWith('wxfile://') ||
    url.startsWith('http://tmp/') ||
    url.startsWith('file://') ||
    url.includes('/tmp/') ||
    url.includes('/temp/')
  )
}

const toFullFileUrl = (url) => {
  if (!url) return ''
  if (isLocalTempFile(url)) return url
  if (/^https?:\/\//.test(url)) return url
  const origin = String(ORIGIN_URL || '').replace(/\/+$/, '')
  const path = url.startsWith('/') ? url : `/${url}`
  return `${origin}${path}`
}

const chooseImageField = (field) => {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      const tempFilePath = res.tempFilePaths?.[0]
      if (!tempFilePath) return

      formData.value[field.key] = tempFilePath
      localFileMap.value[field.key] = {
        tempFilePath,
        name: field.key,
        uploadedUrl: '',
      }
      errors.value[field.key] = ''
    },
  })
}

const previewImage = (url) => {
  const full = toFullFileUrl(url)
  if (!full) return
  uni.previewImage({ urls: [full], current: full })
}

const uploadLocalFilesForOrder = async (targetOrderId, input = {}) => {
  const finalData = { ...input }

  for (const field of fields.value) {
    if (field.type !== 'image' && field.type !== 'file') continue

    const key = field.key
    const currentValue = formData.value[key]
    if (!currentValue || !isLocalTempFile(currentValue)) {
      if (currentValue) finalData[key] = currentValue
      continue
    }

    uploadingMap.value[key] = true
    try {
      const uploadRes = await uploadOrderMaterial(targetOrderId, currentValue, key)
      if (!uploadRes?.url) throw new Error(`${field.label}上传失败`)
      finalData[key] = uploadRes.url
      formData.value[key] = uploadRes.url
    } finally {
      uploadingMap.value[key] = false
    }
  }

  return finalData
}

const loadServiceForm = async () => {
  const [svcRes, formRes] = await Promise.all([
    get(`/v1/client/services/${serviceId.value}`),
    get(`/v1/client/services/${serviceId.value}/init-form`),
  ])

  const svcPayload = unwrapResponse(svcRes)
  const formPayload = unwrapResponse(formRes)

  serviceInfo.value = svcPayload

  const action = formPayload.action || {
    action_name: formPayload.action_name,
    button_label: formPayload.button_label,
    form_fields: formPayload.form_fields || [],
  }

  currentAction.value = action?.action_name ? action : null

  const rawFields =
    action?.form_fields ||
    formPayload.form_fields ||
    formPayload.form_schema?.fields ||
    formPayload.fields ||
    []

  fields.value = normalizeFields(rawFields)
  initFormData(fields.value)
}

const loadOrderActionForm = async () => {
  const [detailRes, actionsRes] = await Promise.all([
    get(`/v1/client/orders/${orderId.value}`),
    get(`/v1/client/orders/${orderId.value}/actions`),
  ])

  const detailPayload = unwrapResponse(detailRes)
  const actionsPayload = unwrapResponse(actionsRes)

  orderInfo.value = detailPayload.order || detailPayload

  const actionList =
    actionsPayload.actions ||
    actionsPayload.action_nodes ||
    actionsPayload.actionNodes ||
    orderInfo.value?.action_nodes ||
    []

  const action = actionList.find((item) => item.action_name === actionName.value)

  if (!action) {
    throw new Error('当前操作不存在或已失效')
  }

  currentAction.value = action
  fields.value = normalizeFields(action.form_fields || action.formFields || [])
  initFormData(fields.value)
}

const goBack = () => {
  const pages = getCurrentPages()
  if (pages.length > 1) {
    uni.navigateBack()
  } else {
    uni.switchTab({ url: '/pages/profile/index' })
  }
}

const submitServiceOrder = async () => {
  const normalData = {}
  fields.value.forEach((field) => {
    if (field.type === 'image' || field.type === 'file') return
    normalData[field.key] = formData.value[field.key]
  })

  const payload = {
    service_id: serviceInfo.value?.id || serviceInfo.value?.service_id || serviceId.value,
    contact_name: normalData.contact_name || client.userInfo?.nickname || '微信客户',
    contact_phone: normalData.contact_phone || '',
    form_data: normalData,
  }

  const res = await post('/v1/client/orders', payload)
  const data = unwrapResponse(res)
  const createdOrder = data.order || data.data?.order || data
  const newOrderId = createdOrder?.id

  if (!newOrderId) {
    throw new Error('订单创建成功，但未返回订单ID')
  }

  await uploadLocalFilesForOrder(newOrderId, {})

  uni.showToast({ title: '订单已提交', icon: 'success' })

  setTimeout(() => {
      uni.redirectTo({ url: `/subpkg/order-detail/index?id=${newOrderId}` })
  }, 800)
}

const submitOrderAction = async () => {
  const inputData = await uploadLocalFilesForOrder(orderId.value, { ...formData.value })

  await post(`/v1/client/orders/${orderId.value}/action`, {
    action_name: currentAction.value.action_name,
    remark: '',
    input_data: inputData,
  })

  uni.showToast({ title: '操作成功', icon: 'success' })

  uni.setStorageSync('yesok_order_detail_need_refresh', {
    order_id: orderId.value,
    ts: Date.now(),
  })

  setTimeout(() => {
    uni.navigateBack()
  }, 600)
}

const submitForm = async () => {
  if (submitting.value) return
  if (!validateAll()) return

  submitting.value = true
  uni.showLoading({ title: '提交中...' })

  try {
    if (isServiceMode.value) {
      await submitServiceOrder()
    } else {
      await submitOrderAction()
    }
  } catch (error) {
    uni.showToast({
      title: error?.message || '提交失败',
      icon: 'none',
    })
  } finally {
    uni.hideLoading()
    submitting.value = false
  }
}

onLoad(async (options) => {
  mode.value = options.mode || ''
  serviceId.value = options.service_id || options.serviceId || ''
  orderId.value = options.order_id || options.orderId || ''
  actionName.value = decodeURIComponent(options.action_name || options.actionName || '')

  loading.value = true

  try {
    if (mode.value === 'service') {
      await loadServiceForm()
    } else if (mode.value === 'order') {
      await loadOrderActionForm()
    } else {
      throw new Error('未知表单模式')
    }
  } catch (error) {
    uni.showToast({
      title: error?.message || '表单加载失败',
      icon: 'none',
    })
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <view class="form-page">
    <view class="page-header">
      <view class="back-btn" @tap="goBack">‹</view>
      <view class="page-title-wrap">
        <text class="page-title">{{ pageTitle }}</text>
        <text class="page-subtitle">
          {{ isServiceMode ? '请填写预约需求，管家将尽快联系您' : '请按要求填写资料后提交' }}
        </text>
      </view>
    </view>

    <view v-if="loading" class="loading-card">
      正在加载表单...
    </view>

    <view v-else class="form-card">
      <view
        v-for="field in fields"
        v-show="isFieldVisible(field)"
        :key="field.key"
        class="field-wrap"
      >
        <view class="field-label">
          <text>{{ field.label }}</text>
          <text v-if="isFieldRequired(field)" class="required">*</text>
        </view>

        <input
          v-if="field.type === 'text' || field.type === 'phone' || field.type === 'number'"
          v-model="formData[field.key]"
          :type="field.type === 'phone' ? 'number' : field.type === 'number' ? 'digit' : 'text'"
          class="field-input"
          :class="{ error: errors[field.key] }"
          :placeholder="field.placeholder || `请输入${field.label}`"
          confirm-type="done"
          cursor-spacing="24"
          @blur="validateField(field)"
        />

        <input
          v-else-if="field.type === 'date'"
          v-model="formData[field.key]"
          class="field-input"
          :class="{ error: errors[field.key] }"
          placeholder="格式：2025-01-15"
          confirm-type="done"
          cursor-spacing="24"
          @blur="validateField(field)"
        />

        <input
          v-else-if="field.type === 'datetime'"
          v-model="formData[field.key]"
          class="field-input"
          :class="{ error: errors[field.key] }"
          placeholder="格式：2025-01-15 14:30"
          confirm-type="done"
          cursor-spacing="24"
          @blur="validateField(field)"
        />

        <textarea
          v-else-if="field.type === 'textarea'"
          v-model="formData[field.key]"
          class="field-textarea"
          :class="{ error: errors[field.key] }"
          :placeholder="field.placeholder || `请输入${field.label}`"
          :auto-height="false"
          cursor-spacing="24"
          @blur="validateField(field)"
        />

        <picker
          v-else-if="field.type === 'select'"
          mode="selector"
          :value="0"
          :range="field.options || []"
          :range-key="'label'"
          @tap="uni.hideKeyboard && uni.hideKeyboard()"
          @change="(e) => {
            const selected = field.options[e.detail.value]
            formData[field.key] = selected?.value ?? selected ?? ''
            validateField(field)
          }"
        >
          <view class="field-picker" :class="{ error: errors[field.key], filled: formData[field.key] }">
            <text>{{ getOptionLabel(field) }}</text>
            <text class="picker-arrow">›</text>
          </view>
        </picker>

        <view v-else-if="field.type === 'image' || field.type === 'file'" class="upload-field">
          <view class="upload-box" :class="{ uploading: uploadingMap[field.key] }" @tap="chooseImageField(field)">
            <image
              v-if="formData[field.key]"
              class="upload-preview"
              :src="toFullFileUrl(formData[field.key])"
              mode="aspectFill"
              @tap.stop="previewImage(formData[field.key])"
            />
            <view v-else class="upload-placeholder">
              <text class="upload-plus">+</text>
              <text class="upload-text">上传图片</text>
            </view>
            <view v-if="uploadingMap[field.key]" class="upload-mask">
              <text>上传中...</text>
            </view>
          </view>
          <text v-if="formData[field.key]" class="upload-change" @tap="chooseImageField(field)">点击更换</text>
        </view>

        <view v-else class="field-input unsupported">
          暂不支持字段类型：{{ field.type }}
        </view>

        <text v-if="errors[field.key]" class="field-error">{{ errors[field.key] }}</text>
      </view>

      <button class="submit-btn" :disabled="submitting || !fields.length" @tap="submitForm">
        {{ submitting ? '提交中...' : '确认提交' }}
      </button>
    </view>
  </view>
</template>

<style scoped>
.form-page {
  min-height: 100vh;
  padding: 24rpx 24rpx 48rpx;
  background: #f2f6f5;
  box-sizing: border-box;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 16rpx 0 28rpx;
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: #fff;
  color: #102a55;
  font-size: 52rpx;
  line-height: 68rpx;
  box-shadow: 0 8rpx 20rpx rgba(0, 0, 0, 0.08);
}

.page-title-wrap {
  flex: 1;
}

.page-title {
  display: block;
  color: #102a55;
  font-size: 36rpx;
  font-weight: 900;
}

.page-subtitle {
  display: block;
  margin-top: 8rpx;
  color: #6b7c78;
  font-size: 24rpx;
  line-height: 1.5;
}

.loading-card,
.form-card {
  border-radius: 32rpx;
  background: #fff;
  box-shadow: 0 18rpx 48rpx rgba(0, 77, 64, 0.05);
}

.loading-card {
  padding: 80rpx 32rpx;
  text-align: center;
  color: #6b7c78;
  font-size: 28rpx;
}

.form-card {
  padding: 32rpx;
}

.field-wrap {
  margin-bottom: 32rpx;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-bottom: 14rpx;
  color: #102a55;
  font-size: 26rpx;
  font-weight: 800;
}

.required {
  color: #e53e3e;
}

.field-input,
.field-picker {
  box-sizing: border-box;
  width: 100%;
  height: 88rpx;
  padding: 0 28rpx;
  border: 2rpx solid rgba(0, 77, 64, 0.15);
  border-radius: 24rpx;
  background: #f8fbfa;
  color: #102a55;
  font-size: 28rpx;
  line-height: 88rpx;
}

.field-textarea {
  box-sizing: border-box;
  width: 100%;
  height: 180rpx;
  min-height: 180rpx;
  padding: 22rpx 28rpx;
  border: 2rpx solid rgba(0, 77, 64, 0.15);
  border-radius: 24rpx;
  background: #f8fbfa;
  color: #102a55;
  font-size: 28rpx;
  line-height: 40rpx;
}

.error {
  border-color: #e53e3e;
  background: #fffafa;
}

.field-picker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #9aa3b5;
}

.field-picker.filled {
  color: #102a55;
}

.picker-arrow {
  color: #9aa3b5;
  font-size: 36rpx;
}

.field-error {
  display: block;
  margin-top: 10rpx;
  color: #e53e3e;
  font-size: 24rpx;
}

.upload-field {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.upload-box {
  position: relative;
  width: 180rpx;
  height: 180rpx;
  border: 2rpx dashed rgba(0, 77, 64, 0.22);
  border-radius: 24rpx;
  overflow: hidden;
  background: #f8fbfa;
}

.upload-preview {
  width: 100%;
  height: 100%;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #7f8f8b;
}

.upload-plus {
  font-size: 48rpx;
  line-height: 1;
}

.upload-text,
.upload-change {
  color: #004d40;
  font-size: 24rpx;
}

.upload-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.42);
  color: #fff;
  font-size: 24rpx;
}

.unsupported {
  display: flex;
  align-items: center;
  color: #9aa3b5;
}

.submit-btn {
  width: 100%;
  height: 92rpx;
  margin-top: 40rpx;
  border: none;
  border-radius: 46rpx;
  background: linear-gradient(135deg, #004d40, #00695c);
  color: #fff;
  font-size: 30rpx;
  font-weight: 900;
  line-height: 92rpx;
}
</style>
