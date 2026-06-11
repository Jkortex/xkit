import { computed, onMounted, onUnmounted, shallowRef } from 'vue';
import { useHotkeyRuntime } from '@xkit/hotkeys';
import { isTypingTarget } from '@/presentation/composables/hotkeys/keyboardScope';
import { useHomeSelectionCommands } from '@/presentation/composables/home/useHomeSelectionCommands';
import type { MemoVM } from '@/presentation/view-models/MemoVM';
import type { Ref } from 'vue';

interface UseHomeKeyboardOptions {
  readonly memos: Ref<MemoVM[]>;
  readonly startEdit: (memo: MemoVM) => void;
  readonly handleDelete: (id: string) => Promise<void>;
}

/** Wires Home view state into the shared hotkey runtime. */
export function useHomeKeyboard(options: UseHomeKeyboardOptions) {
  const runtime = useHotkeyRuntime();
  const snapshot = shallowRef(runtime.getSnapshot());
  const unsubscribeRuntime = runtime.subscribe((nextSnapshot) => {
    snapshot.value = nextSnapshot;
  });

  const selectionCommands = useHomeSelectionCommands({
    memos: options.memos,
    startEdit: options.startEdit,
    handleDelete: options.handleDelete,
  });

  const syncTypingState = (target: EventTarget | null): void => {
    const typing = isTypingTarget(target);
    runtime.setFlag('isTyping', typing);

    if (snapshot.value.mode === 'command') return;
    if (typing) {
      runtime.setMode('insert');
      return;
    }

    runtime.setMode('normal');
  };

  const handleFocusIn = (event: FocusEvent): void => {
    if (snapshot.value.mode === 'command') return;
    syncTypingState(event.target);
  };

  const handleFocusOut = (): void => {
    window.setTimeout(() => {
      if (snapshot.value.mode === 'command') return;
      syncTypingState(document.activeElement);
    }, 0);
  };

  const handleWindowKeydown = async (event: KeyboardEvent): Promise<void> => {
    syncTypingState(event.target);

    if ((event.target as HTMLElement | null)?.dataset.commandInput) {
      const key = event.key;
      if (
        key === 'ArrowDown' ||
        key === 'ArrowUp' ||
        key === 'Enter' ||
        key === 'Escape'
      ) {
        return;
      }
    }

    await runtime.handleKeydown(event);
  };

  onMounted(() => {
    window.addEventListener('keydown', handleWindowKeydown, true);
    window.addEventListener('focusin', handleFocusIn);
    window.addEventListener('focusout', handleFocusOut);
  });

  onUnmounted(() => {
    unsubscribeRuntime();
    window.removeEventListener('keydown', handleWindowKeydown, true);
    window.removeEventListener('focusin', handleFocusIn);
    window.removeEventListener('focusout', handleFocusOut);
    runtime.setMode('normal');
    runtime.setFlag('home.hasSelectedMemo', false);
    runtime.setFlag('home.isEditorOpen', false);
  });

  return {
    activeIndex: selectionCommands.activeIndex,
    currentMode: computed(() => snapshot.value.mode),
  };
}
