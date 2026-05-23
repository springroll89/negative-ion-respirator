<template>
  <div>
    <h2>数据报表</h2>
    <el-row :gutter="20" style="margin-bottom:20px">
      <el-col :span="6"><el-card><el-statistic title="设备总数" :value="dashboard.total_devices" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="在线设备" :value="dashboard.online_devices" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="今日订单" :value="dashboard.today_orders" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="今日营收" :value="dashboard.today_revenue / 100" prefix="¥" /></el-card></el-col>
    </el-row>

    <el-card>
      <template #header>
        <span>设备使用报告</span>
        <el-date-picker v-model="dateRange" type="daterange" range-separator="至"
          start-placeholder="开始日期" end-placeholder="结束日期" style="margin-left:16px" @change="fetchReport" />
      </template>
      <el-table :data="deviceReports">
        <el-table-column prop="device_sn" label="设备序列号" />
        <el-table-column prop="total_sessions" label="使用次数" />
        <el-table-column prop="total_duration" label="总时长(秒)" />
        <el-table-column prop="total_revenue" label="总营收(分)" />
        <el-table-column prop="avg_out_temp" label="平均出口温度(℃)" />
        <el-table-column prop="fault_count" label="故障次数" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })
api.interceptors.request.use(c => { c.headers.Authorization = `Bearer ${localStorage.getItem('token')}`; return c })

const dashboard = reactive({ total_devices: 0, online_devices: 0, today_orders: 0, today_revenue: 0 })
const deviceReports = ref([])
const dateRange = ref()

async function fetchReport() {
  try {
    const params: any = {}
    if (dateRange.value) {
      params.start_date = dateRange.value[0].toISOString().split('T')[0]
      params.end_date = dateRange.value[1].toISOString().split('T')[0]
    }
    const { data } = await api.get('/admin/report', { params })
    if (data.data) {
      if (data.data.dashboard) Object.assign(dashboard, data.data.dashboard)
      deviceReports.value = data.data.device_reports || []
    }
  } catch { /* silent */ }
}

onMounted(fetchReport)
</script>
