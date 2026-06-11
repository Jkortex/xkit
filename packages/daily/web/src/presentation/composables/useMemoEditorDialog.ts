import { nextTick, ref, watch, type Ref } from 'vue';
import type { MemoVM } from '@/presentation/view-models/MemoVM';
import type { ResourceVM } from '@/presentation/view-models/ResourceVM';

interface MemoEditorExpose {
  focusEditor: () => void;
}

interface UseMemoEditorDialogOptions {
  onRefresh: () => void;
}

interface UseMemoEditorDialogResult {
  showEditorDialog: Ref<boolean>;
  editingMemo: Ref<
    { id: string; content: string; resources: ResourceVM[] } | undefined
  >;
  editorExpanded: Ref<boolean>;
  bindMemoEditorRef: (instance: unknown) => void;
  openCreateEditor: () => void;
  startEdit: (memo: MemoVM) => void;
  closeEditor: () => void;
  toggleEditorExpanded: () => void;
  onEditorSuccess: () => void;
}

/** Manages create/edit dialog state and editor focus lifecycle. */
export function useMemoEditorDialog(
  options: UseMemoEditorDialogOptions,
): UseMemoEditorDialogResult {
  const memoEditorRef = ref<MemoEditorExpose | null>(null);
  const showEditorDialog = ref(false);
  const editingMemo = ref<
    { id: string; content: string; resources: ResourceVM[] } | undefined
  >(undefined);
  const editorExpanded = ref(false);

  const openCreateEditor = (): void => {
    editingMemo.value = undefined;
    editorExpanded.value = false;
    showEditorDialog.value = true;
  };

  const startEdit = (memo: MemoVM): void => {
    editingMemo.value = {
      id: memo.uuid,
      content: memo.content,
      resources: memo.resources,
    };
    editorExpanded.value = false;
    showEditorDialog.value = true;
  };

  const closeEditor = (): void => {
    showEditorDialog.value = false;
    editorExpanded.value = false;
  };

  const toggleEditorExpanded = (): void => {
    if (!showEditorDialog.value) return;
    editorExpanded.value = !editorExpanded.value;
  };

  const onEditorSuccess = (): void => {
    closeEditor();
    options.onRefresh();
  };

  const bindMemoEditorRef = (instance: unknown): void => {
    memoEditorRef.value = (instance as MemoEditorExpose | null) ?? null;
  };

  watch(showEditorDialog, async (visible) => {
    if (!visible) {
      editorExpanded.value = false;
      return;
    }
    await nextTick();
    memoEditorRef.value?.focusEditor();
  });

  return {
    showEditorDialog,
    editingMemo,
    editorExpanded,
    bindMemoEditorRef,
    openCreateEditor,
    startEdit,
    closeEditor,
    toggleEditorExpanded,
    onEditorSuccess,
  };
}
