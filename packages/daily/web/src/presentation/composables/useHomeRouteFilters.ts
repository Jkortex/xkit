import { watch, type Ref } from 'vue';
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router';
import type {
  FilterToken,
  HomeFilterRouteFields,
} from '@/presentation/filters/types';
import { tokensToRouteFields } from '@/presentation/filters/filterUtils';
import {
  buildRouteQueryFromFilters,
  hasSameRouteQuery,
  parseFiltersFromRouteQuery,
} from '@/presentation/adapters/homeRouteQueryAdapter';

interface UseHomeRouteFiltersOptions {
  route: RouteLocationNormalizedLoaded;
  router: Router;
  inputText: Ref<string>;
  tokens: Ref<FilterToken[]>;
  applyRouteFilters: (filters: HomeFilterRouteFields) => void;
  onQueryApplied: () => void;
}

interface UseHomeRouteFiltersResult {
  submitSearch: () => Promise<void>;
}

/** Keeps Home filters synchronized with route query params. */
export function useHomeRouteFilters(
  options: UseHomeRouteFiltersOptions,
): UseHomeRouteFiltersResult {
  const submitSearch = async (): Promise<void> => {
    const fields = tokensToRouteFields(
      options.tokens.value,
      options.inputText.value,
    );
    const nextQuery = buildRouteQueryFromFilters(fields);

    if (hasSameRouteQuery(options.route.query, nextQuery)) {
      options.onQueryApplied();
      return;
    }
    await options.router.replace({ path: '/', query: nextQuery });
  };

  watch(
    () => options.route.query,
    (query) => {
      const parsed = parseFiltersFromRouteQuery(query);
      options.applyRouteFilters(parsed);
      options.onQueryApplied();
    },
    { immediate: true, deep: true },
  );

  return {
    submitSearch,
  };
}
