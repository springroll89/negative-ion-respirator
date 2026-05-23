<template>
  <div>
    <h2>系统设置</h2>
    <el-card>
      <template #header>地区/季节默认温度配置</template>
      <el-table :data="configs">
        <el-table-column prop="region_code" label="地区" />
        <el-table-column prop="season" label="季节" />
        <el-table-column prop="max_heat_temp" label="最高加热温度(℃)" />
        <el-table-column prop="target_out_temp" label="目标出口温度(℃)" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="editConfig(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="编辑配置">
      <el-form :model="editForm" label-width="140px">
        <el-form-item label="最高加热温度(℃)"><el-input-number v-model="editForm.max_heat_temp" :min="0" :max="80" /></el-form-item>
        <el-form-item label="目标出口温度(℃)"><el-input-number v-model="editForm.target_out_temp" :min="30" :max="40" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({ baseURL: '/api/v1' })
api.interceptors.request.use(c => { c.headers.Authorization = `Bearer ${localStorage.getItem('token')}`; return c })

const configs = ref([
  { region_code: 'default', season: 'spring', max_heat_temp: 80, target_out_temp: 35 },
  { region_code: 'default', season: 'summer', max_heat_temp: 70, target_out_temp: 32 },
  { region_code: 'default', season: 'autumn', max_heat_temp: 75, target_out_temp: 34 },
  { region_code: 'default', season: 'winter', max_heat_temp: 80, target_out_temp: 40 },
])
const dialogVisible = ref(false)
const editForm = reactive({ region_code: '', season: '', max_heat_temp: 80, target_out_temp: 35 })

function editConfig(row: any) {
  Object.assign(editForm, row)
  dialogVisible.value = true
}

async function saveConfig() {
  try {
    await api.put('/admin/region/config', editForm)
    ElMessage.success('配置已更新')
    dialogVisible.value = false
  } catch { ElMessage.error('保存失败') }
}
</script>
