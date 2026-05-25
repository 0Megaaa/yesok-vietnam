<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const services = ref([])
const categoryDictOptions = ref([])

// 快捷 Emoji 表情库
const commonEmojis = ['🌴', '🎫', '🏦', '🏠', '✈️', '🚗', '📄', '💼', '🩺', '🎓', '🍜', '🌟']

// --- ERP 查询参数 ---
const queryParams = ref({
  service_name: '',
  status: '',
  is_hot: ''
})

// --- 弹窗控制与表单状态 ---
const dialogVisible = ref(false)
const dialogType = ref('add') // 'add' 或 'edit'
const currentId = ref(null)

const serviceForm = ref({
  service_code: '',
  service_name: '',
  display_name: '',
  icon: '🌴',
  cover_image: '',
  description: '',
  base_price: 0,
  unit: '次',
  status: 1,
  is_hot: false,
})

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const resetForm = () => {
  currentId.value = null
  serviceForm.value = {
    service_code: '',
    service_name: '',
    display_name: '',
    icon: '🌴',
    cover_image: '',
    description: '',
    base_price: 0,
    unit: '次',
    status: 1,
    is_hot: false,
  }
}

// --- 加载服务分类字典 ---
const loadCategoryDict = async () => {
  try {
    const res = await request.get('/v1/admin/dict-data?dict_type=service_category')
    // 自动解包: res 为 { code, data: [...], msg } 或直接返回 data
    const rawData = res?.data?.data ?? res?.data ?? res ?? []
    categoryDictOptions.value = Array.isArray(rawData) ? rawData : []
  } catch (error) {
    console.error('加载服务分类字典失败:', error)
  }
}

// --- 【核心安全重构】绝对不卡死，找不到就吐出原始编码 ---
const getCategoryLabel = (row) => {
  if (!row) return '未分类'
  // 优先看后端有没有直接传 category_name 类似的中文，没有的话再用 code 从字典里捞
  const code = row.service_code || ''
  if (!categoryDictOptions.value || categoryDictOptions.value.length === 0) return code || '未分类'
  const found = categoryDictOptions.value.find(item => item.dict_value === code)
  return found ? found.dict_label : (code || '未分类')
}

// --- 加载服务列表 ---
const loadServices = async () => {
  loading.value = true
  try {
    const params = {}
    if (queryParams.value.service_name) params.service_name = queryParams.value.service_name
    if (queryParams.value.status !== '') params.status = queryParams.value.status
    if (queryParams.value.is_hot !== '') params.is_hot = queryParams.value.is_hot

    const res = await request.get('/v1/admin/services', { params })

    // 自动解包各类包装格式
    let targetList = null
    if (res && Array.isArray(res.data)) {
      targetList = res.data
    } else if (res && res.data && Array.isArray(res.data.list)) {
      targetList = res.data.list
    } else if (Array.isArray(res)) {
      targetList = res
    } else if (res && Array.isArray(res.list)) {
      targetList = res.list
    }

    services.value = targetList || []
  } catch (error) {
    showToast(error?.message || '服务列表加载失败', 'error')
  } finally {
    loading.value = false
  }
}

