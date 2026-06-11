import { computed } from 'vue';
import { useMemoStore } from '@/infra/stores/useMemoStore';
import { MemoPresenter } from '@/presentation/presenters/MemoPresenter';
import { toMemoListQuery } from '@/presentation/adapters/memoFilterAdapter';
import type { MemoFilterVM } from '@/presentation/view-models/MemoFilterVM';

const PAGE_SIZE = 20;

/**
 * 极简响应式 Model Hook
 * 读：直接连接 Store
 * 写：通过 Store 的 action
 */
export function useMemoModel() {
  const store = useMemoStore();

  const memos = computed(() => store.memos.map(MemoPresenter.toViewModel));

  const fetchMemos = async (params: MemoFilterVM = {}) => {
    await store.getMemos(
      toMemoListQuery(params, {
        limit: PAGE_SIZE,
        offset: 0,
      }),
    );
  };

  const fetchMore = async (params: MemoFilterVM = {}) => {
    if (store.loading || !store.hasMore) return;
    await store.getMemos(
      toMemoListQuery(params, {
        limit: PAGE_SIZE,
        offset: store.memos.length,
      }),
    );
  };

  return {
    memos,
    loading: computed(() => store.loading),
    loadingMore: computed(() => store.loadingMore),
    hasMore: computed(() => store.hasMore),
    fetchMemos,
    fetchMore,
  };
}
