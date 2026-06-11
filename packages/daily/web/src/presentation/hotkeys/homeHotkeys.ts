import {
  HOTKEY_MODES,
  type HotkeyBindingDefinition,
  type HotkeyCommandDefinition,
  type HotkeyMode,
  type HotkeySnapshot,
} from '@xkit/hotkeys';
import { APP_ROOT_HOTKEY_BINDINGS } from './appRootHotkeys';

const hasFlag = (snapshot: HotkeySnapshot, name: string): boolean => {
  return snapshot.flags[name] ?? false;
};

const whenNoDialog = (snapshot: HotkeySnapshot): boolean => {
  return !hasFlag(snapshot, 'dialog.open');
};

export const HOME_HOTKEY_COMMANDS: readonly HotkeyCommandDefinition[] = [
  {
    id: 'home.memo.create',
    title: '新建笔记',
    category: '笔记',
    aliases: ['new', 'create memo'],
  },
  {
    id: 'home.search.focus',
    title: '聚焦搜索',
    category: '导航',
    aliases: ['search', 'find'],
  },
  {
    id: 'home.memo.select_next',
    title: '下一条笔记',
    category: '笔记',
    aliases: ['next memo', 'down'],
  },
  {
    id: 'home.memo.select_prev',
    title: '上一条笔记',
    category: '笔记',
    aliases: ['prev memo', 'up'],
  },
  {
    id: 'home.memo.edit_selected',
    title: '编辑当前笔记',
    category: '笔记',
    aliases: ['edit memo'],
    isEnabled: (snapshot) => hasFlag(snapshot, 'home.hasSelectedMemo'),
  },
  {
    id: 'home.memo.delete_selected',
    title: '删除当前笔记',
    category: '笔记',
    aliases: ['remove memo'],
    isEnabled: (snapshot) => hasFlag(snapshot, 'home.hasSelectedMemo'),
  },
  {
    id: 'home.shortcuts.toggle',
    title: '显示/隐藏快捷键提示',
    category: '界面',
    aliases: ['hotkeys help'],
  },
  {
    id: 'app.nav.home',
    title: '跳转到首页',
    category: '导航',
    aliases: ['go home'],
  },
  {
    id: 'app.nav.stats',
    title: '跳转到统计页',
    category: '导航',
    aliases: ['go stats'],
  },
  {
    id: 'app.nav.random_walk',
    title: '随机漫步',
    category: '导航',
    aliases: ['random walk', 'random memo'],
  },
  {
    id: 'app.nav.admin_invites',
    title: '打开邀请管理',
    category: '导航',
    aliases: ['invite admin', 'invites'],
  },
  {
    id: 'home.tag_governance.open',
    title: '打开标签治理',
    category: '标签',
    aliases: ['tag governance'],
  },
  {
    id: 'app.account_menu.toggle',
    title: '打开账号菜单',
    category: '菜单',
    aliases: ['account menu', 'avatar menu'],
  },
  {
    id: 'home.backup.import',
    title: '导入备份',
    category: '数据',
    aliases: ['import zip', 'restore'],
  },
  {
    id: 'home.backup.export',
    title: '导出备份',
    category: '数据',
    aliases: ['export zip', 'backup'],
  },
  {
    id: 'app.auth.switch_user',
    title: '切换账号',
    category: '账号',
    aliases: ['logout', 'switch account'],
  },
  {
    id: 'app.command_palette.open',
    title: '打开命令面板',
    category: '模式',
    aliases: ['command palette'],
  },
  {
    id: 'app.mode.normal.enter',
    title: '回到普通模式',
    category: '模式',
    aliases: ['normal mode'],
  },
  {
    id: 'home.editor.close',
    title: '关闭编辑弹窗',
    category: '编辑器',
    aliases: ['close editor'],
    visibleWhen: (snapshot) => hasFlag(snapshot, 'home.isEditorOpen'),
  },
  {
    id: 'home.editor.toggle_expand',
    title: '切换编辑区扩展',
    category: '编辑器',
    aliases: ['expand editor'],
    visibleWhen: (snapshot) => hasFlag(snapshot, 'home.isEditorOpen'),
  },
];

