<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  Drawer as TDrawer,
  Button as TButton,
  Popconfirm as TPopconfirm,
  Loading as TLoading,
  Tabs as TTabs,
  TabPanel as TTabPanel,
} from 'tdesign-vue-next';
import { History, RotateCcw, Eye, Code2 } from 'lucide-vue-next';
import { useMemoStore } from '@/infra/stores/useMemoStore';
import { useMemoHistory } from '../composables/useMemoHistory';
import type { MemoHistoryVM } from '../view-models/MemoHistoryVM';
import { renderMarkdown } from '@/utils/markdown';
import MemoHistoryDiffView from './MemoHistoryDiffView.vue';
import type { MemoDTO } from '@/application/ports/dto/Memo';

const props = defineProps<{
  memoId: string | null;
}>();

const emit = defineEmits<{
  (e: 'rollback-success', memo: MemoDTO): void;
}>();

const visible = ref(false);
const selectedHistory = ref<MemoHistoryVM | null>(null);
const activeTab = ref<'preview' | 'diff'>('diff');
const store = useMemoStore();

const currentMemo = computed(() => {
  if (!props.memoId) return null;
  return store.memos.find((m) => m.uuid === props.memoId) ?? null;
});

const currentMemoContent = computed(() => {
  return currentMemo.value?.content || '';
});

const { historyList, loading, rollingBack, fetchHistory, rollback } =
  useMemoHistory({
    onRollbackSuccess: (memo) => {
      emit('rollback-success', memo);
      visible.value = false;
    },
  });

const open = () => {
  if (props.memoId) {
    visible.value = true;
    selectedHistory.value = null;
    activeTab.value = 'diff';
    fetchHistory(props.memoId);
  }
};

const selectHistory = (item: MemoHistoryVM) => {
  selectedHistory.value = item;
};

const isCurrentSnapshot = computed(() => {
  if (!selectedHistory.value) return false;
  return selectedHistory.value.content === currentMemoContent.value;
});

watch(
  historyList,
  (list) => {
    if (!visible.value) return;
    if (selectedHistory.value) return;
    if (list.length > 0) {
      selectedHistory.value = list[0] ?? null;
    }
  },
  { flush: 'post' },
);

const handleRollback = () => {
  if (props.memoId && selectedHistory.value) {
    rollback(props.memoId, selectedHistory.value.id);
  }
};

defineExpose({ open });
</script>

