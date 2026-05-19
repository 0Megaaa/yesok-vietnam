<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const dictTypes = ref([])
const dictData = ref([])

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const dictTypeForm = ref({ id: null, dict_name: '', dict_code: '', remark: '', status: 1 })

const dictDataForm = ref({
  id: null,
  dict_code: 'article_category',
  dict_label: '',
  dict_value: '',
  sort_order: 10,
  status: 1,
  remark: '',
})

const loadDicts = async () => {
  loading.value = true
  try {
    console.log('[Dict] 发起 GET /v1/admin/dict-types + /v1/admin/dict-data')
    const [typeRes, dataRes] = await Promise.all([
      request.get('/v1/admin/dict-types'),
      request.get('/v1/admin/dict-data'),
    ])
    console.log('[Dict] ✅ dictTypes =', typeRes.data)
    console.log('[Dict] ✅ dictData =', dataRes.data)
    dictTypes.value = typeRes.data.list || []
    dictData.value = dataRes.data.list || []
  } catch (error) {
    console.error('[Dict] ❌ 报错：', error)
    showToast(error?.message || '字典数据加载失败', 'error')
  } finally {
    loading.value = false
  }
}

// 字典类型操作
const resetDictTypeForm = () => {
  dictTypeForm.value = { id: null, dict_name: '', dict_code: '', remark: '', status: 1 }
}

const editDictType = (item) => {
  dictTypeForm.value = { ...item }
}

const saveDictType = async () => {
  try {
    const payload = { ...dictTypeForm.value, status: Number(dictTypeForm.value.status) }
    if (payload.id) {
      await request.put(`/v1/admin/dict-types/${payload.id}`, payload)
    } else {
      await request.post('/v1/admin/dict-types', payload)
    }
    resetDictTypeForm()
    showToast('字典类型已保存', 'success')
    await loadDicts()
  } catch (error) {
    showToast(error?.message || '字典类型保存失败', 'error')
  }
}

const deleteDictType = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-types/${item.id}`)
    showToast('字典类型已删除', 'success')
    await loadDicts()
  } catch (error) {
    showToast(error?.message || '字典类型删除失败', 'error')
  }
}

// 字典数据操作
const resetDictDataForm = () => {
  dictDataForm.value = {
    id: null,
    dict_code: 'article_category',
    dict_label: '',
    dict_value: '',
    sort_order: 10,
    status: 1,
    remark: '',
  }
}

const editDictData = (item) => {
  dictDataForm.value = { ...item }
}

const saveDictData = async () => {
  try {
    const payload = {
      ...dictDataForm.value,
      sort_order: Number(dictDataForm.value.sort_order),
      status: Number(dictDataForm.value.status),
    }
    if (payload.id) {
      await request.put(`/v1/admin/dict-data/${payload.id}`, payload)
    } else {
      await request.post('/v1/admin/dict-data', payload)
    }
    resetDictDataForm()
    showToast('字典数据已保存', 'success')
    await loadDicts()
  } catch (error) {
    showToast(error?.message || '字典数据保存失败', 'error')
  }
}

const deleteDictData = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-data/${item.id}`)
    showToast('字典数据已删除', 'success')
    await loadDicts()
  } catch (error) {
    showToast(error?.message || '字典数据删除失败', 'error')
  }
}

onMounted(() => {
  console.log('[Dict] 组件挂载，调用 loadDicts')
  loadDicts()
})
</script>

