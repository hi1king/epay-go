<!-- web/src/views/merchant/Dashboard.vue -->
<template>
  <div>
    <a-row :gutter="16">
      <a-col :span="6">
        <a-card>
          <a-statistic title="今日订单数" :value="stats.today_order_count" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="今日订单金额" :value="stats.today_order_amount" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="可用余额" :value="stats.balance" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="冻结金额" :value="stats.frozen_balance" :precision="2" prefix="¥" />
        </a-card>
      </a-col>
    </a-row>

    <a-card title="最近订单" style="margin-top: 16px">
      <a-table :data="recentOrders" :loading="loading" :pagination="false">
        <template #columns>
          <a-table-column title="订单号" data-index="trade_no" />
          <a-table-column title="金额" data-index="amount">
            <template #cell="{ record }">¥{{ record.amount }}</template>
          </a-table-column>
          <a-table-column title="状态" data-index="status">
            <template #cell="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'orange'">
                {{ record.status === 1 ? '已支付' : '未支付' }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="创建时间" data-index="created_at" />
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getDashboard } from '@/api/merchant'
import type { Order } from '@/api/types'

const loading = ref(false)
const stats = reactive({
  today_order_count: 0,
  today_order_amount: 0,
  balance: 0,
  frozen_balance: 0,
})
const recentOrders = ref<Order[]>([])

onMounted(async () => {
  loading.value = true
  try {
    const res = await getDashboard()
    Object.assign(stats, res.data)
    recentOrders.value = res.data.recent_orders || []
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})
</script>
