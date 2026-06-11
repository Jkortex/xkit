import { computed } from 'vue';
import { useStatsStore } from '@/infra/stores/useStatsStore';

export function useStats() {
  const store = useStatsStore();

  const stats = computed(() => store.stats);

  const fetchStats = async (options: { reset?: boolean } = {}) => {
    if (options.reset) {
      store.setStats(null);
    }
    await store.fetchStats();
  };

  return {
    stats,
    loading: computed(() => store.loading),
    fetchStats,
  };
}
