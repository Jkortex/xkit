import { describe, expect, it } from 'vitest';
import { transformStats } from './stats-transform';

describe('transformStats', () => {
  it('maps backend stats to frontend dto', () => {
    const result = transformStats({
      memos_total: 2,
      tags_total: 3,
      resources_total: 4,
      heatmap: [{ date: '2026-03-06', count: 1 }],
    });

    expect(result).toEqual({
      memosTotal: 2,
      tagsTotal: 3,
      resourcesTotal: 4,
      heatmap: [{ date: '2026-03-06', count: 1 }],
    });
  });

  it('throws on invalid daily date', () => {
    expect(() =>
      transformStats({
        memos_total: 1,
        tags_total: 1,
        resources_total: 1,
        heatmap: [{ date: '03-06-2026', count: 1 }],
      }),
    ).toThrow(/invalid heatmap date/);
  });
});
