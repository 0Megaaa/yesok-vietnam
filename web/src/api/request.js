import axios from 'axios'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  // Route-aware token selection:
  // - Admin routes use 'admin_token'
  // - All other routes use 'client_token'
  const isAdminRoute = config.url && config.url.startsWith('/v1/admin')
  const key = isAdminRoute ? 'admin_token' : 'client_token'
  const token = localStorage.getItem(key)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const isAdminRoute = error.config?.url && error.config.url.startsWith('/v1/admin')
      const key = isAdminRoute ? 'admin_token' : 'client_token'
      localStorage.removeItem(key)
      // Redirect to home (client) or admin login (admin)
      if (isAdminRoute) {
        window.location.href = '/admin/login'
      } else {
        window.location.href = '/'
      }
    }
    return Promise.reject(error)
  }
)

export default request
