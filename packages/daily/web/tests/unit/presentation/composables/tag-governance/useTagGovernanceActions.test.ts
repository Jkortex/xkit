import { beforeEach, describe, expect, it, vi } from 'vitest';
import { success } from '@/utils/result';
import { setActivePinia, createPinia } from 'pinia';

const { mockMemoGateway } = vi.hoisted(() => ({
  mockMemoGateway: {
    renameTag: vi.fn(),
    mergeTags: vi.fn(),
    upsertTagAlias: vi.fn(),
    listTagAliases: vi.fn(),
    deleteTagAlias: vi.fn(),
    listTagAudits: vi.fn(),
  },
}));

vi.mock('@/infra/gateway/HttpMemoGateway', () => ({
  memoGateway: mockMemoGateway,
}));

vi.mock('@/infra/gateway/HttpTagSetGateway', () => ({
  tagSetGateway: {},
}));

vi.mock('tdesign-vue-next', () => ({
  MessagePlugin: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
  },
}));

import { MessagePlugin } from 'tdesign-vue-next';
import { useTagGovernanceState } from '@/presentation/composables/tag-governance/tagGovernanceState';
import { useTagGovernanceActions } from '@/presentation/composables/tag-governance/useTagGovernanceActions';

describe('useTagGovernanceActions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
  });

  it('loads aliases into state', async () => {
    mockMemoGateway.listTagAliases.mockResolvedValue(
      success([{ alias: 'tmp', canonical: 'test' }]),
    );
    const state = useTagGovernanceState();
    const actions = useTagGovernanceActions(state, { onChanged: vi.fn() });

    await actions.fetchTagAliases();

    expect(state.tagAliases.value).toEqual([
      { alias: 'tmp', canonical: 'test' },
    ]);
  });

  it('warns when rename input is missing', async () => {
    const state = useTagGovernanceState();
    const actions = useTagGovernanceActions(state, { onChanged: vi.fn() });

    await actions.handleRenameTag();

    expect(MessagePlugin.warning).toHaveBeenCalledWith(
      '请填写来源标签和目标标签',
    );
  });

  it('renames tag and closes dialog', async () => {
    mockMemoGateway.renameTag.mockResolvedValue(
      success({
        from: 'a',
        to: 'b',
        affected_memos: 2,
        merged: false,
      }),
    );
    const onChanged = vi.fn();
    const state = useTagGovernanceState();
    state.showTagManageDialog.value = true;
    state.renameFrom.value = 'a';
    state.renameTo.value = 'b';
    const actions = useTagGovernanceActions(state, { onChanged });

    await actions.handleRenameTag();

    expect(mockMemoGateway.renameTag).toHaveBeenCalledWith('a', 'b');
    expect(onChanged).toHaveBeenCalled();
    expect(state.showTagManageDialog.value).toBe(false);
  });
});