<template>
  <TDrawer
    v-model:visible="visible"
    header="版本溯源与对比"
    :footer="false"
    size="80%"
    placement="bottom"
    destroy-on-close
  >
    <div class="history-drawer-body flex h-full gap-6">
      <!-- Left: Timeline List (Professional Style) -->
      <div
        class="history-timeline w-72 flex flex-col border-r border-border pr-6"
      >
        <div class="mb-4 flex items-center justify-between gap-2">
          <div class="text-tiny font-bold text-muted uppercase tracking-widest">
            版本历史 ({{ historyList.length }})
          </div>
        </div>

        <div v-if="loading" class="py-10 flex justify-center">
          <TLoading size="small" text="载入历史..." />
        </div>
        <div
          v-else-if="historyList.length === 0"
          class="py-10 text-center text-muted text-sm bg-page rounded-xl border border-dashed border-border"
        >
          暂无历史记录
        </div>
        <div v-else class="flex-1 overflow-y-auto pr-1 space-y-2">
          <button
            v-for="item in historyList"
            :key="item.id"
            class="w-full text-left rounded-xl border transition-all group relative overflow-hidden timeline-item"
            :class="
              selectedHistory?.id === item.id
                ? 'border-border bg-page border-l-4 border-l-accent p-3'
                : 'border-border/60 bg-surface hover:border-border p-3'
            "
            @click="selectHistory(item)"
          >
            <span
              class="timeline-dot"
              :class="
                selectedHistory?.id === item.id ? 'bg-accent' : 'bg-border'
              "
            ></span>
            <div class="flex justify-between items-start mb-1">
              <span class="text-sm font-bold">{{ item.relativeTime }}</span>
            </div>
            <div class="text-xs-plus text-muted font-mono">
              {{ item.absoluteTime }}
            </div>
          </button>
        </div>
      </div>

      <!-- Right: Detailed Comparison / Preview -->
      <div class="flex-1 flex flex-col min-w-0">
        <Transition name="history-fade" mode="out-in">
          <div
            v-if="!selectedHistory"
            key="empty"
            class="flex-1 flex flex-col items-center justify-center text-muted bg-page rounded-2xl border-2 border-dashed border-border"
          >
            <div
              class="w-16 h-16 rounded-full bg-surface flex items-center justify-center mb-4 shadow-sm"
            >
              <History :size="32" class="opacity-30" />
            </div>
            <p class="text-sm font-medium">请从左侧选择一个版本进行对比</p>
            <p class="text-xs-plus mt-1 opacity-60">
              你可以查看与当前内容的差异并执行回滚
            </p>
          </div>

          <div
            v-else
            :key="selectedHistory.id"
            class="flex-1 flex flex-col min-w-0"
          >
            <!-- Header Bar -->
            <div
              class="flex justify-between items-center mb-6 bg-page p-5 rounded-2xl border border-border shadow-sm"
            >
              <div class="flex items-center gap-4">
                <div class="p-2.5 rounded-xl bg-accent-soft text-accent">
                  <History :size="20" />
                </div>
                <div>
                  <div class="text-sm font-black text-primary-text">
                    版本快照：{{ selectedHistory.relativeTime }}
                  </div>
                  <div class="text-xs-plus text-muted font-mono">
                    ID: {{ selectedHistory.id }}
                  </div>
                </div>
                <span
                  v-if="isCurrentSnapshot"
                  class="text-tiny bg-green-500/15 text-green-700 border border-green-500/30 px-2 py-0.5 rounded-full font-bold"
                >
                  当前版本
                </span>
              </div>
              <TPopconfirm
                content="恢复此版本将覆盖当前笔记内容，确定吗？"
                @confirm="handleRollback"
              >
                <TButton
                  theme="primary"
                  shape="round"
                  :loading="rollingBack"
                  class="px-6!"
                >
                  <template #icon><RotateCcw :size="16" /></template>
                  立即恢复
                </TButton>
              </TPopconfirm>
            </div>

            <!-- Tabs for View Mode -->
            <TTabs
              v-model="activeTab"
              class="ui-history-tabs flex-1 flex flex-col min-h-0"
            >
              <TTabPanel value="diff" label="差异对比">
                <template #label>
                  <div class="flex items-center gap-2 px-2">
                    <Code2 :size="14" /> 差异对比
                  </div>
                </template>
                <div
                  class="mt-4 flex-1 overflow-y-auto rounded-2xl border border-border bg-surface p-4"
                >
                  <div
                    class="mb-4 flex items-center gap-2 text-tiny text-muted uppercase font-bold tracking-tighter"
                  >
                    <span
                      class="inline-block w-2 h-2 rounded-full bg-red-500/50"
                    ></span>
                    历史版本 ({{ selectedHistory.content.length }} 字)
                    <span class="mx-2">vs</span>
                    <span
                      class="inline-block w-2 h-2 rounded-full bg-green-500/50"
                    ></span>
                    当前活跃 ({{ currentMemoContent.length }} 字)
                  </div>
                  <MemoHistoryDiffView
                    :old-text="selectedHistory.content"
                    :new-text="currentMemoContent"
                  />
                </div>
              </TTabPanel>

              <TTabPanel value="preview" label="渲染预览">
                <template #label>
                  <div class="flex items-center gap-2 px-2">
                    <Eye :size="14" /> 渲染预览
                  </div>
                </template>
                <div
                  class="mt-4 flex-1 overflow-y-auto rounded-2xl border border-border bg-surface p-6"
                >
                  <div
                    class="prose prose-sm max-w-none markdown-body"
                    v-html="renderMarkdown(selectedHistory.content)"
                  ></div>

                  <div
                    v-if="selectedHistory.tags.length > 0"
                    class="mt-8 flex flex-wrap gap-2 pt-6 border-t border-border"
                  >
                    <span
                      v-for="tag in selectedHistory.tags"
                      :key="tag"
                      class="inline-flex items-center rounded-full bg-accent-soft px-2.5 py-0.5 text-tiny font-bold text-accent"
                    >
                      #{{ tag }}
                    </span>
                  </div>
                </div>
              </TTabPanel>
            </TTabs>
          </div>
        </Transition>
      </div>
    </div>
  </TDrawer>
</template>

<style scoped>
.history-fade-enter-active,
.history-fade-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}
.history-fade-enter-from,
.history-fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
.history-timeline {
  position: relative;
}
.history-timeline::before {
  content: '';
  position: absolute;
  left: 8px;
  top: 46px;
  bottom: 18px;
  width: 1px;
  background: var(--daily-border);
  opacity: 0.6;
}
.timeline-item {
  padding-left: 2rem;
}
.timeline-dot {
  position: absolute;
  left: 4px;
  top: 1.05rem;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  box-shadow: 0 0 0 2px var(--daily-bg-page);
}
.history-drawer-body {
  max-width: 1100px;
  margin: 0 auto;
  width: 100%;
}
.ui-history-tabs :deep(.t-tabs__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.ui-history-tabs :deep(.t-tabs__panel) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.markdown-body :deep(p) {
  margin-bottom: 0.5rem;
}
</style>
