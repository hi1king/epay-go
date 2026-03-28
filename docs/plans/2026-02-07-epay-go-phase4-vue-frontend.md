# EPay Go 重构 - 阶段四：Vue3 前端

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现管理后台、商户中心、收银台三个前端模块，完成完整的用户界面。

**Architecture:** 使用 Vue3 + TypeScript + Arco Design，Pinia 状态管理，Vue Router 路由，Axios 请求封装。采用模块化目录结构，按功能划分页面。

**Tech Stack:** Vue3, TypeScript, Vite, Arco Design, Pinia, Vue Router, Axios

**前置条件:** 阶段一至三已完成，后端 API 已就绪。

---

## Task 1: 配置项目基础结构

**Files:**
- Modify: `epay-go/web/vite.config.ts`
- Modify: `epay-go/web/src/main.ts`
- Create: `epay-go/web/src/env.d.ts`

**Step 1: 更新 Vite 配置**

```typescript
// web/vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/merchant': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

**Step 2: 更新入口文件**

```typescript
// web/src/main.ts
import { createApp } from 'vue'
import ArcoVue from '@arco-design/web-vue'
import ArcoVueIcon from '@arco-design/web-vue/es/icon'
import '@arco-design/web-vue/dist/arco.css'

import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'

const app = createApp(App)

app.use(ArcoVue)
app.use(ArcoVueIcon)
app.use(createPinia())
app.use(router)

app.mount('#app')
```

**Step 3: 创建类型声明文件**

```typescript
// web/src/env.d.ts
/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
```

**Step 4: 安装额外依赖**

```bash
cd d:/project/payment/epay-go/web
npm install pinia vue-router@4 axios
npm install -D @types/node
```

**Step 5: 提交**

```bash
cd d:/project/payment/epay-go
git add web/
git commit -m "feat(web): configure vite, install dependencies"
```

---

## Task 2: 创建 API 请求封装

**Files:**
- Create: `epay-go/web/src/api/request.ts`
- Create: `epay-go/web/src/api/types.ts`

**Step 1: 创建请求封装**

```typescript
// web/src/api/request.ts
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { Message } from '@arco-design/web-vue'

