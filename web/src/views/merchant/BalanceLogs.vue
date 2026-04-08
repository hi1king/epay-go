<!-- web/src/views/merchant/BalanceLogs.vue -->
<template>
  <div>
    <a-table :data="balanceLogs" :loading="loading" :pagination="pagination" @page-change="handlePageChange">
      <template #columns>
        <a-table-column title="ID" data-index="id" :width="60" />
        <a-table-column title="类型" data-index="action" :width="80">
          <template #cell="{ record }">
            <a-tag :color="record.action === 1 ? 'green' : 'red'">
              {{ record.action === 1 ? '收入' : '支出' }}
            </a-tag>
          </template>
        </a-table-column>
        <a-table-column title="金额" data-index="amount" :width="120">
          <template #cell="{ record }">
            <span :class="record.action === 1 ? 'text-green' : 'text-red'">
              {{ record.action === 1 ? '+' : '-' }}¥{{ record.amount }}
            </span>
          </template>
        </a-table-column>
        <a-table-column title="变动前余额" data-index="before_balance" :width="120">
          <template #cell="{ record }">¥{{ record.before_balance }}</template>
        </a-table-column>
        <a-table-column title="变动后余额" data-index="after_balance" :width="120">
          <template #cell="{ record }">¥{{ record.after_balance }}</template>
        </a-table-column>
        <a-table-column title="业务类型" data-index="type" :width="100">
          <template #cell="{ record }">{{ typeText(record.type) }}</template>
        </a-table-column>
        <a-table-column title="关联单号" data-index="trade_no" :width="200" />
        <a-table-column title="时间" data-index="created_at" :width="180" />
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getBalanceLogs } from '@/api/merchant'
import type { MerchantBalanceLog } from '@/api/types'

const loading = ref(false)
const balanceLogs = ref<MerchantBalanceLog[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

const typeText = (type: string) => {
  const map: Record<string, string> = {
    order_income: '订单收入',
    withdraw_freeze: '提现冻结',
    withdraw_unfreeze: '提现退回',
    refund: '退款',
    adjust: '调账',
  }
  return map[type] || type
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getBalanceLogs({ page: pagination.current, page_size: pagination.pageSize })
    balanceLogs.value = res.data.list
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

<style scoped>
.text-green {
  color: #00b42a;
}
.text-red {
  color: #f53f3f;
}
</style>
