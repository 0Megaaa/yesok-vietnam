<script setup>
import { ref, onMounted } from 'vue'
import { getOrderList, updateOrder } from '@/api/admin/orders'

const orders = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const statusFilter = ref('')
const loading = ref(true)

const statuses = ['pending', 'confirmed', 'in_progress', 'completed', 'cancelled', 'failed', 'refunded']

async function fetchOrders(p = 1) {
  loading.value = true
  try {
    const params = { page: p, limit: limit.value }
    if (statusFilter.value) params.status = statusFilter.value
    const data = await getOrderList(params)
    orders.value = data.orders
    total.value = data.total
    page.value = data.page
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchOrders())

async function handleStatusChange(orderId, newStatus) {
  try {
    await updateOrder(orderId, { status: newStatus })
    const o = orders.value.find((x) => x.id === orderId)
    if (o) o.status = newStatus
  } catch (e) {
    alert('更新失败: ' + (e.response?.data?.error || e.message))
  }
}

function formatAmount(order) {
  return `${order.amount} ${order.currency}`
}

function formatDate(ts) {
  return new Date(ts).toLocaleString()
}
</script>

<template>
  <div class="order-manage">
    <h1>订单管理</h1>
    <div class="filters">
      <label>
        状态筛选：
        <select v-model="statusFilter" @change="fetchOrders(1)">
          <option value="">全部</option>
          <option v-for="s in statuses" :key="s" :value="s">{{ s }}</option>
        </select>
      </label>
      <span class="total-hint">共 {{ total }} 条订单</span>
    </div>
    <div v-if="loading && orders.length === 0" class="loading">加载中...</div>
    <table v-else class="order-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>订单号</th>
          <th>用户ID</th>
          <th>金额</th>
          <th>状态</th>
          <th>创建时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="order in orders" :key="order.id">
          <td class="mono">{{ order.id }}</td>
          <td class="mono order-no">{{ order.order_no }}</td>
          <td class="mono">{{ order.user_id }}</td>
          <td class="mono">{{ formatAmount(order) }}</td>
          <td>
            <span :class="['status-badge', `status-${order.status}`]">{{ order.status }}</span>
          </td>
          <td class="mono">{{ formatDate(order.created_at) }}</td>
          <td>
            <select :value="order.status" @change="handleStatusChange(order.id, $event.target.value)">
              <option v-for="s in statuses" :key="s" :value="s">{{ s }}</option>
            </select>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="orders.length === 0 && !loading" class="empty">暂无订单</div>
    <div class="pagination">
      <button :disabled="page <= 1" @click="fetchOrders(page - 1)">上一页</button>
      <span>第 {{ page }} 页</span>
      <button @click="fetchOrders(page + 1)">下一页</button>
    </div>
  </div>
</template>

<style scoped>
.order-manage h1 { margin-bottom: 1.25rem; }
.filters {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  margin-bottom: 1rem;
}
.filters select {
  padding: 0.3rem 0.6rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 0.875rem;
}
.total-hint { font-size: 0.875rem; color: var(--text); opacity: 0.6; }
.order-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}
.order-table th,
.order-table td {
  text-align: left;
  padding: 0.5rem 0.625rem;
  border-bottom: 1px solid var(--border);
}
.order-table th {
  font-size: 0.8rem;
  color: var(--text);
  font-weight: 500;
  background: var(--code-bg);
}
.order-table select {
  padding: 0.2rem 0.4rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 0.8125rem;
}
.mono { font-family: var(--mono); font-size: 0.8125rem; }
.order-no { max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  background: var(--code-bg);
  color: var(--text);
}
.status-pending { background: #fef9c3; color: #854d0e; }
.status-confirmed { background: #dbeafe; color: #1e40af; }
.status-in_progress { background: #ede9fe; color: #5b21b6; }
.status-completed { background: #dcfce7; color: #166534; }
.status-cancelled, .status-failed { background: #fee2e2; color: #991b1b; }
.status-refunded { background: #f3f4f6; color: #374151; }
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
.loading, .empty { color: var(--text); opacity: 0.6; }
</style>