const request: AxiosInstance = axios.create({
  baseURL: '',
  timeout: 30000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    if (data.code === 0) {
      return data
    }
    // 业务错误
    Message.error(data.msg || '请求失败')
    return Promise.reject(new Error(data.msg))
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        Message.error('登录已过期，请重新登录')
        localStorage.removeItem('token')
        window.location.href = '/login'
      } else if (status === 403) {
        Message.error('无权访问')
      } else {
        Message.error(data?.msg || '服务器错误')
      }
    } else {
      Message.error('网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
```

**Step 2: 创建类型定义**

```typescript
// web/src/api/types.ts
// 通用响应
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 分页响应
export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// 商户
export interface Merchant {
  id: number
  username: string
  email: string
  phone: string
  api_key: string
  balance: string
  frozen_balance: string
  status: number
  created_at: string
}

// 订单
export interface Order {
  id: number
  trade_no: string
  out_trade_no: string
  merchant_id: number
  channel_id: number
  pay_type: string
  amount: string
  real_amount: string
  fee: string
  name: string
  status: number
  notify_status: number
  paid_at: string | null
  created_at: string
}

// 通道
export interface Channel {
  id: number
  name: string
  plugin: string
  pay_types: string
  config: any
  rate: string
  daily_limit: string
  status: number
  sort: number
}

// 结算
export interface Settlement {
  id: number
  settle_no: string
  merchant_id: number
  amount: string
  fee: string
  actual_amount: string
  account_type: string
  account_no: string
  account_name: string
  status: number
  remark: string
  created_at: string
}

// 资金记录
export interface BalanceRecord {
  id: number
  merchant_id: number
  action: number
  amount: string
  before_balance: string
  after_balance: string
  type: string
  trade_no: string
  created_at: string
}
```

**Step 3: 提交**

```bash
git add web/src/api/
git commit -m "feat(web): add axios request wrapper and types"
```

---

## Task 3: 创建管理后台 API

**Files:**
- Create: `epay-go/web/src/api/admin.ts`

**Step 1: 创建管理后台 API**

```typescript
// web/src/api/admin.ts
import request from './request'
import type { ApiResponse, PageData, Merchant, Order, Channel, Settlement } from './types'

// 登录
export function adminLogin(data: { username: string; password: string }) {
  return request.post<any, ApiResponse>('/admin/auth/login', data)
}

// 仪表盘
export function getDashboard() {
  return request.get<any, ApiResponse>('/admin/dashboard')
}

// 商户列表
export function getMerchants(params: { page: number; page_size: number; status?: number }) {
  return request.get<any, ApiResponse<PageData<Merchant>>>('/admin/merchants', { params })
}

// 商户详情
export function getMerchant(id: number) {
  return request.get<any, ApiResponse<Merchant>>(`/admin/merchants/${id}`)
}

// 更新商户状态
export function updateMerchantStatus(id: number, status: number) {
  return request.patch<any, ApiResponse>(`/admin/merchants/${id}/status`, { status })
}

// 订单列表
export function getOrders(params: { page: number; page_size: number; merchant_id?: number; status?: number }) {
  return request.get<any, ApiResponse<PageData<Order>>>('/admin/orders', { params })
}

// 订单详情
export function getOrder(tradeNo: string) {
  return request.get<any, ApiResponse<Order>>(`/admin/orders/${tradeNo}`)
}

// 重发通知
export function renotifyOrder(tradeNo: string) {
  return request.post<any, ApiResponse>(`/admin/orders/${tradeNo}/renotify`)
}

// 通道列表
export function getChannels(params: { page: number; page_size: number }) {
  return request.get<any, ApiResponse<PageData<Channel>>>('/admin/channels', { params })
}

// 创建通道
export function createChannel(data: Partial<Channel>) {
  return request.post<any, ApiResponse>('/admin/channels', data)
}

// 更新通道
export function updateChannel(id: number, data: Partial<Channel>) {
  return request.put<any, ApiResponse>(`/admin/channels/${id}`, data)
}

// 删除通道
export function deleteChannel(id: number) {
  return request.delete<any, ApiResponse>(`/admin/channels/${id}`)
}

// 结算列表
export function getSettlements(params: { page: number; page_size: number; merchant_id?: number; status?: number }) {
  return request.get<any, ApiResponse<PageData<Settlement>>>('/admin/settlements', { params })
}

// 审核通过
export function approveSettlement(id: number) {
  return request.patch<any, ApiResponse>(`/admin/settlements/${id}/approve`)
}

// 驳回
export function rejectSettlement(id: number, remark: string) {
  return request.patch<any, ApiResponse>(`/admin/settlements/${id}/reject`, { remark })
}
```

**Step 2: 提交**

```bash
git add web/src/api/admin.ts
git commit -m "feat(web): add admin api"
```

---

## Task 4: 创建商户端 API

**Files:**
- Create: `epay-go/web/src/api/merchant.ts`

**Step 1: 创建商户端 API**

```typescript
// web/src/api/merchant.ts
import request from './request'
import type { ApiResponse, PageData, Merchant, Order, Settlement, BalanceRecord } from './types'

// 注册
export function register(data: { username: string; password: string; email?: string; phone?: string }) {
  return request.post<any, ApiResponse>('/merchant/auth/register', data)
}

// 登录
export function login(data: { username: string; password: string }) {
  return request.post<any, ApiResponse>('/merchant/auth/login', data)
}

// 仪表盘
export function getDashboard() {
  return request.get<any, ApiResponse>('/merchant/dashboard')
}

// 获取个人信息
export function getProfile() {
  return request.get<any, ApiResponse<Merchant>>('/merchant/profile')
}

// 更新个人信息
export function updateProfile(data: { email?: string; phone?: string }) {
  return request.put<any, ApiResponse>('/merchant/profile', data)
}

// 修改密码
export function updatePassword(data: { old_password: string; new_password: string }) {
  return request.put<any, ApiResponse>('/merchant/profile/password', data)
}

// 重置API密钥
export function resetApiKey() {
  return request.post<any, ApiResponse<{ api_key: string }>>('/merchant/profile/reset-key')
}

// 订单列表
export function getOrders(params: { page: number; page_size: number; status?: number }) {
  return request.get<any, ApiResponse<PageData<Order>>>('/merchant/orders', { params })
}

// 订单详情
export function getOrder(tradeNo: string) {
  return request.get<any, ApiResponse<Order>>(`/merchant/orders/${tradeNo}`)
}

// 结算列表
export function getSettlements(params: { page: number; page_size: number; status?: number }) {
  return request.get<any, ApiResponse<PageData<Settlement>>>('/merchant/settlements', { params })
}

// 申请结算
export function applySettlement(data: {
  amount: string
  account_type: string
  account_no: string
  account_name: string
}) {
  return request.post<any, ApiResponse>('/merchant/settlements', data)
}

// 资金记录
export function getRecords(params: { page: number; page_size: number }) {
  return request.get<any, ApiResponse<PageData<BalanceRecord>>>('/merchant/records', { params })
}
```

**Step 2: 提交**

```bash
git add web/src/api/merchant.ts
git commit -m "feat(web): add merchant api"
```

---

## Task 5: 创建路由配置

**Files:**
- Create: `epay-go/web/src/router/index.ts`

**Step 1: 创建路由配置**

```typescript
// web/src/router/index.ts
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  // 首页重定向
  {
    path: '/',
    redirect: '/merchant/login',
  },

  // 管理后台
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/Login.vue'),
    meta: { title: '管理员登录' },
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, authType: 'admin' },
    children: [
      {
        path: '',
        redirect: '/admin/dashboard',
      },
      {
        path: 'dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'merchants',
        name: 'AdminMerchants',
        component: () => import('@/views/admin/Merchants.vue'),
        meta: { title: '商户管理' },
      },
      {
        path: 'orders',
        name: 'AdminOrders',
        component: () => import('@/views/admin/Orders.vue'),
        meta: { title: '订单管理' },
      },
      {
        path: 'channels',
        name: 'AdminChannels',
        component: () => import('@/views/admin/Channels.vue'),
        meta: { title: '通道管理' },
      },
      {
        path: 'settlements',
        name: 'AdminSettlements',
        component: () => import('@/views/admin/Settlements.vue'),
        meta: { title: '结算管理' },
      },
    ],
  },

  // 商户中心
  {
    path: '/merchant/login',
    name: 'MerchantLogin',
    component: () => import('@/views/merchant/Login.vue'),
    meta: { title: '商户登录' },
  },
  {
    path: '/merchant/register',
    name: 'MerchantRegister',
    component: () => import('@/views/merchant/Register.vue'),
    meta: { title: '商户注册' },
  },
  {
    path: '/merchant',
    component: () => import('@/layouts/MerchantLayout.vue'),
    meta: { requiresAuth: true, authType: 'merchant' },
    children: [
      {
        path: '',
        redirect: '/merchant/dashboard',
      },
      {
        path: 'dashboard',
        name: 'MerchantDashboard',
        component: () => import('@/views/merchant/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'orders',
        name: 'MerchantOrders',
        component: () => import('@/views/merchant/Orders.vue'),
        meta: { title: '订单管理' },
      },
      {
        path: 'settlements',
        name: 'MerchantSettlements',
        component: () => import('@/views/merchant/Settlements.vue'),
        meta: { title: '结算管理' },
      },
      {
        path: 'records',
        name: 'MerchantRecords',
        component: () => import('@/views/merchant/Records.vue'),
        meta: { title: '资金记录' },
      },
      {
        path: 'profile',
        name: 'MerchantProfile',
        component: () => import('@/views/merchant/Profile.vue'),
        meta: { title: '个人信息' },
      },
    ],
  },

  // 收银台
  {
    path: '/cashier/:tradeNo',
    name: 'Cashier',
    component: () => import('@/views/cashier/Index.vue'),
    meta: { title: '收银台' },
  },

  // 404
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = (to.meta.title as string) || 'EPay'

  if (to.meta.requiresAuth) {
    const authType = to.meta.authType as string
    const tokenKey = authType === 'admin' ? 'admin_token' : 'merchant_token'
    const token = localStorage.getItem(tokenKey)

    if (!token) {
      const loginPath = authType === 'admin' ? '/admin/login' : '/merchant/login'
      next({ path: loginPath, query: { redirect: to.fullPath } })
      return
    }
  }

  next()
})

