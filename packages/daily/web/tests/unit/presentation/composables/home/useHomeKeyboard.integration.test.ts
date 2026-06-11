// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, nextTick, ref, type Ref } from 'vue';
import { mount, type VueWrapper } from '@vue/test-utils';
import {
  createHotkeyPlugin,
  HotkeyCommand,
  HotkeyContext,
  HOTKEY_MODES,
  useCmd,
  useCtx,
} from '@xkit/hotkeys';
import { useHomeKeyboard } from '@/presentation/composables/home/useHomeKeyboard';
import { useAppCommandPalette } from '@/presentation/composables/useAppCommandPalette';
import {
  HOME_HOTKEY_BINDINGS,
  HOME_HOTKEY_COMMANDS,
} from '@/presentation/hotkeys/homeHotkeys';
import { APP_ROOT_HOTKEY_BINDINGS } from '@/presentation/hotkeys/appRootHotkeys';
import CommandPalette from '@/presentation/components/CommandPalette.vue';
import HomeMemoEditorDialog from '@/presentation/components/HomeMemoEditorDialog.vue';
import type { MemoVM } from '@/presentation/view-models/MemoVM';

interface HookApi {
  readonly commandPaletteOpen: Ref<boolean>;
  readonly currentMode: Ref<string>;
  readonly openCommandPalette: () => void;
  readonly editorDialog: Ref<InstanceType<typeof HomeMemoEditorDialog> | null>;
}

interface MountHookOptions {
  readonly onEditorSuccess?: () => void;
  readonly focusSearchInput?: ReturnType<typeof vi.fn>;
  readonly showEditorDialogInitial?: boolean;
  readonly onSetup?: () => void;
}

const createMemo = (id: number): MemoVM => ({
  content: `memo-${id}`,
  displayDate: '2026/3/8',
  id,
  relativeTime: '刚刚',
  resources: [],
  tags: [],
});

// Mock MemoEditor to avoid deep dependency issues in integration tests
vi.mock('@/presentation/components/MemoEditor.vue', () => ({
  default: defineComponent({
    name: 'MemoEditor',
    props: ['editMemo', 'expanded'],
    emits: ['cancel-edit', 'success'],
    setup(props, { emit: _emit, expose }) {
      const textareaRef = ref<HTMLTextAreaElement | null>(null);
      expose({
        focusEditor: () => textareaRef.value?.focus(),
      });
      return () =>
        h('textarea', {
          ref: textareaRef,
          class: 'mock-editor-textarea',
        });
    },
  }),
}));

