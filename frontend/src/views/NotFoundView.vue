<template>
  <div class="not-found">
    <n-result
      status="404"
      title="404 Not Found"
      description="抱歉，您访问的页面不存在"
    >
      <template #footer>
        <n-button type="primary" @click="router.push('/')">
          返回首页
        </n-button>
      </template>
    </n-result>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.not-found {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}
</style>