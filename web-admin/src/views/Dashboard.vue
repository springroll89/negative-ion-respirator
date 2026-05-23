<script setup lang="ts">
import { ref } from 'vue'

const stats = ref([
  { label: '设备总数', value: 0 },
  { label: '在线设备', value: 0 },
  { label: '今日使用次数', value: 0 },
  { label: '告警数量', value: 0 },
])
</script>

<template>
  <div class="dashboard">
    <h1>仪表盘</h1>
    <el-row :gutter="20">
      <el-col v-for="(stat, index) in stats" :key="index" :span="6">
        <el-card>
          <template #header>
            <span>{{ stat.label }}</span>
          </template>
          <div class="stat-value">{{ stat.value }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 20px;
}
h1 {
  margin-bottom: 20px;
}
.stat-value {
  font-size: 32px;
  font-weight: bold;
  text-align: center;
  color: #409eff;
}
</style>
