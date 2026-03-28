# 完成前端实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 修复 App.vue 并补全缺失的管理后台页面，使整个前端系统可以正常运行。

**架构：** Vue 3 + TypeScript + Arco Design + Vue Router。需要修复路由渲染机制，并为管理后台创建 CRUD 页面（商户管理、订单管理、通道管理、结算管理）。

**技术栈：** Vue 3, TypeScript, Arco Design Vue, Vue Router, Axios

---

## Task 1: 修复 App.vue 启用路由系统

**文件：**
- Modify: `web/src/App.vue`

**步骤 1: 备份并查看当前 App.vue**

```bash
cat web/src/App.vue
```

预期：看到 HelloWorld 组件和 Vite 默认模板

**步骤 2: 替换 App.vue 为路由视图**

```vue
<script setup lang="ts">
// App.vue - 主应用入口
</script>

<template>
  <router-view />
</template>

<style>
#app {
  min-height: 100vh;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
</style>
```

**步骤 3: 删除无用的 HelloWorld 组件**

```bash
rm web/src/components/HelloWorld.vue
```

**步骤 4: 测试路由是否生效**

```bash
# 在 web 目录启动开发服务器
cd web && npm run dev
```

预期：访问 http://localhost:5173 应该重定向到 /merchant/login，显示商户登录页面（而不是计时器）

**步骤 5: 提交修复**

```bash
git add web/src/App.vue web/src/components/HelloWorld.vue
git commit -m "fix: enable router-view in App.vue

- Replace default Vite template with router-view
- Remove unused HelloWorld component
- Fix routing system not rendering pages

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: 创建商户管理页面

**文件：**
- Create: `web/src/views/admin/Merchants.vue`

**步骤 1: 创建商户管理页面**

```vue
<!-- web/src/views/admin/Merchants.vue -->
<template>
  <div class="merchants-page">
    <div class="page-header">
      <h2>商户管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <template #icon><icon-plus /></template>
        新建商户
      </a-button>
    </div>

    <a-card>
      <a-table
        :columns="columns"
        :data="merchants"
        :loading="loading"
        :pagination="pagination"
        @page-change="handlePageChange"
      >
        <template #status="{ record }">
          <a-tag :color="record.status === 'active' ? 'green' : 'red'">
            {{ record.status === 'active' ? '启用' : '禁用' }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space>
            <a-button size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button
              size="small"
              status="danger"
              @click="handleToggleStatus(record)"
            >
              {{ record.status === 'active' ? '禁用' : '启用' }}
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 创建/编辑商户对话框 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="editingId ? '编辑商户' : '新建商户'"
      @ok="handleSubmit"
      @cancel="handleCancel"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item field="name" label="商户名称" required>
          <a-input v-model="form.name" placeholder="请输入商户名称" />
        </a-form-item>
        <a-form-item field="contact" label="联系人" required>
          <a-input v-model="form.contact" placeholder="请输入联系人" />
        </a-form-item>
        <a-form-item field="email" label="邮箱" required>
          <a-input v-model="form.email" placeholder="请输入邮箱" />
        </a-form-item>
        <a-form-item field="phone" label="手机号" required>
          <a-input v-model="form.phone" placeholder="请输入手机号" />
        </a-form-item>
        <a-form-item field="password" label="密码" :required="!editingId">
          <a-input-password
            v-model="form.password"
            :placeholder="editingId ? '留空则不修改' : '请输入密码'"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'
import { getMerchants, createMerchant, updateMerchant, toggleMerchantStatus } from '@/api/admin'
import type { Merchant } from '@/api/types'

const loading = ref(false)
const merchants = ref<Merchant[]>([])
const modalVisible = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  contact: '',
  email: '',
  phone: '',
  password: '',
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '商户名称', dataIndex: 'name' },
  { title: '联系人', dataIndex: 'contact' },
  { title: '邮箱', dataIndex: 'email' },
  { title: '手机号', dataIndex: 'phone' },
  { title: '状态', slotName: 'status', width: 100 },
  { title: '创建时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', slotName: 'operations', width: 150 },
]

