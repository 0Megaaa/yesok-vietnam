<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAdminStore } from '@/store'

const router = useRouter()
const adminStore = useAdminStore()

const loginForm = ref({ username: 'admin', password: '123456' })
const submitting = ref(false)

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const handleLogin = async () => {
  if (!loginForm.value.username || !loginForm.value.password) {
    showToast('请输入账号和密码', 'warning')
    return
  }
  submitting.value = true
  try {
    console.log('[Login] 开始登录，credentials =', loginForm.value)
    await adminStore.login(loginForm.value.username, loginForm.value.password)
    console.log('[Login] ✅ 登录成功，token 已写入 store')
    showToast('管家登录成功', 'success')
    router.push('/dashboard')
  } catch (error) {
    console.error('[Login] ❌ 登录失败：', error)
    showToast(error?.message || '登录失败', 'error')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <span class="eyebrow">YESOK COMMAND CENTER</span>
      <span class="login-title">管家后台登录</span>
      <span class="login-desc">使用 sys_users 种子账号进入商业闭环后台。</span>
      <el-input
        v-model="loginForm.username"
        class="input"
        placeholder="账号 admin"
        @keyup.enter="handleLogin"
      />
      <el-input
        v-model="loginForm.password"
        class="input"
        type="password"
        placeholder="密码 123456"
        show-password
        @keyup.enter="handleLogin"
      />
      <el-button
        class="primary-btn"
        :loading="submitting"
        type="default"
        @click="handleLogin"
      >
        进入后台
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f2f6f5;
}

.login-card {
  width: min(440px, calc(100% - 32px));
  padding: 34px;
  border-radius: 36px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 26px 80px rgba(0, 77, 64, 0.12);
  backdrop-filter: blur(18px);
}

.eyebrow {
  display: block;
  color: #c5a059;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 1.8px;
}

.login-title {
  display: block;
  margin-top: 10px;
  font-size: 28px;
  font-weight: 900;
  color: #12312c;
}

.login-desc {
  display: block;
  margin-top: 10px;
  color: #6b7c78;
  font-size: 13px;
  line-height: 1.7;
}

.input {
  box-sizing: border-box;
  width: 100%;
  margin-top: 12px;
}

.primary-btn {
  width: 100%;
  height: 46px;
  margin-top: 16px;
  color: #fff;
  background: #004d40;
  border: 0;
  border-radius: 999px;
  font-weight: 900;
  font-size: 15px;
}
</style>