export default router
```

**Step 2: 提交**

```bash
git add web/src/router/
git commit -m "feat(web): add vue router with admin and merchant routes"
```

---

## Task 6: 创建布局组件

**Files:**
- Create: `epay-go/web/src/layouts/AdminLayout.vue`
- Create: `epay-go/web/src/layouts/MerchantLayout.vue`

**Step 1: 创建管理后台布局**

```vue
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
        <a-menu-item key="settlements">
          <template #icon><icon-swap /></template>
          结算管理
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
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)

const currentRoute = computed(() => {
  const path = route.path.split('/')[2] || 'dashboard'
  return path
})

const handleMenuClick = (key: string) => {
  router.push(`/admin/${key}`)
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
```

**Step 2: 创建商户中心布局**

```vue
<!-- web/src/layouts/MerchantLayout.vue -->
<template>
  <a-layout class="layout">
    <a-layout-sider collapsible v-model:collapsed="collapsed" :width="220">
      <div class="logo">
        <span v-if="!collapsed">EPay 商户中心</span>
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
        <a-menu-item key="orders">
          <template #icon><icon-file /></template>
          订单管理
        </a-menu-item>
        <a-menu-item key="settlements">
          <template #icon><icon-swap /></template>
          结算管理
        </a-menu-item>
        <a-menu-item key="records">
          <template #icon><icon-history /></template>
          资金记录
        </a-menu-item>
        <a-menu-item key="profile">
          <template #icon><icon-settings /></template>
          个人信息
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="header">
        <div class="header-right">
          <span class="balance">余额: ¥{{ balance }}</span>
          <a-dropdown>
            <a-button type="text">
              <icon-user /> {{ username }}
            </a-button>
            <template #content>
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getProfile } from '@/api/merchant'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)
const username = ref('')
const balance = ref('0.00')

const currentRoute = computed(() => {
  const path = route.path.split('/')[2] || 'dashboard'
  return path
})

const handleMenuClick = (key: string) => {
  router.push(`/merchant/${key}`)
}

const handleLogout = () => {
  localStorage.removeItem('merchant_token')
  router.push('/merchant/login')
}

onMounted(async () => {
  try {
    const res = await getProfile()
    username.value = res.data.username
    balance.value = res.data.balance
  } catch (e) {
    // ignore
  }
})
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
  gap: 16px;
}
.balance {
  color: #165dff;
  font-weight: 500;
}
.content {
  margin: 24px;
  padding: 24px;
  background: #fff;
  min-height: calc(100vh - 64px - 48px);
}
</style>
```

**Step 3: 提交**

```bash
git add web/src/layouts/
git commit -m "feat(web): add admin and merchant layout components"
```

---

## Task 7: 创建登录页面

**Files:**
- Create: `epay-go/web/src/views/admin/Login.vue`
- Create: `epay-go/web/src/views/merchant/Login.vue`
- Create: `epay-go/web/src/views/merchant/Register.vue`

**Step 1: 创建管理员登录页**

```vue
<!-- web/src/views/admin/Login.vue -->
<template>
  <div class="login-container">
    <div class="login-box">
      <h2>EPay 管理后台</h2>
      <a-form :model="form" @submit="handleSubmit" layout="vertical">
        <a-form-item field="username" label="用户名" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model="form.username" placeholder="请输入用户名" size="large" />
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model="form.password" placeholder="请输入密码" size="large" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" long size="large" :loading="loading">
            登录
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { adminLogin } from '@/api/admin'

