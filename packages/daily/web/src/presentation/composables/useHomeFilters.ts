import { ref, type Ref } from 'vue';
import type { MemoFilterVM } from '@/presentation/view-models/MemoFilterVM';
import type {
  FilterToken,
  HomeFilterRouteFields,
} from '@/presentation/filters/types';
import * as FilterUtils from '@/presentation/filters/filterUtils';

// Re-export types for backward compat — importers should migrate to '@/presentation/filters/types'
export type {
  SortMode,
  FilterTokenType,
  FilterToken,
  HomeFilterRouteFields,
} from '@/presentation/filters/types';

interface UseHomeFiltersResult {
  inputText: Ref<string>;
  tokens: Ref<FilterToken[]>;
  buildQuery: () => MemoFilterVM;
  addToken: (token: Omit<FilterToken, 'id'>) => void;
  removeToken: (id: string) => void;
  clearAll: () => void;
  applyRouteFilters: (filters: HomeFilterRouteFields) => void;
}

export function useHomeFilters(): UseHomeFiltersResult {
  const inputText = ref('');
  const tokens = ref<FilterToken[]>([]);

  const addToken = (token: Omit<FilterToken, 'id'>) => {
    tokens.value = FilterUtils.addToken(tokens.value, token);
  };

  const removeToken = (id: string) => {
    tokens.value = FilterUtils.removeToken(tokens.value, id);
  };

  const clearAll = () => {
    const cleared = FilterUtils.clearAll();
    tokens.value = cleared.tokens;
    inputText.value = cleared.inputText;
  };

  const applyRouteFilters = (filters: HomeFilterRouteFields) => {
    const result = FilterUtils.applyRouteFilters(filters);
    tokens.value = result.tokens;
    inputText.value = result.inputText;
  };

  const buildQuery = (): MemoFilterVM => {
    return FilterUtils.buildQuery(tokens.value, inputText.value);
  };

  return {
    inputText,
    tokens,
    buildQuery,
    addToken,
    removeToken,
    clearAll,
    applyRouteFilters,
  };
}