<template>
  <div class="dict-page">
    <div class="page-header">
      <span class="page-title">字典管理</span>
      <span class="page-sub">服务分类、资讯分类与订单状态统一维护</span>
      <el-button class="refresh-btn" type="default" :loading="loading" @click="loadDicts">
        刷新
      </el-button>
    </div>

    <div class="dual-grid">
      <!-- 字典类型 -->
      <div class="sub-panel">
        <span class="sub-title">字典类型</span>

        <div class="service-form two-col">
          <el-input v-model="dictTypeForm.dict_name" class="input" placeholder="字典名称" />
          <el-input v-model="dictTypeForm.dict_code" class="input" placeholder="dict_code" />
          <el-input v-model="dictTypeForm.remark" class="input span-2" placeholder="备注" />
          <el-button class="primary-btn" type="default" @click="saveDictType">
            {{ dictTypeForm.id ? '更新类型' : '新增类型' }}
          </el-button>
          <el-button class="ghost-btn" type="default" @click="resetDictTypeForm">清空</el-button>
        </div>

        <div class="table-scroll">
          <div class="table-list">
            <div
              v-for="item in dictTypes"
              :key="item.id"
              class="table-row"
            >
              <span class="dict-name">{{ item.dict_name }} · {{ item.dict_code }}</span>
              <span class="dict-status" :class="{ active: item.status === 1 }">
                {{ item.status === 1 ? '启用' : '停用' }}
              </span>
              <el-button class="ghost-btn" type="default" @click="editDictType(item)">编辑</el-button>
              <el-button class="danger-btn" type="default" @click="deleteDictType(item)">删除</el-button>
            </div>
            <div v-if="!dictTypes.length" class="empty">暂无字典类型</div>
          </div>
        </div>
      </div>

      <!-- 字典数据 -->
      <div class="sub-panel">
        <span class="sub-title">字典数据</span>

        <div class="service-form two-col">
          <el-input v-model="dictDataForm.dict_code" class="input" placeholder="dict_code" />
          <el-input v-model="dictDataForm.dict_label" class="input" placeholder="标签" />
          <el-input v-model="dictDataForm.dict_value" class="input" placeholder="值" />
          <el-input v-model="dictDataForm.sort_order" class="input" type="number" placeholder="排序" />
          <el-input v-model="dictDataForm.remark" class="input span-2" placeholder="备注" />
          <el-button class="primary-btn" type="default" @click="saveDictData">
            {{ dictDataForm.id ? '更新数据' : '新增数据' }}
          </el-button>
          <el-button class="ghost-btn" type="default" @click="resetDictDataForm">清空</el-button>
        </div>

        <div class="table-scroll">
          <div class="table-list">
            <div
              v-for="item in dictData"
              :key="item.id"
              class="table-row"
            >
              <span class="dict-name">{{ item.dict_code }} · {{ item.dict_label }}</span>
              <span class="dict-value">{{ item.dict_value }}</span>
              <el-button class="ghost-btn" type="default" @click="editDictData(item)">编辑</el-button>
              <el-button class="danger-btn" type="default" @click="deleteDictData(item)">删除</el-button>
            </div>
            <div v-if="!dictData.length" class="empty">暂无字典数据</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dict-page {
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
  flex: 1;
  color: #6b7c78;
  font-size: 13px;
}

.refresh-btn {
  height: 38px;
  padding: 0 18px;
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.dual-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.sub-panel {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
}

.sub-title {
  display: block;
  margin-bottom: 12px;
  color: #12312c;
  font-size: 18px;
  font-weight: 900;
}

.service-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.input {
  width: 100%;
}

.span-2 {
  grid-column: span 2;
}

.primary-btn {
  height: 46px;
  color: #fff;
  background: #004d40;
  border: 0;
  border-radius: 999px;
  font-weight: 900;
  font-size: 14px;
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

.danger-btn {
  height: 32px;
  padding: 0 14px;
  color: #b42318;
  background: rgba(180, 35, 24, 0.09);
  border: 0;
  border-radius: 999px;
  font-weight: 700;
}

.table-scroll {
  width: 100%;
  overflow-x: auto;
  margin-top: 16px;
}

.table-list {
  min-width: 500px;
}

.table-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
}

.dict-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: #12312c;
}

.dict-status {
  min-width: 50px;
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  text-align: center;
  color: #6b7c78;
  background: rgba(0, 77, 64, 0.08);
}

.dict-status.active {
  color: #004d40;
  background: rgba(0, 77, 64, 0.12);
}

.dict-value {
  min-width: 60px;
  color: #6b7c78;
  font-size: 12px;
  text-align: center;
}

.empty {
  padding: 20px;
  color: #6b7c78;
  text-align: center;
  font-size: 13px;
}

@media (max-width: 1024px) {
  .dual-grid {
    grid-template-columns: 1fr;
  }
}
</style>
