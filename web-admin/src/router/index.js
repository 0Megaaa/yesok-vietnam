import { createRouter, createWebHistory } from 'vue-router'

// isTokenExpired 检查本地存储的 admin_token_expire 是否已过期。
const isTokenExpired = () => {
  if (typeof localStorage === 'undefined') return true
  const expireStr = localStorage.getItem('admin_token_expire')
  if (!expireStr) return true
  const expire = parseInt(expireStr, 10)
  if (isNaN(expire)) return true
  return Date.now() / 1000 > expire
}

// requireAuth 检查 admin_token 存在且未过期。
const requireAuth = () => {
  if (typeof localStorage === 'undefined') return false
  const token = localStorage.getItem('admin_token')
  if (!token) return false
  if (isTokenExpired()) {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_token_expire')
    return false
  }
  return true
}

// vite.base = /admin/，router 也必须以 /admin/ 为 base。
// createWebHistory(import.meta.env.BASE_URL) 等于 /admin/。
// 所以 next('/login') 实际跳 /admin/login，next('/dashboard') 实际跳 /admin/dashboard。
const routes = [
  {
    path: '/',
    name: 'Home',
    redirect: '/dashboard',
  },
  {
    path: '/',
    component: () => import('@/layout/BasicLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/admin/dashboard/index.vue'),
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/admin/order/index.vue'),
      },
      {
        path: 'services',
        name: 'Services',
        component: () => import('@/views/admin/service/index.vue'),
      },
      {
        path: 'articles',
        name: 'Articles',
        component: () => import('@/views/admin/article/index.vue'),
      },
      {
        path: 'finance',
        name: 'Finance',
        component: () => import('@/views/admin/finance/index.vue'),
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/admin/user/index.vue'),
      },
      {
        path: 'system/dict',
        name: 'SystemDict',
        component: () => import('@/views/admin/system/dict/index.vue'),
      },
      {
        path: 'order/:id',
        name: 'OrderDetail',
        component: () => import('@/views/admin/order/detail.vue'),
        props: true,
      },
    ],
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/admin/Login.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL || '/admin/'),
  routes,
})

// 全局路由守卫
router.beforeEach((to, from, next) => {
  if (to.meta?.requiresAuth && !requireAuth()) {
    next('/login')
  } else if (to.path === '/login' && requireAuth()) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
