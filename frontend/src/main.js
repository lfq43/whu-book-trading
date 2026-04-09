import { createApp } from 'vue'
import { createPinia } from 'pinia'  // Pinia 状态管理
import ElementPlus from 'element-plus'  // Element Plus UI 组件库
import 'element-plus/dist/index.css'  // Element Plus 样式
import * as ElementPlusIconsVue from '@element-plus/icons-vue'  // 图标

import App from './App.vue'
import router from './router'

const app = createApp(App)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}

app.use(createPinia())  // 使用 Pinia 状态管理 token等信息
app.use(router)         // 使用路由
app.use(ElementPlus)    // 使用 Element Plus

app.mount('#app')