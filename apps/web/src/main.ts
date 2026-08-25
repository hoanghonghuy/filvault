import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const savedTheme = localStorage.getItem('filvault.theme')
if (savedTheme === 'dark') {
  document.documentElement.dataset.theme = 'dark'
} else {
  delete document.documentElement.dataset.theme
}

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