// --- 图片上传占位 ---
const handleImageUpload = async (fileOptions) => {
  const formData = new FormData()
  formData.append('file', fileOptions.file)
  try {
    const res = await request.post('/v1/admin/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const url = res.data?.url || res.url
    if (url) {
      serviceForm.value.cover_image = url
      showToast('图片上传成功', 'success')
    }
  } catch (error) {
    console.error('由于后端上传接口待打通，目前支持手动输入相对或绝对图片托管路径', error)
  }
}

const handleQuery = () => { loadServices() }
const resetQuery = () => {
  queryParams.value = { service_name: '', status: '', is_hot: '' }
  loadServices()
}

const openAddDialog = () => {
  dialogType.value = 'add'
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row) => {
  dialogType.value = 'edit'
  currentId.value = row.id
  serviceForm.value = JSON.parse(JSON.stringify(row))
  dialogVisible.value = true
}

const saveService = async () => {
  if (!serviceForm.value.service_name) {
    return showToast('系统服务名称不能为空', 'warning')
  }
  try {
    const payload = {
      service_code: serviceForm.value.service_code,
      service_name: serviceForm.value.service_name,
      display_name: serviceForm.value.display_name,
      icon: serviceForm.value.icon,
      cover_image: serviceForm.value.cover_image,
      description: serviceForm.value.description,
      base_price: Number(serviceForm.value.base_price || 0),
      unit: serviceForm.value.unit,
      status: serviceForm.value.status,
      is_hot: serviceForm.value.is_hot,
      currency: 'CNY',
      sort_order: 0,
      form_schema: '{"fields":[]}'
    }

    if (dialogType.value === 'add') {
      await request.post('/v1/admin/services', payload)
      showToast('新服务建立并上架成功', 'success')
    } else {
      await request.put(`/v1/admin/services/${currentId.value}`, payload)
      showToast('服务配置更新成功', 'success')
    }
    dialogVisible.value = false
    resetForm()
    await loadServices()
  } catch (error) {
    showToast(error?.message || '服务保存失败', 'error')
  }
}

const toggleService = async (row) => {
  const targetStatus = row.status === 1 ? 0 : 1
  try {
    await request.put(`/v1/admin/services/${row.id}`, {
      ...row,
      status: targetStatus,
    })
    showToast(`服务已${targetStatus === 1 ? '上架' : '下架'}`, 'success')
    await loadServices()
  } catch (error) {
    showToast('更新服务状态失败', 'error')
  }
}

const selectEmoji = (emoji) => { serviceForm.value.icon = emoji }

onMounted(async () => {
  await loadCategoryDict()
  await loadServices()
})
</script>

<template>
  <div class="services-page-container">
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

    <div class="action-bar-wrapper">
      <el-button type="primary" class="erp-add-btn" @click="openAddDialog">
        + 新增服务
      </el-button>
    </div>

    <div class="table-wrapper card-layout">
      <el-table v-loading="loading" :data="services" border stripe style="width: 100%">
        <el-table-column label="服务图标" width="90" align="center">
          <template #default="scope">
            <span class="service-icon-preview">{{ scope.row.icon || '🌴' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="服务类型" width="160">
          <template #default="scope">
            <el-tag type="info" effect="plain">
              {{ getCategoryLabel(scope.row) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="service_name" label="业务核心名称" width="180" />
        <el-table-column prop="display_name" label="客户端展示名称" width="180" />

        <el-table-column label="价格" width="150" align="right">
          <template #default="scope">
            <span class="price-text">
              ￥{{ scope.row.base_price.toLocaleString() }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="计价单位" width="100" align="center" />

        <el-table-column label="热门推荐" width="110" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.is_hot ? 'danger' : 'info'" effect="light">
              {{ scope.row.is_hot ? '🔥 热门' : '常规' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="服务状态" width="110" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'info'" effect="dark">
              {{ scope.row.status === 1 ? '上架中' : '已下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="服务简介" show-overflow-tooltip min-width="240" />

        <el-table-column label="核心调度操作" width="180" fixed="right" align="center">
          <template #default="scope">
            <el-button link type="primary" size="small" @click="openEditDialog(scope.row)">
              编辑服务
            </el-button>
            <el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" size="small" @click="toggleService(scope.row)">
              {{ scope.row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '✨ 新增服务' : '编辑服务'" width="600px" destroy-on-close @closed="resetForm">
      <el-form :model="serviceForm" label-width="110px" label-position="right">

        <el-form-item label="服务类型" required>
          <el-select
            v-model="serviceForm.service_code"
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
          <div style="font-size: 11px; color: #909399; margin-top: 4px;">注：需保持与业务流程逻辑链一致</div>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="服务名称" required>
              <el-input v-model="serviceForm.service_name" placeholder="例: 越南商务签证" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="客户端展示名称">
              <el-input v-model="serviceForm.display_name" placeholder="前端呈现给客户的名字" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="系统图标">
          <div class="emoji-picker-container">
            <el-input v-model="serviceForm.icon" placeholder="选定系统图标" style="width: 160px; margin-bottom: 8px;" />
            <div class="emoji-tray">
              <span v-for="emoji in commonEmojis" :key="emoji" class="emoji-item" :class="{ selected: serviceForm.icon === emoji }" @click="selectEmoji(emoji)">
                {{ emoji }}
              </span>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="服务封面图">
          <div class="upload-integration-box">
            <el-upload action="" :http-request="handleImageUpload" :show-file-list="false">
              <el-button size="small">点击异步上传文件</el-button>
            </el-upload>
            <el-input v-model="serviceForm.cover_image" placeholder="或直接输入托管的绝对图片 URL 路径" style="margin-top: 6px;" />
          </div>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="14">
            <el-form-item label="价格" required>
              <el-input v-model="serviceForm.base_price" type="number" placeholder="单位：分">
                <template #append>元 (￥)</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="计价单位" required>
              <el-input v-model="serviceForm.unit" placeholder="次 / 人 / 单" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="是否热门推荐">
              <el-switch v-model="serviceForm.is_hot" active-text="🔥 开启爆款标签" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="服务状态">
              <el-radio-group v-model="serviceForm.status">
                <el-radio :value="1">上架</el-radio>
                <el-radio :value="0">下架</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="服务简介">
          <el-input v-model="serviceForm.description" type="textarea" :rows="3" maxlength="100" show-word-limit placeholder="简述该项业务包含的范畴与告知" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveService">确认发布</el-button>
        </span>
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
.erp-add-btn { background-color: #004d40 !important; border-color: #004d40 !important; font-weight: bold; }
.service-icon-preview { font-size: 20px; }
.price-text { font-family: monospace; font-weight: bold; color: #e6a23c; }
.emoji-tray { display: flex; gap: 8px; flex-wrap: wrap; padding: 8px; background: #f5f7fa; border-radius: 6px; border: 1px dashed #dcdfe6; }
.emoji-item { font-size: 20px; cursor: pointer; padding: 4px; border-radius: 4px; transition: all 0.2s; }
.emoji-item:hover, .emoji-item.selected { background: #e0f2f1; transform: scale(1.2); }
.dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }
</style>