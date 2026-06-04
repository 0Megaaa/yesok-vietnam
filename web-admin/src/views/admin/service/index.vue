<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import { getDictHelp } from '@/utils/workflowDictMeta'

const loading = ref(true)
const services = ref([])
const categoryDictOptions = ref([])

const commonEmojis = ['🌴', '🎫', '🏦', '🏠', '✈️', '🚗', '📄', '💼', '🩺', '🎓', '🍜', '🌟']

// --- 查询参数 ---
const queryParams = ref({ service_name: '', status: '', is_hot: '' })

// --- 弹窗与表单状态 ---
const dialogVisible = ref(false)
const dialogType = ref('add')
const currentId = ref(null)
const submitting = ref(false)

// Wizard 当前步骤（0=基础信息，1=工作流节点）
const currentStep = ref(0)

// 复合表单对象
const form = ref({
  service_info: blankServiceInfo(),
  workflow_nodes: [],
})

function blankServiceInfo() {
  return {
    id: 0,
    service_code: '',
    service_name: '',
    display_name: '',
    icon: '🌴',
    cover_image: '',
    description: '',
    base_price: 0,
    currency: 'VND',
    unit: '次',
    sort_order: 0,
    status: 1,
    is_hot: false,
    form_schema: '{"fields":[]}',
  }
}

// --- 工具函数 ---
const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const resetForm = () => {
  currentId.value = null
  currentStep.value = 0
  form.value = {
    service_info: blankServiceInfo(),
    workflow_nodes: [],
  }
}

// --- 字典加载 ---
const loadCategoryDict = async () => {
  try {
    const res = await request.get('/v1/admin/dict-data?dict_type=service_category')
    const rawData = res?.data?.data ?? res?.data ?? res ?? []
    categoryDictOptions.value = Array.isArray(rawData) ? rawData : []
  } catch (e) {
    console.error('加载服务分类字典失败', e)
  }
}

const getCategoryLabel = (row) => {
  if (!row) return '未分类'
  const code = row.service_code || ''
  if (!categoryDictOptions.value?.length) return code || '未分类'
  const found = categoryDictOptions.value.find((item) => item.dict_value === code)
  return found ? found.dict_label : (code || '未分类')
}

// --- 列表加载 ---
const loadServices = async () => {
  loading.value = true
  try {
    const params = {}
    if (queryParams.value.service_name) params.service_name = queryParams.value.service_name
    if (queryParams.value.status !== '') params.status = queryParams.value.status
    if (queryParams.value.is_hot !== '') params.is_hot = queryParams.value.is_hot
    const res = await request.get('/v1/admin/services', { params })
    let targetList = null
    if (res && Array.isArray(res.data)) targetList = res.data
    else if (res?.data?.list) targetList = res.data.list
    else if (Array.isArray(res)) targetList = res
    else if (res?.list) targetList = res.list
    services.value = targetList || []
  } catch (e) {
    showToast(e?.message || '服务列表加载失败', 'error')
  } finally {
    loading.value = false
  }
}

