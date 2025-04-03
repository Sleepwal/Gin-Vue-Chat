<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <transition name="fade" mode="out-in">
            <router-view />
          </transition>
        </n-dialog-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { ref, provide } from 'vue'
import { darkTheme, NConfigProvider, NMessageProvider, NNotificationProvider, NDialogProvider } from 'naive-ui'

// 主题设置（默认为亮色主题）
const theme = ref(null)

// 提供主题切换功能
const toggleTheme = () => {
  theme.value = theme.value ? null : darkTheme
}

// 将主题切换函数提供给子组件
provide('toggleTheme', toggleTheme)
</script>

<style>
/* 全局样式已在main.css中定义 */

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>