<!-- web/src/views/merchant/Orders.vue -->
<template>
  <div>
    <a-table :data="orders" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="订单号" data-index="trade_no" :width="200" />
        <a-table-column title="商户订单号" data-index="out_trade_no" :width="180" />
        <a-table-column title="金额" data-index="amount">
          <template #cell="{ record }">¥{{ record.amount }}</template>
        </a-table-column>
        <a-table-column title="实收" data-index="real_amount">
          <template #cell="{ record }">¥{{ record.real_amount }}</template>
        </a-table-column>
        <a-table-column title="手续费" data-index="fee">
          <template #cell="{ record }">¥{{ record.fee }}</template>
        </a-table-column>
        <a-table-column title="支付类型" data-index="pay_type" :width="100" />
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

onMounted(() => {
  fetchData()
})
</script>
