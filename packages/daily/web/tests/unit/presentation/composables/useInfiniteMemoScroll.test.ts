// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import { mountComposable } from '../../support/mountComposable';
import { useInfiniteMemoScroll } from '@/presentation/composables/useInfiniteMemoScroll';

describe('useInfiniteMemoScroll lifecycle', () => {
  let mounted: ReturnType<typeof mountComposable<void>> | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(document.documentElement, 'scrollHeight', {
      configurable: true,
      value: 1000,
    });
    Object.defineProperty(document.documentElement, 'scrollTop', {
      configurable: true,
      value: 850,
    });
    Object.defineProperty(document.documentElement, 'clientHeight', {
      configurable: true,
      value: 100,
    });
  });

  afterEach(() => {
    mounted?.unmount();
    mounted = null;
  });

  it('detaches scroll listener after unmount', async () => {
    const fetchMore = vi.fn(async () => {});
    const buildQuery = vi.fn(() => ({ search: 'x' }));
    mounted = mountComposable(() =>
      useInfiniteMemoScroll({
        hasMore: ref(true),
        loading: ref(false),
        loadingMore: ref(false),
        buildQuery,
        fetchMore,
      }),
    );

    window.dispatchEvent(new Event('scroll'));
    await Promise.resolve();
    expect(fetchMore).toHaveBeenCalledTimes(1);

    mounted.unmount();
    mounted = null;
    fetchMore.mockClear();

    window.dispatchEvent(new Event('scroll'));
    await Promise.resolve();
    expect(fetchMore).not.toHaveBeenCalled();
  });
});
