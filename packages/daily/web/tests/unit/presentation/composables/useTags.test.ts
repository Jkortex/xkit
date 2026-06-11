// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { success } from '@/utils/result';
import { mountComposable } from '../../support/mountComposable';
import { setActivePinia, createPinia } from 'pinia';

const { mockMemoGateway } = vi.hoisted(() => ({
  mockMemoGateway: {
    getTags: vi.fn(),
  },
}));

vi.mock('@/infra/gateway/HttpMemoGateway', () => ({
  memoGateway: mockMemoGateway,
}));

import { useTags } from '@/presentation/composables/useTags';

describe('useTags lifecycle', () => {
  let mounted: ReturnType<
    typeof mountComposable<ReturnType<typeof useTags>>
  > | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
  });

  afterEach(() => {
    mounted?.unmount();
    mounted = null;
  });

  it('reflects tags from store and fetches successfully', async () => {
    mockMemoGateway.getTags.mockResolvedValue(
      success([{ name: 'ops', count: 2 }]),
    );

    mounted = mountComposable(() => useTags());

    // Initial empty tags
    expect(mounted.getApi().tags.value).toEqual([]);

    // Fetch tags
    await mounted.getApi().fetchTags();

    expect(mockMemoGateway.getTags).toHaveBeenCalled();
    expect(mounted.getApi().tags.value).toEqual([{ name: 'ops', count: 2 }]);
  });
});
