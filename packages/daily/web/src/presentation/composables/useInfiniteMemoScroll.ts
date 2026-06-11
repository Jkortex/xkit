import { onMounted, onUnmounted, type Ref } from 'vue';
import type { MemoFilterVM } from '@/presentation/view-models/MemoFilterVM';

interface UseInfiniteMemoScrollOptions {
  hasMore: Ref<boolean>;
  loading: Ref<boolean>;
  loadingMore: Ref<boolean>;
  buildQuery: () => MemoFilterVM;
  fetchMore: (params: MemoFilterVM) => Promise<void>;
}

/** Triggers memo pagination when the viewport nears the document bottom. */
export function useInfiniteMemoScroll(
  options: UseInfiniteMemoScrollOptions,
): void {
  const handleScroll = (): void => {
    const scrollHeight = document.documentElement.scrollHeight;
    const scrollTop = document.documentElement.scrollTop;
    const clientHeight = document.documentElement.clientHeight;
    if (scrollHeight - scrollTop - clientHeight > 120) return;
    if (
      !options.hasMore.value ||
      options.loading.value ||
      options.loadingMore.value
    ) {
      return;
    }
    void options.fetchMore(options.buildQuery());
  };

  onMounted(() => {
    window.addEventListener('scroll', handleScroll);
  });

  onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll);
  });
}