// --- 图片上传 ---
const handleImageUpload = async (fileOptions) => {
  const fd = new FormData()
  fd.append('file', fileOptions.file)
  try {
    const res = await request.post('/v1/admin/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
    const url = res.data?.url || res.url
    if (url) { form.value.service_info.cover_image = url; showToast('图片上传成功', 'success') }
  } catch (e) {
    showToast('图片上传失败', 'error')
  }
}

// --- 工作流节点 CRUD（对齐 sys_workflow_nodes 新 Schema）---
const blankNode = () => ({
  id: 0,
  stage_code: '',
  stage_name: '',
  executor_role: 'admin',
  action_name: '',
  button_label: '',
  action_type: 'button_click',
  form_fields: [],
  target_status: '',
  macro_status: '',
  notify_type: '',
  need_audit: false,
  audit_reject_status: '',
  sort_order: 0,
})

const addNode = () => {
  const node = blankNode()
  node.sort_order = form.value.workflow_nodes.length + 1
  form.value.workflow_nodes.push(node)
}

const removeNode = (index) => {
  form.value.workflow_nodes.splice(index, 1)
}

// form_fields 动态编辑器状态（行内展开）
const fieldEditor = ref({ visible: false, nodeIndex: -1 })

const openFieldEditor = (nodeIndex) => {
  fieldEditor.value = { visible: true, nodeIndex }
}

const closeFieldEditor = () => {
  fieldEditor.value = { visible: false, nodeIndex: -1 }
}

const currentEditNode = computed(() => {
  if (fieldEditor.value.nodeIndex < 0) return null
  return form.value.workflow_nodes[fieldEditor.value.nodeIndex] || null
})

const addFieldRow = () => {
  if (!currentEditNode.value) return
  currentEditNode.value.form_fields = currentEditNode.value.form_fields || []
  currentEditNode.value.form_fields.push({ key: '', label: '', type: 'text', required: false, options: [] })
}

const removeFieldRow = (fieldIndex) => {
  if (!currentEditNode.value) return
  currentEditNode.value.form_fields.splice(fieldIndex, 1)
}

const fieldTypeOptions = [
  { label: '文本输入', value: 'text' },
  { label: '多行文本', value: 'textarea' },
  { label: '数字', value: 'number' },
  { label: '日期', value: 'date' },
  { label: '日期时间', value: 'datetime' },
  { label: '下拉选择', value: 'select' },
  { label: '图片上传', value: 'image' },
  { label: '文件上传', value: 'file' },
  { label: '手机号', value: 'phone' },
]

const openFieldOptionsDialog = (field, fieldIndex) => {
  const raw = JSON.stringify(field.options || [], null, 2)
  const input = window.prompt(`字段「${field.label}」的 options JSON（格式：[{"label":"选项1","value":"val1"}]）`, raw)
  if (input === null) return
  try {
    currentEditNode.value.form_fields[fieldIndex].options = JSON.parse(input)
  } catch {
    window.alert('JSON 格式错误')
  }
}

// --- 提交前置校验 ---
const validateWorkflowNodes = () => {
  const nodes = form.value.workflow_nodes
  // 检查重复：同一 service_id + stage_code + executor_role + action_name 不允许重复
  const seen = new Map()
  for (let i = 0; i < nodes.length; i++) {
    const n = nodes[i]
    // 基础必填校验
    if (!n.stage_code?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「当前节点」不能为空` }
    if (!n.stage_name?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「节点名称」不能为空` }
    if (!n.executor_role?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「执行角色」不能为空` }
    if (!n.action_name?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「动作名称」不能为空` }
    if (!n.button_label?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「按钮文案」不能为空` }
    if (!n.action_type?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「动作类型」不能为空` }
    if (!n.target_status?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「下一节点」不能为空` }
    if (!n.macro_status?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行「执行后主状态」不能为空` }

    // form_input 类型的 form_fields 校验
    if (n.action_type === 'form_input') {
      if (!n.form_fields || n.form_fields.length === 0) {
        return { ok: false, index: i, msg: `第 ${i + 1} 行「form_input」类型必须有表单字段` }
      }
      for (let j = 0; j < n.form_fields.length; j++) {
        const f = n.form_fields[j]
        if (!f.key?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行第 ${j + 1} 个字段「key」不能为空` }
        if (!f.label?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行第 ${j + 1} 个字段「label」不能为空` }
        if (!f.type?.trim()) return { ok: false, index: i, msg: `第 ${i + 1} 行第 ${j + 1} 个字段「type」不能为空` }
      }
    } else {
      // 非 form_input 类型，form_fields 置空
      n.form_fields = []
    }

    // notify_type 默认为 none
    if (!n.notify_type?.trim()) {
      n.notify_type = 'none'
    }

    // 重复检查
    const key = `${n.stage_code}_${n.executor_role}_${n.action_name}`
    if (seen.has(key)) {
      return { ok: false, index: i, msg: `第 ${i + 1} 行与第 ${seen.get(key) + 1} 行：同一节点同一角色下动作重复，请调整` }
    }
    seen.set(key, i)
  }
  return { ok: true }
}

// --- Wizard 导航 ---
const nextStep = () => {
  if (currentStep.value === 0) {
    if (!form.value.service_info.service_name?.trim()) {
      return showToast('服务名称不能为空', 'warning')
    }
    if (!form.value.service_info.service_code) {
      return showToast('服务类型不能为空', 'warning')
    }
    currentStep.value = 1
  }
}

const prevStep = () => {
  if (currentStep.value > 0) currentStep.value--
}

// --- 打开新增弹窗 ---
const openAddDialog = () => {
  dialogType.value = 'add'
  resetForm()
  dialogVisible.value = true
}

// --- 打开编辑弹窗：调用聚合接口获取完整配置 ---
const openEditDialog = async (row) => {
  dialogType.value = 'edit'
  currentId.value = row.id
  submitting.value = true
  try {
    // 单次调用聚合接口获取 service_info + workflow_nodes
    const res = await request.get(`/v1/admin/services/${row.id}`)
    const payload = res.data || res

    const svc = payload.service_info || {}
    const nodes = payload.workflow_nodes || []

    form.value = {
      service_info: {
        id: svc.id || 0,
        service_code: svc.service_code || '',
        service_name: svc.service_name || '',
        display_name: svc.display_name || '',
        icon: svc.icon || '🌴',
        cover_image: svc.cover_image || '',
        description: svc.description || '',
        base_price: svc.base_price || 0,
        currency: svc.currency || 'VND',
        unit: svc.unit || '次',
        sort_order: svc.sort_order || 0,
        status: svc.status ?? 1,
        is_hot: svc.is_hot ?? false,
        form_schema: typeof svc.form_schema === 'string' ? svc.form_schema : JSON.stringify(svc.form_schema || { fields: [] }),
      },
      workflow_nodes: nodes.map((n) => ({
        id: n.id || 0,
        stage_code: n.stage_code || '',
        stage_name: n.stage_name || '',
        executor_role: n.executor_role || 'admin',
        action_name: n.action_name || '',
        button_label: n.button_label || '',
        action_type: n.action_type || 'button_click',
        form_fields: n.form_fields || [],
        target_status: n.target_status || '',
        macro_status: n.macro_status || '',
        notify_type: n.notify_type || '',
        need_audit: n.need_audit ?? false,
        audit_reject_status: n.audit_reject_status || '',
        sort_order: n.sort_order ?? 0,
      })),
    }
    currentStep.value = 0
    dialogVisible.value = true
  } catch (e) {
    showToast('加载服务详情失败', 'error')
  } finally {
    submitting.value = false
  }
}

// --- 保存服务（含节点）：一条龙复合 Payload ---
const saveService = async () => {
  // 工作流节点校验（空数组允许提交）
  const v = validateWorkflowNodes()
  if (!v.ok) return showToast(v.msg, 'warning')

  submitting.value = true
  try {
    const payload = {
      service_info: {
        ...form.value.service_info,
        base_price: Number(form.value.service_info.base_price) || 0,
      },
      workflow_nodes: form.value.workflow_nodes.map((node) => ({
        ...node,
        need_audit: !!node.need_audit,
        audit_reject_status: node.need_audit
          ? (node.audit_reject_status || 'wait_supplement')
          : (node.audit_reject_status || ''),
      })),
    }
    if (dialogType.value === 'add') {
      await request.post('/v1/admin/services', payload)
    } else {
      await request.put(`/v1/admin/services/${currentId.value}`, payload)
    }
    showToast(dialogType.value === 'add' ? '服务创建成功' : '服务更新成功', 'success')
    dialogVisible.value = false
    resetForm()
    await loadServices()
  } catch (e) {
    showToast(e?.message || '服务保存失败', 'error')
  } finally {
    submitting.value = false
  }
}

// --- 上下架 ---
const toggleService = async (row) => {
  const targetStatus = row.status === 1 ? 0 : 1
  try {
    await request.put(`/v1/admin/services/${row.id}`, {
      service_info: { ...row, status: targetStatus },
      workflow_nodes: [],
    })
    showToast(`服务已${targetStatus === 1 ? '上架' : '下架'}`, 'success')
    await loadServices()
  } catch (e) {
    showToast('更新服务状态失败', 'error')
  }
}

// 工作流节点字典数据
const macroStatusOptions = ref([])
const nodeStageOptions = ref([])
const workflowActionOptions = ref([])
const actionTypeOptions = ref([])
const executorRoleOptions = ref([])
const notifyTypeOptions = ref([])

const loadWorkflowDicts = async () => {
  const [ms, ns, ac, at, er, nt] = await Promise.all([
    request.get('/v1/admin/dict-data?dict_type=macro_status&status=1&pageSize=200').catch(() => null),
    request.get('/v1/admin/dict-data?dict_type=node_stage&status=1&pageSize=200').catch(() => null),
    request.get('/v1/admin/dict-data?dict_type=workflow_action&status=1&pageSize=200').catch(() => null),
    request.get('/v1/admin/dict-data?dict_type=action_type&status=1&pageSize=200').catch(() => null),
    request.get('/v1/admin/dict-data?dict_type=executor_role&status=1&pageSize=200').catch(() => null),
    request.get('/v1/admin/dict-data?dict_type=notify_type&status=1&pageSize=200').catch(() => null),
  ])
  const unwrap = (res) => {
    const raw = res?.data?.data ?? res?.data ?? res ?? []
    return Array.isArray(raw) ? raw : []
  }
  // 过滤掉 page_jump 类型
  macroStatusOptions.value = unwrap(ms)
  nodeStageOptions.value = unwrap(ns)
  workflowActionOptions.value = unwrap(ac).filter(o => o.dict_value !== 'page_jump')
  actionTypeOptions.value = unwrap(at).filter(o => ['button_click', 'form_input', 'wx_pay'].includes(o.dict_value))
  executorRoleOptions.value = unwrap(er)
  notifyTypeOptions.value = unwrap(nt)
}

const selectEmoji = (emoji) => { form.value.service_info.icon = emoji }

const handleQuery = () => loadServices()
const resetQuery = () => {
  queryParams.value = { service_name: '', status: '', is_hot: '' }
  loadServices()
}

onMounted(async () => {
  await Promise.all([loadCategoryDict(), loadWorkflowDicts(), loadServices()])
})

onUnmounted(() => { loading.value = false })
</script>

<template>
  <div class="services-page-container">
    <!-- 筛选区 -->
    <div class="filter-wrapper card-layout">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="服务名称">
          <el-input v-model="queryParams.service_name" placeholder="请输入检索的服务名" clearable @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item label="服务状态">
          <el-select v-model="queryParams.status" placeholder="全部状态" clearable style="width: 140px">
            <el-option label="已上架" :value="1" />
            <el-option label="已下架" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="热门推荐">
          <el-select v-model="queryParams.is_hot" placeholder="全部" clearable style="width: 120px">
            <el-option label="🔥 热门" :value="true" />
            <el-option label="常规" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar-wrapper">
      <el-button type="primary" class="erp-add-btn" @click="openAddDialog">+ 新增服务</el-button>
    </div>

    <!-- 表格 -->
    <div class="table-wrapper card-layout">
      <el-table v-loading="loading" :data="services" border stripe style="width: 100%">
        <el-table-column label="服务图标" width="90" align="center">
          <template #default="{ row }">
            <span class="service-icon-preview">{{ row.icon || '🌴' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="服务类型" width="160">
          <template #default="{ row }">
            <el-tag type="info" effect="plain">{{ getCategoryLabel(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="service_name" label="服务名称" width="180" />
        <el-table-column label="价格" width="150" align="right">
          <template #default="{ row }">
            <span class="price-text">￥{{ (Number(row.base_price) / 100).toLocaleString() }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="计价单位" width="100" align="center" />
        <el-table-column label="热门推荐" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_hot ? 'danger' : 'info'" effect="light">
              {{ row.is_hot ? '🔥 热门' : '常规' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="服务状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" effect="dark">
              {{ row.status === 1 ? '上架中' : '已下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="服务简介" show-overflow-tooltip min-width="240" />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEditDialog(row)">编辑服务</el-button>
            <el-button link :type="row.status === 1 ? 'danger' : 'success'" size="small" @click="toggleService(row)">
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 弹窗：向导式配置 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '✨ 新增服务' : '⚙️ 编辑服务'"
      width="760px"
      destroy-on-close
      :close-on-click-modal="false"
      @closed="resetForm"
    >
      <!-- Wizard 步骤条 -->
      <el-steps :active="currentStep" finish-status="success" style="margin-bottom: 20px">
        <el-step title="基础信息" />
        <el-step title="工作流节点" />
      </el-steps>

      <!-- 步骤 0：基础信息表单 -->
      <div v-show="currentStep === 0">
        <el-form :model="form.service_info" label-width="120px" label-position="right">
          <el-form-item label="服务类型" required>
            <el-select
              v-model="form.service_info.service_code"
              placeholder="请选择服务类型"
              style="width: 100%"
              clearable
            >
              <el-option
                v-for="item in categoryDictOptions"
                :key="item.dict_value"
                :label="item.dict_label"
                :value="item.dict_value"
              />
            </el-select>
          </el-form-item>

          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="服务名称" required>
                <el-input v-model="form.service_info.service_name" placeholder="例：越南商务签证" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="客户端展示名">
                <el-input v-model="form.service_info.display_name" placeholder="前端呈现给客户的名字" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="系统图标">
            <div class="emoji-picker-container">
              <el-input v-model="form.service_info.icon" style="width: 140px; margin-bottom: 6px" readonly />
              <div class="emoji-tray">
                <span
                  v-for="emoji in commonEmojis"
                  :key="emoji"
                  class="emoji-item"
                  :class="{ selected: form.service_info.icon === emoji }"
                  @click="selectEmoji(emoji)"
                >{{ emoji }}</span>
              </div>
            </div>
          </el-form-item>

          <el-form-item label="服务封面图">
            <div>
              <el-upload action="" :http-request="handleImageUpload" :show-file-list="false">
                <el-button size="small">上传封面图</el-button>
              </el-upload>
              <el-input
                v-model="form.service_info.cover_image"
                placeholder="或直接输入图片 URL"
                style="margin-top: 6px"
              />
            </div>
          </el-form-item>

          <el-row :gutter="20">
            <el-col :span="14">
              <el-form-item label="价格">
                <el-input v-model="form.service_info.base_price" type="number" placeholder="单位：分">
                  <template #append>元 (￥)</template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="10">
              <el-form-item label="价格单位">
                <el-input v-model="form.service_info.unit" placeholder="次 / 人 / 单" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="热门推荐">
                <el-switch v-model="form.service_info.is_hot" active-text="🔥 开启爆款" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="服务状态">
                <el-radio-group v-model="form.service_info.status">
                  <el-radio :value="1">上架</el-radio>
                  <el-radio :value="0">下架</el-radio>
                </el-radio-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="服务简介">
            <el-input
              v-model="form.service_info.description"
              type="textarea"
              :rows="3"
              maxlength="100"
              show-word-limit
              placeholder="简述业务包含的范畴"
            />
          </el-form-item>
        </el-form>
      </div>

      <!-- 步骤 1：工作流节点配置 -->
      <div v-show="currentStep === 1">
        <div class="node-tip">
          每个节点描述订单在某一阶段的状态及可执行的操作按钮。
          工作流节点非必填，可留空后直接保存服务。
        </div>
        <div class="node-actions">
          <el-button type="primary" size="small" @click="addNode">+ 添加节点</el-button>
        </div>

        <el-table
          :data="form.workflow_nodes"
          border
          stripe
          size="small"
          style="margin-top: 10px"
        >
          <!-- 排序 -->
          <el-table-column label="排序" width="70" align="center">
            <template #default="{ row }">
              <el-input-number
                v-model="row.sort_order"
                :min="1" :max="999"
                size="small"
                controls-position="right"
                style="width: 60px"
              />
            </template>
          </el-table-column>

          <!-- 当前节点 -->
          <el-table-column label="当前节点 *" min-width="140">
            <template #default="{ row }">
              <el-select v-model="row.stage_code" placeholder="选择节点" size="small" style="width: 100%" clearable allow-create filterable @change="(v) => { if(!row.stage_name && v) row.stage_name = nodeStageOptions.find(o => o.dict_value === v)?.dict_label || v }">
                <el-option v-for="o in nodeStageOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
              </el-select>
            </template>
          </el-table-column>

          <!-- 节点名称 -->
          <el-table-column label="节点名称 *" min-width="120">
            <template #default="{ row }">
              <el-input v-model="row.stage_name" placeholder="如 提交资料" size="small" />
            </template>
          </el-table-column>

          <!-- 执行角色 -->
          <el-table-column label="执行角色 *" width="110" align="center">
            <template #default="{ row }">
              <el-select v-model="row.executor_role" size="small" style="width: 100%">
                <el-option v-for="o in executorRoleOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
              </el-select>
            </template>
          </el-table-column>

          <!-- 动作名称 -->
          <el-table-column label="动作名称 *" min-width="180">
            <template #default="{ row }">
              <div style="display: flex; flex-direction: column; gap: 4px">
                <el-select v-model="row.action_name" placeholder="选择动作" size="small" clearable filterable allow-create>
                  <el-option v-for="o in workflowActionOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
                </el-select>
                <el-input v-model="row.button_label" placeholder="按钮文案，如 提交资料" size="small" />
              </div>
            </template>
          </el-table-column>

          <!-- 动作类型 -->
          <el-table-column label="动作类型 *" width="150">
            <template #default="{ row, $index }">
              <div style="display: flex; flex-direction: column; gap: 4px">
                <el-select v-model="row.action_type" size="small" style="width: 100%" @change="() => { if(row.action_type !== 'form_input') row.form_fields = [] }">
                  <el-option v-for="o in actionTypeOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
                </el-select>
                <template v-if="row.action_type === 'form_input'">
                  <el-button size="mini" type="primary" plain @click="openFieldEditor($index)">
                    {{ row.form_fields?.length ? `编辑 ${row.form_fields.length} 个字段` : '+ 添加字段' }}
                  </el-button>
                </template>
              </div>
            </template>
          </el-table-column>

          <!-- 下一节点 -->
          <el-table-column label="下一节点 *" width="120">
            <template #default="{ row }">
              <el-select v-model="row.target_status" placeholder="选择目标" size="small" style="width: 100%" clearable allow-create filterable>
                <el-option v-for="o in nodeStageOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
              </el-select>
            </template>
          </el-table-column>

          <!-- 执行后主状态 -->
          <el-table-column label="执行后主状态 *" width="130">
            <template #default="{ row }">
              <el-select v-model="row.macro_status" placeholder="选择状态" size="small" style="width: 100%">
                <el-option v-for="o in macroStatusOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
              </el-select>
            </template>
          </el-table-column>

          <!-- 通知类型 -->
          <el-table-column label="通知类型" width="120">
            <template #default="{ row }">
              <el-select v-model="row.notify_type" placeholder="无通知" size="small" style="width: 100%" clearable>
                <el-option v-for="o in notifyTypeOptions" :key="o.dict_value" :label="o.dict_label" :value="o.dict_value" />
              </el-select>
            </template>
          </el-table-column>

          <!-- 需人工审核 -->
          <el-table-column label="需审核" width="70" align="center">
            <template #default="{ row }">
              <el-tooltip content="需人工审核后才推进" placement="top">
                <el-switch v-model="row.need_audit" size="small" :disabled="row.action_type === 'wx_pay'" />
              </el-tooltip>
            </template>
          </el-table-column>

          <!-- 审核失败回退节点 -->
          <el-table-column label="审核失败回退" min-width="140">
            <template #default="{ row }">
              <el-select
                v-model="row.audit_reject_status"
                clearable
                placeholder="默认 wait_supplement"
                size="small"
                style="width: 100%"
                :disabled="!row.need_audit"
              >
                <el-option
                  v-for="item in nodeStageOptions"
                  :key="item.dict_value || item.value"
                  :label="item.dict_label || item.label"
                  :value="item.dict_value || item.value"
                />
              </el-select>
            </template>
          </el-table-column>

          <!-- 操作 -->
          <el-table-column label="操作" width="70" align="center">
            <template #default="{ $index }">
              <el-button link type="danger" size="small" @click="removeNode($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="!form.workflow_nodes.length" description="暂无节点配置（允许空），可跳过直接保存服务" style="padding: 20px 0" />
      </div>

      <template #footer>
        <span class="dialog-footer">
          <!-- Wizard 导航按钮 -->
          <template v-if="currentStep === 0">
            <el-button @click="dialogVisible = false" :disabled="submitting">取消</el-button>
            <el-button type="primary" :disabled="submitting" @click="nextStep">
              下一步
            </el-button>
          </template>
          <template v-else>
            <el-button @click="prevStep" :disabled="submitting">上一步</el-button>
            <el-button type="primary" :loading="submitting" @click="saveService">
              {{ submitting ? '保存中...' : '保存服务' }}
            </el-button>
          </template>
        </span>
      </template>
    </el-dialog>

    <!-- 动态表单字段编辑器弹窗 -->
    <el-dialog
      v-model="fieldEditor.visible"
      :title="`编辑表单字段：${currentEditNode?.button_label || ''}`"
      width="680px"
      destroy-on-close
      append-to-body
    >
      <div v-if="currentEditNode" class="field-editor">
        <div class="field-editor-tip">
          为动作「{{ currentEditNode.button_label }}」配置收集字段，用户提交后将作为订单资料存档。
        </div>
        <el-table :data="currentEditNode.form_fields || []" border size="small" style="margin-bottom: 12px">
          <el-table-column label="字段 Key" min-width="140">
            <template #default="{ row }">
              <el-input v-model="row.key" placeholder="如 passport_img" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="显示名称" min-width="120">
            <template #default="{ row }">
              <el-input v-model="row.label" placeholder="如 护照首页" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="控件类型" width="130">
            <template #default="{ row }">
              <el-select v-model="row.type" size="small" style="width: 100%" @change="() => { if(row.type !== 'select') row.options = [] }">
                <el-option v-for="o in fieldTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="必填" width="60" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.required" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="选项配置" width="110" align="center">
            <template #default="{ row, $index }">
              <el-button v-if="row.type === 'select'" size="small" type="primary" plain @click="openFieldOptionsDialog(row, $index)">
                配置选项
              </el-button>
              <span v-else class="no-options">—</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="60" align="center">
            <template #default="{ $index }">
              <el-button link type="danger" size="small" @click="removeFieldRow($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-button type="primary" plain size="small" @click="addFieldRow">+ 新增字段</el-button>
      </div>
      <template #footer>
        <el-button @click="closeFieldEditor">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.services-page-container {
  width: 100%;
  padding: 4px;
  box-sizing: border-box;
}

.card-layout {
  padding: 20px;
  border: 1px solid rgba(0, 77, 64, 0.06);
  border-radius: 12px;
  background: #ffffff;
  box-shadow: 0 4px 20px rgba(0, 50, 40, 0.02);
  margin-bottom: 16px;
}

.filter-wrapper .el-form-item { margin-bottom: 0; }

.action-bar-wrapper { margin-bottom: 14px; }

.erp-add-btn {
  background-color: #004d40 !important;
  border-color: #004d40 !important;
  font-weight: bold;
}

.service-icon-preview { font-size: 20px; }

.price-text { font-family: monospace; font-weight: bold; color: #e6a23c; }

.emoji-tray {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 6px;
  border: 1px dashed #dcdfe6;
}

.emoji-item { font-size: 20px; cursor: pointer; padding: 4px; border-radius: 4px; transition: all 0.2s; }
.emoji-item:hover, .emoji-item.selected { background: #e0f2f1; transform: scale(1.2); }

.dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }

/* 工作流节点样式 */
.node-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.6;
  padding: 8px 12px;
  background: #fafafa;
  border-radius: 6px;
  border-left: 3px solid #004d40;
  margin-bottom: 10px;
}

.node-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-bottom: 0;
}

.field-editor { display: flex; flex-direction: column; }

.field-editor-tip {
  font-size: 12px;
  color: #909399;
  background: #f5f7fa;
  border-radius: 6px;
  padding: 8px 12px;
  margin-bottom: 12px;
  border-left: 3px solid #004d40;
  line-height: 1.6;
}

.no-options { color: #c0c4cc; }
</style>
