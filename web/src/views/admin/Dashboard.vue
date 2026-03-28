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
  today_order_amount: 0,
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
