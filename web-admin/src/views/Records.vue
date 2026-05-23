<script setup lang="ts">
import { ref } from 'vue'

interface Record {
  id: string
  deviceName: string
  startTime: string
  endTime: string
  duration: number
}

const records = ref<Record[]>([])

const columns = [
  { prop: 'id', label: '记录编号' },
  { prop: 'deviceName', label: '设备名称' },
  { prop: 'startTime', label: '开始时间' },
  { prop: 'endTime', label: '结束时间' },
  { prop: 'duration', label: '使用时长(分钟)' },
]
</script>

<template>
  <div class="records">
    <h1>使用记录</h1>
    <el-table :data="records" empty-text="暂无使用记录">
      <el-table-column v-for="col in columns" :key="col.prop" :prop="col.prop" :label="col.label" />
    </el-table>
  </div>
</template>

<style scoped>
.records {
  padding: 20px;
}
h1 {
  margin-bottom: 20px;
}
</style>
