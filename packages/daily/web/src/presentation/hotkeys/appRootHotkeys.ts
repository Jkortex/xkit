import { HOTKEY_MODES, type HotkeyBindingDefinition } from '@xkit/hotkeys';

const whenNoDialog = (snapshot: {
  flags: Readonly<Record<string, boolean>>;
}) => {
  return !snapshot.flags['dialog.open'];
};

export const APP_ROOT_HOTKEY_BINDINGS: readonly HotkeyBindingDefinition[] = [
  {
    commandId: 'app.nav.home',
    context: { id: 'root' },
    keys: 'g h',
    when: whenNoDialog,
  },
  {
    commandId: 'app.nav.stats',
    context: { id: 'root' },
    keys: 'g s',
    when: whenNoDialog,
  },
  {
    commandId: 'app.nav.admin_invites',
    context: { id: 'root' },
    keys: 'g i',
    when: whenNoDialog,
  },
  {
    commandId: 'app.auth.switch_user',
    context: { id: 'root' },
    keys: 'g u',
    when: whenNoDialog,
  },
  {
    commandId: 'app.nav.random_walk',
    context: { id: 'root' },
    keys: 'g r',
    when: whenNoDialog,
  },
  {
    commandId: 'app.account_menu.toggle',
    context: { id: 'root' },
    keys: 'alt+m',
    mode: HOTKEY_MODES.NORMAL,
    when: whenNoDialog,
  },
  {
    commandId: 'app.command_palette.open',
    context: { id: 'root' },
    keys: ':',
    mode: HOTKEY_MODES.NORMAL,
    when: whenNoDialog,
  },
  {
    commandId: 'app.command_palette.open',
    context: { id: 'root' },
    keys: 'ctrl+k',
    mode: HOTKEY_MODES.NORMAL,
    when: whenNoDialog,
  },
  {
    commandId: 'app.command_palette.open',
    context: { id: 'root' },
    keys: 'ctrl+k',
    mode: HOTKEY_MODES.INSERT,
    when: whenNoDialog,
  },
  {
    commandId: 'app.command_palette.open',
    context: { id: 'root' },
    keys: 'meta+k',
    mode: HOTKEY_MODES.NORMAL,
    when: whenNoDialog,
  },
  {
    commandId: 'app.command_palette.open',
    context: { id: 'root' },
    keys: 'meta+k',
    mode: HOTKEY_MODES.INSERT,
    when: whenNoDialog,
  },
];