const router = useRouter()
const route = useRoute()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const handleSubmit = async () => {
  loading.value = true
  try {
    const res = await adminLogin(form)
    localStorage.setItem('admin_token', res.data.token)
    localStorage.setItem('token', res.data.token)
    Message.success('登录成功')
    const redirect = route.query.redirect as string || '/admin/dashboard'
    router.push(redirect)
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}
.login-box h2 {
  text-align: center;
  margin-bottom: 32px;
  color: #1d2129;
}
</style>
```

**Step 2: 创建商户登录页**

```vue
<!-- web/src/views/merchant/Login.vue -->
<template>
  <div class="login-container">
    <div class="login-box">
      <h2>EPay 商户中心</h2>
      <a-form :model="form" @submit="handleSubmit" layout="vertical">
        <a-form-item field="username" label="用户名" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model="form.username" placeholder="请输入用户名" size="large" />
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model="form.password" placeholder="请输入密码" size="large" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" long size="large" :loading="loading">
            登录
          </a-button>
        </a-form-item>
        <div class="register-link">
          还没有账号？<router-link to="/merchant/register">立即注册</router-link>
        </div>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { login } from '@/api/merchant'

const router = useRouter()
const route = useRoute()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const handleSubmit = async () => {
  loading.value = true
  try {
    const res = await login(form)
    localStorage.setItem('merchant_token', res.data.token)
    localStorage.setItem('token', res.data.token)
    Message.success('登录成功')
    const redirect = route.query.redirect as string || '/merchant/dashboard'
    router.push(redirect)
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}
.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}
.login-box h2 {
  text-align: center;
  margin-bottom: 32px;
  color: #1d2129;
}
.register-link {
  text-align: center;
  margin-top: 16px;
  color: #86909c;
}
.register-link a {
  color: #165dff;
}
</style>
```

**Step 3: 创建商户注册页**

```vue
<!-- web/src/views/merchant/Register.vue -->
<template>
  <div class="login-container">
    <div class="login-box">
      <h2>商户注册</h2>
      <a-form :model="form" @submit="handleSubmit" layout="vertical">
        <a-form-item field="username" label="用户名" :rules="[{ required: true, message: '请输入用户名' }, { minLength: 4, message: '用户名至少4个字符' }]">
          <a-input v-model="form.username" placeholder="请输入用户名" size="large" />
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="[{ required: true, message: '请输入密码' }, { minLength: 6, message: '密码至少6个字符' }]">
          <a-input-password v-model="form.password" placeholder="请输入密码" size="large" />
        </a-form-item>
        <a-form-item field="email" label="邮箱">
          <a-input v-model="form.email" placeholder="请输入邮箱（选填）" size="large" />
        </a-form-item>
        <a-form-item field="phone" label="手机号">
          <a-input v-model="form.phone" placeholder="请输入手机号（选填）" size="large" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" long size="large" :loading="loading">
            注册
          </a-button>
        </a-form-item>
        <div class="register-link">
          已有账号？<router-link to="/merchant/login">立即登录</router-link>
        </div>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { register } from '@/api/merchant'

const router = useRouter()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  email: '',
  phone: '',
})

const handleSubmit = async () => {
  loading.value = true
  try {
    const res = await register(form)
    localStorage.setItem('merchant_token', res.data.token)
    localStorage.setItem('token', res.data.token)
    Message.success('注册成功')
    router.push('/merchant/dashboard')
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}
.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}
.login-box h2 {
  text-align: center;
  margin-bottom: 32px;
  color: #1d2129;
}
.register-link {
  text-align: center;
  margin-top: 16px;
  color: #86909c;
}
.register-link a {
  color: #165dff;
}
</style>
```

**Step 4: 提交**

```bash
git add web/src/views/
git commit -m "feat(web): add login and register pages"
```

---

## Task 8: 创建管理后台页面

**Files:**
- Create: `epay-go/web/src/views/admin/Dashboard.vue`
- Create: `epay-go/web/src/views/admin/Merchants.vue`
- Create: `epay-go/web/src/views/admin/Orders.vue`
- Create: `epay-go/web/src/views/admin/Channels.vue`
- Create: `epay-go/web/src/views/admin/Settlements.vue`

**Step 1: 创建仪表盘页面**

```vue
<!-- web/src/views/admin/Dashboard.vue -->
<template>
  <div class="dashboard">
    <a-row :gutter="24">
      <a-col :span="8">
        <a-card>
          <a-statistic title="今日订单数" :value="stats.today_order_count" />
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card>
          <a-statistic title="今日交易额" :value="stats.today_order_amount" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card>
          <a-statistic title="商户总数" :value="stats.total_merchants" />
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboard } from '@/api/admin'

const stats = ref({
  today_order_count: 0,
  today_order_amount: '0.00',
  total_merchants: 0,
})

onMounted(async () => {
  try {
    const res = await getDashboard()
    stats.value = res.data
  } catch (e) {
    // ignore
  }
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}
</style>
```

**Step 2: 创建商户管理页面**

```vue
<!-- web/src/views/admin/Merchants.vue -->
<template>
  <div>
    <a-table :data="merchants" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="ID" data-index="id" :width="80" />
        <a-table-column title="用户名" data-index="username" />
        <a-table-column title="邮箱" data-index="email" />
        <a-table-column title="余额" data-index="balance">
          <template #cell="{ record }">¥{{ record.balance }}</template>
        </a-table-column>
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '正常' : '禁用' }}
            </a-tag>
          </template>
        </a-table-column>
        <a-table-column title="注册时间" data-index="created_at" :width="180" />
        <a-table-column title="操作" :width="120">
          <template #cell="{ record }">
            <a-button type="text" size="small" @click="toggleStatus(record)">
              {{ record.status === 1 ? '禁用' : '启用' }}
            </a-button>
          </template>
        </a-table-column>
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getMerchants, updateMerchantStatus } from '@/api/admin'
import type { Merchant } from '@/api/types'

