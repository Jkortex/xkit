// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { success } from '@/utils/result';
import { mountComposable } from '../../support/mountComposable';
import { setActivePinia, createPinia } from 'pinia';

const { mockStatsGateway } = vi.hoisted(() => ({
  mockStatsGateway: {
    getStats: vi.fn(),
  },
}));

vi.mock('@/infra/gateway/HttpStatsGateway', () => ({
  statsGateway: mockStatsGateway,
}));

import { useStats } from '@/presentation/composables/useStats';

describe('useStats lifecycle', () => {
  let mounted: ReturnType<
    typeof mountComposable<ReturnType<typeof useStats>>
  > | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
  });

  afterEach(() => {
    mounted?.unmount();
    mounted = null;
  });

  it('reflects stats from store and fetches successfully', async () => {
    const mockStats = {
      memosTotal: 2,
      tagsTotal: 1,
      resourcesTotal: 0,
      heatmap: [],
    };
    mockStatsGateway.getStats.mockResolvedValue(success(mockStats));

    mounted = mountComposable(() => useStats());

    // Initial empty stats
    expect(mounted.getApi().stats.value).toBeNull();

    // Fetch stats
    await mounted.getApi().fetchStats();

    expect(mockStatsGateway.getStats).toHaveBeenCalled();
    expect(mounted.getApi().stats.value?.memosTotal).toBe(2);
  });
});
