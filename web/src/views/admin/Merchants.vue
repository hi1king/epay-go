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