const loading = ref(false)
const merchants = ref<Merchant[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getMerchants({ page: pagination.current, page_size: pagination.pageSize })
    merchants.value = res.data.list
    pagination.total = res.data.total
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

const toggleStatus = async (record: Merchant) => {
  const newStatus = record.status === 1 ? 0 : 1
  try {
    await updateMerchantStatus(record.id, newStatus)
    Message.success('操作成功')
    fetchData()
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 3: 创建订单管理页面**

```vue
<!-- web/src/views/admin/Orders.vue -->
<template>
  <div>
    <a-table :data="orders" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="订单号" data-index="trade_no" :width="200" />
        <a-table-column title="商户订单号" data-index="out_trade_no" :width="180" />
        <a-table-column title="金额" data-index="amount">
          <template #cell="{ record }">¥{{ record.amount }}</template>
        </a-table-column>
        <a-table-column title="支付类型" data-index="pay_type" :width="100" />
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="创建时间" data-index="created_at" :width="180" />
        <a-table-column title="操作" :width="100">
          <template #cell="{ record }">
            <a-button type="text" size="small" @click="handleRenotify(record)" v-if="record.status === 1">
              重发通知
            </a-button>
          </template>
        </a-table-column>
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getOrders, renotifyOrder } from '@/api/admin'
import type { Order } from '@/api/types'

const loading = ref(false)
const orders = ref<Order[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

const statusText = (status: number) => {
  const map: Record<number, string> = { 0: '未支付', 1: '已支付', 2: '已退款' }
  return map[status] || '未知'
}

const statusColor = (status: number) => {
  const map: Record<number, string> = { 0: 'orange', 1: 'green', 2: 'red' }
  return map[status] || 'gray'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getOrders({ page: pagination.current, page_size: pagination.pageSize })
    orders.value = res.data.list
    pagination.total = res.data.total
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

const handleRenotify = async (record: Order) => {
  try {
    await renotifyOrder(record.trade_no)
    Message.success('通知已发送')
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 4: 创建通道管理页面**

```vue
<!-- web/src/views/admin/Channels.vue -->
<template>
  <div>
    <div style="margin-bottom: 16px;">
      <a-button type="primary" @click="showCreateModal = true">新增通道</a-button>
    </div>
    <a-table :data="channels" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="ID" data-index="id" :width="80" />
        <a-table-column title="名称" data-index="name" />
        <a-table-column title="插件" data-index="plugin" :width="100" />
        <a-table-column title="费率" data-index="rate">
          <template #cell="{ record }">{{ (parseFloat(record.rate) * 100).toFixed(2) }}%</template>
        </a-table-column>
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '启用' : '禁用' }}
            </a-tag>
          </template>
        </a-table-column>
        <a-table-column title="操作" :width="150">
          <template #cell="{ record }">
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-popconfirm content="确定删除吗？" @ok="handleDelete(record.id)">
              <a-button type="text" size="small" status="danger">删除</a-button>
            </a-popconfirm>
          </template>
        </a-table-column>
      </template>
    </a-table>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="showCreateModal" :title="editingChannel ? '编辑通道' : '新增通道'" @ok="handleSave">
      <a-form :model="channelForm" layout="vertical">
        <a-form-item field="name" label="名称">
          <a-input v-model="channelForm.name" />
        </a-form-item>
        <a-form-item field="plugin" label="插件">
          <a-select v-model="channelForm.plugin">
            <a-option value="alipay">支付宝</a-option>
            <a-option value="wechat">微信支付</a-option>
            <a-option value="paypal">PayPal</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="pay_types" label="支付方式">
          <a-input v-model="channelForm.pay_types" placeholder="alipay,wxpay" />
        </a-form-item>
        <a-form-item field="rate" label="费率">
          <a-input-number v-model="channelForm.rate" :precision="4" :min="0" :max="1" />
        </a-form-item>
        <a-form-item field="status" label="状态">
          <a-switch v-model="channelForm.status" :checked-value="1" :unchecked-value="0" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getChannels, createChannel, updateChannel, deleteChannel } from '@/api/admin'
import type { Channel } from '@/api/types'

