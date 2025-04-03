import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// 导入Naive UI
import naive from 'naive-ui'

// 导入全局样式
import './assets/main.css'

const app = createApp(App)

// 使用Pinia状态管理
app.use(createPinia())

// 使用路由
app.use(router)

// 使用Naive UI
app.use(naive)

app.mount('#app')