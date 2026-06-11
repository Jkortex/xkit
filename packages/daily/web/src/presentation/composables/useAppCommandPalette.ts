import { computed, onUnmounted, shallowRef } from 'vue';
import {
  useCommandPalette,
  useHotkeyRuntime,
  type HotkeyMode,
} from '@xkit/hotkeys';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';
import { getHomeCommandShortcutLabel } from '@/presentation/hotkeys/homeHotkeys';

interface AppCommandPaletteItem {
  readonly category: string;
  readonly id: string;
  readonly shortcut?: string;
  readonly title: string;
}

interface UseAppCommandPaletteResult {
  readonly close: () => void;
  readonly currentMode: Readonly<{ value: HotkeyMode }>;
  readonly executeActive: () => Promise<boolean>;
  readonly executeById: (id: string) => Promise<boolean>;
  readonly isOpen: Readonly<{ value: boolean }>;
  readonly items: Readonly<{ value: readonly AppCommandPaletteItem[] }>;
  readonly moveSelection: (delta: number) => void;
  readonly activeIndex: Readonly<{ value: number }>;
  readonly open: () => void;
  readonly query: Readonly<{ value: string }>;
  readonly setQuery: (query: string) => void;
  readonly setOpen: (visible: boolean) => void;
}

/** Owns the single app-level command palette instance. */
export const useAppCommandPalette = (): UseAppCommandPaletteResult => {
  const runtime = useHotkeyRuntime();
  const commandPalette = useCommandPalette();
  const snapshot = shallowRef(runtime.getSnapshot());

  useDialogOpenFlag(computed(() => commandPalette.isOpen.value));

  const unsubscribeRuntime = runtime.subscribe((nextSnapshot) => {
    snapshot.value = nextSnapshot;
  });

  onUnmounted(() => {
    unsubscribeRuntime();
  });

  return {
    activeIndex: commandPalette.activeIndex,
    close: commandPalette.close,
    currentMode: computed(() => snapshot.value.mode),
    executeActive: commandPalette.executeActive,
    executeById: commandPalette.executeCommand,
    isOpen: commandPalette.isOpen,
    items: computed(() => {
      return commandPalette.items.value.map((item) => ({
        category: item.category ?? '通用',
        id: item.id,
        shortcut: getHomeCommandShortcutLabel(item.id),
        title: item.title,
      }));
    }),
    moveSelection: commandPalette.moveSelection,
    open: commandPalette.open,
    query: commandPalette.query,
    setQuery: (query: string) => {
      commandPalette.query.value = query;
    },
    setOpen: (visible: boolean) => {
      if (visible) {
        commandPalette.open();
        return;
      }
      commandPalette.close();
    },
  };
};
