import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  pushMock,
  memoState,
  fetchMemosMock,
  fetchMoreMock,
  deleteMemoMock,
  buildQueryMock,
  submitSearchMock,
  useInfiniteScrollMock,
} = vi.hoisted(() => ({
  pushMock: vi.fn(),
  memoState: {
    memos: { value: [] },
    loading: { value: false },
    loadingMore: { value: false },
    hasMore: { value: true },
  },
  fetchMemosMock: vi.fn(),
  fetchMoreMock: vi.fn(),
  deleteMemoMock: vi.fn(),
  buildQueryMock: vi.fn(() => ({ search: 'k' })),
  submitSearchMock: vi.fn(async () => undefined),
  useInfiniteScrollMock: vi.fn(),
}));

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({ push: pushMock, replace: vi.fn() }),
}));

vi.mock('tdesign-vue-next', () => ({
  MessagePlugin: {
    success: vi.fn(),
  },
}));

vi.mock('@/presentation/composables/useMemoModel', () => ({
  useMemoModel: () => ({
    memos: memoState.memos,
    fetchMemos: fetchMemosMock,
    fetchMore: fetchMoreMock,
    loading: memoState.loading,
    loadingMore: memoState.loadingMore,
    hasMore: memoState.hasMore,
  }),
}));

vi.mock('@/presentation/composables/useMemoActions', () => ({
  useMemoActions: () => ({
    deleteMemo: deleteMemoMock,
  }),
}));

vi.mock('@/presentation/composables/useHomeFilters', () => ({
  useHomeFilters: () => ({
    searchText: { value: '' },
    tagAny: { value: '' },
    tagAll: { value: '' },
    fromDate: { value: '' },
    toDate: { value: '' },
    sortMode: { value: 'created_at_desc' },
    hasResource: { value: undefined },
    buildQuery: buildQueryMock,
  }),
}));

vi.mock('@/presentation/composables/useMemoEditorDialog', () => ({
  useMemoEditorDialog: () => ({
    showEditorDialog: { value: false },
    editingMemo: { value: undefined },
    editorExpanded: { value: false },
    bindMemoEditorRef: vi.fn(),
    openCreateEditor: vi.fn(),
    startEdit: vi.fn(),
    closeEditor: vi.fn(),
    toggleEditorExpanded: vi.fn(),
    onEditorSuccess: vi.fn(),
  }),
}));

vi.mock('@/presentation/composables/useHomeRouteFilters', () => ({
  useHomeRouteFilters: (options: { onQueryApplied: () => void }) => {
    options.onQueryApplied();
    return {
      submitSearch: submitSearchMock,
    };
  },
}));

vi.mock('@/presentation/composables/useInfiniteMemoScroll', () => ({
  useInfiniteMemoScroll: useInfiniteScrollMock,
}));

import { MessagePlugin } from 'tdesign-vue-next';
import { useHomeViewData } from '@/presentation/composables/home/useHomeViewData';

describe('useHomeViewData', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    memoState.memos.value = [];
  });

  it('focuses search input via command deck ref binding', () => {
    const focusSearchInput = vi.fn();
    const data = useHomeViewData();

    data.bindCommandDeckRef({ focusSearchInput });
    data.focusSearchInput();

    expect(focusSearchInput).toHaveBeenCalledTimes(1);
  });

  it('shows success message when deleting memo succeeds', async () => {
    deleteMemoMock.mockResolvedValue(true);
    const data = useHomeViewData();

    await data.handleDelete(7);

    expect(deleteMemoMock).toHaveBeenCalledWith(7);
    expect(MessagePlugin.success).toHaveBeenCalledWith('删除成功');
  });

  it('wires initial query apply and infinite scroll setup', async () => {
    const data = useHomeViewData();

    await data.submitSearch();

    expect(buildQueryMock).toHaveBeenCalled();
    expect(fetchMemosMock).toHaveBeenCalledWith({ search: 'k' });
    expect(useInfiniteScrollMock).toHaveBeenCalledTimes(1);
    expect(submitSearchMock).toHaveBeenCalledTimes(1);
  });
});
