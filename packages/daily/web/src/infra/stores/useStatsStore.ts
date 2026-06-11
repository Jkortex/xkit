import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { StatsDTO } from '@/application/ports/dto/Stats';
import { statsGateway } from '@/infra/gateway/HttpStatsGateway';

export const useStatsStore = defineStore('stats', () => {
  const stats = ref<StatsDTO | null>(null);
  const loading = ref(false);

  const setStats = (data: StatsDTO | null) => {
    stats.value = data;
  };

  const fetchStats = async () => {
    loading.value = true;
    const result = await statsGateway.getStats();
    loading.value = false;
    if (result.kind === 'success') {
      stats.value = result.value;
    }
  };

  return {
    stats,
    loading,
    setStats,
    fetchStats,
  };
});
