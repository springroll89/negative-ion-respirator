<template>
  <div>
    <h2>批量配置</h2>
    <el-card>
      <el-form :model="form" label-width="120px">
        <el-form-item label="目标类型">
          <el-radio-group v-model="form.target_type">
            <el-radio value="region">按地区</el-radio>
            <el-radio value="device">按设备</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择地区" v-if="form.target_type === 'region'">
          <el-select v-model="form.target_regions" multiple placeholder="选择地区">
            <el-option label="默认" value="default" />
            <el-option label="北方" value="north" />
            <el-option label="南方" value="south" />
          </el-select>
        </el-form-item>
        <el-form-item label="最高加热温度(℃)">
          <el-input-number v-model="form.max_heat_temp" :min="0" :max="80" />
        </el-form-item>
        <el-form-item label="目标出口温度(℃)">
          <el-input-number v-model="form.target_out_temp" :min="30" :max="40" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitBatch" :loading="submitting">提交批量任务</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top:20px">
      <template #header>任务历史</template>
      <el-table :data="tasks">
        <el-table-column prop="id" label="任务ID" width="180" />
        <el-table-column prop="task_type" label="类型" width="100" />
        <el-table-column prop="target_type" label="目标" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'running' ? 'warning' : 'info'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="进度">
          <template #default="{ row }">
            <el-progress :percentage="row.total ? Math.round(row.progress / row.total * 100) : 0" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({ baseURL: '/api/v1' })
api.interceptors.request.use(c => { c.headers.Authorization = `Bearer ${localStorage.getItem('token')}`; return c })

const form = reactive({ target_type: 'region', target_regions: [] as string[], max_heat_temp: 80, target_out_temp: 35 })
const submitting = ref(false)
const tasks = ref<any[]>([])

async function submitBatch() {
  submitting.value = true
  try {
    const { data } = await api.post('/admin/batch/config', form)
    ElMessage.success(`任务已创建: ${data.data.id}`)
    tasks.value.unshift(data.data)
  } catch { ElMessage.error('提交失败') }
  finally { submitting.value = false }
}
</script>
