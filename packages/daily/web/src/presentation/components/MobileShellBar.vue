<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  Button as TButton,
  Drawer as TDrawer,
  Loading as TLoading,
} from 'tdesign-vue-next';
import {
  BarChart2,
  Download,
  LogOut,
  Menu,
  Moon,
  Sparkles,
  Sun,
  Tag as TagIcon,
  Upload,
  UserCog,
} from 'lucide-vue-next';
import TagCloudPanel from './TagCloudPanel.vue';
import TagSetQuickPanel from './TagSetQuickPanel.vue';
import ImportReportDialog from './ImportReportDialog.vue';
import { useBackup } from '../composables/useBackup';
import { useStats } from '../composables/useStats';
import { useTags } from '../composables/useTags';
import { useTheme } from '../composables/useTheme';
import { useAuthStore } from '@/infra/stores/useAuthStore';

const emit = defineEmits<{
  (e: 'random-walk'): void;
}>();

const router = useRouter();
const auth = useAuthStore();
const drawerVisible = ref(false);
const { stats, fetchStats, loading: statsLoading } = useStats();
const { tags, fetchTags, loading: tagsLoading } = useTags();
const { isDark, toggleTheme } = useTheme();
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

const tagsSummary = computed(() => `${tags.value.length} 个标签`);
const memosSummary = computed(() =>
  stats.value ? `${stats.value.memosTotal} 条记录` : '记录统计加载中',
);

const openDrawer = () => {
  drawerVisible.value = true;
  if (!stats.value) {
    void fetchStats();
  }
  if (tags.value.length === 0) {
    void fetchTags();
  }
};

const closeDrawer = () => {
  drawerVisible.value = false;
};

const navigateTo = async (path: string) => {
  closeDrawer();
  await router.push(path);
};

const handleTagClick = async (tagName: string) => {
  closeDrawer();
  await router.push({ path: '/', query: { tagAny: tagName } });
};

const handleSwitchUser = async () => {
  await auth.logout();
  closeDrawer();
  await router.replace('/login');
};

const handleOpenAdmin = async () => {
  closeDrawer();
  await router.push('/admin/invites');
};

const handleRandomWalk = () => {
  closeDrawer();
  emit('random-walk');
};

const handleImportReportVisibleChange = (value: boolean) => {
  showImportReport.value = value;
};
</script>

<template>
  <div
    class="sticky top-0 z-40 border-b border-border bg-surface/95 px-4 py-3 backdrop-blur lg:hidden"
  >
    <div class="flex items-center gap-3">
      <button
        class="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border bg-page text-secondary"
        @click="openDrawer"
      >
        <Menu :size="18" />
      </button>
      <div class="min-w-0 flex-1">
        <div class="text-sm font-black text-primary-text">Daily</div>
        <div class="truncate text-xs text-muted">
          {{ memosSummary }} · {{ tagsSummary }}
        </div>
      </div>
      <button
        class="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border bg-page text-secondary"
        @click="emit('random-walk')"
      >
        <Sparkles :size="18" />
      </button>
    </div>
  </div>

  <TDrawer
    v-model:visible="drawerVisible"
    header="工作台"
    placement="left"
    size="88%"
    destroy-on-close
  >
    <div class="space-y-6 pb-6">
      <div
        class="flex items-center justify-between gap-3 rounded-2xl border border-border bg-page p-4"
      >
        <div>
          <div class="text-sm font-black text-primary-text">保持轻量操作</div>
          <div class="text-xs text-muted">导航、账号和标签入口集中到这里</div>
        </div>
        <TButton
          variant="text"
          shape="circle"
          size="small"
          @click="toggleTheme"
        >
          <template #icon>
            <Sun v-if="isDark" :size="18" class="text-amber-400" />
            <Moon v-else :size="18" class="text-slate-600" />
          </template>
        </TButton>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <button class="ui-surface-card p-4 text-left" @click="navigateTo('/')">
          <div class="mb-2 flex items-center gap-2 text-sm font-bold">
            <TagIcon :size="16" /> 全部笔记
          </div>
          <div class="text-xs text-muted">回到记录主列表</div>
        </button>
        <button
          class="ui-surface-card p-4 text-left"
          @click="navigateTo('/stats')"
        >
          <div class="mb-2 flex items-center gap-2 text-sm font-bold">
            <BarChart2 :size="16" /> 统计回顾
          </div>
          <div class="text-xs text-muted">看节奏，不离开主应用</div>
        </button>
        <button class="ui-surface-card p-4 text-left" @click="handleRandomWalk">
          <div class="mb-2 flex items-center gap-2 text-sm font-bold">
            <Sparkles :size="16" /> 随机漫步
          </div>
          <div class="text-xs text-muted">用旧笔记打断思维惯性</div>
        </button>
        <button
          v-if="auth.isAdmin"
          class="ui-surface-card p-4 text-left"
          @click="handleOpenAdmin"
        >
          <div class="mb-2 flex items-center gap-2 text-sm font-bold">
            <UserCog :size="16" /> 邀请管理
          </div>
          <div class="text-xs text-muted">管理成员入口</div>
        </button>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <button
          class="ui-surface-card p-4 text-left"
          @click="fileInput?.click()"
        >
          <div class="mb-2 flex items-center gap-2 text-sm font-bold">
            <Upload :size="16" /> 导入备份
          </div>
          <div class="text-xs text-muted">
            {{ importing ? '处理中...' : '恢复已有数据' }}
          </div>
        </button>
        <button class="ui-surface-card p-4 text-left" @click="handleExport">
          <div class="mb-2 flex items-center gap-2 text-sm font-bold">
            <Download :size="16" /> 导出备份
          </div>
          <div class="text-xs text-muted">
            {{ exporting ? '导出中...' : '下载当前快照' }}
          </div>
        </button>
      </div>

      <button
        class="ui-surface-card flex w-full items-center gap-3 p-4 text-left"
        @click="handleSwitchUser"
      >
        <div class="rounded-full bg-accent-soft p-2 text-accent">
          <LogOut :size="16" />
        </div>
        <div>
          <div class="text-sm font-bold">切换账号</div>
          <div class="text-xs text-muted">
            当前用户：{{ auth.user?.username || '未登录' }}
          </div>
        </div>
      </button>

      <div class="rounded-2xl border border-border bg-page">
        <TagSetQuickPanel />
      </div>

      <div class="rounded-2xl border border-border bg-page p-4">
        <div class="mb-3 flex items-center justify-between gap-2">
          <div>
            <div class="text-sm font-bold">标签入口</div>
            <div class="text-xs text-muted">直接点标签回到首页筛选</div>
          </div>
          <TLoading v-if="statsLoading || tagsLoading" size="small" />
        </div>
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
  </TDrawer>

  <ImportReportDialog
    :visible="showImportReport"
    :report="latestImportReport"
    :humanize-reason="humanizeReason"
    @update:visible="handleImportReportVisibleChange"
  />
</template>
