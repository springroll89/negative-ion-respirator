<template>
  <div v-if="$route.path === '/login'">
    <router-view />
  </div>
  <el-container v-else style="min-height:100vh">
    <el-aside width="200px" style="background:#304156">
      <div style="padding:20px;color:#fff;font-size:16px;text-align:center">呼吸器管理后台</div>
      <el-menu :default-active="$route.path" router background-color="#304156" text-color="#bfcbd9" active-text-color="#409EFF">
        <el-menu-item index="/"><el-icon><DataAnalysis /></el-icon> 仪表盘</el-menu-item>
        <el-menu-item index="/devices"><el-icon><Monitor /></el-icon> 设备管理</el-menu-item>
        <el-menu-item index="/batch"><el-icon><Setting /></el-icon> 批量配置</el-menu-item>
        <el-menu-item index="/orders"><el-icon><Document /></el-icon> 订单管理</el-menu-item>
        <el-menu-item index="/reports"><el-icon><TrendCharts /></el-icon> 数据报表</el-menu-item>
        <el-menu-item index="/settings"><el-icon><Tools /></el-icon> 系统设置</el-menu-item>
      </el-menu>
      <div style="position:absolute;bottom:20px;width:100%;text-align:center">
        <el-button text style="color:#bfcbd9" @click="logout">退出登录</el-button>
      </div>
    </el-aside>
    <el-main><router-view /></el-main>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
function logout() { localStorage.removeItem('token'); router.push('/login') }
</script>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
}
</style>
