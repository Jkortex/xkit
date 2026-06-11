<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { Dialog as TDialog } from 'tdesign-vue-next';
import { HotkeyCommand, HotkeyContext, useHotkeyRuntime } from '@xkit/hotkeys';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';
import MemoEditor from '@/presentation/components/MemoEditor.vue';
import type { MemoVM } from '@/presentation/view-models/MemoVM';
import type { ResourceVM } from '@/presentation/view-models/ResourceVM';

interface EditMemoValue {
  readonly uuid: string;
  readonly content: string;
  readonly tags: string[];
  readonly resources: ResourceVM[];
}

interface MemoEditorExpose {
  focusEditor: () => void;
}

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const visible = ref(false);
const expanded = ref(false);
const editingMemo = ref<EditMemoValue | undefined>(undefined);
const memoEditorRef = ref<MemoEditorExpose | null>(null);

useDialogOpenFlag(visible);
const runtime = useHotkeyRuntime();
watch(
  visible,
  (val) => {
    runtime.setFlag('home.isEditorOpen', val);
  },
  { immediate: true },
);

const openCreate = (): void => {
  editingMemo.value = undefined;
  expanded.value = false;
  visible.value = true;
};

const openEdit = (memo: MemoVM): void => {
  editingMemo.value = {
    uuid: memo.uuid,
    content: memo.content,
    tags: memo.tags,
    resources: memo.resources,
  };
  expanded.value = false;
  visible.value = true;
};

const close = (): void => {
  visible.value = false;
  expanded.value = false;
};

const toggleExpand = (): void => {
  if (!visible.value) return;
  expanded.value = !expanded.value;
};

const handleSuccess = (): void => {
  close();
  emit('success');
};

const bindMemoEditorRef = (instance: unknown): void => {
  memoEditorRef.value = (instance as MemoEditorExpose | null) ?? null;
};

watch(visible, async (val) => {
  if (val) {
    await nextTick();
    memoEditorRef.value?.focusEditor();
  } else {
    expanded.value = false;
  }
});

defineExpose({
  openCreate,
  openEdit,
  close,
  toggleExpand,
  visible,
});
</script>

<template>
  <HotkeyContext id="memo-editor" :active="visible">
    <HotkeyCommand id="home.editor.close" @run="close" />
    <HotkeyCommand id="home.editor.toggle_expand" @run="toggleExpand" />

    <TDialog
      v-model:visible="visible"
      :header="editingMemo ? '编辑笔记' : '新建笔记'"
      :footer="false"
      :close-on-esc-keydown="false"
      :width="expanded ? 'min(92vw, 980px)' : '620px'"
      destroy-on-close
    >
      <MemoEditor
        :ref="bindMemoEditorRef"
        :edit-memo="editingMemo"
        :expanded="expanded"
        @cancel-edit="close"
        @success="handleSuccess"
      />
    </TDialog>
  </HotkeyContext>
</template>
