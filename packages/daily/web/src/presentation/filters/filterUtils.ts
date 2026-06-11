import type { FilterToken, HomeFilterRouteFields } from './types';
import type { FilterTokenType } from './types';
import type { MemoFilterVM } from '@/presentation/view-models/MemoFilterVM';

const genId = () => Math.random().toString(36).substring(2, 9);

const SINGLETON_TYPES: FilterTokenType[] = [
  'from',
  'to',
  'has_resource',
  'sort',
];

export function addToken(
  tokens: FilterToken[],
  token: Omit<FilterToken, 'id'>,
): FilterToken[] {
  if (SINGLETON_TYPES.includes(token.type)) {
    return [
      ...tokens.filter((t) => t.type !== token.type),
      { ...token, id: genId() },
    ];
  }
  return [...tokens, { ...token, id: genId() }];
}

export function removeToken(tokens: FilterToken[], id: string): FilterToken[] {
  return tokens.filter((t) => t.id !== id);
}

export function clearAll(): { tokens: FilterToken[]; inputText: string } {
  return { tokens: [], inputText: '' };
}

export function applyRouteFilters(filters: HomeFilterRouteFields): {
  tokens: FilterToken[];
  inputText: string;
} {
  const newTokens: FilterToken[] = [];

  if (filters.tagAny) {
    filters.tagAny.split(',').forEach((t) => {
      const val = t.trim();
      if (val)
        newTokens.push({
          id: genId(),
          type: 'tag',
          value: val,
          label: `tag:${val}`,
        });
    });
  }
  if (filters.tagAll) {
    filters.tagAll.split(',').forEach((t) => {
      const val = t.trim();
      if (val)
        newTokens.push({
          id: genId(),
          type: 'tags_all',
          value: val,
          label: `tag+:${val}`,
        });
    });
  }
  if (filters.tagExclude) {
    filters.tagExclude.split(',').forEach((t) => {
      const val = t.trim();
      if (val)
        newTokens.push({
          id: genId(),
          type: 'tags_exclude',
          value: val,
          label: `tag-:${val}`,
        });
    });
  }
  if (filters.fromDate) {
    newTokens.push({
      id: genId(),
      type: 'from',
      value: filters.fromDate,
      label: `from:${filters.fromDate}`,
    });
  }
  if (filters.toDate) {
    newTokens.push({
      id: genId(),
      type: 'to',
      value: filters.toDate,
      label: `to:${filters.toDate}`,
    });
  }
  if (filters.hasResource !== undefined) {
    newTokens.push({
      id: genId(),
      type: 'has_resource',
      value: filters.hasResource,
      label: filters.hasResource ? 'has:resource' : 'has:no-resource',
    });
  }
  if (filters.sortMode && filters.sortMode !== 'created_at_desc') {
    newTokens.push({
      id: genId(),
      type: 'sort',
      value: filters.sortMode,
      label: `sort:${filters.sortMode}`,
    });
  }

  return { tokens: newTokens, inputText: filters.searchText || '' };
}

export function buildQuery(
  tokens: FilterToken[],
  inputText: string,
): MemoFilterVM {
  const query: MemoFilterVM = {
    includeResources: true,
    sort: 'created_at_desc',
  };

  const searchTerms: string[] = [];
  if (inputText.trim()) {
    searchTerms.push(inputText.trim());
  }

  const tagsAny: string[] = [];
  const tagsAll: string[] = [];
  const tagsExclude: string[] = [];

  tokens.forEach((token) => {
    switch (token.type) {
      case 'search':
      case 'text':
        searchTerms.push(token.value);
        break;
      case 'tag':
        tagsAny.push(token.value);
        break;
      case 'tags_any':
        if (Array.isArray(token.value)) tagsAny.push(...token.value);
        else tagsAny.push(token.value);
        break;
      case 'tags_all':
        if (Array.isArray(token.value)) tagsAll.push(...token.value);
        else tagsAll.push(token.value);
        break;
      case 'tags_exclude':
        if (Array.isArray(token.value)) tagsExclude.push(...token.value);
        else tagsExclude.push(token.value);
        break;
      case 'from':
        query.from = token.value;
        break;
      case 'to':
        query.to = token.value;
        break;
      case 'has_resource':
        query.hasResource = token.value;
        break;
      case 'sort':
        query.sort = token.value;
        break;
    }
  });

  if (searchTerms.length > 0) {
    query.search = searchTerms.join(' ');
  }
  if (tagsAny.length > 0) {
    query.tagsAny = [...new Set(tagsAny)];
  }
  if (tagsAll.length > 0) {
    query.tagsAll = [...new Set(tagsAll)];
  }
  if (tagsExclude.length > 0) {
    query.tagsExclude = [...new Set(tagsExclude)];
  }

  return query;
}

export function tokensToRouteFields(
  tokens: FilterToken[],
  inputText: string,
): HomeFilterRouteFields {
  const fields: HomeFilterRouteFields = {
    searchText: inputText,
    tagAny: '',
    tagAll: '',
    tagExclude: '',
    fromDate: '',
    toDate: '',
    sortMode: 'created_at_desc',
    hasResource: undefined,
  };

  const tagsAny: string[] = [];
  const tagsAll: string[] = [];
  const tagsExclude: string[] = [];

  tokens.forEach((t) => {
    if (t.type === 'tag') tagsAny.push(t.value);
    else if (t.type === 'tags_any') tagsAny.push(t.value);
    else if (t.type === 'tags_all') tagsAll.push(t.value);
    else if (t.type === 'tags_exclude') tagsExclude.push(t.value);
    else if (t.type === 'from') fields.fromDate = t.value;
    else if (t.type === 'to') fields.toDate = t.value;
    else if (t.type === 'has_resource') fields.hasResource = t.value;
    else if (t.type === 'sort') fields.sortMode = t.value;
  });

  fields.tagAny = tagsAny.join(',');
  fields.tagAll = tagsAll.join(',');
  fields.tagExclude = tagsExclude.join(',');

  return fields;
}
