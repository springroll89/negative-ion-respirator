<template>
  <div>
    <h2>仪表盘</h2>
    <el-row :gutter="20">
      <el-col :span="6"><el-card><el-statistic title="设备总数" :value="stats.total" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="在线设备" :value="stats.online" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="今日订单" :value="stats.todayOrders" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="今日营收" :value="stats.todayRevenue" prefix="¥" /></el-card></el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const stats = reactive({ total: 0, online: 0, todayOrders: 0, todayRevenue: 0 })
</script>
