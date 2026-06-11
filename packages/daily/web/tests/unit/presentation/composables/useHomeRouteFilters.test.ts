import { beforeEach, describe, expect, it, vi } from 'vitest';
import { reactive, ref } from 'vue';
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router';
import { useHomeRouteFilters } from '@/presentation/composables/useHomeRouteFilters';
import type {
  FilterToken,
  HomeFilterRouteFields,
} from '@/presentation/filters/types';

describe('useHomeRouteFilters', () => {
  const createDeps = () => {
    const route = reactive({
      query: {
        search: 'init',
        tagAny: 'ops',
      },
    }) as RouteLocationNormalizedLoaded;
    const replace = vi.fn(async () => undefined);
    const router = { replace } as unknown as Router;

    const inputText = ref('');
    const tokens = ref<FilterToken[]>([]);

    const applyRouteFilters = vi.fn((filters: HomeFilterRouteFields) => {
      inputText.value = filters.searchText || '';
      tokens.value = [];
      if (filters.tagAny) {
        tokens.value.push({
          id: '1',
          type: 'tag',
          value: filters.tagAny,
          label: `tag:${filters.tagAny}`,
        });
      }
    });

    const onQueryApplied = vi.fn();

    const hook = useHomeRouteFilters({
      route,
      router,
      inputText,
      tokens,
      applyRouteFilters,
      onQueryApplied,
    });

    return {
      route,
      replace,
      inputText,
      tokens,
      applyRouteFilters,
      onQueryApplied,
      submitSearch: hook.submitSearch,
    };
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('initially syncs refs from route query', () => {
    const deps = createDeps();

    expect(deps.inputText.value).toBe('init');
    expect(deps.tokens.value).toEqual([
      { id: '1', type: 'tag', value: 'ops', label: 'tag:ops' },
    ]);
    expect(deps.onQueryApplied).toHaveBeenCalledTimes(1);
  });

  it('submits router replace when query changed', async () => {
    const deps = createDeps();
    deps.inputText.value = 'changed';

    await deps.submitSearch();

    expect(deps.replace).toHaveBeenCalledWith({
      path: '/',
      query: { search: 'changed', tagAny: 'ops' },
    });
  });

  it('only reapplies when query is unchanged', async () => {
    const deps = createDeps();
    deps.inputText.value = 'init';
    deps.tokens.value = [
      { id: '1', type: 'tag', value: 'ops', label: 'tag:ops' },
    ];
    deps.onQueryApplied.mockClear();

    await deps.submitSearch();

    expect(deps.replace).not.toHaveBeenCalled();
    expect(deps.onQueryApplied).toHaveBeenCalledTimes(1);
  });
});
