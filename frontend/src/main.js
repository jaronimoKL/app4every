import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from '@/router'
import App from './App.vue'
import './index.css'

const app = createApp(App)

app.use(createPinia()) // State management (хранилища Pinia)
app.use(router)        // Маршрутизация (Vue Router)

app.mount('#app')
