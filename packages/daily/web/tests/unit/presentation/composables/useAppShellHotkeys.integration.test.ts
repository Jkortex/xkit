// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { mount, type VueWrapper } from '@vue/test-utils';
import {
  createHotkeyPlugin,
  HotkeyCommand,
  HotkeyContext,
  type HotkeyRuntime,
} from '@xkit/hotkeys';
import type { RouteLocationNormalizedLoaded } from 'vue-router';
import CommandPalette from '@/presentation/components/CommandPalette.vue';
import { useAppCommandPalette } from '@/presentation/composables/useAppCommandPalette';
import { useAppShellHotkeys } from '@/presentation/composables/useAppShellHotkeys';
import { HOME_HOTKEY_COMMANDS } from '@/presentation/hotkeys/homeHotkeys';
import { APP_ROOT_HOTKEY_BINDINGS } from '@/presentation/hotkeys/appRootHotkeys';

describe('useAppShellHotkeys integration', () => {
  let wrapper: VueWrapper<unknown> | null = null;
  let runtime: HotkeyRuntime | null = null;

  const mountHook = (
    options: {
      dialogOpen?: boolean;
      routeName?: string;
      routePath?: string;
    } = {},
  ) => {
    const push = vi.fn().mockResolvedValue(undefined);
    const onAccountMenuToggle = vi.fn();
    const onRandomWalk = vi.fn();
    const onSwitchUser = vi.fn();
    const route = ref({
      fullPath: options.routePath ?? '/stats',
      name: options.routeName ?? 'stats',
      path: options.routePath ?? '/stats',
    }) as { value: RouteLocationNormalizedLoaded };
    const hotkeyPlugin = createHotkeyPlugin({
      bindings: APP_ROOT_HOTKEY_BINDINGS,
      commands: HOME_HOTKEY_COMMANDS,
    });
    runtime = hotkeyPlugin.runtime;

    const hookComponent = defineComponent({
      setup() {
        const commandPalette = useAppCommandPalette();
        useAppShellHotkeys({
          route: route.value,
        });

        if (options.dialogOpen) {
          runtime?.setFlag('dialog.open', true);
        }

        return () =>
          h(HotkeyContext, { id: 'root' }, [
            h(HotkeyCommand, {
              id: 'app.nav.home',
              onRun: () => void push('/'),
            }),
            h(HotkeyCommand, {
              id: 'app.nav.stats',
              onRun: () => void push('/stats'),
            }),
            h(HotkeyCommand, {
              id: 'app.nav.admin_invites',
              onRun: () => void push('/admin/invites'),
            }),
            h(HotkeyCommand, {
              id: 'app.auth.switch_user',
              onRun: () => void onSwitchUser(),
            }),
            h(HotkeyCommand, {
              id: 'app.nav.random_walk',
              onRun: onRandomWalk,
            }),
            h(HotkeyCommand, {
              id: 'app.account_menu.toggle',
              onRun: onAccountMenuToggle,
            }),
            h(HotkeyCommand, {
              id: 'app.command_palette.open',
              onRun: commandPalette.open,
            }),
            h(HotkeyCommand, {
              id: 'app.mode.normal.enter',
              onRun: () => {
                commandPalette.close();
                const active = document.activeElement as HTMLElement | null;
                active?.blur();
                runtime?.setMode('normal');
                runtime?.setFlag('isTyping', false);
              },
            }),
            h(CommandPalette, {
              visible: commandPalette.isOpen.value,
              query: commandPalette.query.value,
              activeIndex: commandPalette.activeIndex.value,
              items: [...commandPalette.items.value],
              'onUpdate:visible': commandPalette.setOpen,
              'onUpdate:query': commandPalette.setQuery,
              onClose: commandPalette.close,
              onExecuteActive: () => void commandPalette.executeActive(),
              onExecuteItem: (id: string) =>
                void commandPalette.executeById(id),
              onMove: commandPalette.moveSelection,
            }),
          ]);
      },
    });

    wrapper = mount(hookComponent, {
      global: {
        plugins: [hotkeyPlugin],
      },
    });

    return {
      onRandomWalk,
      push,
      route,
    };
  };

  beforeEach(() => {
    wrapper = null;
    runtime = null;
    document.body.innerHTML = '';
    vi.restoreAllMocks();
  });

  afterEach(() => {
    wrapper?.unmount();
    wrapper = null;
    runtime = null;
    document.body.innerHTML = '';
  });

  it('executes app-root navigation bindings on non-home shell routes', async () => {
    const { push } = mountHook();

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'g',
      }),
    );
    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'h',
      }),
    );
    await nextTick();

    expect(push).toHaveBeenCalledWith('/');
  });

  it('opens the home command palette from ctrl+k on non-home shell routes', async () => {
    const { push } = mountHook();

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        ctrlKey: true,
        key: 'k',
      }),
    );
    await nextTick();

    expect(push).not.toHaveBeenCalled();
    expect(wrapper?.find('[data-command-input="true"]').exists()).toBe(true);
  });

  it('blocks app-root bindings when dialog.open is active', async () => {
    const { onRandomWalk, push } = mountHook({
      dialogOpen: true,
    });

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'g',
      }),
    );
    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'r',
      }),
    );
    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        ctrlKey: true,
        key: 'k',
      }),
    );
    await nextTick();

    expect(onRandomWalk).not.toHaveBeenCalled();
    expect(push).not.toHaveBeenCalled();
  });
});
