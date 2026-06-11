import { describe, expect, it } from 'vitest';
import {
  addToken,
  removeToken,
  clearAll,
  applyRouteFilters,
  buildQuery,
  tokensToRouteFields,
} from './filterUtils';
import type { FilterToken } from './types';

const makeToken = (
  overrides: {
    type: FilterToken['type'];
    value: unknown;
  } & Partial<FilterToken>,
): FilterToken => ({
  id: overrides.id ?? 'test-id',
  type: overrides.type,
  value: overrides.value,
  label: overrides.label ?? `${overrides.type}:${overrides.value}`,
});

describe('addToken', () => {
  it('appends a non-singleton token', () => {
    const result = addToken([], {
      type: 'tag',
      value: 'vue',
      label: 'tag:vue',
    });
    expect(result).toHaveLength(1);
    expect(result[0].type).toBe('tag');
    expect(result[0].value).toBe('vue');
    expect(result[0].id).toBeDefined();
  });

  it('replaces existing token for singleton types (from/to/has_resource/sort)', () => {
    const existing = [makeToken({ type: 'from', value: '2024-01-01' })];
    const result = addToken(existing, {
      type: 'from',
      value: '2024-02-01',
      label: 'from:2024-02-01',
    });
    expect(result).toHaveLength(1);
    expect(result[0].value).toBe('2024-02-01');
  });

  it('keeps non-singleton tokens and adds new one', () => {
    const existing = [makeToken({ type: 'tag', value: 'vue' })];
    const result = addToken(existing, {
      type: 'tag',
      value: 'react',
      label: 'tag:react',
    });
    expect(result).toHaveLength(2);
  });
});

describe('removeToken', () => {
  it('removes token by id', () => {
    const tokens = [
      makeToken({ id: 'a', type: 'tag', value: 'vue' }),
      makeToken({ id: 'b', type: 'tag', value: 'react' }),
    ];
    expect(removeToken(tokens, 'a')).toHaveLength(1);
    expect(removeToken(tokens, 'a')[0].id).toBe('b');
  });

  it('returns all tokens when id not found', () => {
    const tokens = [makeToken({ id: 'a', type: 'tag', value: 'vue' })];
    expect(removeToken(tokens, 'nonexistent')).toHaveLength(1);
  });
});

describe('clearAll', () => {
  it('returns empty tokens and inputText', () => {
    const result = clearAll();
    expect(result).toEqual({ tokens: [], inputText: '' });
  });
});

