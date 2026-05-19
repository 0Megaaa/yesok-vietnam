<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const services = ref([])

const serviceForm = ref({
  service_code: '',
  service_name: '',
  display_name: '',
  icon: '🌴',
  description: '',
  base_price: 0,
  currency: 'VND',
  unit: '次',
  sort_order: 10,
  status: 1,
  is_hot: false,
  form_schema: '{"fields":[]}',
})

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const resetForm = () => {
  serviceForm.value = {
    service_code: '',
    service_name: '',
    display_name: '',
    icon: '🌴',
    description: '',
    base_price: 0,
    currency: 'VND',
    unit: '次',
    sort_order: 10,
    status: 1,
    is_hot: false,
    form_schema: '{"fields":[]}',
  }
}

const loadServices = async () => {
  loading.value = true
  try {
    console.log('[Services] 发起 GET /v1/admin/services')
    const res = await request.get('/v1/admin/services')
    console.log('[Services] ✅ 返回：', res.data)
    services.value = res.data.list || []
  } catch (error) {
    console.error('[Services] ❌ 报错：', error)
    showToast(error?.message || '服务列表加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const saveService = async () => {
  try {
    await request.post('/v1/admin/services', {
      ...serviceForm.value,
      base_price: Number(serviceForm.value.base_price),
      sort_order: Number(serviceForm.value.sort_order),
    })
    resetForm()
    showToast('服务配置已保存，C端将动态更新', 'success')
    await loadServices()
  } catch (error) {
    showToast(error?.message || '服务保存失败', 'error')
  }
}

const toggleService = async (service) => {
  try {
    await request.put(`/v1/admin/services/${service.id}`, {
      ...service,
      status: service.status === 1 ? 0 : 1,
    })
    await loadServices()
    showToast(`服务已${service.status === 1 ? '下架' : '上架'}`, 'success')
  } catch (error) {
    showToast(error?.message || '服务状态更新失败', 'error')
  }
}

onMounted(() => {
  console.log('[Services] 组件挂载，调用 loadServices')
  loadServices()
})
</script>

<template>
  <div class="services-page">
    <div class="page-header">
      <span class="page-title">服务配置</span>
      <span class="page-sub">服务价格与热门标签会立即驱动 C 端首页</span>
    </div>

    <!-- 服务表单 -->
    <div class="content-card">
      <div class="service-form">
        <el-input v-model="serviceForm.service_code" class="input" placeholder="service_code" />
        <el-input v-model="serviceForm.service_name" class="input" placeholder="服务名称" />
        <el-input v-model="serviceForm.display_name" class="input" placeholder="展示名称" />
        <el-input v-model="serviceForm.icon" class="input" placeholder="图标" />
        <el-input v-model="serviceForm.base_price" class="input" type="number" placeholder="价格（分）" />
        <el-input v-model="serviceForm.unit" class="input" placeholder="单位" />
        <el-input v-model="serviceForm.description" class="input span-2" placeholder="服务描述" />
        <el-button class="primary-btn" type="default" @click="saveService">保存服务</el-button>
      </div>
    </div>

    <!-- 服务列表 -->
    <div class="content-card">
      <div class="table-scroll">
        <div class="table-list">
          <div
            v-for="service in services"
            :key="service.id"
            class="table-row"
          >
            <span class="service-name">{{ service.icon }} {{ service.display_name || service.service_name }}</span>
            <span class="service-price">
              {{ Math.round(service.base_price / 100) }} {{ service.currency }}/{{ service.unit }}
            </span>
            <span class="service-status" :class="{ active: service.status === 1 }">
              {{ service.status === 1 ? '启用' : '停用' }}
            </span>
            <el-button
              class="ghost-btn"
              type="default"
              @click="toggleService(service)"
            >
              {{ service.status === 1 ? '下架' : '上架' }}
            </el-button>
          </div>
          <div v-if="!services.length" class="empty">暂无服务配置</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.services-page {
  width: 100%;
  min-width: 800px;
}

.page-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 900;
  color: #12312c;
}

.page-sub {
  color: #6b7c78;
  font-size: 13px;
}

.content-card {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
  margin-bottom: 16px;
}

.service-form {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.input {
  width: 100%;
}

.span-2 {
  grid-column: span 2;
}

.primary-btn {
  grid-column: span 3;
  height: 46px;
  color: #fff;
  background: #004d40;
  border: 0;
  border-radius: 999px;
  font-weight: 900;
  font-size: 15px;
}

.table-scroll {
  width: 100%;
  overflow-x: auto;
}

.table-list {
  min-width: 800px;
  margin-top: 16px;
}

.table-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
}

.service-name {
  flex: 1;
  font-size: 14px;
  font-weight: 700;
  color: #12312c;
}

.service-price {
  min-width: 100px;
  color: #6b7c78;
  font-size: 13px;
}

.service-status {
  min-width: 60px;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  color: #6b7c78;
  background: rgba(0, 77, 64, 0.08);
  text-align: center;
}

.service-status.active {
  color: #004d40;
  background: rgba(0, 77, 64, 0.12);
}

.ghost-btn {
  height: 32px;
  padding: 0 14px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.empty {
  padding: 20px;
  color: #6b7c78;
  text-align: center;
  font-size: 13px;
}

@media (max-width: 768px) {
  .service-form {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .span-2 {
    grid-column: span 2;
  }

  .primary-btn {
    grid-column: span 2;
  }
}
</style>
