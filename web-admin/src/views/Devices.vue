<script setup lang="ts">
import { ref } from 'vue'

interface Device {
  id: string
  name: string
  status: string
  lastSeen: string
}

const devices = ref<Device[]>([])

const columns = [
  { prop: 'id', label: '设备编号' },
  { prop: 'name', label: '设备名称' },
  { prop: 'status', label: '状态' },
  { prop: 'lastSeen', label: '最后在线' },
]
</script>

<template>
  <div class="devices">
    <h1>设备管理</h1>
    <el-table :data="devices" empty-text="暂无设备数据">
      <el-table-column v-for="col in columns" :key="col.prop" :prop="col.prop" :label="col.label" />
    </el-table>
  </div>
</template>

<style scoped>
.devices {
  padding: 20px;
}
h1 {
  margin-bottom: 20px;
}
</style>
