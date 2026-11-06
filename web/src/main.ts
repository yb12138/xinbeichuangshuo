import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import './styles/battle-theme.css'
import './styles/decision-controls.css'
import App from './App.vue'
import { installAudioUnlock } from './audio/sfx'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.mount('#app')
// 浏览器要求用户交互后才能出声，首次点击 / 按键时解锁音频
installAudioUnlock()