describe('useHomeKeyboard integration', () => {
  let wrapper: VueWrapper<unknown> | null = null;
  let api: HookApi | null = null;

  const mountHook = (
    options: MountHookOptions = {},
  ): { focusSearchInput: ReturnType<typeof vi.fn> } => {
    const hotkeyPlugin = createHotkeyPlugin({
      bindings: [...HOME_HOTKEY_BINDINGS, ...APP_ROOT_HOTKEY_BINDINGS],
      commands: HOME_HOTKEY_COMMANDS,
    });
    const focusSearchInput = options.focusSearchInput ?? vi.fn();
    const onEditorSuccess = options.onEditorSuccess ?? vi.fn();

    const keyboardComponent = defineComponent({
      setup() {
        const commandPalette = useAppCommandPalette();
        const editorDialog = ref<InstanceType<
          typeof HomeMemoEditorDialog
        > | null>(null);

        const hook = useHomeKeyboard({
          focusSearchInput,
          handleDelete: vi.fn().mockResolvedValue(undefined),
          memos: ref([createMemo(1)]),
          openCreateEditor: () => editorDialog.value?.openCreate(),
          requestBackupExport: vi.fn(),
          requestBackupImport: vi.fn(),
          startEdit: (memo) => editorDialog.value?.openEdit(memo),
        });

        options.onSetup?.();

        api = {
          commandPaletteOpen: commandPalette.isOpen,
          currentMode: hook.currentMode,
          openCommandPalette: commandPalette.open,
          editorDialog,
        };

        return () =>
          h(HotkeyContext, { id: 'home' }, [
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
                const runtime = hotkeyPlugin.runtime;
                runtime.setMode('normal');
                runtime.setFlag('isTyping', false);
              },
            }),
            h(HotkeyCommand, {
              id: 'home.memo.create',
              onRun: () => editorDialog.value?.openCreate(),
            }),
            h(HotkeyCommand, {
              id: 'home.search.focus',
              onRun: () => {
                focusSearchInput();
                hotkeyPlugin.runtime.setMode(HOTKEY_MODES.INSERT);
              },
            }),
            h(HotkeyCommand, {
              id: 'home.shortcuts.toggle',
              onRun: vi.fn(),
            }),
            h(HotkeyCommand, {
              id: 'home.tag_governance.open',
              onRun: vi.fn(),
            }),
            h(HotkeyCommand, {
              id: 'home.backup.import',
              onRun: vi.fn(),
            }),
            h(HotkeyCommand, {
              id: 'home.backup.export',
              onRun: vi.fn(),
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
            h(HomeMemoEditorDialog, {
              ref: editorDialog,
              onSuccess: onEditorSuccess,
            }),
          ]);
      },
    });

    const rootComponent = defineComponent({
      setup() {
        useCtx({
          id: 'root',
        });

        return () => h(keyboardComponent);
      },
    });

    wrapper = mount(rootComponent, {
      global: {
        plugins: [hotkeyPlugin],
      },
    });

    if (options.showEditorDialogInitial) {
      api?.editorDialog.value?.openCreate();
    }

    return {
      focusSearchInput,
    };
  };

  beforeEach(() => {
    api = null;
    wrapper = null;
    document.body.innerHTML = '';
    vi.restoreAllMocks();
  });

  afterEach(() => {
    wrapper?.unmount();
    wrapper = null;
    document.body.innerHTML = '';
  });

  it('opens the command palette on colon from normal mode', async () => {
    mountHook();

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: ':',
      }),
    );
    await nextTick();

    expect(api?.commandPaletteOpen.value).toBe(true);
    expect(api?.currentMode.value).toBe(HOTKEY_MODES.COMMAND);
  });

  it('focuses search on ctrl+k from an input target', async () => {
    const { focusSearchInput } = mountHook();
    const input = document.createElement('input');
    document.body.appendChild(input);
    input.focus();

    input.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        ctrlKey: true,
        key: 'k',
      }),
    );
    await nextTick();

    expect(api?.commandPaletteOpen.value).toBe(true);
    expect(api?.currentMode.value).toBe(HOTKEY_MODES.COMMAND);
    expect(focusSearchInput).not.toHaveBeenCalled();
  });

  it('executes home-level sequence bindings only when home is the deepest context', async () => {
    const onRandomWalk = vi.fn();
    mountHook({
      onSetup: () => {
        useCmd('app.nav.random_walk', onRandomWalk);
      },
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
    await nextTick();

    expect(onRandomWalk).toHaveBeenCalledTimes(1);
  });

  it('blocks home-level bindings when editor is the deepest context', async () => {
    const onRandomWalk = vi.fn();
    mountHook({
      onSetup: () => {
        useCmd('app.nav.random_walk', onRandomWalk);
      },
      showEditorDialogInitial: true,
    });
    await nextTick(); // Wait for visibility state to propagate

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: ':',
      }),
    );
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
    await nextTick();

    expect(api?.commandPaletteOpen.value).toBe(false);
    expect(onRandomWalk).not.toHaveBeenCalled();
  });

  it('lets the command palette take over input until it closes', async () => {
    const onRandomWalk = vi.fn();
    mountHook({
      onSetup: () => {
        useCmd('app.nav.random_walk', onRandomWalk);
      },
    });

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: ':',
      }),
    );
    await nextTick();

    const paletteInput = wrapper?.find('[data-command-input="true"]');
    expect(api?.commandPaletteOpen.value).toBe(true);
    expect(paletteInput?.exists()).toBe(true);

    paletteInput?.element.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'g',
      }),
    );
    paletteInput?.element.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'r',
      }),
    );
    await nextTick();

    expect(onRandomWalk).not.toHaveBeenCalled();

    paletteInput?.element.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'Escape',
      }),
    );
    await nextTick();

    expect(api?.commandPaletteOpen.value).toBe(false);

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
    await nextTick();

    expect(onRandomWalk).toHaveBeenCalledTimes(1);
  });

  it('closes only the palette on escape when both palette and editor are open', async () => {
    mountHook({
      showEditorDialogInitial: true,
    });
    await nextTick();

    api?.openCommandPalette();
    await nextTick();

    const paletteInput = wrapper?.find('[data-command-input="true"]');
    expect(api?.commandPaletteOpen.value).toBe(true);

    paletteInput?.element.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'Escape',
      }),
    );
    await nextTick();

    expect(api?.commandPaletteOpen.value).toBe(false);
    expect(api?.editorDialog.value?.visible).toBe(true);
  });

  it('closes the editor when escape is pressed from a typing target', async () => {
    mountHook({
      showEditorDialogInitial: true,
    });
    await nextTick();

    const textarea = wrapper?.find('.mock-editor-textarea');
    expect(textarea?.exists()).toBe(true);
    const textareaElement = textarea?.element as
      | HTMLTextAreaElement
      | undefined;
    textareaElement?.focus();
    await nextTick();

    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        bubbles: true,
        cancelable: true,
        key: 'Escape',
      }),
    );
    await nextTick();

    expect(api?.editorDialog.value?.visible).toBe(false);
    expect(api?.currentMode.value).toBe(HOTKEY_MODES.NORMAL);
  });
});
