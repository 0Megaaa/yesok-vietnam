<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(true)
const appUsers = ref([])
const sysUsers = ref([])

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const loadUsers = async () => {
  loading.value = true
  try {
    console.log('[Users] 发起 GET /v1/admin/app-users + /v1/admin/sys-users')
    const [appRes, sysRes] = await Promise.all([
      request.get('/v1/admin/app-users'),
      request.get('/v1/admin/sys-users'),
    ])
    console.log('[Users] ✅ appUsers =', appRes.data)
    console.log('[Users] ✅ sysUsers =', sysRes.data)
    appUsers.value = appRes.data.list || []
    sysUsers.value = sysRes.data.list || []
  } catch (error) {
    console.error('[Users] ❌ 报错：', error)
    showToast(error?.message || '用户列表加载失败', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  console.log('[Users] 组件挂载，调用 loadUsers')
  loadUsers()
})
</script>

<template>
  <div class="users-page">
    <div class="page-header">
      <span class="page-title">用户矩阵</span>
      <el-button class="refresh-btn" type="default" :loading="loading" @click="loadUsers">
        刷新
      </el-button>
    </div>

    <!-- C 端客户 -->
    <div class="panel-grid">
      <div class="glass-panel wide">
        <span class="panel-title">C 端客户矩阵</span>
        <div class="user-table">
          <div class="user-header">
            <span>昵称</span>
            <span>手机号 / OpenID</span>
            <span>注册时间</span>
          </div>
          <div
            v-for="user in appUsers"
            :key="user.id"
            class="user-line"
          >
            <span>{{ user.nickname || '未命名客户' }}</span>
            <span class="strong">{{ user.phone || user.wechat_open_id || '无联系方式' }}</span>
            <span class="muted">{{ user.created_at || user.registered_at || '-' }}</span>
          </div>
          <div v-if="!appUsers.length" class="empty">暂无 C 端客户</div>
        </div>
      </div>

      <!-- B 端员工 -->
      <div class="glass-panel wide">
        <span class="panel-title">B 端员工矩阵</span>
        <div class="user-table">
          <div class="user-header">
            <span>账号</span>
            <span>姓名</span>
            <span>角色</span>
          </div>
          <div
            v-for="user in sysUsers"
            :key="user.id"
            class="user-line"
          >
            <span>{{ user.username }}</span>
            <span>{{ user.real_name || '-' }}</span>
            <span class="role-badge">{{ user.role || '未知' }}</span>
          </div>
          <div v-if="!sysUsers.length" class="empty">暂无 B 端员工</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.users-page {
  width: 100%;
  min-width: 800px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 900;
  color: #12312c;
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

.panel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.glass-panel {
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 30px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 20px 60px rgba(0, 77, 64, 0.08);
  backdrop-filter: blur(15px);
}

.panel-title {
  display: block;
  margin-bottom: 16px;
  color: #12312c;
  font-size: 15px;
  font-weight: 700;
}

.user-table {
  min-width: 0;
}

.user-header {
  display: grid;
  grid-template-columns: 1fr 2fr 1.5fr;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 2px solid rgba(0, 77, 64, 0.1);
  color: #6b7c78;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.user-line {
  display: grid;
  grid-template-columns: 1fr 2fr 1.5fr;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 77, 64, 0.08);
  font-size: 13px;
  color: #12312c;
  align-items: center;
}

.user-line:last-child {
  border-bottom: none;
}

.strong {
  font-weight: 700;
  color: #004d40;
}

.muted {
  color: #6b7c78;
  font-size: 12px;
}

.role-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  background: rgba(0, 77, 64, 0.08);
  color: #004d40;
  font-size: 12px;
  font-weight: 700;
}

.empty {
  padding: 20px 0;
  color: #6b7c78;
  text-align: center;
  font-size: 13px;
}

@media (max-width: 1024px) {
  .panel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
