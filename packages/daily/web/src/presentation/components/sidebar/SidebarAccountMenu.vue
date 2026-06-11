<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue';
import {
  ChevronDown,
  Download,
  Key,
  LogOut,
  Settings,
  Shield,
  Upload,
} from 'lucide-vue-next';

interface SidebarAccountMenuProps {
  username?: string;
  isAdmin?: boolean;
  importing?: boolean;
  exporting?: boolean;
}

const props = withDefaults(defineProps<SidebarAccountMenuProps>(), {
  username: '',
  isAdmin: false,
  importing: false,
  exporting: false,
});

const emit = defineEmits<{
  (e: 'import'): void;
  (e: 'export'): void;
  (e: 'open-admin'): void;
  (e: 'open-api-keys'): void;
  (e: 'switch-user'): void;
}>();

const open = ref(false);
const rootRef = ref<HTMLElement | null>(null);
const menuPanelRef = ref<HTMLElement | null>(null);
const userInitial = computed(
  () => props.username.trim().charAt(0).toUpperCase() || 'U',
);

const toggleMenu = () => {
  open.value = !open.value;
};

const closeMenu = () => {
  open.value = false;
};

const getMenuActionButtons = (): HTMLButtonElement[] => {
  if (!menuPanelRef.value) return [];
  return Array.from(
    menuPanelRef.value.querySelectorAll<HTMLButtonElement>(
      '[data-menu-action="true"]',
    ),
  );
};

const focusMenuItem = (index: number) => {
  const items = getMenuActionButtons();
  if (items.length === 0) return;
  const safeIndex = Math.max(0, Math.min(index, items.length - 1));
  items[safeIndex]?.focus();
};

const openMenuAndFocus = (index = 0) => {
  open.value = true;
  void nextTick(() => {
    focusMenuItem(index);
  });
};

const triggerImport = () => {
  emit('import');
  closeMenu();
};

const triggerExport = () => {
  emit('export');
  closeMenu();
};

const triggerOpenAdmin = () => {
  emit('open-admin');
  closeMenu();
};

const triggerOpenApiKeys = () => {
  emit('open-api-keys');
  closeMenu();
};

const triggerSwitchUser = () => {
  emit('switch-user');
  closeMenu();
};

const handleDocumentPointerDown = (event: PointerEvent) => {
  if (!open.value) return;
  const target = event.target as Node | null;
  if (!target) return;
  if (!rootRef.value?.contains(target)) {
    closeMenu();
  }
};

const handleDocumentKeydown = (event: KeyboardEvent) => {
  if (!open.value) return;
  if (event.key === 'Escape') {
    closeMenu();
  }
};

const handleTriggerKeydown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowDown') {
    event.preventDefault();
    openMenuAndFocus(0);
    return;
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault();
    openMenuAndFocus(getMenuActionButtons().length - 1);
  }
};

const handleMenuKeydown = (event: KeyboardEvent) => {
  const items = getMenuActionButtons();
  if (items.length === 0) return;
  const current = document.activeElement as HTMLButtonElement | null;
  const currentIndex = Math.max(items.indexOf(current as HTMLButtonElement), 0);
  if (event.key === 'ArrowDown') {
    event.preventDefault();
    focusMenuItem((currentIndex + 1) % items.length);
    return;
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault();
    focusMenuItem((currentIndex - 1 + items.length) % items.length);
    return;
  }
  if (event.key === 'Home') {
    event.preventDefault();
    focusMenuItem(0);
    return;
  }
  if (event.key === 'End') {
    event.preventDefault();
    focusMenuItem(items.length - 1);
  }
};

defineExpose({
  toggleFromShortcut: () => {
    if (open.value) closeMenu();
    else openMenuAndFocus(0);
  },
});

onMounted(() => {
  document.addEventListener('pointerdown', handleDocumentPointerDown);
  document.addEventListener('keydown', handleDocumentKeydown);
});

onUnmounted(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown);
  document.removeEventListener('keydown', handleDocumentKeydown);
});
</script>

<template>
  <div ref="rootRef" class="relative">
    <button
      data-testid="account-menu-trigger"
      class="flex items-center gap-2 rounded-full border border-border bg-page px-1.5 py-1 hover:border-accent"
      @click="toggleMenu"
      @keydown="handleTriggerKeydown"
    >
      <span
        class="flex h-7 w-7 items-center justify-center rounded-full bg-accent text-tiny font-semibold text-white"
      >
        {{ userInitial }}
      </span>
      <ChevronDown :size="14" class="text-muted" />
    </button>

    <div
      v-if="open"
      ref="menuPanelRef"
      data-testid="account-menu"
      class="absolute right-0 z-20 mt-2 w-52 rounded-lg border border-border bg-surface p-1 shadow-md"
      @keydown="handleMenuKeydown"
    >
      <div class="px-2 py-1 text-xs-plus text-muted">
        <div class="font-semibold text-primary-text">
          {{ props.username || '当前用户' }}
        </div>
        <div>数据仅作用于当前账号</div>
      </div>
      <button
        data-testid="account-menu-import"
        data-menu-action="true"
        class="ui-sidebar-action text-left text-sm"
        @click="triggerImport"
      >
        <Upload :size="14" />
        {{ props.importing ? '正在导入...' : '导入备份 (ZIP)' }}
      </button>
      <button
        data-testid="account-menu-export"
        data-menu-action="true"
        class="ui-sidebar-action text-left text-sm"
        @click="triggerExport"
      >
        <Download :size="14" />
        导出备份 (ZIP)
        <span v-if="props.exporting" class="ml-auto text-xs">处理中...</span>
      </button>
      <button
        v-if="props.isAdmin"
        data-testid="account-menu-admin"
        data-menu-action="true"
        class="ui-sidebar-action text-left text-sm"
        @click="triggerOpenAdmin"
      >
        <Shield :size="14" /> 邀请管理
      </button>
      <button
        data-testid="account-menu-api-keys"
        data-menu-action="true"
        class="ui-sidebar-action text-left text-sm"
        @click="triggerOpenApiKeys"
      >
        <Key :size="14" /> API Key 管理
      </button>
      <button
        data-testid="account-menu-switch-user"
        data-menu-action="true"
        class="ui-sidebar-action text-left text-sm"
        @click="triggerSwitchUser"
      >
        <LogOut :size="14" /> 切换账号
      </button>
      <button
        data-menu-action="true"
        class="ui-sidebar-action text-left text-sm"
        @click="closeMenu"
      >
        <Settings :size="14" /> 收起菜单
      </button>
    </div>
  </div>
</template>
