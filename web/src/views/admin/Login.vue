<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '@/store/admin'

const router = useRouter()
const admin = useAdminStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await admin.login(username.value, password.value)
    router.push('/admin')
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>管理后台登录</h1>
      <form @submit.prevent="handleLogin">
        <div class="field">
          <label for="username">用户名</label>
          <input id="username" v-model="username" type="text" placeholder="admin" required autocomplete="username" />
        </div>
        <div class="field">
          <label for="password">密码</label>
          <input id="password" v-model="password" type="password" placeholder="••••••••" required autocomplete="current-password" />
        </div>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
      <p class="hint">默认账号：admin / admin123</p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}
.login-card {
  width: 100%;
  max-width: 360px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 2rem;
  box-shadow: var(--shadow);
}
.login-card h1 {
  font-size: 1.25rem;
  margin-bottom: 1.5rem;
  text-align: center;
}
.field {
  margin-bottom: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
.field label {
  font-size: 0.875rem;
  color: var(--text);
}
.field input {
  padding: 0.625rem 0.875rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 0.9375rem;
  outline: none;
  transition: border-color 0.2s;
}
.field input:focus {
  border-color: var(--accent);
}
.error-msg {
  color: #ef4444;
  font-size: 0.875rem;
  margin-bottom: 0.75rem;
}
.submit-btn {
  width: 100%;
  padding: 0.75rem;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: opacity 0.2s;
  margin-top: 0.5rem;
}
.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.hint {
  text-align: center;
  font-size: 0.8125rem;
  color: var(--text);
  opacity: 0.5;
  margin-top: 1rem;
}
</style>
