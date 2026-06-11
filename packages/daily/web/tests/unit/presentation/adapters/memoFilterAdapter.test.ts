import { describe, expect, it } from 'vitest';
import { toMemoListQuery } from '@/presentation/adapters/memoFilterAdapter';

describe('memoFilterAdapter', () => {
  it('maps filter vm fields to list query', () => {
    const query = toMemoListQuery({
      search: 'hello',
      from: '2026-03-01',
      to: '2026-03-08',
      sort: 'updated_at_desc',
      hasResource: true,
      tagsAny: ['ops', 'infra'],
      tagsAll: ['team-a'],
      includeResources: true,
    });

    expect(query).toEqual({
      search: 'hello',
      tag: undefined,
      from: '2026-03-01',
      to: '2026-03-08',
      hasResource: true,
      tagsAny: ['ops', 'infra'],
      tagsAll: ['team-a'],
      sort: 'updated_at_desc',
      includeResources: true,
      limit: undefined,
      offset: undefined,
    });
  });

  it('adds paging arguments when provided', () => {
    const query = toMemoListQuery({ search: 'x' }, { limit: 20, offset: 40 });

    expect(query.limit).toBe(20);
    expect(query.offset).toBe(40);
  });
});
