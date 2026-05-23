<template>
  <div>
    <h2>订单管理</h2>
    <el-table :data="orders" v-loading="loading">
      <el-table-column prop="id" label="订单ID" width="80" />
      <el-table-column prop="tid" label="交易ID" width="200" />
      <el-table-column prop="user_id" label="用户ID" width="80" />
      <el-table-column prop="device_id" label="设备ID" width="80" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'completed' ? 'success' : 'warning'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="duration" label="时长(秒)" width="100" />
      <el-table-column prop="amount" label="金额(分)" width="100" />
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString() }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })
api.interceptors.request.use(c => { c.headers.Authorization = `Bearer ${localStorage.getItem('token')}`; return c })

const orders = ref([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await api.get('/admin/orders')
    orders.value = data.data || []
  } finally { loading.value = false }
})
</script>
