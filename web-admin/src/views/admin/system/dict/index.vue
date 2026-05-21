<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const dataLoading = ref(false)
const dictTypes = ref([])
const dictData = ref([])

const currentTypeId = ref(null)
const currentDictCode = ref('')

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

// 加载字典类型列表（仅类型，不含数据）
const loadDicts = async () => {
  loading.value = true
  try {
    const res = await request.get('/v1/admin/dict-types')
    dictTypes.value = res.data.list || []
  } catch (error) {
    showToast('字典类型加载失败', 'error')
  } finally {
    loading.value = false
  }
}

// 根据选中的 dict_code 获取字典数据列表
const fetchDictData = async () => {
  // 1. 守卫：如果没值，别乱发请求
  if (!currentDictCode.value) {
    dictData.value = []
    return
  }

  dataLoading.value = true
  try {
    // 2. 使用模板字符串强制拼接参数，URL 绝对不会出错
    // 使用 encodeURIComponent 确保特殊字符不会导致请求报错
    const url = `/v1/admin/dict-data?dict_type=${encodeURIComponent(currentDictCode.value)}`

    // 3. 直接调用，不传第二个参数
    const res = await request.get(url)

    // 4. 处理数据
    dictData.value = res.data.data || []
  } catch (error) {
    showToast('字典数据加载失败', 'error')
    console.error('字典加载错误:', error)
  } finally {
    dataLoading.value = false
  }
}

// 点击左侧字典类型 → 同步更新状态并立即请求数据
const selectType = async (item) => {
  currentTypeId.value = item.id
  currentDictCode.value = item.dict_code
  await fetchDictData(item.dict_code)
}

// 字典类型操作
const dictTypeForm = ref({ id: null, dict_name: '', dict_code: '', remark: '', status: 1 })

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
    if (currentTypeId.value === item.id) {
      currentTypeId.value = null
      currentDictCode.value = ''
      dictData.value = []
    }
    await loadDicts()
  } catch (error) {
    showToast(error?.message || '字典类型删除失败', 'error')
  }
}

// 字典数据操作
const dictDataForm = ref({
  id: null,
  dict_code: '',
  dict_label: '',
  dict_value: '',
  sort_order: 10,
  status: 1,
  remark: '',
})

const resetDictDataForm = () => {
  dictDataForm.value = {
    id: null,
    dict_code: currentDictCode.value,
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
      dict_code: currentDictCode.value,
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
    await fetchDictData()
  } catch (error) {
    showToast(error?.message || '字典数据保存失败', 'error')
  }
}

const deleteDictData = async (item) => {
  try {
    await request.delete(`/v1/admin/dict-data/${item.id}`)
    showToast('字典数据已删除', 'success')
    await fetchDictData()
  } catch (error) {
    showToast(error?.message || '字典数据删除失败', 'error')
  }
}

