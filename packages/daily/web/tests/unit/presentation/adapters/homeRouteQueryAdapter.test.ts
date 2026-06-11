import { describe, expect, it } from 'vitest';
import {
  buildRouteQueryFromFilters,
  hasSameRouteQuery,
  parseFiltersFromRouteQuery,
} from '@/presentation/adapters/homeRouteQueryAdapter';

describe('homeRouteQueryAdapter', () => {
  it('builds compact route query from filters', () => {
    const query = buildRouteQueryFromFilters({
      searchText: ' hello ',
      tagAny: 'ops',
      tagAll: '',
      tagExclude: '',
      fromDate: '',
      toDate: '2026-03-08',
      sortMode: 'updated_at_desc',
      hasResource: false,
    });

    expect(query).toEqual({
      search: 'hello',
      tagAny: 'ops',
      to: '2026-03-08',
      sort: 'updated_at_desc',
      hasResource: '0',
    });
  });

  it('parses filter fields from route query', () => {
    const parsed = parseFiltersFromRouteQuery({
      search: 'k',
      tagAny: 'a,b',
      sort: 'created_at_asc',
      hasResource: '1',
    });

    expect(parsed.searchText).toBe('k');
    expect(parsed.tagAny).toBe('a,b');
    expect(parsed.sortMode).toBe('created_at_asc');
    expect(parsed.hasResource).toBe(true);
  });

  it('compares normalized query values', () => {
    const same = hasSameRouteQuery(
      { search: ['x'], hasResource: '1' },
      { search: 'x', hasResource: '1' },
    );

    expect(same).toBe(true);
  });
});
