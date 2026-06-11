// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { mount, type VueWrapper } from '@vue/test-utils';
import { createHotkeyPlugin, useCtx, useHotkeyRuntime } from '@xkit/hotkeys';
import { HOME_HOTKEY_COMMANDS } from '@/presentation/hotkeys/homeHotkeys';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';

describe('useDialogOpenFlag', () => {
  let wrapper: VueWrapper<unknown> | null = null;

  beforeEach(() => {
    wrapper = null;
  });

  afterEach(() => {
    wrapper?.unmount();
    wrapper = null;
  });

  it('keeps dialog.open active until all tracked dialogs close', async () => {
    const firstVisible = ref(false);
    const secondVisible = ref(false);
    const snapshots: boolean[] = [];
    const hotkeyPlugin = createHotkeyPlugin({
      bindings: [],
      commands: HOME_HOTKEY_COMMANDS,
    });

    const DialogTracker = defineComponent({
      setup() {
        useDialogOpenFlag(firstVisible);
        useDialogOpenFlag(secondVisible);
        const runtime = useHotkeyRuntime();
        snapshots.push(Boolean(runtime.getSnapshot().flags['dialog.open']));
        runtime.subscribe((snapshot) => {
          snapshots.push(Boolean(snapshot.flags['dialog.open']));
        });
        return () => h('div');
      },
    });

    const Root = defineComponent({
      setup() {
        useCtx({
          id: 'root',
        });
        return () => h(DialogTracker);
      },
    });

    wrapper = mount(Root, {
      global: {
        plugins: [hotkeyPlugin],
      },
    });

    firstVisible.value = true;
    await nextTick();
    expect(
      hotkeyPlugin.runtime.getSnapshot().flags['dialog.open'],
    ).toBeTruthy();

    secondVisible.value = true;
    await nextTick();
    firstVisible.value = false;
    await nextTick();
    expect(
      hotkeyPlugin.runtime.getSnapshot().flags['dialog.open'],
    ).toBeTruthy();

    secondVisible.value = false;
    await nextTick();
    expect(hotkeyPlugin.runtime.getSnapshot().flags['dialog.open']).toBeFalsy();

    expect(snapshots).toContain(true);
    expect(snapshots.at(-1)).toBe(false);
  });
});
