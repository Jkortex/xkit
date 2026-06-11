<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Aside as TAside } from 'tdesign-vue-next';
import ApiKeyDialog from './ApiKeyDialog.vue';
import ImportReportDialog from './ImportReportDialog.vue';
import TagCloudPanel from './TagCloudPanel.vue';
import TagSetQuickPanel from './TagSetQuickPanel.vue';
import SidebarAccountMenu from './sidebar/SidebarAccountMenu.vue';
import SidebarHeader from './sidebar/SidebarHeader.vue';
import SidebarMenu from './sidebar/SidebarMenu.vue';
import SidebarStatsPanel from './sidebar/SidebarStatsPanel.vue';
import { useBackup } from '../composables/useBackup';
import { useStats } from '../composables/useStats';
import { useTheme } from '../composables/useTheme';
import { useTags } from '../composables/useTags';
import { useAuthStore } from '@/infra/stores/useAuthStore';
import { uiCommandBus } from '../ui-command/uiCommandBus';

const router = useRouter();
const auth = useAuthStore();
const accountMenuRef = ref<{
  toggleFromShortcut: () => void;
} | null>(null);
const { stats, fetchStats } = useStats();
const { isDark, toggleTheme } = useTheme();
const { tags, fetchTags } = useTags();

const {
  fileInput,
  importing,
  exporting,
  showImportReport,
  latestImportReport,
  humanizeReason,
  handleImport,
  handleExport,
} = useBackup({
  onImported: () => {
    fetchStats({ reset: true });
    fetchTags({ reset: true });
  },
});

onMounted(() => {
  fetchStats();
  fetchTags();
});

const handleTagClick = (tagName: string) => {
  router.push({ path: '/', query: { tagAny: tagName } });
};

const handleImportReportVisibleChange = (visible: boolean) => {
  showImportReport.value = visible;
};

const handleSwitchUser = async () => {
  await auth.logout();
  await router.replace('/login');
};

const handleOpenAdmin = () => {
  router.push('/admin/invites');
};

const handleOpenApiKeys = () => {
  uiCommandBus.emit('OpenApiKeyManager', {});
};

const sidebarStats = computed(() => {
  if (!stats.value) return null;
  return {
    ...stats.value,
    tagsTotal: tags.value.length,
  };
});

onUnmounted(() => {});

defineExpose({
  requestBackupExport: () => {
    void handleExport();
  },
  requestBackupImport: () => {
    fileInput.value?.click();
  },
  toggleAccountMenu: () => {
    accountMenuRef.value?.toggleFromShortcut();
  },
});

defineEmits<{
  (e: 'random-walk'): void;
}>();
</script>

<template>
  <TAside class="border-r border-border w-72! hidden lg:block">
    <div class="p-4 flex flex-col sticky top-0 h-screen bg-surface">
      <SidebarHeader :is-dark="isDark" @toggle-theme="toggleTheme">
        <template #account>
          <SidebarAccountMenu
            ref="accountMenuRef"
            :username="auth.user?.username"
            :is-admin="auth.isAdmin"
            :importing="importing"
            :exporting="exporting"
            @import="fileInput?.click()"
            @export="handleExport"
            @open-admin="handleOpenAdmin"
            @open-api-keys="handleOpenApiKeys"
            @switch-user="handleSwitchUser"
          />
        </template>
      </SidebarHeader>

      <SidebarMenu @random-walk="$emit('random-walk')" />

      <div v-if="sidebarStats" class="shrink-0">
        <SidebarStatsPanel :stats="sidebarStats" />
      </div>

      <div class="border-t border-border my-1" />

      <div class="shrink-0">
        <TagSetQuickPanel />
      </div>

      <div class="border-t border-border my-1 shrink-0" />

      <div class="flex-1 min-h-0 overflow-y-auto overscroll-contain">
        <TagCloudPanel
          :tags="tags"
          @select-tag="handleTagClick"
          @changed="fetchTags"
        />
      </div>
    </div>
    <input
      ref="fileInput"
      type="file"
      accept=".zip"
      class="hidden"
      @change="handleImport"
    />
  </TAside>

  <ApiKeyDialog />

  <ImportReportDialog
    :visible="showImportReport"
    :report="latestImportReport"
    :humanize-reason="humanizeReason"
    @update:visible="handleImportReportVisibleChange"
  />
</template>

<style scoped>
.overscroll-contain {
  overscroll-behavior: contain;
}
</style>
