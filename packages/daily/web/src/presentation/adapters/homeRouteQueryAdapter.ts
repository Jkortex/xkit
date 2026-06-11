import type { LocationQuery, LocationQueryRaw } from 'vue-router';
import type {
  SortMode,
  HomeFilterRouteFields,
} from '@/presentation/filters/types';

const normalizeQueryValue = (
  value: LocationQuery[string] | LocationQueryRaw[string],
): string => {
  const v = Array.isArray(value) ? value[0] : value;
  return String(v ?? '');
};

export const buildRouteQueryFromFilters = (
  filters: HomeFilterRouteFields,
): LocationQueryRaw => {
  const query: LocationQueryRaw = {};
  const search = filters.searchText.trim();
  const any = filters.tagAny.trim();
  const all = filters.tagAll.trim();
  const exclude = filters.tagExclude.trim();
  if (search) query.search = search;
  if (any) query.tagAny = any;
  if (all) query.tagAll = all;
  if (exclude) query.tagExclude = exclude;
  if (filters.fromDate) query.from = filters.fromDate;
  if (filters.toDate) query.to = filters.toDate;
  if (filters.sortMode !== 'created_at_desc') query.sort = filters.sortMode;
  if (filters.hasResource === true) query.hasResource = '1';
  if (filters.hasResource === false) query.hasResource = '0';
  return query;
};

export const hasSameRouteQuery = (
  currentQuery: LocationQuery,
  nextQuery: LocationQueryRaw,
): boolean => {
  const keys = new Set([
    ...Object.keys(currentQuery),
    ...Object.keys(nextQuery),
  ]);
  for (const key of keys) {
    const a = currentQuery[key];
    const b = nextQuery[key];
    if (normalizeQueryValue(a) !== normalizeQueryValue(b)) return false;
  }
  return true;
};

export const parseFiltersFromRouteQuery = (
  query: LocationQuery,
): HomeFilterRouteFields => {
  const querySort = (query.sort as string) || 'created_at_desc';
  const sortMode: SortMode =
    querySort === 'created_at_asc' || querySort === 'updated_at_desc'
      ? querySort
      : 'created_at_desc';
  const queryHasResource = query.hasResource as string | undefined;
  const hasResource =
    queryHasResource === '1'
      ? true
      : queryHasResource === '0'
        ? false
        : undefined;

  return {
    searchText: (query.search as string) || '',
    tagAny: (query.tagAny as string) || '',
    tagAll: (query.tagAll as string) || '',
    tagExclude: (query.tagExclude as string) || '',
    fromDate: (query.from as string) || '',
    toDate: (query.to as string) || '',
    sortMode,
    hasResource,
  };
};
