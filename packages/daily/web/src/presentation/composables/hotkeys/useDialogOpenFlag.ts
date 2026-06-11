import { onUnmounted, watch, type MaybeRefOrGetter } from 'vue';
import { toValue } from 'vue';
import { tryUseHotkeyRuntime, type HotkeyRuntime } from '@xkit/hotkeys';

const dialogCounts = new WeakMap<HotkeyRuntime, number>();

const syncDialogFlag = (runtime: HotkeyRuntime, delta: number): void => {
  const current = dialogCounts.get(runtime) ?? 0;
  const next = Math.max(0, current + delta);
  dialogCounts.set(runtime, next);
  runtime.setFlag('dialog.open', next > 0);
};

/**
 * Mirrors a dialog/overlay visibility source into the shared `dialog.open` flag.
 * Multiple dialogs can opt in at once; the flag is released only after all close.
 */
export const useDialogOpenFlag = (visible: MaybeRefOrGetter<boolean>): void => {
  const runtime = tryUseHotkeyRuntime();
  if (!runtime) return;
  let applied = false;

  const setApplied = (nextVisible: boolean): void => {
    if (nextVisible === applied) return;
    syncDialogFlag(runtime, nextVisible ? 1 : -1);
    applied = nextVisible;
  };

  watch(
    () => Boolean(toValue(visible)),
    (nextVisible) => {
      setApplied(nextVisible);
    },
    { immediate: true },
  );

  onUnmounted(() => {
    if (!applied) return;
    syncDialogFlag(runtime, -1);
    applied = false;
  });
};
