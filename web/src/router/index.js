import { createRouter, createWebHistory } from 'vue-router'
import { useClientStore } from '@/store/client'
import { useAdminStore } from '@/store/admin'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // ── Client routes (mobile: TG WebApp / WeChat / APP) ──────────────────
    {
      path: '/',
      component: () => import('../layouts/ClientLayout.vue'),
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('../views/client/Home.vue'),
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('../views/client/Profile.vue'),
        },
      ],
    },

    // ── Admin routes (PC browser) ────────────────────────────────────────
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),
      meta: { requiresAdmin: true },
      children: [
        {
          path: '',
          name: 'admin-dashboard',
          component: () => import('../views/admin/Dashboard.vue'),
        },
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('../views/admin/UserManage.vue'),
        },
        {
          path: 'orders',
          name: 'admin-orders',
          component: () => import('../views/admin/OrderManage.vue'),
        },
      ],
    },

    // ── Admin login (no layout) ───────────────────────────────────────────
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('../views/admin/Login.vue'),
    },
  ],
})

// Navigation guard: protect /admin/* with admin token
router.beforeEach((to) => {
  // Skip guard for login page
  if (to.name === 'admin-login') return

  if (to.matched.some((r) => r.meta?.requiresAdmin)) {
    const admin = useAdminStore()
    if (!admin.isLoggedIn) {
      return { name: 'admin-login' }
    }
  }
})

export default router
