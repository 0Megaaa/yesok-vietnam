<script setup>
import { ref, onMounted } from 'vue'
import { getUserList, updateUserRole, deleteUser } from '@/api/admin/users'

const users = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const loading = ref(true)
const error = ref('')

const roles = ['user', 'admin', 'worker']

async function fetchUsers(p = 1) {
  loading.value = true
  error.value = ''
  try {
    const data = await getUserList({ page: p, limit: limit.value })
    users.value = data.users
    total.value = data.total
    page.value = data.page
  } catch (e) {
    error.value = '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchUsers())

async function handleRoleChange(userId, newRole) {
  try {
    await updateUserRole(userId, newRole)
    const u = users.value.find((x) => x.id === userId)
    if (u) u.role = newRole
  } catch (e) {
    alert('更新角色失败: ' + (e.response?.data?.error || e.message))
  }
}

async function handleDelete(userId) {
  if (!confirm('确定删除该用户？')) return
  try {
    await deleteUser(userId)
    users.value = users.value.filter((u) => u.id !== userId)
  } catch (e) {
    alert('删除失败: ' + (e.response?.data?.error || e.message))
  }
}
</script>

<template>
  <div class="user-manage">
    <h1>用户管理</h1>
    <div v-if="loading && users.length === 0" class="loading">加载中...</div>
    <div v-else-if="error && users.length === 0" class="error">{{ error }}</div>
    <template v-else>
      <p class="total-hint">共 {{ total }} 位用户</p>
      <table class="user-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>Telegram ID</th>
            <th>角色</th>
            <th>余额</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td class="mono">{{ user.id }}</td>
            <td>{{ user.username || '—' }}</td>
            <td class="mono">{{ user.tg_id }}</td>
            <td>
              <select :value="user.role" @change="handleRoleChange(user.id, $event.target.value)">
                <option v-for="r in roles" :key="r" :value="r">{{ r }}</option>
              </select>
            </td>
            <td class="mono">{{ user.balance }}</td>
            <td>
              <button class="delete-btn" @click="handleDelete(user.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="pagination">
        <button :disabled="page <= 1" @click="fetchUsers(page - 1)">上一页</button>
        <span>第 {{ page }} 页</span>
        <button @click="fetchUsers(page + 1)">下一页</button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.user-manage h1 { margin-bottom: 1.25rem; }
.total-hint { font-size: 0.875rem; color: var(--text); opacity: 0.6; margin-bottom: 1rem; }
.user-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}
.user-table th,
.user-table td {
  text-align: left;
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--border);
}
.user-table th {
  font-size: 0.8125rem;
  color: var(--text);
  font-weight: 500;
  background: var(--code-bg);
}
.user-table select {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 0.875rem;
}
.delete-btn {
  background: none;
  border: 1px solid #ef4444;
  color: #ef4444;
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8125rem;
  transition: background 0.2s;
}
.delete-btn:hover { background: #fef2f2; }
.mono { font-family: var(--mono); font-size: 0.875rem; }
.pagination {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1.25rem;
}
.pagination button {
  padding: 0.375rem 1rem;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
}
.pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
.loading, .error { color: var(--text); opacity: 0.6; }
</style>
