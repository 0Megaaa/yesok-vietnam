<script setup>
import { computed, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useClientStore } from '@/store/client'
import { get, post } from '@/api/request'
import AuthPopup from '@/components/AuthPopup.vue'
import { useGlobalShare } from '@/composables/useGlobalShare'

const client = useClientStore()
const serviceId = ref('')
const serviceData = ref(null)
const loading = ref(true)
const submitting = ref(false)

// 动态表单相关
const formSchema = ref({ fields: [] }) // {fields: [{name, label, type, required, placeholder, options}]}
const formValues = ref({})              // 用户填写的表单数据
const formErrors = ref({})              // 字段级错误信息
const showForm = ref(false)             // 是否展示下单表单弹窗
const submitAction = ref(null)         // 当前提交动作信息

// 调试日志：监控 showForm 变化
watch(showForm, (val) => {
  console.log('[service-detail] showForm changed:', val)
})

const includes = computed(() => [
  '需求沟通与方案确认',
  '材料清单整理与中文说明',
  '本地资源匹配与预约协调',
  '办理进度实时反馈',
  '交付后注意事项提醒',
])

const steps = computed(() => [
  { no: '1', title: '填写需求', desc: '按服务类型填写专属表单，管家实时接收。' },
  { no: '2', title: '管家评估', desc: '专属管家确认可行性、材料清单与报价。' },
  { no: '3', title: '开始办理', desc: '进入动态工作流节点，按状态推进订单。' },
  { no: '4', title: '交付验收', desc: '完成服务后提交结果并同步后续注意事项。' },
])

useGlobalShare({ title: 'YesOK越南管家｜服务详情', path: '/pages/home/index' })

// unwrapResponse 解包 request.js 返回的 {data, status} 包装结构
const unwrapResponse = (res) => {
  if (!res) return {}
  return res.data ?? res
}

// formatPrice 格式化金额为人民币展示
const formatPrice = (amount) => {
  const n = Number(amount || 0)
  if (!n) return '面议'
  return `¥${n.toLocaleString('zh-CN')}`
}

// normalizeServiceDetail 以 /client/services/:id 接口返回为准规范化服务详情
const normalizeServiceDetail = (raw = {}) => {
  const serviceName = raw.service_name || raw.serviceName || raw.display_name || raw.name || '服务详情'
  const displayName = raw.display_name || raw.displayName || serviceName

  return {
    ...raw,
    id: raw.id || raw.service_id,
    service_id: raw.service_id || raw.id,
    service_code: raw.service_code || raw.serviceCode || raw.code || '',
    service_name: serviceName,
    display_name: displayName,
    title: serviceName,
    icon: raw.icon || '📋',
    cover_image: raw.cover_image || raw.coverImage || '',
    description: raw.description || raw.desc || '该服务由 YesOK 越南管家提供中文一站式协助。',
    base_price: Number(raw.base_price ?? raw.basePrice ?? 0),
    price_text: formatPrice(raw.base_price ?? raw.basePrice),
    unit: raw.unit || '次',
  }
}

onLoad((options) => {
  serviceId.value = options.id || ''
  loadService(serviceId.value)
})

// loadService 加载服务详情，同时从 init-form 接口获取下单表单配置。
// 表单数据来源：优先取服务 start 节点的 form_fields，回退到 sys_services.form_schema。
const loadService = async (id) => {
  loading.value = true
  try {
    const [svcRes, formRes] = await Promise.all([
      get(`/v1/client/services/${id}`),
      get(`/v1/client/services/${id}/init-form`).catch(() => ({ data: { form_fields: [], source: 'none' } })),
    ])

    const svcPayload = unwrapResponse(svcRes)
    const formPayload = unwrapResponse(formRes)

    serviceData.value = normalizeServiceDetail(svcPayload)
    parseFormSchema(formPayload)
  } catch {
    serviceData.value = {
      id: 'unknown',
      icon: '📋',
      bg: '#E3F0FF',
      name: '服务详情',
      price: '面议',
      unit: '',
      tags: ['专属管家'],
      description: '该服务详情正在准备中，请联系管家获取更多信息。',
    }
    formSchema.value = { fields: [] }
  } finally {
    loading.value = false
  }
}

// normalizeFieldOptions 规范化字段选项
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

// normalizeFields 规范化字段列表
const normalizeFields = (fields = []) => {
  if (!Array.isArray(fields)) return []
  return fields.map((f) => {
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
  }).filter(f => f.key)
}

