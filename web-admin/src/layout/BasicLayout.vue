<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAdminStore } from '@/store'

const route = useRoute()
const router = useRouter()
const adminStore = useAdminStore()

const sidebarCollapsed = ref(false)

// 侧边栏折叠/展开
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// 路由跳转（替代原来的 selectPanel）
const navigateTo = (path) => {
  router.push(path)
}

// 判断当前路由是否激活
const isActive = (path) => route.path === path || route.path.startsWith(path + '/')

const navItems = [
  { label: '数据看板', path: '/admin/dashboard' },
  { label: '订单中心', path: '/admin/orders' },
  { label: '服务配置', path: '/admin/services' },
  { label: '资讯管理', path: '/admin/articles' },
  { label: '财务流水', path: '/admin/finance' },
  { label: '用户矩阵', path: '/admin/users' },
  { label: '系统管理', path: '/admin/system/dict' },
]

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  await adminStore.logout()
  router.push('/login')
}

// 响应式：移动端自动折叠
const handleResize = () => {
  sidebarCollapsed.value = window.innerWidth < 768
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="admin-shell" :class="{ collapsed: sidebarCollapsed }">
    <!-- 移动端展开按钮 -->
    <el-button class="collapse-toggle" type="default" @click="toggleSidebar">
      {{ sidebarCollapsed ? '展开菜单' : '收起菜单' }}
    </el-button>

    <!-- 侧边栏 -->
    <div class="side-nav">
      <span class="brand">Yesok 2.0</span>

      <div
        v-for="item in navItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        @click="navigateTo(item.path)"
      >
        {{ item.label }}
      </div>

      <div class="nav-exit" @click="handleLogout">
        退出登录
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="workspace">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
  </div>
</template>

<style scoped>
.admin-shell {
  display: flex;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: #f2f6f5;
  color: #12312c;
}

.collapse-toggle {
  position: fixed;
  top: 18px;
  left: 18px;
  z-index: 12;
  height: 38px;
  padding: 0 16px;
  color: #12312c;
  background: #f5d98f;
  border: 0;
  border-radius: 999px;
  font-weight: 900;
  box-shadow: 0 12px 32px rgba(0, 77, 64, 0.16);
  display: none;
}

.side-nav {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 8;
  width: 218px;
  padding: 72px 18px 28px;
  overflow-y: auto;
  background: linear-gradient(180deg, #07362f, #004d40);
  box-shadow: 24px 0 70px rgba(0, 77, 64, 0.16);
  transition: transform 0.24s ease;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.admin-shell.collapsed .side-nav {
  transform: translateX(-256px);
}

.brand {
  display: block;
  margin-bottom: 28px;
  color: #fff;
  font-size: 24px;
  font-weight: 900;
  font-family: 'Montserrat', sans-serif;
}

.nav-item {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 18px;
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.72);
  background: transparent;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  white-space: nowrap;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.nav-item.active {
  color: #12312c;
  background: #f5d98f;
}

.nav-exit {
  margin-top: auto;
  padding: 10px 18px;
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.65);
  background: rgba(229, 57, 53, 0.18);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  text-align: center;
  border: 1px solid rgba(229, 57, 53, 0.3);
}

.nav-exit:hover {
  background: rgba(229, 57, 53, 0.35);
  color: #fff;
}

.workspace {
  box-sizing: border-box;
  flex: 1;
  width: 100%;
  height: 100vh;
  min-width: 0;
  margin-left: 218px;
  overflow-x: auto;
  overflow-y: auto;
  padding: 24px;
  transition: margin-left 0.24s ease;
}

.admin-shell.collapsed .workspace {
  margin-left: 0;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .collapse-toggle {
    display: flex;
    align-items: center;
  }

  .side-nav {
    width: 218px;
  }

  .workspace {
    margin-left: 0;
    padding-top: 74px;
  }
}
</style>