onMounted(() => {
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

    <div class="master-detail">
      <!-- 左侧：字典类型列表 -->
      <div class="type-panel">
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
              class="type-row"
              :class="{ active: currentTypeId === item.id }"
              @click="selectType(item)"
            >
              <div class="type-info">
                <span class="dict-name">{{ item.dict_name }}</span>
                <span class="dict-code">{{ item.dict_code }}</span>
              </div>
              <div class="type-actions">
                <span class="dict-status" :class="{ active: item.status === 1 }">
                  {{ item.status === 1 ? '启用' : '停用' }}
                </span>
                <el-button class="ghost-btn" type="default" @click.stop="editDictType(item)">编辑</el-button>
                <el-button class="danger-btn" type="default" @click.stop="deleteDictType(item)">删除</el-button>
              </div>
            </div>
            <div v-if="!dictTypes.length && !loading" class="empty">暂无字典类型</div>
            <div v-if="loading" class="empty">加载中…</div>
          </div>
        </div>
      </div>

      <!-- 右侧：字典数据明细 -->
      <div class="data-panel">
        <span class="sub-title">
          字典数据
          <span v-if="currentDictCode" class="data-hint">（当前类型：{{ currentDictCode }}）</span>
        </span>

        <div class="service-form two-col">
<!--          <el-input v-model="dictDataForm.dict_label" class="input" placeholder="标签" />-->
<!--          <el-input v-model="dictDataForm.dict_value" class="input" placeholder="值" />-->
<!--          <el-input v-model="dictDataForm.sort_order" class="input" type="number" placeholder="排序" />-->
<!--          <el-input v-model="dictDataForm.remark" class="input span-2" placeholder="备注" />-->
          <el-button class="primary-btn" type="default" @click="saveDictData">
            {{ dictDataForm.id ? '更新数据' : '新增数据' }}
          </el-button>
<!--          <el-button class="ghost-btn" type="default" @click="resetDictDataForm">清空</el-button>-->
        </div>

        <div class="table-scroll">
          <div class="table-list">
            <div v-if="!currentTypeId && !dataLoading" class="empty hint">
              请先在左侧选择一个字典类型
            </div>
            <div v-else-if="dataLoading" class="empty">数据加载中…</div>
            <div
              v-else
              v-for="item in dictData"
              :key="item.id"
              class="table-row"
            >
              <span class="dict-name">{{ item.dict_label }}</span>
              <span class="dict-value">{{ item.dict_value }}</span>
              <span class="dict-status" :class="{ active: item.status === 1 }">
                {{ item.status === 1 ? '启用' : '停用' }}
              </span>
              <el-button class="ghost-btn" type="default" @click="editDictData(item)">编辑</el-button>
              <el-button class="danger-btn" type="default" @click="deleteDictData(item)">删除</el-button>
            </div>
            <div v-if="currentTypeId && !dataLoading && !dictData.length" class="empty">
              该类型暂无数据
            </div>
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

/* 主从联动布局 */
.master-detail {
  display: flex;
  gap: 20px;
  align-items: stretch;
}

/* 两侧面板统一卡片样式 */
.type-panel,
.data-panel {
  min-height: 580px;
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
}

/* 左右两侧均分 50/50 */
.type-panel,
.data-panel {
  flex: 1;
  min-width: 0;
}

.sub-title {
  display: block;
  margin-bottom: 12px;
  color: #12312c;
  font-size: 18px;
  font-weight: 900;
}

.data-hint {
  font-size: 13px;
  font-weight: 500;
  color: #6b7c78;
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

/* 字典类型列表项 */
.type-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 14px 16px;
  margin-bottom: 6px;
  border-radius: 16px;
  cursor: pointer;
  transition: background 0.15s;
  border: 1.5px solid transparent;
}

/* 左侧激活指示条 — 默认隐藏 */
.type-row::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 0;
  border-radius: 0 3px 3px 0;
  background: #f5d98f;
  transition: height 0.2s ease;
}

/* 定位用于 ::before 绝对定位 */
.type-row {
  position: relative;
  overflow: hidden;
}

.type-row:hover {
  background: rgba(0, 77, 64, 0.05);
}

/* 激活态：浅绿背景 + 左侧金色条 */
.type-row.active {
  background: rgba(0, 77, 64, 0.10);
  border-color: rgba(245, 217, 143, 0.5);
}

.type-row.active::before {
  height: 70%;
}

.type-row.active .dict-name {
  color: #004d40;
  font-weight: 800;
}

.type-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 0;
}

.dict-name {
  font-size: 14px;
  font-weight: 600;
  color: #12312c;
  line-height: 1.3;
}

.dict-code {
  font-size: 11px;
  color: #6b7c78;
}

.type-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.table-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
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

.empty.hint {
  color: #9ab;
  font-size: 14px;
}

@media (max-width: 900px) {
  .master-detail {
    flex-direction: column;
  }
  .type-panel {
    flex: none;
    width: 100%;
    min-height: auto;
  }
  .data-panel {
    min-height: auto;
  }
}
</style>
