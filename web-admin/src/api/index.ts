import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export const authAPI = {
  login: (data: { username: string; password: string }) => api.post('/auth/login', data),
}

export const deviceAPI = {
  list: (params: { page: number; page_size: number }) => api.get('/admin/devices', { params }),
  get: (id: number) => api.get(`/admin/device/${id}`),
  register: (data: { device_sn: string; device_name: string; region_code: string }) =>
    api.post('/admin/device/register', data),
  updateConfig: (data: { device_id: number; max_heat_temp: number; target_out_temp: number }) =>
    api.put('/admin/device/config', data),
}

export default api
