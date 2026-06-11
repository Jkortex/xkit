import { describe, expect, it } from 'vitest';
import { MemoHistoryPresenter } from '@/presentation/presenters/MemoHistoryPresenter';

describe('MemoHistoryPresenter', () => {
  it('formats history DTO into view model', () => {
    const dto = {
      id: 'h1',
      memo_uuid: 'm1',
      content: 'hello',
      tags: ['T1'],
      resource_ids: ['R1'],
      created_at: '2026-03-10T10:00:00Z',
    };

    const vm = MemoHistoryPresenter.toViewModel(dto);

    expect(vm.id).toBe('h1');
    expect(vm.content).toBe('hello');
    expect(vm.tags).toEqual(['T1']);
    expect(vm.resourceIds).toEqual(['R1']);
    expect(vm.absoluteTime).toBeTruthy();
    expect(vm.relativeTime).toBeTruthy();
  });

  it('formats relative time correctly for today', () => {
    const now = new Date();
    const tenMinutesAgo = new Date(now.getTime() - 10 * 60 * 1000);

    const rel = MemoHistoryPresenter.formatRelativeTime(tenMinutesAgo);
    expect(rel).toBe('10分钟前');
  });
});