describe('applyRouteFilters', () => {
  it('maps tagAny to tag tokens', () => {
    const { tokens, inputText } = applyRouteFilters({
      searchText: '',
      tagAny: 'vue,react',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
    expect(tokens).toHaveLength(2);
    expect(tokens[0].type).toBe('tag');
    expect(tokens[0].value).toBe('vue');
    expect(tokens[1].type).toBe('tag');
    expect(tokens[1].value).toBe('react');
    expect(inputText).toBe('');
  });

  it('ignores empty tagAny entries after splitting', () => {
    const { tokens } = applyRouteFilters({
      searchText: '',
      tagAny: 'vue,,react',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
    expect(tokens).toHaveLength(2);
  });

  it('maps tagAll to tags_all tokens', () => {
    const { tokens } = applyRouteFilters({
      searchText: '',
      tagAny: '',
      tagAll: 'typescript,zod',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
    expect(tokens).toHaveLength(2);
    expect(tokens[0].type).toBe('tags_all');
    expect(tokens[0].value).toBe('typescript');
  });

  it('maps from/to dates', () => {
    const { tokens } = applyRouteFilters({
      searchText: '',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '2024-01-01',
      toDate: '2024-12-31',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
    expect(tokens).toHaveLength(2);
    expect(tokens.find((t) => t.type === 'from')?.value).toBe('2024-01-01');
    expect(tokens.find((t) => t.type === 'to')?.value).toBe('2024-12-31');
  });

  it('maps hasResource true/false', () => {
    const r1 = applyRouteFilters({
      searchText: '',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: true,
    });
    expect(r1.tokens[0].type).toBe('has_resource');
    expect(r1.tokens[0].label).toBe('has:resource');

    const r2 = applyRouteFilters({
      searchText: '',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: false,
    });
    expect(r2.tokens[0].label).toBe('has:no-resource');
  });

  it('skips sort token when sortMode is default', () => {
    const { tokens } = applyRouteFilters({
      searchText: '',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
    expect(tokens.find((t) => t.type === 'sort')).toBeUndefined();
  });

  it('adds sort token for non-default sortMode', () => {
    const { tokens } = applyRouteFilters({
      searchText: '',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_asc',
      hasResource: undefined,
    });
    expect(tokens.find((t) => t.type === 'sort')?.value).toBe('created_at_asc');
  });

  it('passes searchText as inputText', () => {
    const { inputText } = applyRouteFilters({
      searchText: 'hello world',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
    expect(inputText).toBe('hello world');
  });
});

describe('buildQuery', () => {
  it('empty tokens + empty input returns default query', () => {
    const q = buildQuery([], '');
    expect(q).toEqual({
      includeResources: true,
      sort: 'created_at_desc',
    });
  });

  it('includes inputText as search term', () => {
    const q = buildQuery([], 'hello');
    expect(q.search).toBe('hello');
  });

  it('collects tag tokens into tagsAny', () => {
    const tokens = [
      makeToken({ type: 'tag', value: 'vue' }),
      makeToken({ type: 'tag', value: 'react' }),
    ];
    const q = buildQuery(tokens, '');
    expect(q.tagsAny).toEqual(['vue', 'react']);
  });

  it('deduplicates tagsAny', () => {
    const tokens = [
      makeToken({ id: 'a', type: 'tag', value: 'vue' }),
      makeToken({ id: 'b', type: 'tag', value: 'vue' }),
    ];
    const q = buildQuery(tokens, '');
    expect(q.tagsAny).toEqual(['vue']);
  });

  it('collects tags_all tokens into tagsAll', () => {
    const tokens = [makeToken({ type: 'tags_all', value: 'typescript' })];
    const q = buildQuery(tokens, '');
    expect(q.tagsAll).toEqual(['typescript']);
  });

  it('sets from/to/hasResource/sort from tokens', () => {
    const tokens = [
      makeToken({ type: 'from', value: '2024-01-01' }),
      makeToken({ type: 'to', value: '2024-12-31' }),
      makeToken({ type: 'has_resource', value: true }),
      makeToken({ type: 'sort', value: 'created_at_asc' }),
    ];
    const q = buildQuery(tokens, '');
    expect(q.from).toBe('2024-01-01');
    expect(q.to).toBe('2024-12-31');
    expect(q.hasResource).toBe(true);
    expect(q.sort).toBe('created_at_asc');
  });

  it('merges inputText with search/text tokens', () => {
    const tokens = [makeToken({ type: 'search', value: 'keyword' })];
    const q = buildQuery(tokens, 'inputtext');
    expect(q.search).toBe('inputtext keyword');
  });

  it('handles tags_any and tags_all as arrays', () => {
    const tokens = [
      makeToken({ type: 'tags_any', value: ['a', 'b'] }),
      makeToken({ type: 'tags_all', value: ['c', 'd'] }),
    ];
    const q = buildQuery(tokens, '');
    expect(q.tagsAny).toEqual(['a', 'b']);
    expect(q.tagsAll).toEqual(['c', 'd']);
  });
});

describe('tokensToRouteFields', () => {
  it('empty tokens returns default fields', () => {
    const fields = tokensToRouteFields([], '');
    expect(fields).toEqual({
      searchText: '',
      tagAny: '',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '',
      sortMode: 'created_at_desc',
      hasResource: undefined,
    });
  });

  it('maps tag and tags_any to tagAny', () => {
    const tokens = [
      makeToken({ type: 'tag', value: 'vue' }),
      makeToken({ type: 'tags_any', value: 'react' }),
    ];
    const fields = tokensToRouteFields(tokens, '');
    expect(fields.tagAny).toBe('vue,react');
  });

  it('maps tags_all to tagAll', () => {
    const tokens = [makeToken({ type: 'tags_all', value: 'typescript' })];
    const fields = tokensToRouteFields(tokens, '');
    expect(fields.tagAll).toBe('typescript');
  });

  it('maps from/to/has_resource/sort', () => {
    const tokens = [
      makeToken({ type: 'from', value: '2024-01-01' }),
      makeToken({ type: 'to', value: '2024-12-31' }),
      makeToken({ type: 'has_resource', value: false }),
      makeToken({ type: 'sort', value: 'created_at_asc' }),
    ];
    const fields = tokensToRouteFields(tokens, '');
    expect(fields.fromDate).toBe('2024-01-01');
    expect(fields.toDate).toBe('2024-12-31');
    expect(fields.hasResource).toBe(false);
    expect(fields.sortMode).toBe('created_at_asc');
  });
});
