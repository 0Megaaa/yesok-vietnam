import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../store'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/Home.vue'),
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../views/Profile.vue'),
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/Admin.vue'),
      meta: { requireAdmin: true },
    },
  ],
})

router.beforeEach((to) => {
  if (to.meta.requireAdmin) {
    const userStore = useUserStore()
    if (userStore.userInfo?.role !== 'admin') {
      alert('无权限访问管理后台')
      return '/'
    }
  }
})

export default router