export const HOME_HOTKEY_BINDINGS: readonly HotkeyBindingDefinition[] = [
  {
    commandId: 'home.shortcuts.toggle',
    context: { id: 'home' },
    keys: '?',
    when: whenNoDialog,
  },
  {
    commandId: 'home.memo.create',
    context: { id: 'home' },
    keys: 'n',
    when: whenNoDialog,
  },
  {
    commandId: 'home.search.focus',
    context: { id: 'home' },
    keys: '/',
    when: whenNoDialog,
  },
  {
    commandId: 'home.memo.select_next',
    context: { id: 'home' },
    keys: 'j',
    when: whenNoDialog,
  },
  {
    commandId: 'home.memo.select_prev',
    context: { id: 'home' },
    keys: 'k',
    when: whenNoDialog,
  },
  {
    commandId: 'home.memo.edit_selected',
    context: { id: 'home' },
    keys: 'e',
    when: whenNoDialog,
  },
  {
    commandId: 'home.memo.delete_selected',
    context: { id: 'home' },
    keys: 'd',
    when: whenNoDialog,
  },
  {
    commandId: 'home.tag_governance.open',
    context: { id: 'home' },
    keys: 'g t',
    when: whenNoDialog,
  },
  {
    commandId: 'home.editor.close',
    context: { id: 'memo-editor' },
    keys: 'escape',
    mode: [HOTKEY_MODES.NORMAL, HOTKEY_MODES.INSERT],
  },
  {
    commandId: 'app.mode.normal.enter',
    keys: 'escape',
    mode: HOTKEY_MODES.INSERT,
  },
  {
    commandId: 'app.mode.normal.enter',
    keys: 'escape',
    mode: HOTKEY_MODES.COMMAND,
  },
  {
    commandId: 'home.editor.toggle_expand',
    context: { id: 'memo-editor' },
    keys: 'ctrl+shift+f',
    mode: [HOTKEY_MODES.NORMAL, HOTKEY_MODES.INSERT],
  },
  {
    commandId: 'home.editor.toggle_expand',
    context: { id: 'memo-editor' },
    keys: 'meta+shift+f',
    mode: [HOTKEY_MODES.NORMAL, HOTKEY_MODES.INSERT],
  },
];

const MODIFIER_LABELS: Record<string, string> = {
  alt: 'Alt',
  ctrl: 'Ctrl',
  meta: 'Cmd',
  shift: 'Shift',
};

const normalizeKeyLabel = (key: string): string => {
  if (MODIFIER_LABELS[key]) return MODIFIER_LABELS[key];
  if (key === 'escape') return 'Esc';
  if (key === 'space') return 'Space';
  return key.length === 1 ? key.toUpperCase() : key.toUpperCase();
};

const formatShortcutStroke = (stroke: string): string => {
  return stroke
    .split('+')
    .map((segment) => normalizeKeyLabel(segment))
    .join('+');
};

const formatShortcut = (keys: string): string => {
  return keys
    .trim()
    .split(/\s+/)
    .map((stroke) => formatShortcutStroke(stroke))
    .join(' ');
};

const matchesMode = (
  bindingMode: HotkeyMode | readonly HotkeyMode[] | undefined,
  targetMode: HotkeyMode,
): boolean => {
  if (!bindingMode) return targetMode === HOTKEY_MODES.NORMAL;
  if (Array.isArray(bindingMode)) {
    return (bindingMode as readonly HotkeyMode[]).includes(targetMode);
  }
  return bindingMode === targetMode;
};

export const getHomeCommandShortcutLabel = (
  commandId: string,
): string | undefined => {
  const labels = [...HOME_HOTKEY_BINDINGS, ...APP_ROOT_HOTKEY_BINDINGS]
    .filter((binding) => {
      return (
        binding.commandId === commandId &&
        (matchesMode(binding.mode, HOTKEY_MODES.NORMAL) ||
          matchesMode(binding.mode, HOTKEY_MODES.INSERT))
      );
    })
    .map((binding) => formatShortcut(binding.keys));

  const uniqueLabels = [...new Set(labels)];
  if (uniqueLabels.length === 0) return undefined;

  return uniqueLabels.join(' / ');
};
