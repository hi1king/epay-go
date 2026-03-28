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
        <a-table-column title="操作" :width="180">
          <template #cell="{ record }">
            <a-button type="text" size="small" @click="handleRenotify(record)" v-if="record.status === 1">
              重发通知
            </a-button>
            <a-button type="text" size="small" @click="openRefundModal(record)" v-if="record.status === 1">
              发起退款
            </a-button>
          </template>
        </a-table-column>
      </template>
    </a-table>

    <a-modal
      v-model:visible="refundVisible"
      title="发起退款"
      :confirm-loading="refundSubmitting"
      @ok="handleCreateRefund"
      @cancel="handleRefundCancel"
    >
      <a-form :model="refundForm" layout="vertical">
        <a-form-item label="订单号">
          <a-input :model-value="currentOrder?.trade_no || ''" readonly />
        </a-form-item>
        <a-form-item label="订单金额">
          <a-input :model-value="currentOrder ? `¥${currentOrder.amount}` : ''" readonly />
        </a-form-item>
        <a-form-item label="退款金额" required>
          <a-input v-model="refundForm.amount" placeholder="请输入退款金额" />
        </a-form-item>
        <a-form-item label="退款原因">
          <a-textarea v-model="refundForm.reason" placeholder="请输入退款原因" :auto-size="{ minRows: 3, maxRows: 5 }" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { createRefund, getOrders, renotifyOrder } from '@/api/admin'
import type { Order } from '@/api/types'

const loading = ref(false)
const orders = ref<Order[]>([])
const refundVisible = ref(false)
const refundSubmitting = ref(false)
const currentOrder = ref<Order | null>(null)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})
const refundForm = reactive({
  amount: '',
  reason: '',
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

const openRefundModal = (record: Order) => {
  currentOrder.value = record
  refundForm.amount = record.amount
  refundForm.reason = ''
  refundVisible.value = true
}

const handleRefundCancel = () => {
  refundVisible.value = false
  currentOrder.value = null
  refundForm.amount = ''
  refundForm.reason = ''
}

const handleCreateRefund = async () => {
  if (!currentOrder.value) {
    return
  }

  if (!refundForm.amount) {
    Message.warning('请输入退款金额')
    return
  }

  refundSubmitting.value = true
  try {
    await createRefund({
      trade_no: currentOrder.value.trade_no,
      amount: refundForm.amount,
      reason: refundForm.reason || undefined,
    })
    Message.success('退款申请已创建')
    handleRefundCancel()
    await fetchData()
  } catch (e) {
    // ignore
  } finally {
    refundSubmitting.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>
