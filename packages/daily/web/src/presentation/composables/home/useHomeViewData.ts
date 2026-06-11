import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { useMemoActions } from '@/presentation/composables/useMemoActions';
import { useHomeFilters } from '@/presentation/composables/useHomeFilters';
import { useHomeRouteFilters } from '@/presentation/composables/useHomeRouteFilters';
import { useInfiniteMemoScroll } from '@/presentation/composables/useInfiniteMemoScroll';
import { useMemoModel } from '@/presentation/composables/useMemoModel';

export interface HomeCommandDeckExpose {
  focusSearchInput: () => void;
}

export function useHomeViewData() {
  const route = useRoute();
  const router = useRouter();
  const commandDeckRef = ref<HomeCommandDeckExpose | null>(null);
  const bindCommandDeckRef = (instance: unknown): void => {
    commandDeckRef.value = (instance as HomeCommandDeckExpose) ?? null;
  };

  const { memos, fetchMemos, fetchMore, loading, loadingMore, hasMore } =
    useMemoModel();
  const { deleteMemo } = useMemoActions();

  const {
    inputText,
    tokens,
    buildQuery,
    addToken,
    removeToken,
    clearAll,
    applyRouteFilters,
  } = useHomeFilters();

  const fetchWithCurrentFilters = (): void => {
    void fetchMemos(buildQuery());
  };

  const focusSearchInput = (): void => {
    commandDeckRef.value?.focusSearchInput();
  };

  const applySingleTagFilter = async (tagName: string): Promise<void> => {
    clearAll();
    addToken({ type: 'tag', value: tagName, label: `tag:${tagName}` });
    await submitSearch();
  };

  const handleDelete = async (id: string): Promise<void> => {
    const success = await deleteMemo(id);
    if (success) MessagePlugin.success('删除成功');
  };

  const { submitSearch } = useHomeRouteFilters({
    route,
    router,
    inputText,
    tokens,
    applyRouteFilters,
    onQueryApplied: fetchWithCurrentFilters,
  });

  useInfiniteMemoScroll({
    hasMore,
    loading,
    loadingMore,
    buildQuery,
    fetchMore,
  });

  return {
    bindCommandDeckRef,
    memos,
    loading,
    loadingMore,
    hasMore,
    inputText,
    tokens,
    focusSearchInput,
    applySingleTagFilter,
    handleDelete,
    submitSearch,
    fetchMemos,
    buildQuery,
    addToken,
    removeToken,
    clearAll,
  };
}
