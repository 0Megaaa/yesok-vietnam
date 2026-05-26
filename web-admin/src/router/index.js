import { createRouter, createWebHistory } from 'vue-router'

// 路由懒加载守卫：检查 admin_token
const requireAuth = () => {
  const token =
    typeof localStorage !== 'undefined'
      ? localStorage.getItem('admin_token')
      : ''
  return !!token
}

const routes = [
  {
    path: '/',
    name: 'Home',
    redirect: '/admin',
  },
  {
    path: '/admin',
    component: () => import('@/layout/BasicLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/admin/dashboard',
      },
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
    redirect: '/admin/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 全局路由守卫
router.beforeEach((to, from, next) => {
  if (to.meta?.requiresAuth && !requireAuth()) {
    next('/login')
  } else if (to.path === '/login' && requireAuth()) {
    next('/admin')
  } else {
    next()
  }
})

export default router