// parseFormSchema 从 init-form 接口响应中解析 form_fields。
// 兼容多种返回格式：
// 1. 平铺格式：{action_name, button_label, form_fields}
// 2. 包装格式：{action: {action_name, form_fields}, form_fields}
// 3. form_schema 格式：{form_schema: {fields: []}}
const parseFormSchema = (formPayload = {}) => {
  const payload = unwrapResponse(formPayload)

  const action = payload.action || {
    action_name: payload.action_name,
    button_label: payload.button_label,
    action_type: payload.action_type,
    target_status: payload.target_status,
    macro_status: payload.macro_status,
    notify_type: payload.notify_type,
    form_fields: payload.form_fields || [],
  }

  submitAction.value = action?.action_name ? action : null

  const rawFields =
    action?.form_fields ||
    payload.form_fields ||
    payload.form_schema?.fields ||
    payload.fields ||
    []

  const fields = normalizeFields(rawFields)

  formSchema.value = { fields }

  const values = {}
  const errors = {}
  fields.forEach((f) => {
    values[f.key] = ''
    errors[f.key] = ''
  })
  formValues.value = values
  formErrors.value = errors
}

// validateField 校验单个字段。
const validateField = (field) => {
  const k = field.key || field.name
  const val = (formValues.value[k] || '').toString().trim()
  if (field.required && !val) {
    formErrors.value[k] = `请填写 ${field.label}`
    return false
  }
  delete formErrors.value[k]
  return true
}

// validateAll 校验所有必填字段。
// 返回 true 表示全部通过。
const validateAll = () => {
  let ok = true
  ;(formSchema.value.fields || []).forEach((f) => {
    if (!validateField(f)) ok = false
  })
  return ok
}

// getSelectedLabel 获取 select 字段的显示文本
const getSelectedLabel = (field) => {
  const key = field.key || field.name
  const value = formValues.value[key]
  const option = (field.options || []).find((item) => item.value === value)
  return option?.label || value || field.placeholder || `请选择${field.label}`
}

// openOrderForm 打开下单表单弹窗。
const openOrderForm = () => {
  const serviceName = serviceData.value?.service_name || serviceData.value?.display_name || '该服务'
  if (!client.checkAuth(`预约「${serviceName}」`)) return
  showForm.value = true
}

// closeForm 关闭下单表单弹窗。
const closeForm = () => {
  showForm.value = false
  formErrors.value = {}
}

// submitOrder 提交订单。
const submitOrder = async () => {
  if (!validateAll()) return
  submitting.value = true
  try {
    const payload = {
      service_id: serviceData.value.id,
      contact_name: formValues.value.contact_name || client.userInfo?.nickname || '微信客户',
      contact_phone: formValues.value.contact_phone || '',
      form_data: { ...formValues.value },
    }
    const res = await post('/v1/client/orders', payload)
    // 正确提取订单 ID，只使用数字 ID 跳转
    const order = res.order || res.data?.order || {}
    const orderId = order.id

    client.addOrder(order)
    showForm.value = false
    safeToast('订单已提交，管家即将联系您', 'success')

    // 只用数字 ID 跳转，因为后端按数字 ID 查询
    const uniApi = typeof uni !== 'undefined' ? uni : null
    if (uniApi?.navigateTo && orderId) {
      setTimeout(() => {
        uniApi.navigateTo({ url: `/pages/order-detail/index?id=${orderId}` })
      }, 1500)
    }

    // 重置表单
    const values = {}
    ;(formSchema.value.fields || []).forEach((f) => { values[f.name || f.key] = '' })
    formValues.value = values
    formErrors.value = {}
  } catch (err) {
    safeToast(err?.message || '订单提交失败', 'error')
  } finally {
    submitting.value = false
  }
}

const goBack = () => {
  const pages = getCurrentPages()
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (pages.length > 1) {
    if (uniApi?.navigateBack) uniApi.navigateBack()
  } else {
    if (uniApi?.switchTab) uniApi.switchTab({ url: '/pages/home/index' })
  }
}

const contactManager = () => {
  const serviceName = serviceData.value?.service_name || serviceData.value?.display_name || '服务详情'
  if (!client.checkAuth(`咨询「${serviceName}」`)) return
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.navigateTo) uniApi.navigateTo({ url: `/pages/chat/index?svc=${encodeURIComponent(serviceName)}` })
}

const safeToast = (title, icon = 'info') => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showToast) {
    uniApi.showToast({ title, icon })
  } else {
    console.info(`[Toast] ${title}`)
  }
}
</script>

