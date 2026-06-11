import { computed, ref, watch, type ComputedRef, type Ref } from 'vue';
import type { MemoVM } from '@/presentation/view-models/MemoVM';

interface UseHomeSelectionStateOptions {
  readonly memos: Ref<MemoVM[]>;
}

interface UseHomeSelectionStateResult {
  readonly activeIndex: Ref<number>;
  readonly selectedMemo: ComputedRef<MemoVM | null>;
  readonly moveSelection: (delta: number) => void;
}

/** Manages the active memo index for the home list. */
export function useHomeSelectionState(
  options: UseHomeSelectionStateOptions,
): UseHomeSelectionStateResult {
  const activeIndex = ref(0);

  const selectedMemo = computed<MemoVM | null>(() => {
    return options.memos.value[activeIndex.value] ?? null;
  });

  const moveSelection = (delta: number): void => {
    if (options.memos.value.length === 0) return;
    const nextIndex = activeIndex.value + delta;

    if (nextIndex < 0) {
      activeIndex.value = 0;
      return;
    }

    if (nextIndex >= options.memos.value.length) {
      activeIndex.value = options.memos.value.length - 1;
      return;
    }

    activeIndex.value = nextIndex;
  };

  watch(
    () => options.memos.value.length,
    (length) => {
      if (length === 0) {
        activeIndex.value = 0;
        return;
      }

      if (activeIndex.value >= length) {
        activeIndex.value = length - 1;
      }
    },
  );

  return {
    activeIndex,
    selectedMemo,
    moveSelection,
  };
}
