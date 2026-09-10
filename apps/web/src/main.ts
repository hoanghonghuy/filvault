import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { installCallModalFocusGuard } from '@/lib/callModalFocusGuard'
import { installRouteFocus } from '@/lib/routeFocus'
import { installSharedTabKeyboard } from '@/lib/sharedTabKeyboard'
import { hydrateAppearance } from '@/lib/theme'

hydrateAppearance()
installCallModalFocusGuard()
installRouteFocus(router)
installSharedTabKeyboard()

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