<template>
  <view class="detail-page">
    <!-- Hero -->
    <view class="hero" :style="{ background: serviceData?.bg || serviceData?.cover_image || '#E3F0FF' }">
      <view class="nav-back" @click="goBack">‹</view>
      <text class="hero-icon">{{ serviceData?.icon || '📋' }}</text>
    </view>

    <!-- 服务信息 -->
    <view class="title-card">
      <view class="title-row">
        <view class="service-icon-wrap" :style="{ background: serviceData?.bg || '#E3F0FF' }">
          <text class="service-icon">{{ serviceData?.icon || '📋' }}</text>
        </view>
        <view class="title-main">
          <text class="service-title">{{ serviceData?.service_name || serviceData?.display_name || '服务详情' }}</text>
          <view class="tag-row">
            <text v-for="tag in (serviceData?.tags || [])" :key="tag" class="tag">{{ tag }}</text>
          </view>
        </view>
      </view>
      <view class="price-row">
        <text class="price">{{ serviceData?.price_text || '面议' }}</text>
        <text class="unit">/{{ serviceData?.unit || '次' }}</text>
      </view>
      <text class="desc">{{ serviceData?.description || serviceData?.desc || '该服务由 YesOK 越南管家提供中文一站式协助。' }}</text>
    </view>

    <!-- 包含服务 -->
    <view class="section-card">
      <view class="section-title"><view class="section-bar"></view>包含服务</view>
      <view v-for="item in includes" :key="item" class="include-item">
        <text class="check">✓</text>
        <text>{{ item }}</text>
      </view>
    </view>

    <!-- 办理流程 -->
    <view class="section-card">
      <view class="section-title"><view class="section-bar"></view>办理流程</view>
      <view v-for="step in steps" :key="step.no" class="step-item">
        <view class="step-no">{{ step.no }}</view>
        <view class="step-content">
          <text class="step-title">{{ step.title }}</text>
          <text class="step-desc">{{ step.desc }}</text>
        </view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar">
      <button class="ghost-btn" @click="contactManager">去咨询</button>
      <button class="primary-btn" @click="openOrderForm">立即预约</button>
    </view>

    <!-- 下单表单弹窗 -->
    <view
      v-if="showForm"
      class="form-overlay"
      @touchmove.stop.prevent
    >
      <view class="form-mask" @tap="closeForm"></view>

      <view
        class="form-sheet"
        @tap.stop
        @click.stop
      >
        <!-- 弹窗头部 -->
        <view class="form-header">
          <text class="form-title">填写预约信息</text>
          <view class="form-close" @tap.stop="closeForm">✕</view>
        </view>

        <!-- 动态表单字段 -->
        <scroll-view
          scroll-y
          class="form-body"
          :enhanced="true"
          :show-scrollbar="false"
          @tap.stop
          @click.stop
        >
          <view v-if="!formSchema.fields.length" class="empty-form">
            当前服务暂未配置预约表单，请联系管家处理。
          </view>
          <view v-for="field in formSchema.fields" :key="field.key || field.name" class="field-wrap">
            <!-- 标签 -->
            <view class="field-label">
              <text>{{ field.label }}</text>
              <text v-if="field.required" class="required-star">*</text>
            </view>

            <!-- text / phone / number -->
            <input
              v-if="field.type === 'text' || field.type === 'phone'"
              v-model="formValues[field.key || field.name]"
              :type="field.type === 'phone' ? 'number' : 'text'"
              class="field-input"
              :class="{ error: formErrors[field.key || field.name] }"
              :placeholder="field.placeholder || `请输入${field.label}`"
              :adjust-position="true"
              cursor-spacing="80"
              @tap.stop
              @click.stop
              @blur="validateField(field)"
            />
            <input
              v-else-if="field.type === 'number'"
              v-model="formValues[field.key || field.name]"
              type="digit"
              class="field-input"
              :class="{ error: formErrors[field.key || field.name] }"
              :placeholder="field.placeholder || `请输入${field.label}`"
              :adjust-position="true"
              cursor-spacing="80"
              @tap.stop
              @click.stop
              @blur="validateField(field)"
            />

            <!-- date -->
            <input
              v-else-if="field.type === 'date'"
              v-model="formValues[field.key || field.name]"
              class="field-input"
              :class="{ error: formErrors[field.key || field.name] }"
              placeholder="格式：2025-01-15"
              :adjust-position="true"
              cursor-spacing="80"
              @tap.stop
              @click.stop
              @blur="validateField(field)"
            />

            <!-- datetime -->
            <input
              v-else-if="field.type === 'datetime'"
              v-model="formValues[field.key || field.name]"
              class="field-input"
              :class="{ error: formErrors[field.key || field.name] }"
              placeholder="格式：2025-01-15 14:30"
              :adjust-position="true"
              cursor-spacing="80"
              @tap.stop
              @click.stop
              @blur="validateField(field)"
            />

            <!-- textarea -->
            <textarea
              v-else-if="field.type === 'textarea'"
              v-model="formValues[field.key || field.name]"
              class="field-textarea"
              :class="{ error: formErrors[field.key || field.name] }"
              :placeholder="field.placeholder || `请输入${field.label}`"
              :auto-height="true"
              cursor-spacing="80"
              @tap.stop
              @click.stop
              @blur="validateField(field)"
            />

            <!-- select -->
            <picker
              v-else-if="field.type === 'select'"
              mode="selector"
              :value="0"
              :range="field.options || []"
              :range-key="'label'"
              @tap.stop
              @click.stop
              @change="(e) => {
                const selected = field.options[e.detail.value]
                formValues[field.key || field.name] = selected?.value ?? selected ?? ''
                validateField(field)
              }"
            >
              <view class="field-picker" :class="{ error: formErrors[field.key || field.name], filled: formValues[field.key || field.name] }">
                <text>{{ getSelectedLabel(field) }}</text>
                <text class="picker-arrow">›</text>
              </view>
            </picker>

            <!-- file (simple URL input) -->
            <view
              v-else-if="field.type === 'file'"
              class="field-file"
            >
              <input
                v-model="formValues[field.key || field.name]"
                class="field-input"
                style="height: auto; padding: 10px 14px;"
                placeholder="请输入文件URL，或联系管家上传"
                :adjust-position="true"
                cursor-spacing="80"
                @tap.stop
                @click.stop
              />
            </view>

            <!-- image (URL input) -->
            <view
              v-else-if="field.type === 'image'"
              class="field-file"
            >
              <input
                v-model="formValues[field.key || field.name]"
                class="field-input"
                style="height: auto; padding: 10px 14px;"
                placeholder="请输入图片URL，或联系管家上传"
                :adjust-position="true"
                cursor-spacing="80"
                @tap.stop
                @click.stop
              />
            </view>

            <!-- 错误提示 -->
            <text v-if="formErrors[field.key || field.name]" class="field-error">{{ formErrors[field.key || field.name] }}</text>
          </view>
        </scroll-view>

        <!-- 提交按钮 -->
        <view class="form-footer" @tap.stop @click.stop>
          <button
            class="submit-btn"
            :disabled="submitting || !formSchema.fields.length"
            @tap.stop="submitOrder"
          >
            {{ submitting ? '提交中...' : '确认提交' }}
          </button>
        </view>
      </view>
    </view>

    <AuthPopup />
  </view>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding-bottom: 92px;
  background: #f2f6f5;
}

.hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 220px;
}

.nav-back {
  position: absolute;
  top: 16px;
  left: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  color: #102a55;
  font-size: 28px;
  line-height: 34px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.hero-icon { font-size: 88px; }

.title-card,
.section-card {
  margin: 10px 12px 0;
  padding: 16px;
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 1px 8px rgba(0, 0, 0, 0.05);
}

.title-row { display: flex; gap: 12px; align-items: flex-start; }

.service-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 50px;
  height: 50px;
  border-radius: 14px;
  font-size: 24px;
}

.title-main { flex: 1; }

.service-title {
  display: block;
  margin-bottom: 8px;
  color: #102a55;
  font-size: 19px;
  font-weight: 800;
}

.tag-row { display: flex; flex-wrap: wrap; gap: 6px; }

.tag {
  padding: 3px 8px;
  border-radius: 12px;
  background: #f0f6ff;
  color: #0d47a1;
  font-size: 10px;
  font-weight: 700;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin: 14px 0 10px;
}

.price {
  color: #e97832;
  font-size: 23px;
  font-weight: 900;
}

.unit { color: #9aa3b5; font-size: 13px; }

.desc {
  display: block;
  color: #4a5568;
  font-size: 13px;
  line-height: 1.8;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  color: #102a55;
  font-size: 16px;
  font-weight: 800;
}

.section-bar {
  width: 4px;
  height: 16px;
  border-radius: 4px;
  background: #004d40;
}

.include-item {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f6ff;
  color: #4a5568;
  font-size: 13px;
}

.check { color: #2e7d32; font-weight: 900; }

.step-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f6ff;
}

.step-no {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #004d40;
  color: #fff;
  font-size: 13px;
  font-weight: 800;
}

.step-content { flex: 1; }

.step-title {
  display: block;
  color: #102a55;
  font-size: 14px;
  font-weight: 800;
}

.step-desc {
  display: block;
  margin-top: 4px;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.6;
}

/* 底部操作栏 */
.bottom-bar {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 20;
  display: flex;
  gap: 12px;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 -2px 16px rgba(0, 0, 0, 0.08);
}

.ghost-btn,
.primary-btn {
  flex: 1;
  height: 44px;
  margin: 0;
  border-radius: 22px;
  font-size: 15px;
  font-weight: 800;
  line-height: 44px;
}

.ghost-btn {
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
}

.primary-btn {
  color: #fff;
  background: linear-gradient(135deg, #004d40, #00695c);
}

/* 下单表单弹窗 */
.form-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  background: transparent;
}

.form-mask {
  position: absolute;
  inset: 0;
  z-index: 1;
  background: rgba(0, 0, 0, 0.36);
}

.form-sheet {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: 88vh;
  border-radius: 28px 28px 0 0;
  background: #fff;
  box-shadow: 0 -18px 60px rgba(0, 0, 0, 0.18);
}

.form-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 20px 20px 16px;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
}

