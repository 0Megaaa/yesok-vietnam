<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
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
  { label: '财务管理', path: '/admin/finance' },
  { label: '用户管理', path: '/admin/users' },
]

const systemSubItems = [
  { label: '字典设置', path: '/admin/system/dict' },
  // 未来可追加：{ label: '菜单管理', path: '/admin/system/menu' },
]

// 系统管理：支持手动展开/收起，同时在子路由激活时自动展开
const isSystemOpen = ref(false)

const isSubActive = (path) => route.path === path

// 点击父级切换展开状态
const toggleSystemMenu = () => {
  isSystemOpen.value = !isSystemOpen.value
}

// 路由变化时自动展开（支持页面刷新后仍保持展开）
watch(
  () => route.path,
  (path) => {
    if (path.startsWith('/admin/system')) {
      isSystemOpen.value = true
    }
  },
  { immediate: true }
)

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

      <!-- 系统管理（折叠子菜单） -->
      <div class="nav-group">
        <div
          class="nav-item nav-group-parent"
          :class="{ open: isSystemOpen }"
          @click="toggleSystemMenu"
        >
          <span>系统管理</span>
          <span class="nav-arrow" :class="{ rotated: isSystemOpen }">›</span>
        </div>

        <div v-show="isSystemOpen" class="nav-sub-items">
          <div
            v-for="item in systemSubItems"
            :key="item.path"
            class="nav-item nav-sub-item"
            :class="{ active: isSubActive(item.path) }"
            @click="navigateTo(item.path)"
          >
            {{ item.label }}
          </div>
        </div>
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

/* 系统管理父级 */
.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-group-parent {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  font-weight: 500;
  cursor: default;
}

/* 父级激活（子菜单已展开）时高亮 */
.nav-group-parent.open {
  color: rgba(255, 255, 255, 0.9);
  background: rgba(255, 255, 255, 0.06);
}

.nav-arrow {
  font-size: 18px;
  font-weight: 400;
  transition: transform 0.2s ease;
  transform: rotate(90deg);
}

.nav-arrow.rotated {
  transform: rotate(-90deg);
}

/* 子菜单容器 */
.nav-sub-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-left: 10px;
}

/* 子菜单项 */
.nav-sub-item {
  height: 38px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.55);
  font-weight: 400;
}

.nav-sub-item:hover {
  color: rgba(255, 255, 255, 0.9);
  background: rgba(255, 255, 255, 0.1);
}

/* 子菜单激活态 — 与父级保持一致的高亮色 */
.nav-sub-item.active {
  color: #12312c;
  background: #f5d98f;
  font-weight: 600;
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
