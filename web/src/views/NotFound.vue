<!-- web/src/views/NotFound.vue -->
<template>
  <div class="not-found">
    <a-result status="404" title="404" subtitle="抱歉，您访问的页面不存在">
      <template #extra>
        <a-space>
          <a-button type="primary" @click="$router.push('/')">返回首页</a-button>
          <a-button @click="$router.back()">返回上一页</a-button>
        </a-space>
      </template>
    </a-result>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.not-found {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f7f8fa;
}
</style>