.form-title {
  color: #102a55;
  font-size: 17px;
  font-weight: 900;
}

.form-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #f2f6f5;
  color: #6b7c78;
  font-size: 14px;
}

.form-body {
  flex: 1;
  padding: 16px 18px;
  max-height: 60vh;
}

/* 动态字段 */
.field-wrap {
  margin-bottom: 18px;
}

.empty-form {
  padding: 40px 20px;
  text-align: center;
  color: #6b7c78;
  font-size: 14px;
  line-height: 1.8;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 8px;
  color: #102a55;
  font-size: 13px;
  font-weight: 800;
}

.required-star { color: #e53e3e; }

.field-input {
  box-sizing: border-box;
  width: 100%;
  height: 44px;
  padding: 0 14px;
  border: 1.5px solid rgba(0, 77, 64, 0.15);
  border-radius: 14px;
  background: #f8fbfa;
  color: #102a55;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.field-input:focus { border-color: #004d40; }
.field-input.error { border-color: #e53e3e; }

.field-textarea {
  box-sizing: border-box;
  width: 100%;
  min-height: 80px;
  padding: 10px 14px;
  border: 1.5px solid rgba(0, 77, 64, 0.15);
  border-radius: 14px;
  background: #f8fbfa;
  color: #102a55;
  font-size: 14px;
  outline: none;
  resize: none;
  transition: border-color 0.2s;
}

.field-textarea:focus { border-color: #004d40; }
.field-textarea.error { border-color: #e53e3e; }

.field-picker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  width: 100%;
  height: 44px;
  padding: 0 14px;
  border: 1.5px solid rgba(0, 77, 64, 0.15);
  border-radius: 14px;
  background: #f8fbfa;
  color: #9aa3b5;
  font-size: 14px;
  transition: border-color 0.2s;
}

.field-picker.filled { color: #102a55; }
.field-picker.error { border-color: #e53e3e; }

.field-file {
  padding: 12px;
  border: 1.5px dashed rgba(0, 77, 64, 0.2);
  border-radius: 14px;
  background: #f8fbfa;
}

.file-placeholder {
  color: #9aa3b5;
  font-size: 12px;
}

.picker-arrow {
  color: #9aa3b5;
  font-size: 18px;
  font-weight: 700;
}

.field-error {
  display: block;
  margin-top: 5px;
  color: #e53e3e;
  font-size: 11px;
}

/* 提交按钮 */
.form-footer {
  flex-shrink: 0;
  padding: 12px 18px calc(12px + env(safe-area-inset-bottom));
  border-top: 1px solid rgba(0, 77, 64, 0.08);
}

.submit-btn {
  width: 100%;
  height: 46px;
  border: none;
  border-radius: 23px;
  background: linear-gradient(135deg, #004d40, #00695c);
  color: #fff;
  font-size: 15px;
  font-weight: 900;
  line-height: 46px;
}

.submit-btn[disabled] {
  opacity: 0.6;
}
</style>