const loadMerchants = async () => {
  loading.value = true
  try {
    const res = await getMerchants({
      page: pagination.current,
      page_size: pagination.pageSize,
    })
    merchants.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  editingId.value = null
  Object.assign(form, {
    name: '',
    contact: '',
    email: '',
    phone: '',
    password: '',
  })
  modalVisible.value = true
}

const handleEdit = (record: Merchant) => {
  editingId.value = record.id
  Object.assign(form, {
    name: record.name,
    contact: record.contact,
    email: record.email,
    phone: record.phone,
    password: '',
  })
  modalVisible.value = true
}

const handleSubmit = async () => {
  try {
    if (editingId.value) {
      await updateMerchant(editingId.value, form)
      Message.success('更新成功')
    } else {
      await createMerchant(form)
      Message.success('创建成功')
    }
    modalVisible.value = false
    loadMerchants()
  } catch (e) {
    // error handled by interceptor
  }
}

const handleCancel = () => {
  modalVisible.value = false
}

const handleToggleStatus = async (record: Merchant) => {
  try {
    await toggleMerchantStatus(record.id)
    Message.success('状态更新成功')
    loadMerchants()
  } catch (e) {
    // error handled by interceptor
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  loadMerchants()
}

onMounted(() => {
  loadMerchants()
})
</script>

<style scoped>
.merchants-page {
  padding: 0;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
</style>
```

**步骤 2: 在 admin API 中添加商户管理方法**

修改 `web/src/api/admin.ts`，添加以下方法（如果不存在）：

```typescript
// 商户管理
export const getMerchants = (params: { page: number; page_size: number }) =>
  request.get<{ list: Merchant[]; total: number }>('/admin/merchants', { params })

export const createMerchant = (data: any) =>
  request.post('/admin/merchants', data)

export const updateMerchant = (id: number, data: any) =>
  request.put(`/admin/merchants/${id}`, data)

export const toggleMerchantStatus = (id: number) =>
  request.put(`/admin/merchants/${id}/toggle-status`)
```

**步骤 3: 在 types.ts 中添加 Merchant 类型**

修改 `web/src/api/types.ts`，添加：

```typescript
export interface Merchant {
  id: number
  name: string
  contact: string
  email: string
  phone: string
  status: 'active' | 'inactive'
  created_at: string
}
```

**步骤 4: 测试商户管理页面**

访问 http://localhost/admin/merchants（需要先登录管理后台）

预期：看到商户列表表格、新建按钮、编辑/禁用操作按钮

**步骤 5: 提交**

```bash
git add web/src/views/admin/Merchants.vue web/src/api/admin.ts web/src/api/types.ts
git commit -m "feat: add merchant management page

- Create Merchants.vue with CRUD operations
- Add merchant API methods (list/create/update/toggle)
- Add Merchant type definition

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: 创建订单管理页面

**文件：**
- Create: `web/src/views/admin/Orders.vue`

**步骤 1: 创建订单管理页面**

```vue
<!-- web/src/views/admin/Orders.vue -->
<template>
  <div class="orders-page">
    <div class="page-header">
      <h2>订单管理</h2>
    </div>

    <a-card>
      <!-- 搜索表单 -->
      <a-form :model="searchForm" layout="inline" style="margin-bottom: 16px">
        <a-form-item field="trade_no" label="订单号">
          <a-input v-model="searchForm.trade_no" placeholder="请输入订单号" style="width: 200px" />
        </a-form-item>
        <a-form-item field="merchant_id" label="商户ID">
          <a-input-number v-model="searchForm.merchant_id" placeholder="商户ID" style="width: 150px" />
        </a-form-item>
        <a-form-item field="status" label="状态">
          <a-select v-model="searchForm.status" placeholder="全部" style="width: 120px" allow-clear>
            <a-option value="pending">待支付</a-option>
            <a-option value="paid">已支付</a-option>
            <a-option value="closed">已关闭</a-option>
            <a-option value="refunding">退款中</a-option>
            <a-option value="refunded">已退款</a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">查询</a-button>
            <a-button @click="handleReset">重置</a-button>
          </a-space>
        </a-form-item>
      </a-form>

      <a-table
        :columns="columns"
        :data="orders"
        :loading="loading"
        :pagination="pagination"
        @page-change="handlePageChange"
      >
        <template #status="{ record }">
          <a-tag :color="getStatusColor(record.status)">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>
        <template #amount="{ record }">
          ¥{{ record.amount }}
        </template>
        <template #operations="{ record }">
          <a-button size="small" @click="handleViewDetail(record)">详情</a-button>
        </template>
      </a-table>
    </a-card>

    <!-- 订单详情对话框 -->
    <a-modal
      v-model:visible="detailVisible"
      title="订单详情"
      :footer="false"
      width="600px"
    >
      <a-descriptions :data="detailData" :column="1" bordered />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getOrders } from '@/api/admin'
import type { Order } from '@/api/types'

const loading = ref(false)
const orders = ref<Order[]>([])
const detailVisible = ref(false)

const searchForm = reactive({
  trade_no: '',
  merchant_id: undefined as number | undefined,
  status: '',
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const columns = [
  { title: '订单号', dataIndex: 'trade_no', width: 180 },
  { title: '商户ID', dataIndex: 'merchant_id', width: 100 },
  { title: '金额', slotName: 'amount', width: 120 },
  { title: '支付通道', dataIndex: 'channel_code', width: 120 },
  { title: '状态', slotName: 'status', width: 100 },
  { title: '创建时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', slotName: 'operations', width: 100 },
]

const detailData = ref<any[]>([])

const loadOrders = async () => {
  loading.value = true
  try {
    const res = await getOrders({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...searchForm,
    })
    orders.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  loadOrders()
}

const handleReset = () => {
  Object.assign(searchForm, {
    trade_no: '',
    merchant_id: undefined,
    status: '',
  })
  handleSearch()
}

const handlePageChange = (page: number) => {
  pagination.current = page
  loadOrders()
}

const handleViewDetail = (record: Order) => {
  detailData.value = [
    { label: '订单号', value: record.trade_no },
    { label: '商户ID', value: record.merchant_id },
    { label: '商户订单号', value: record.out_trade_no },
    { label: '金额', value: `¥${record.amount}` },
    { label: '支付通道', value: record.channel_code },
    { label: '状态', value: getStatusText(record.status) },
    { label: '回调地址', value: record.notify_url || '-' },
    { label: '返回地址', value: record.return_url || '-' },
    { label: '创建时间', value: record.created_at },
    { label: '支付时间', value: record.paid_at || '-' },
  ]
  detailVisible.value = true
}

const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    pending: 'orange',
    paid: 'green',
    closed: 'gray',
    refunding: 'blue',
    refunded: 'purple',
  }
  return colorMap[status] || 'gray'
}

const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    closed: '已关闭',
    refunding: '退款中',
    refunded: '已退款',
  }
  return textMap[status] || status
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 0;
}
.page-header {
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
</style>
```

**步骤 2: 在 admin API 中添加订单查询方法**

修改 `web/src/api/admin.ts`：

```typescript
// 订单管理
export const getOrders = (params: any) =>
  request.get<{ list: Order[]; total: number }>('/admin/orders', { params })
```

**步骤 3: 在 types.ts 中添加 Order 类型**

修改 `web/src/api/types.ts`：

```typescript
export interface Order {
  id: number
  trade_no: string
  out_trade_no: string
  merchant_id: number
  amount: string
  channel_code: string
  status: 'pending' | 'paid' | 'closed' | 'refunding' | 'refunded'
  notify_url?: string
  return_url?: string
  created_at: string
  paid_at?: string
}
```

**步骤 4: 测试订单管理页面**

访问 http://localhost/admin/orders

预期：看到订单列表、搜索表单、状态筛选、详情按钮

**步骤 5: 提交**

```bash
git add web/src/views/admin/Orders.vue web/src/api/admin.ts web/src/api/types.ts
git commit -m "feat: add order management page

- Create Orders.vue with search and detail view
- Add order query API method
- Add Order type definition
- Support filtering by trade_no, merchant_id, status

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: 创建通道管理页面

**文件：**
- Create: `web/src/views/admin/Channels.vue`

**步骤 1: 创建通道管理页面**

```vue
<!-- web/src/views/admin/Channels.vue -->
<template>
  <div class="channels-page">
    <div class="page-header">
      <h2>通道管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <template #icon><icon-plus /></template>
        新建通道
      </a-button>
    </div>

    <a-card>
      <a-table
        :columns="columns"
        :data="channels"
        :loading="loading"
        :pagination="pagination"
        @page-change="handlePageChange"
      >
        <template #status="{ record }">
          <a-tag :color="record.status === 'active' ? 'green' : 'red'">
            {{ record.status === 'active' ? '启用' : '禁用' }}
          </a-tag>
        </template>
        <template #operations="{ record }">
          <a-space>
            <a-button size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button
              size="small"
              status="danger"
              @click="handleToggleStatus(record)"
            >
              {{ record.status === 'active' ? '禁用' : '启用' }}
            </a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 创建/编辑通道对话框 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="editingId ? '编辑通道' : '新建通道'"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="600px"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item field="code" label="通道编码" required>
          <a-input
            v-model="form.code"
            placeholder="如: alipay_scan, wechat_h5"
            :disabled="!!editingId"
          />
        </a-form-item>
        <a-form-item field="name" label="通道名称" required>
          <a-input v-model="form.name" placeholder="如: 支付宝扫码支付" />
        </a-form-item>
        <a-form-item field="type" label="通道类型" required>
          <a-select v-model="form.type" placeholder="请选择通道类型">
            <a-option value="alipay">支付宝</a-option>
            <a-option value="wechat">微信支付</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="fee_rate" label="费率(%)" required>
          <a-input-number
            v-model="form.fee_rate"
            :precision="2"
            :min="0"
            :max="100"
            placeholder="如: 0.6"
          />
        </a-form-item>
        <a-form-item field="config" label="通道配置(JSON)" required>
          <a-textarea
            v-model="form.config"
            :rows="6"
            placeholder='{"app_id":"xxx","private_key":"xxx"}'
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'
import { getChannels, createChannel, updateChannel, toggleChannelStatus } from '@/api/admin'
import type { Channel } from '@/api/types'

const loading = ref(false)
const channels = ref<Channel[]>([])
const modalVisible = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  code: '',
  name: '',
  type: '',
  fee_rate: 0,
  config: '',
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '通道编码', dataIndex: 'code', width: 150 },
  { title: '通道名称', dataIndex: 'name' },
  { title: '类型', dataIndex: 'type', width: 120 },
  { title: '费率(%)', dataIndex: 'fee_rate', width: 100 },
  { title: '状态', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'operations', width: 150 },
]

const loadChannels = async () => {
  loading.value = true
  try {
    const res = await getChannels({
      page: pagination.current,
      page_size: pagination.pageSize,
    })
    channels.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  editingId.value = null
  Object.assign(form, {
    code: '',
    name: '',
    type: '',
    fee_rate: 0,
    config: '',
  })
  modalVisible.value = true
}

const handleEdit = (record: Channel) => {
  editingId.value = record.id
  Object.assign(form, {
    code: record.code,
    name: record.name,
    type: record.type,
    fee_rate: record.fee_rate,
    config: JSON.stringify(record.config, null, 2),
  })
  modalVisible.value = true
}

const handleSubmit = async () => {
  try {
    // 验证 JSON 格式
    JSON.parse(form.config)

    const data = {
      ...form,
      config: JSON.parse(form.config),
    }

    if (editingId.value) {
      await updateChannel(editingId.value, data)
      Message.success('更新成功')
    } else {
      await createChannel(data)
      Message.success('创建成功')
    }
    modalVisible.value = false
    loadChannels()
  } catch (e: any) {
    if (e instanceof SyntaxError) {
      Message.error('配置 JSON 格式错误')
    }
  }
}

const handleCancel = () => {
  modalVisible.value = false
}

const handleToggleStatus = async (record: Channel) => {
  try {
    await toggleChannelStatus(record.id)
    Message.success('状态更新成功')
    loadChannels()
  } catch (e) {
    // error handled by interceptor
  }
}

const handlePageChange = (page: number) => {
  pagination.current = page
  loadChannels()
}

onMounted(() => {
  loadChannels()
})
</script>

<style scoped>
.channels-page {
  padding: 0;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
</style>
```

**步骤 2: 在 admin API 中添加通道管理方法**

修改 `web/src/api/admin.ts`：

```typescript
// 通道管理
export const getChannels = (params: { page: number; page_size: number }) =>
  request.get<{ list: Channel[]; total: number }>('/admin/channels', { params })

export const createChannel = (data: any) =>
  request.post('/admin/channels', data)

export const updateChannel = (id: number, data: any) =>
  request.put(`/admin/channels/${id}`, data)

export const toggleChannelStatus = (id: number) =>
  request.put(`/admin/channels/${id}/toggle-status`)
```

**步骤 3: 在 types.ts 中添加 Channel 类型**

修改 `web/src/api/types.ts`：

```typescript
export interface Channel {
  id: number
  code: string
  name: string
  type: 'alipay' | 'wechat'
  fee_rate: number
  config: Record<string, any>
  status: 'active' | 'inactive'
}
```

**步骤 4: 测试通道管理页面**

访问 http://localhost/admin/channels

预期：看到通道列表、新建按钮、编辑/禁用操作、JSON 配置表单

**步骤 5: 提交**

```bash
git add web/src/views/admin/Channels.vue web/src/api/admin.ts web/src/api/types.ts
git commit -m "feat: add channel management page

- Create Channels.vue with CRUD operations
- Add channel API methods (list/create/update/toggle)
- Add Channel type definition
- Support JSON config editing with validation

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: 创建结算管理页面

**文件：**
- Create: `web/src/views/admin/Settlements.vue`

**步骤 1: 创建结算管理页面**

```vue
<!-- web/src/views/admin/Settlements.vue -->
<template>
  <div class="settlements-page">
    <div class="page-header">
      <h2>结算管理</h2>
    </div>

    <a-card>
      <!-- 搜索表单 -->
      <a-form :model="searchForm" layout="inline" style="margin-bottom: 16px">
        <a-form-item field="merchant_id" label="商户ID">
          <a-input-number v-model="searchForm.merchant_id" placeholder="商户ID" style="width: 150px" />
        </a-form-item>
        <a-form-item field="status" label="状态">
          <a-select v-model="searchForm.status" placeholder="全部" style="width: 120px" allow-clear>
            <a-option value="pending">待审核</a-option>
            <a-option value="processing">处理中</a-option>
            <a-option value="completed">已完成</a-option>
            <a-option value="rejected">已拒绝</a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">查询</a-button>
            <a-button @click="handleReset">重置</a-button>
          </a-space>
        </a-form-item>
      </a-form>

      <a-table
        :columns="columns"
        :data="settlements"
        :loading="loading"
        :pagination="pagination"
        @page-change="handlePageChange"
      >
        <template #status="{ record }">
          <a-tag :color="getStatusColor(record.status)">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>
        <template #amount="{ record }">
          ¥{{ record.amount }}
        </template>
        <template #operations="{ record }">
          <a-space>
            <a-button
              v-if="record.status === 'pending'"
              size="small"
              type="primary"
              @click="handleApprove(record)"
            >
              审核通过
            </a-button>
            <a-button
              v-if="record.status === 'pending'"
              size="small"
              status="danger"
              @click="handleReject(record)"
            >
              拒绝
            </a-button>
            <a-button size="small" @click="handleViewDetail(record)">详情</a-button>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- 详情对话框 -->
    <a-modal
      v-model:visible="detailVisible"
      title="结算详情"
      :footer="false"
      width="600px"
    >
      <a-descriptions :data="detailData" :column="1" bordered />
    </a-modal>

    <!-- 拒绝原因对话框 -->
    <a-modal
      v-model:visible="rejectVisible"
      title="拒绝结算"
      @ok="handleRejectSubmit"
      @cancel="rejectVisible = false"
    >
      <a-form-item label="拒绝原因" required>
        <a-textarea v-model="rejectReason" :rows="4" placeholder="请输入拒绝原因" />
      </a-form-item>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { getSettlements, approveSettlement, rejectSettlement } from '@/api/admin'
import type { Settlement } from '@/api/types'

const loading = ref(false)
const settlements = ref<Settlement[]>([])
const detailVisible = ref(false)
const rejectVisible = ref(false)
const rejectReason = ref('')
const currentSettlement = ref<Settlement | null>(null)

const searchForm = reactive({
  merchant_id: undefined as number | undefined,
  status: '',
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '商户ID', dataIndex: 'merchant_id', width: 100 },
  { title: '结算金额', slotName: 'amount', width: 120 },
  { title: '手续费', dataIndex: 'fee', width: 100 },
  { title: '实际到账', dataIndex: 'actual_amount', width: 120 },
  { title: '状态', slotName: 'status', width: 100 },
  { title: '申请时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', slotName: 'operations', width: 200 },
]

const detailData = ref<any[]>([])

const loadSettlements = async () => {
  loading.value = true
  try {
    const res = await getSettlements({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...searchForm,
    })
    settlements.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  loadSettlements()
}

const handleReset = () => {
  Object.assign(searchForm, {
    merchant_id: undefined,
    status: '',
  })
  handleSearch()
}

const handlePageChange = (page: number) => {
  pagination.current = page
  loadSettlements()
}

const handleApprove = (record: Settlement) => {
  Modal.confirm({
    title: '确认审核通过',
    content: `确认通过商户 ${record.merchant_id} 的结算申请吗？金额：¥${record.amount}`,
    onOk: async () => {
      try {
        await approveSettlement(record.id)
        Message.success('审核通过')
        loadSettlements()
      } catch (e) {
        // error handled by interceptor
      }
    },
  })
}

const handleReject = (record: Settlement) => {
  currentSettlement.value = record
  rejectReason.value = ''
  rejectVisible.value = true
}

const handleRejectSubmit = async () => {
  if (!rejectReason.value.trim()) {
    Message.warning('请输入拒绝原因')
    return
  }

  try {
    await rejectSettlement(currentSettlement.value!.id, { reason: rejectReason.value })
    Message.success('已拒绝')
    rejectVisible.value = false
    loadSettlements()
  } catch (e) {
    // error handled by interceptor
  }
}

const handleViewDetail = (record: Settlement) => {
  detailData.value = [
    { label: 'ID', value: record.id },
    { label: '商户ID', value: record.merchant_id },
    { label: '结算金额', value: `¥${record.amount}` },
    { label: '手续费', value: `¥${record.fee}` },
    { label: '实际到账', value: `¥${record.actual_amount}` },
    { label: '状态', value: getStatusText(record.status) },
    { label: '银行卡号', value: record.bank_card || '-' },
    { label: '开户行', value: record.bank_name || '-' },
    { label: '申请时间', value: record.created_at },
    { label: '处理时间', value: record.processed_at || '-' },
    { label: '备注', value: record.remark || '-' },
  ]
  detailVisible.value = true
}

const getStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    pending: 'orange',
    processing: 'blue',
    completed: 'green',
    rejected: 'red',
  }
  return colorMap[status] || 'gray'
}

const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待审核',
    processing: '处理中',
    completed: '已完成',
    rejected: '已拒绝',
  }
  return textMap[status] || status
}

onMounted(() => {
  loadSettlements()
})
</script>

<style scoped>
.settlements-page {
  padding: 0;
}
.page-header {
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
</style>
```

**步骤 2: 在 admin API 中添加结算管理方法**

修改 `web/src/api/admin.ts`：

```typescript
// 结算管理
export const getSettlements = (params: any) =>
  request.get<{ list: Settlement[]; total: number }>('/admin/settlements', { params })

export const approveSettlement = (id: number) =>
  request.put(`/admin/settlements/${id}/approve`)

export const rejectSettlement = (id: number, data: { reason: string }) =>
  request.put(`/admin/settlements/${id}/reject`, data)
```

**步骤 3: 在 types.ts 中添加 Settlement 类型**

修改 `web/src/api/types.ts`：

```typescript
export interface Settlement {
  id: number
  merchant_id: number
  amount: string
  fee: string
  actual_amount: string
  status: 'pending' | 'processing' | 'completed' | 'rejected'
  bank_card?: string
  bank_name?: string
  remark?: string
  created_at: string
  processed_at?: string
}
```

**步骤 4: 测试结算管理页面**

访问 http://localhost/admin/settlements

预期：看到结算申请列表、审核通过/拒绝按钮、详情查看

**步骤 5: 提交**

```bash
git add web/src/views/admin/Settlements.vue web/src/api/admin.ts web/src/api/types.ts
git commit -m "feat: add settlement management page

- Create Settlements.vue with approval workflow
- Add settlement API methods (list/approve/reject)
- Add Settlement type definition
- Support filtering by merchant_id and status

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: 最终测试和验证

**步骤 1: 构建前端检查编译错误**

```bash
cd web && npm run build
```

预期：构建成功，无 TypeScript 错误

**步骤 2: 启动开发服务器**

```bash
cd web && npm run dev
```

预期：服务器启动在 http://localhost:5173

**步骤 3: 测试所有页面路由**

访问以下路径并验证：
- `/` → 重定向到 `/merchant/login` ✓
- `/admin/login` → 管理员登录页面 ✓
- `/admin/dashboard` → 管理后台仪表盘 ✓
- `/admin/merchants` → 商户管理页面 ✓
- `/admin/orders` → 订单管理页面 ✓
- `/admin/channels` → 通道管理页面 ✓
- `/admin/settlements` → 结算管理页面 ✓
- `/merchant/login` → 商户登录页面 ✓

**步骤 4: 测试路由守卫**

尝试直接访问 `/admin/dashboard` 而不登录，应该重定向到 `/admin/login`

**步骤 5: 最终提交**

```bash
git add .
git commit -m "chore: complete frontend implementation

- Fix App.vue to enable router rendering
- Add all admin management pages (merchants/orders/channels/settlements)
- Complete API integration and type definitions
- All routes and navigation working correctly

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## 总结

实现完成后，前端系统将包含：

1. ✅ **路由系统正常工作**（不再显示计时器）
2. ✅ **商户管理页面** - 创建/编辑/启用/禁用商户
3. ✅ **订单管理页面** - 查询订单、查看详情
4. ✅ **通道管理页面** - 管理支付通道和配置
5. ✅ **结算管理页面** - 审核结算申请

所有页面都使用 Arco Design 组件库，与现有商户中心页面保持一致的设计风格。
