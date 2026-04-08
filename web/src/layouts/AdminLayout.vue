<!-- web/src/layouts/AdminLayout.vue -->
<template>
  <a-layout class="layout">
    <a-layout-sider collapsible v-model:collapsed="collapsed" :width="220">
      <div class="logo">
        <span v-if="!collapsed">EPay 管理后台</span>
        <span v-else>EP</span>
      </div>
      <a-menu
        :selected-keys="[currentRoute]"
        @menu-item-click="handleMenuClick"
      >
        <a-menu-item key="dashboard">
          <template #icon><icon-dashboard /></template>
          仪表盘
        </a-menu-item>
        <a-menu-item key="merchants">
          <template #icon><icon-user-group /></template>
          商户管理
        </a-menu-item>
        <a-menu-item key="orders">
          <template #icon><icon-file /></template>
          订单管理
        </a-menu-item>
        <a-menu-item key="channels">
          <template #icon><icon-apps /></template>
          通道管理
        </a-menu-item>
        <a-menu-item key="withdrawals">
          <template #icon><icon-swap /></template>
          提现管理
        </a-menu-item>
        <a-menu-item key="refunds">
          <template #icon><icon-undo /></template>
          退款管理
        </a-menu-item>
        <a-menu-item key="test-payment">
          <template #icon><icon-check-circle /></template>
          测试支付
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="header">
        <div class="header-right">
          <a-dropdown>
            <a-button type="text">
              <icon-user /> 管理员
            </a-button>
            <template #content>
              <a-doption @click="openPasswordModal">修改密码</a-doption>
              <a-doption @click="handleLogout">退出登录</a-doption>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      <a-layout-content class="content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>

  <a-modal v-model:visible="passwordVisible" title="修改密码" @before-ok="handleUpdatePassword">
    <a-form :model="passwordForm" layout="vertical">
      <a-form-item label="原密码">
        <a-input-password v-model="passwordForm.old_password" placeholder="请输入原密码" />
      </a-form-item>
      <a-form-item label="新密码">
        <a-input-password v-model="passwordForm.new_password" placeholder="请输入新密码" />
      </a-form-item>
      <a-form-item label="确认密码">
        <a-input-password v-model="passwordForm.confirm_password" placeholder="请再次输入新密码" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { updateAdminPassword } from '@/api/admin'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)
const passwordVisible = ref(false)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const currentRoute = computed(() => {
  const path = route.path.split('/')[2] || 'dashboard'
  return path
})

const handleMenuClick = (key: string) => {
  router.push(`/admin/${key}`)
}

const resetPasswordForm = () => {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}

const openPasswordModal = () => {
  resetPasswordForm()
  passwordVisible.value = true
}

const handleUpdatePassword = async () => {
  if (!passwordForm.old_password || !passwordForm.new_password) {
    Message.warning('请填写完整')
    return false
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    Message.warning('两次密码输入不一致')
    return false
  }
  if (passwordForm.new_password.length < 6) {
    Message.warning('新密码至少6个字符')
    return false
  }
  await updateAdminPassword({
    old_password: passwordForm.old_password,
    new_password: passwordForm.new_password,
  })
  Message.success('密码已修改，请牢记新密码')
  resetPasswordForm()
  return true
}

const handleLogout = () => {
  localStorage.removeItem('admin_token')
  router.push('/admin/login')
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
}
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}
.header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}
.header-right {
  display: flex;
  align-items: center;
}
.content {
  margin: 24px;
  padding: 24px;
  background: #fff;
  min-height: calc(100vh - 64px - 48px);
}
</style>
