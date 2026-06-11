<script setup lang="ts">
import { defineAsyncComponent, nextTick, ref } from 'vue';
import { HotkeyCommand, HotkeyContext, useHotkeyRuntime } from '@xkit/hotkeys';
import { Layout as TLayout } from 'tdesign-vue-next';
import HomeCreateMemoButton from '@/presentation/components/HomeCreateMemoButton.vue';
import HomeMemoContent from '@/presentation/components/HomeMemoContent.vue';
import HomeMemoEditorDialog from '@/presentation/components/HomeMemoEditorDialog.vue';
import MemoHistoryDrawer from '@/presentation/components/MemoHistoryDrawer.vue';
import { useActiveMemoViewport } from '@/presentation/composables/home/useActiveMemoViewport';
import { useHomeViewData } from '@/presentation/composables/home/useHomeViewData';
import { useHomeKeyboard } from '@/presentation/composables/home/useHomeKeyboard';
import { uiCommandBus } from '@/presentation/ui-command/uiCommandBus';

const HomeCommandDeck = defineAsyncComponent(
  () => import('@/presentation/components/HomeCommandDeck.vue'),
);

const props = withDefaults(
  defineProps<{
    requestBackupExport?: () => void;
    requestBackupImport?: () => void;
  }>(),
  {
    requestBackupExport: () => undefined,
    requestBackupImport: () => undefined,
  },
);

const editorDialogRef = ref<InstanceType<typeof HomeMemoEditorDialog> | null>(
  null,
);
const hotkeyRuntime = useHotkeyRuntime();
const showShortcuts = ref(false);

const {
  bindCommandDeckRef,
  memos,
  fetchMemos,
  buildQuery,
  loading,
  loadingMore,
  hasMore,
  inputText,
  tokens,
  focusSearchInput,
  applySingleTagFilter,
  handleDelete,
  submitSearch,
  addToken,
  removeToken,
  clearAll,
} = useHomeViewData();

const { activeIndex, currentMode } = useHomeKeyboard({
  memos,
  startEdit: (memo) => editorDialogRef.value?.openEdit(memo),
  handleDelete,
});

const { bindMemoAnchor } = useActiveMemoViewport({
  memos,
  activeIndex,
});

const historyDrawerRef = ref<InstanceType<typeof MemoHistoryDrawer> | null>(
  null,
);
const historyMemoId = ref<string | null>(null);

const openHistory = (id: string) => {
  historyMemoId.value = id;
  void nextTick(() => {
    historyDrawerRef.value?.open();
  });
};

const handleRollbackSuccess = async () => {
  await fetchMemos(buildQuery());
};

const handleCreateMemo = (): void => {
  editorDialogRef.value?.openCreate();
};

const handleFocusSearch = (): void => {
  focusSearchInput();
  hotkeyRuntime.setMode('insert');
};

const handleToggleShortcuts = (): void => {
  showShortcuts.value = !showShortcuts.value;
};

const handleOpenTagGovernance = (): void => {
  uiCommandBus.emit('OpenTagGovernance', {});
};
</script>

<template>
  <TLayout>
    <HotkeyContext id="home">
      <HotkeyCommand id="home.memo.create" @run="handleCreateMemo" />
      <HotkeyCommand id="home.search.focus" @run="handleFocusSearch" />
      <HotkeyCommand id="home.shortcuts.toggle" @run="handleToggleShortcuts" />
      <HotkeyCommand
        id="home.tag_governance.open"
        @run="handleOpenTagGovernance"
      />
      <HotkeyCommand id="home.backup.import" @run="props.requestBackupImport" />
      <HotkeyCommand id="home.backup.export" @run="props.requestBackupExport" />

      <HomeCommandDeck
        :ref="bindCommandDeckRef"
        :current-mode="currentMode"
        v-model:input-text="inputText"
        v-model:tokens="tokens"
        v-model:show-shortcuts="showShortcuts"
        @open-create="() => editorDialogRef?.openCreate()"
        @submit-search="submitSearch"
        @add-token="addToken"
        @remove-token="removeToken"
        @clear-all="clearAll"
      />

      <HomeMemoContent
        :memos="memos"
        :loading="loading"
        :loading-more="loadingMore"
        :has-more="hasMore"
        :active-index="activeIndex"
        :bind-memo-anchor="bindMemoAnchor"
        @edit="(memo) => editorDialogRef?.openEdit(memo)"
        @delete="handleDelete"
        @select-tag="(tag) => applySingleTagFilter(tag)"
        @view-history="openHistory"
      />

      <div class="ui-layer-fab fixed bottom-7 right-7 md:hidden">
        <HomeCreateMemoButton @open="() => editorDialogRef?.openCreate()" />
      </div>

      <HomeMemoEditorDialog
        ref="editorDialogRef"
        @success="() => void fetchMemos(buildQuery())"
      />

      <MemoHistoryDrawer
        ref="historyDrawerRef"
        :memo-id="historyMemoId"
        @rollback-success="() => void handleRollbackSuccess()"
      />
    </HotkeyContext>
  </TLayout>
</template>
