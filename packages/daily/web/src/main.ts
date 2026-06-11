import { createApp } from 'vue';
import { createPinia } from 'pinia';
import 'tdesign-vue-next/es/style/index.css';
import './style.css';
import App from './App.vue';
import router from './router';
import { createHotkeyPlugin } from '@xkit/hotkeys';
import { initTheme } from './presentation/theme/themeManager';
import { APP_ROOT_HOTKEY_BINDINGS } from './presentation/hotkeys/appRootHotkeys';
import {
  HOME_HOTKEY_BINDINGS,
  HOME_HOTKEY_COMMANDS,
} from './presentation/hotkeys/homeHotkeys';

initTheme();

const app = createApp(App);
const pinia = createPinia();

const hotkeyPlugin = createHotkeyPlugin({
  bindings: [...HOME_HOTKEY_BINDINGS, ...APP_ROOT_HOTKEY_BINDINGS],
  commands: HOME_HOTKEY_COMMANDS,
});

app.use(pinia);
app.use(router);
app.use(hotkeyPlugin);

app.mount('#app');
