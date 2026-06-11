import { computed } from 'vue';
import { useTagStore } from '@/infra/stores/useTagStore';

export function useTags() {
  const store = useTagStore();

  const tags = computed(() => store.tags);

  const fetchTags = async (options: { reset?: boolean } = {}) => {
    if (options.reset) {
      store.setTags([]);
    }
    await store.fetchTags();
  };

  return {
    tags,
    loading: computed(() => store.loading),
    fetchTags,
  };
}
