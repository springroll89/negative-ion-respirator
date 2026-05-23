import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '仪表盘' },
  },
  {
    path: '/devices',
    name: 'devices',
    component: () => import('@/views/Devices.vue'),
    meta: { title: '设备管理' },
  },
  {
    path: '/records',
    name: 'records',
    component: () => import('@/views/Records.vue'),
    meta: { title: '使用记录' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
