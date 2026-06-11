import { watch, type Ref } from 'vue';
import { useCmd, useHotkeyRuntime } from '@xkit/hotkeys';
import { useHomeSelectionState } from '@/presentation/composables/home/useHomeSelectionState';
import type { MemoVM } from '@/presentation/view-models/MemoVM';

interface UseHomeSelectionCommandsOptions {
  readonly memos: Ref<MemoVM[]>;
  readonly startEdit: (memo: MemoVM) => void;
  readonly handleDelete: (id: string) => Promise<void>;
}

interface UseHomeSelectionCommandsResult {
  readonly activeIndex: Ref<number>;
}

/** Wires selection-driven home commands into the hotkey runtime. */
export function useHomeSelectionCommands(
  options: UseHomeSelectionCommandsOptions,
): UseHomeSelectionCommandsResult {
  const runtime = useHotkeyRuntime();
  const selectionState = useHomeSelectionState({
    memos: options.memos,
  });

  watch(
    () => Boolean(selectionState.selectedMemo.value),
    (hasSelection) => {
      runtime.setFlag('home.hasSelectedMemo', hasSelection);
    },
    { immediate: true },
  );

  useCmd('home.memo.select_next', () => {
    selectionState.moveSelection(1);
  });

  useCmd('home.memo.select_prev', () => {
    selectionState.moveSelection(-1);
  });

  useCmd('home.memo.edit_selected', () => {
    const memo = selectionState.selectedMemo.value;
    if (!memo) return;
    options.startEdit(memo);
  });

  useCmd('home.memo.delete_selected', async () => {
    const memo = selectionState.selectedMemo.value;
    if (!memo) return;

    const ok = window.confirm(`删除笔记 #${memo.uuid}？`);
    if (!ok) return;

    await options.handleDelete(memo.uuid);
  });

  return {
    activeIndex: selectionState.activeIndex,
  };
}
