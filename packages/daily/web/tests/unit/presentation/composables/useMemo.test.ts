// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { success } from '@/utils/result';
import { mountComposable } from '../../support/mountComposable';
import { setActivePinia, createPinia } from 'pinia';
import { useMemoStore } from '@/infra/stores/useMemoStore';

const { mockMemoGateway } = vi.hoisted(() => ({
  mockMemoGateway: {
    getMemos: vi.fn(),
    createMemo: vi.fn(),
    updateMemo: vi.fn(),
    deleteMemo: vi.fn(),
    getRandomMemo: vi.fn(),
    listMemoHistory: vi.fn(),
    rollbackMemo: vi.fn(),
    getTags: vi.fn(),
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

import { useMemoModel } from '@/presentation/composables/useMemoModel';

describe('useMemoModel query adapter', () => {
  let mounted: ReturnType<
    typeof mountComposable<ReturnType<typeof useMemoModel>>
  > | null = null;
  let store: ReturnType<typeof useMemoStore>;

  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
    store = useMemoStore();
  });

  afterEach(() => {
    mounted?.unmount();
    mounted = null;
  });

  it('maps filter vm and default pagination for fetchMemos', async () => {
    mockMemoGateway.getMemos.mockResolvedValue(success([]));
    mounted = mountComposable(() => useMemoModel());

    await mounted.getApi().fetchMemos({
      search: 'hello',
      tagsAny: ['ops'],
      tagsAll: ['platform'],
      hasResource: true,
      includeResources: true,
      sort: 'updated_at_desc',
    });

    expect(mockMemoGateway.getMemos).toHaveBeenCalledWith({
      search: 'hello',
      tag: undefined,
      from: undefined,
      to: undefined,
      hasResource: true,
      tagsAny: ['ops'],
      tagsAll: ['platform'],
      sort: 'updated_at_desc',
      includeResources: true,
      limit: 20,
      offset: 0,
    });
  });

  it('maps offset from current store size for fetchMore', async () => {
    store.$patch({
      memos: [
        {
          id: 1,
          uuid: '1',
          content: 'x',
          status: 'normal',
          tags: [],
          resources: [],
          createdAt: '2026-06-02T10:00:00Z',
          updatedAt: '2026-06-02T10:00:00Z',
        },
        {
          id: 2,
          uuid: '2',
          content: 'y',
          status: 'normal',
          tags: [],
          resources: [],
          createdAt: '2026-06-02T10:00:00Z',
          updatedAt: '2026-06-02T10:00:00Z',
        },
      ],
      hasMore: true,
    });

    mockMemoGateway.getMemos.mockResolvedValue(success([]));
    mounted = mountComposable(() => useMemoModel());

    await mounted.getApi().fetchMore({ search: 'abc' });

    expect(mockMemoGateway.getMemos).toHaveBeenCalledWith({
      search: 'abc',
      tag: undefined,
      from: undefined,
      to: undefined,
      hasResource: undefined,
      tagsAny: undefined,
      tagsAll: undefined,
      sort: undefined,
      includeResources: undefined,
      limit: 20,
      offset: 2,
    });
  });
});
