<script setup lang="ts">
import { computed, ref } from 'vue';
import { HotkeyCommand, HotkeyContext, useHotkeyRuntime } from '@xkit/hotkeys';
import { useRoute, useRouter } from 'vue-router';
import { Layout as TLayout } from 'tdesign-vue-next';
import CommandPalette from './presentation/components/CommandPalette.vue';
import MobileShellBar from './presentation/components/MobileShellBar.vue';
import Sidebar from './presentation/components/Sidebar.vue';
import RandomMemoDialog from './presentation/components/RandomMemoDialog.vue';
import { useAppCommandPalette } from './presentation/composables/useAppCommandPalette';
import { useAppShellHotkeys } from '@/presentation/composables/useAppShellHotkeys';
import { useAuthStore } from '@/infra/stores/useAuthStore';

const randomMemoDialog = ref<InstanceType<typeof RandomMemoDialog> | null>(
  null,
);
const sidebarRef = ref<InstanceType<typeof Sidebar> | null>(null);
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const hotkeyRuntime = useHotkeyRuntime();
const useAppShell = computed(() => {
  return route.name !== 'login' && route.name !== 'invite-register';
});

const handleRandomWalk = () => {
  randomMemoDialog.value?.open();
};

const handleAccountMenuToggle = () => {
  sidebarRef.value?.toggleAccountMenu();
};

const handleBackupExport = () => {
  sidebarRef.value?.requestBackupExport();
};

const handleBackupImport = () => {
  sidebarRef.value?.requestBackupImport();
};

const commandPalette = useAppCommandPalette();

const handleSwitchUser = async () => {
  try {
    await auth.logout();
    await router.replace('/login');
  } catch {
    // logout failure is non-fatal
  }
};

const handleCommandPaletteOpen = (): void => {
  commandPalette.open();
};

const handleEnterNormalMode = (): void => {
  commandPalette.close();
  const active = document.activeElement as HTMLElement | null;
  active?.blur();
  hotkeyRuntime.setMode('normal');
  hotkeyRuntime.setFlag('isTyping', false);
};

useAppShellHotkeys({
  route,
});
</script>

<template>
  <HotkeyContext id="root" :active="useAppShell">
    <RouterView v-if="!useAppShell" />

    <TLayout v-else class="min-h-screen bg-page">
      <Sidebar ref="sidebarRef" @random-walk="handleRandomWalk" />
      <TLayout>
        <MobileShellBar @random-walk="handleRandomWalk" />
        <RouterView v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component
              :is="Component"
              :request-backup-export="handleBackupExport"
              :request-backup-import="handleBackupImport"
            />
          </transition>
        </RouterView>
      </TLayout>

      <RandomMemoDialog ref="randomMemoDialog" />

      <CommandPalette
        :visible="commandPalette.isOpen.value"
        :query="commandPalette.query.value"
        :items="commandPalette.items.value"
        :active-index="commandPalette.activeIndex.value"
        @update:visible="commandPalette.setOpen"
        @update:query="commandPalette.setQuery"
        @close="commandPalette.close"
        @execute-active="() => void commandPalette.executeActive()"
        @execute-item="(id) => void commandPalette.executeById(id)"
        @move="commandPalette.moveSelection"
      />
    </TLayout>

    <HotkeyCommand id="app.nav.home" @run="() => void router.push('/')" />
    <HotkeyCommand id="app.nav.stats" @run="() => void router.push('/stats')" />
    <HotkeyCommand
      id="app.nav.admin_invites"
      @run="() => void router.push('/admin/invites')"
    />
    <HotkeyCommand
      id="app.auth.switch_user"
      @run="() => void handleSwitchUser()"
    />
    <HotkeyCommand id="app.nav.random_walk" @run="handleRandomWalk" />
    <HotkeyCommand
      id="app.account_menu.toggle"
      @run="handleAccountMenuToggle"
    />
    <HotkeyCommand
      id="app.command_palette.open"
      @run="handleCommandPaletteOpen"
    />
    <HotkeyCommand id="app.mode.normal.enter" @run="handleEnterNormalMode" />
  </HotkeyContext>
</template>

<style>
.t-layout {
  background: transparent !important;
}

.t-aside {
  background: var(--daily-bg-surface) !important;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