const loading = ref(false)
const channels = ref<Channel[]>([])
const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
const showCreateModal = ref(false)
const editingChannel = ref<Channel | null>(null)
const channelForm = reactive({
  name: '',
  plugin: 'alipay',
  pay_types: '',
  rate: 0.006,
  status: 1,
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getChannels({ page: pagination.current, page_size: pagination.pageSize })
    channels.value = res.data.list
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

const handleEdit = (record: Channel) => {
  editingChannel.value = record
  Object.assign(channelForm, record)
  showCreateModal.value = true
}

const handleSave = async () => {
  try {
    if (editingChannel.value) {
      await updateChannel(editingChannel.value.id, channelForm)
    } else {
      await createChannel(channelForm)
    }
    Message.success('保存成功')
    showCreateModal.value = false
    editingChannel.value = null
    fetchData()
  } catch (e) {
    // ignore
  }
}

const handleDelete = async (id: number) => {
  try {
    await deleteChannel(id)
    Message.success('删除成功')
    fetchData()
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 5: 创建结算管理页面**

```vue
<!-- web/src/views/admin/Settlements.vue -->
<template>
  <div>
    <a-table :data="settlements" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="结算单号" data-index="settle_no" :width="200" />
        <a-table-column title="商户ID" data-index="merchant_id" :width="100" />
        <a-table-column title="金额" data-index="amount">
          <template #cell="{ record }">¥{{ record.amount }}</template>
        </a-table-column>
        <a-table-column title="手续费" data-index="fee">
          <template #cell="{ record }">¥{{ record.fee }}</template>
        </a-table-column>
        <a-table-column title="实际到账" data-index="actual_amount">
          <template #cell="{ record }">¥{{ record.actual_amount }}</template>
        </a-table-column>
        <a-table-column title="收款账号" data-index="account_no" />
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="操作" :width="150">
          <template #cell="{ record }">
            <template v-if="record.status === 0">
              <a-button type="text" size="small" @click="handleApprove(record.id)">通过</a-button>
              <a-button type="text" size="small" status="danger" @click="showRejectModal(record)">驳回</a-button>
            </template>
          </template>
        </a-table-column>
      </template>
    </a-table>

    <!-- 驳回弹窗 -->
    <a-modal v-model:visible="rejectModalVisible" title="驳回结算" @ok="handleReject">
      <a-textarea v-model="rejectRemark" placeholder="请输入驳回原因" />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getSettlements, approveSettlement, rejectSettlement } from '@/api/admin'
import type { Settlement } from '@/api/types'

const loading = ref(false)
const settlements = ref<Settlement[]>([])
const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
const rejectModalVisible = ref(false)
const rejectingId = ref(0)
const rejectRemark = ref('')

const statusText = (status: number) => {
  const map: Record<number, string> = { 0: '待审核', 1: '处理中', 2: '已完成', 3: '已驳回' }
  return map[status] || '未知'
}

const statusColor = (status: number) => {
  const map: Record<number, string> = { 0: 'orange', 1: 'blue', 2: 'green', 3: 'red' }
  return map[status] || 'gray'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getSettlements({ page: pagination.current, page_size: pagination.pageSize })
    settlements.value = res.data.list
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

const handleApprove = async (id: number) => {
  try {
    await approveSettlement(id)
    Message.success('审核通过')
    fetchData()
  } catch (e) {
    // ignore
  }
}

const showRejectModal = (record: Settlement) => {
  rejectingId.value = record.id
  rejectRemark.value = ''
  rejectModalVisible.value = true
}

const handleReject = async () => {
  if (!rejectRemark.value) {
    Message.warning('请输入驳回原因')
    return
  }
  try {
    await rejectSettlement(rejectingId.value, rejectRemark.value)
    Message.success('已驳回')
    rejectModalVisible.value = false
    fetchData()
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 6: 提交**

```bash
git add web/src/views/admin/
git commit -m "feat(web): add admin dashboard, merchants, orders, channels, settlements pages"
```

---

## Task 9: 创建商户端页面

**Files:**
- Create: `epay-go/web/src/views/merchant/Dashboard.vue`
- Create: `epay-go/web/src/views/merchant/Orders.vue`
- Create: `epay-go/web/src/views/merchant/Settlements.vue`
- Create: `epay-go/web/src/views/merchant/Records.vue`
- Create: `epay-go/web/src/views/merchant/Profile.vue`

**Step 1: 创建商户仪表盘**

```vue
<!-- web/src/views/merchant/Dashboard.vue -->
<template>
  <div class="dashboard">
    <a-row :gutter="24">
      <a-col :span="6">
        <a-card>
          <a-statistic title="可用余额" :value="stats.balance" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="冻结余额" :value="stats.frozen_balance" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="今日订单" :value="stats.today_order_count" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="今日交易额" :value="stats.today_order_amount" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboard } from '@/api/merchant'

const stats = ref({
  balance: '0.00',
  frozen_balance: '0.00',
  today_order_count: 0,
  today_order_amount: '0.00',
})

onMounted(async () => {
  try {
    const res = await getDashboard()
    stats.value = res.data
  } catch (e) {
    // ignore
  }
})
</script>
```

**Step 2: 创建订单页面**

```vue
<!-- web/src/views/merchant/Orders.vue -->
<template>
  <div>
    <a-table :data="orders" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="订单号" data-index="trade_no" :width="200" />
        <a-table-column title="商户订单号" data-index="out_trade_no" :width="180" />
        <a-table-column title="商品名称" data-index="name" />
        <a-table-column title="金额" data-index="amount">
          <template #cell="{ record }">¥{{ record.amount }}</template>
        </a-table-column>
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="创建时间" data-index="created_at" :width="180" />
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getOrders } from '@/api/merchant'
import type { Order } from '@/api/types'

const loading = ref(false)
const orders = ref<Order[]>([])
const pagination = reactive({ current: 1, pageSize: 20, total: 0 })

const statusText = (status: number) => {
  const map: Record<number, string> = { 0: '未支付', 1: '已支付', 2: '已退款' }
  return map[status] || '未知'
}

const statusColor = (status: number) => {
  const map: Record<number, string> = { 0: 'orange', 1: 'green', 2: 'red' }
  return map[status] || 'gray'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getOrders({ page: pagination.current, page_size: pagination.pageSize })
    orders.value = res.data.list
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 3: 创建结算页面**

```vue
<!-- web/src/views/merchant/Settlements.vue -->
<template>
  <div>
    <div style="margin-bottom: 16px;">
      <a-button type="primary" @click="showApplyModal = true">申请结算</a-button>
    </div>
    <a-table :data="settlements" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="结算单号" data-index="settle_no" :width="200" />
        <a-table-column title="金额" data-index="amount">
          <template #cell="{ record }">¥{{ record.amount }}</template>
        </a-table-column>
        <a-table-column title="手续费" data-index="fee" />
        <a-table-column title="实际到账" data-index="actual_amount" />
        <a-table-column title="收款账号" data-index="account_no" />
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
          </template>
        </a-table-column>
        <a-table-column title="备注" data-index="remark" />
        <a-table-column title="申请时间" data-index="created_at" :width="180" />
      </template>
    </a-table>

    <!-- 申请结算弹窗 -->
    <a-modal v-model:visible="showApplyModal" title="申请结算" @ok="handleApply">
      <a-form :model="applyForm" layout="vertical">
        <a-form-item field="amount" label="结算金额">
          <a-input-number v-model="applyForm.amount" :precision="2" :min="10" style="width: 100%" />
        </a-form-item>
        <a-form-item field="account_type" label="收款方式">
          <a-radio-group v-model="applyForm.account_type">
            <a-radio value="alipay">支付宝</a-radio>
            <a-radio value="bank">银行卡</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item field="account_no" label="收款账号">
          <a-input v-model="applyForm.account_no" />
        </a-form-item>
        <a-form-item field="account_name" label="收款人姓名">
          <a-input v-model="applyForm.account_name" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getSettlements, applySettlement } from '@/api/merchant'
import type { Settlement } from '@/api/types'

const loading = ref(false)
const settlements = ref<Settlement[]>([])
const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
const showApplyModal = ref(false)
const applyForm = reactive({
  amount: 100,
  account_type: 'alipay',
  account_no: '',
  account_name: '',
})

const statusText = (status: number) => {
  const map: Record<number, string> = { 0: '待审核', 1: '处理中', 2: '已完成', 3: '已驳回' }
  return map[status] || '未知'
}

const statusColor = (status: number) => {
  const map: Record<number, string> = { 0: 'orange', 1: 'blue', 2: 'green', 3: 'red' }
  return map[status] || 'gray'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getSettlements({ page: pagination.current, page_size: pagination.pageSize })
    settlements.value = res.data.list
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

const handleApply = async () => {
  try {
    await applySettlement({
      amount: applyForm.amount.toString(),
      account_type: applyForm.account_type,
      account_no: applyForm.account_no,
      account_name: applyForm.account_name,
    })
    Message.success('申请成功')
    showApplyModal.value = false
    fetchData()
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 4: 创建资金记录页面**

```vue
<!-- web/src/views/merchant/Records.vue -->
<template>
  <div>
    <a-table :data="records" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="时间" data-index="created_at" :width="180" />
        <a-table-column title="类型" data-index="action" :width="80">
          <template #cell="{ record }">
            <a-tag :color="record.action === 1 ? 'green' : 'red'">
              {{ record.action === 1 ? '收入' : '支出' }}
            </a-tag>
          </template>
        </a-table-column>
        <a-table-column title="金额" data-index="amount">
          <template #cell="{ record }">
            <span :style="{ color: record.action === 1 ? '#00b42a' : '#f53f3f' }">
              {{ record.action === 1 ? '+' : '-' }}¥{{ record.amount }}
            </span>
          </template>
        </a-table-column>
        <a-table-column title="变动前余额" data-index="before_balance">
          <template #cell="{ record }">¥{{ record.before_balance }}</template>
        </a-table-column>
        <a-table-column title="变动后余额" data-index="after_balance">
          <template #cell="{ record }">¥{{ record.after_balance }}</template>
        </a-table-column>
        <a-table-column title="说明" data-index="type" />
        <a-table-column title="关联单号" data-index="trade_no" />
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getRecords } from '@/api/merchant'
import type { BalanceRecord } from '@/api/types'

const loading = ref(false)
const records = ref<BalanceRecord[]>([])
const pagination = reactive({ current: 1, pageSize: 20, total: 0 })

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getRecords({ page: pagination.current, page_size: pagination.pageSize })
    records.value = res.data.list
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>
```

**Step 5: 创建个人信息页面**

```vue
<!-- web/src/views/merchant/Profile.vue -->
<template>
  <div>
    <a-card title="账户信息" style="margin-bottom: 24px;">
      <a-descriptions :column="2">
        <a-descriptions-item label="商户ID">{{ profile.id }}</a-descriptions-item>
        <a-descriptions-item label="用户名">{{ profile.username }}</a-descriptions-item>
        <a-descriptions-item label="邮箱">{{ profile.email || '-' }}</a-descriptions-item>
        <a-descriptions-item label="手机号">{{ profile.phone || '-' }}</a-descriptions-item>
        <a-descriptions-item label="可用余额">¥{{ profile.balance }}</a-descriptions-item>
        <a-descriptions-item label="冻结余额">¥{{ profile.frozen_balance }}</a-descriptions-item>
      </a-descriptions>
    </a-card>

    <a-card title="API 密钥" style="margin-bottom: 24px;">
      <a-space>
        <a-input :model-value="profile.api_key" readonly style="width: 400px;" />
        <a-popconfirm content="确定重置API密钥吗？重置后原密钥将立即失效。" @ok="handleResetKey">
          <a-button type="primary" status="warning">重置密钥</a-button>
        </a-popconfirm>
      </a-space>
    </a-card>

    <a-card title="修改密码">
      <a-form :model="passwordForm" @submit="handleUpdatePassword" style="max-width: 400px;">
        <a-form-item field="old_password" label="原密码">
          <a-input-password v-model="passwordForm.old_password" />
        </a-form-item>
        <a-form-item field="new_password" label="新密码">
          <a-input-password v-model="passwordForm.new_password" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit">修改密码</a-button>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getProfile, resetApiKey, updatePassword } from '@/api/merchant'

const profile = ref({
  id: 0,
  username: '',
  email: '',
  phone: '',
  api_key: '',
  balance: '0.00',
  frozen_balance: '0.00',
})

const passwordForm = reactive({
  old_password: '',
  new_password: '',
})

const fetchProfile = async () => {
  try {
    const res = await getProfile()
    profile.value = res.data
  } catch (e) {
    // ignore
  }
}

const handleResetKey = async () => {
  try {
    const res = await resetApiKey()
    profile.value.api_key = res.data.api_key
    Message.success('密钥已重置')
  } catch (e) {
    // ignore
  }
}

const handleUpdatePassword = async () => {
  try {
    await updatePassword(passwordForm)
    Message.success('密码已修改')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchProfile()
})
</script>
```

**Step 6: 提交**

```bash
git add web/src/views/merchant/
git commit -m "feat(web): add merchant dashboard, orders, settlements, records, profile pages"
```

---

## Task 10: 创建收银台和404页面

**Files:**
- Create: `epay-go/web/src/views/cashier/Index.vue`
- Create: `epay-go/web/src/views/NotFound.vue`
- Modify: `epay-go/web/src/App.vue`

**Step 1: 创建收银台页面**

```vue
<!-- web/src/views/cashier/Index.vue -->
<template>
  <div class="cashier-container">
    <div class="cashier-box">
      <div class="header">
        <h2>订单支付</h2>
      </div>
      <div class="content" v-if="order">
        <div class="order-info">
          <div class="row">
            <span class="label">商品名称</span>
            <span class="value">{{ order.name }}</span>
          </div>
          <div class="row">
            <span class="label">订单号</span>
            <span class="value">{{ order.trade_no }}</span>
          </div>
          <div class="row amount-row">
            <span class="label">支付金额</span>
            <span class="amount">¥{{ order.amount }}</span>
          </div>
        </div>
        <div class="qrcode" v-if="qrcodeUrl">
          <img :src="qrcodeUrl" alt="支付二维码" />
          <p>请使用{{ payTypeName }}扫码支付</p>
        </div>
        <div class="actions" v-if="payUrl">
          <a-button type="primary" size="large" @click="goToPay">前往支付</a-button>
        </div>
      </div>
      <div class="loading" v-else>
        <a-spin />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const order = ref<any>(null)
const qrcodeUrl = ref('')
const payUrl = ref('')

const payTypeName = computed(() => {
  const map: Record<string, string> = {
    alipay: '支付宝',
    wxpay: '微信',
    qqpay: 'QQ钱包',
  }
  return map[order.value?.pay_type] || '手机'
})

const goToPay = () => {
  if (payUrl.value) {
    window.location.href = payUrl.value
  }
}

onMounted(async () => {
  const tradeNo = route.params.tradeNo as string
  // TODO: 获取订单信息和支付参数
  // 这里模拟数据
  order.value = {
    trade_no: tradeNo,
    name: '测试商品',
    amount: '10.00',
    pay_type: 'alipay',
  }
})
</script>

<style scoped>
.cashier-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}
.cashier-box {
  width: 400px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}
.header {
  background: linear-gradient(135deg, #165dff 0%, #722ed1 100%);
  color: #fff;
  padding: 24px;
  text-align: center;
}
.header h2 {
  margin: 0;
}
.content {
  padding: 24px;
}
.order-info .row {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}
.order-info .label {
  color: #86909c;
}
.order-info .amount-row {
  border-bottom: none;
  padding-top: 24px;
}
.order-info .amount {
  font-size: 28px;
  font-weight: bold;
  color: #f53f3f;
}
.qrcode {
  text-align: center;
  padding: 24px 0;
}
.qrcode img {
  width: 200px;
  height: 200px;
}
.qrcode p {
  margin-top: 12px;
  color: #86909c;
}
.actions {
  text-align: center;
  padding: 24px 0;
}
.loading {
  padding: 48px;
  text-align: center;
}
</style>
```

**Step 2: 创建404页面**

```vue
<!-- web/src/views/NotFound.vue -->
<template>
  <div class="not-found">
    <a-result status="404" title="404" subtitle="页面不存在">
      <template #extra>
        <a-button type="primary" @click="$router.push('/')">返回首页</a-button>
      </template>
    </a-result>
  </div>
</template>

<style scoped>
.not-found {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
```

**Step 3: 更新 App.vue**

```vue
<!-- web/src/App.vue -->
<template>
  <router-view />
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
html, body, #app {
  height: 100%;
}
</style>
```

**Step 4: 提交**

```bash
git add web/src/views/ web/src/App.vue
git commit -m "feat(web): add cashier page and 404 page"
```

---

## Task 11: 最终构建验证

**Step 1: 验证前端构建**

```bash
cd d:/project/payment/epay-go/web
npm run build
```

Expected: 构建成功，生成 `dist/` 目录

**Step 2: 提交所有更改**

```bash
cd d:/project/payment/epay-go
git add .
git commit -m "feat(web): complete vue3 frontend with admin, merchant, cashier"
```

---

## 阶段四完成检查清单

- [ ] Vite 配置和依赖安装
- [ ] API 请求封装 (`src/api/request.ts`)
- [ ] 类型定义 (`src/api/types.ts`)
- [ ] 管理后台 API (`src/api/admin.ts`)
- [ ] 商户端 API (`src/api/merchant.ts`)
- [ ] 路由配置 (`src/router/index.ts`)
- [ ] 管理后台布局 (`src/layouts/AdminLayout.vue`)
- [ ] 商户端布局 (`src/layouts/MerchantLayout.vue`)
- [ ] 登录/注册页面
- [ ] 管理后台5个页面（仪表盘、商户、订单、通道、结算）
- [ ] 商户端5个页面（仪表盘、订单、结算、记录、个人信息）
- [ ] 收银台页面
- [ ] 404页面
- [ ] 构建验证通过

---

## 页面汇总

| 模块 | 页面 | 路由 |
|------|------|------|
| **管理后台** | 登录 | `/admin/login` |
| | 仪表盘 | `/admin/dashboard` |
| | 商户管理 | `/admin/merchants` |
| | 订单管理 | `/admin/orders` |
| | 通道管理 | `/admin/channels` |
| | 结算管理 | `/admin/settlements` |
| **商户中心** | 登录 | `/merchant/login` |
| | 注册 | `/merchant/register` |
| | 仪表盘 | `/merchant/dashboard` |
| | 订单管理 | `/merchant/orders` |
| | 结算管理 | `/merchant/settlements` |
| | 资金记录 | `/merchant/records` |
| | 个人信息 | `/merchant/profile` |
| **收银台** | 支付页 | `/cashier/:tradeNo` |

---

**下一阶段：** 阶段五将完成 Docker 部署优化、集成测试和文档编写。
