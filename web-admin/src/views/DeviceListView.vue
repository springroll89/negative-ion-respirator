<template>
  <div>
    <h2>设备管理</h2>
    <el-table :data="devices" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="device_sn" label="序列号" />
      <el-table-column prop="device_name" label="名称" />
      <el-table-column prop="region_code" label="地区" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'online' ? 'success' : 'info'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="showConfig(row)">配置</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="configVisible" title="设备配置">
      <el-form :model="configForm" label-width="120px">
        <el-form-item label="最高加热温度(℃)"><el-input-number v-model="configForm.max_heat_temp" :min="0" :max="80" /></el-form-item>
        <el-form-item label="目标出口温度(℃)"><el-input-number v-model="configForm.target_out_temp" :min="30" :max="40" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { deviceAPI } from '@/api'
import { ElMessage } from 'element-plus'

const devices = ref([])
const loading = ref(false)
const configVisible = ref(false)
const configForm = ref({ device_id: 0, max_heat_temp: 80, target_out_temp: 35 })

async function fetchDevices() {
  loading.value = true
  try {
    const { data } = await deviceAPI.list({ page: 1, page_size: 100 })
    devices.value = data.data
  } finally { loading.value = false }
}

function showConfig(row: any) {
  configForm.value = { device_id: row.id, max_heat_temp: 80, target_out_temp: 35 }
  configVisible.value = true
}

async function saveConfig() {
  try {
    await deviceAPI.updateConfig(configForm.value)
    ElMessage.success('配置已保存')
    configVisible.value = false
  } catch { ElMessage.error('保存失败') }
}

onMounted(fetchDevices)
</script>
